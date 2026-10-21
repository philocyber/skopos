package server

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Philocyber/skopos/locale"
	"github.com/Philocyber/skopos/privatefs"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	jwtKeyFilename      = "jwt.key"
	authPassKey         = "auth.password_hash"
	jwtTTL              = 7 * 24 * time.Hour
	keyChars            = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	maxAuthRequestBytes = 8 << 10
	authBodyReadTimeout = 10 * time.Second
)

// bootstrapToken proves access to the local installation without disclosing its
// JWT signing key. Domain separation prevents using this value to sign sessions.
func bootstrapToken(key []byte) string {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte("Skopos first-owner enrollment v1"))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// ReadSetupToken is used only by the local CLI. It does not create a signing key
// or contact the database; start the backend once before requesting this token.
func ReadSetupToken(keyDir string) (string, error) {
	key, err := os.ReadFile(filepath.Join(keyDir, jwtKeyFilename))
	if err != nil {
		return "", errors.New("cannot read local signing key; start the backend before requesting a setup token")
	}
	key = []byte(strings.TrimSpace(string(key)))
	if len(key) < 32 {
		return "", errors.New("local signing key is invalid")
	}
	return bootstrapToken(key), nil
}

// decodeAuthRequest bounds only credential requests. A server-wide read timeout
// would also affect uploads and long-lived streaming routes.
func decodeAuthRequest(w http.ResponseWriter, r *http.Request, value any) bool {
	controller := http.NewResponseController(w)
	if err := controller.SetReadDeadline(time.Now().Add(authBodyReadTimeout)); err != nil && !errors.Is(err, http.ErrNotSupported) {
		writeErr(w, http.StatusServiceUnavailable, locale.Text(responseLanguage(w), "Authentication request unavailable"))
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxAuthRequestBytes)
	defer func() {
		// Close while the deadline still applies, including invalid partial bodies.
		_ = r.Body.Close()
		_ = controller.SetReadDeadline(time.Time{})
	}()
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(value)
	if err == nil {
		// Consume bounded trailing whitespace and reject a second JSON value.
		// Otherwise a valid prefix could avoid detecting an oversized body.
		var extra any
		if trailing := decoder.Decode(&extra); trailing != io.EOF {
			if trailing == nil {
				err = errors.New("multiple authentication request values")
			} else {
				err = trailing
			}
		}
	}
	if err == nil {
		return true
	}
	var tooLarge *http.MaxBytesError
	var timeout net.Error
	switch {
	case errors.As(err, &tooLarge):
		writeErr(w, http.StatusRequestEntityTooLarge, locale.Text(responseLanguage(w), "Authentication request is too large"))
	case errors.As(err, &timeout) && timeout.Timeout():
		writeErr(w, http.StatusRequestTimeout, locale.Text(responseLanguage(w), "Authentication request timed out"))
	default:
		writeErr(w, http.StatusBadRequest, locale.Text(responseLanguage(w), "Invalid request format"))
	}
	return false
}

// loadOrCreateJWTKey reads the 32-byte signing key from keyDir/jwt.key. keyDir is
// the project base dir (next to the executable), NOT the browsable workspace root
// (dataDir) — the signing key must never be listable/downloadable via the file
// manager. Legacy installs kept it at dataDir/jwt.key; if present there and not yet
// at the new location, it is migrated (key preserved, so sessions stay valid) and
// the old file removed so it disappears from the workspace. On first run a random
// key is generated and persisted. Matching leftover legacy copies are retired
// on restart; conflicting copies fail closed without choosing or rotating keys.
func loadOrCreateJWTKey(keyDir, dataDir string) ([]byte, error) {
	path := filepath.Join(keyDir, jwtKeyFilename)
	legacy := filepath.Join(dataDir, jwtKeyFilename)
	// Directory aliases (including case variants/junctions on Windows) must not
	// turn cleanup into deleting the active key under another spelling.
	keyParent, err := os.Stat(keyDir)
	if err != nil {
		return nil, err
	}
	legacyParent, err := os.Stat(dataDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil && os.SameFile(keyParent, legacyParent) {
		legacy = path
	}
	read := func(name string) ([]byte, error) {
		if err := privatefs.ProtectFile(name); err != nil {
			return nil, err
		}
		data, err := privatefs.ReadFile(name)
		if err != nil {
			return nil, err
		}
		key := []byte(strings.TrimSpace(string(data)))
		if len(key) < 32 {
			return nil, locale.Errorf("Local signing key is invalid; refusing to replace it")
		}
		return key, nil
	}
	readLegacy := func() ([]byte, error) {
		for attempt := 0; ; attempt++ {
			key, err := read(legacy)
			if !temporaryJWTFileError(err) || attempt == 49 {
				return key, err
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	retireLegacy := func(key []byte) ([]byte, error) {
		if legacy == path {
			return key, nil
		}
		for attempt := 0; ; attempt++ {
			old, err := readLegacy()
			if os.IsNotExist(err) {
				return key, nil
			}
			if err != nil {
				return nil, err
			}
			if !hmac.Equal(key, old) {
				return nil, locale.Errorf("Legacy signing key conflicts with the active key; resolve the obsolete workspace key before starting")
			}
			err = os.Remove(legacy)
			if err == nil || os.IsNotExist(err) {
				return key, nil
			}
			if !temporaryJWTFileError(err) || attempt == 49 {
				return nil, err
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	if key, err := read(path); !os.IsNotExist(err) {
		if err != nil {
			return nil, err
		}
		return retireLegacy(key)
	}
	// Preserve the legacy key and fail closed on unreadable/invalid key material.
	if legacy != path {
		data, err := readLegacy()
		if err == nil {
			if err := privatefs.PublishFile(path, data); err != nil {
				if os.IsExist(err) {
					key, err := read(path)
					if err != nil {
						return nil, err
					}
					return retireLegacy(key)
				}
				return nil, err
			}
			if _, err := retireLegacy(data); err != nil {
				return nil, err
			}
			log.Printf(locale.Text(locale.ServerDefault(), "[auth] JWT key migrated from %s to %s (outside the browsable workspace)"), legacy, path)
			return data, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	buf := make([]byte, 32)
	for i := range buf {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(keyChars))))
		if err != nil {
			return nil, locale.Errorf("generate jwt key: %w", err)
		}
		buf[i] = keyChars[n.Int64()]
	}
	if err := privatefs.PublishFile(path, buf); err != nil {
		if os.IsExist(err) {
			key, err := read(path)
			if err != nil {
				return nil, err
			}
			return retireLegacy(key)
		}
		return nil, locale.Errorf("write jwt key: %w", err)
	}
	log.Printf(locale.Text(locale.ServerDefault(), "[auth] New JWT key written to %s"), path)
	return buf, nil
}

// signJWT issues a 7-day HS256 token for user ARTEX.
func signJWT(key []byte) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "SKOPOS",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}).SignedString(key)
}

// sessionVersion binds a session to one password generation without disclosing
// the password hash. Changing or resetting that hash invalidates prior sessions.
func sessionVersion(key []byte, hash string) string {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte("Skopos credential-bound session v1\x00"))
	_, _ = mac.Write([]byte(hash))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

type credentialClaims struct {
	jwt.RegisteredClaims
	Version string `json:"auth_version"`
}

func signCredentialJWT(key []byte, hash string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, credentialClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "SKOPOS", ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtTTL)), IssuedAt: jwt.NewNumericDate(time.Now()),
		},
		Version: sessionVersion(key, hash),
	}).SignedString(key)
}

func verifyCredentialJWT(token string, key []byte, hash string) bool {
	claims := &credentialClaims{}
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) { return key, nil },
		jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
	return err == nil && t.Valid && claims.Subject == "SKOPOS" && hash != "" &&
		hmac.Equal([]byte(claims.Version), []byte(sessionVersion(key, hash)))
}

// authenticatedPassword checks the current database state on each request, so
// the local reset helper takes effect without a server restart or cache expiry.
func (s *Server) authenticatedPassword(w http.ResponseWriter, token string) (string, bool) {
	if !verifyJWT(token, s.jwtKey) {
		writeErr(w, http.StatusUnauthorized, locale.Text(responseLanguage(w), "Invalid or expired token"))
		return "", false
	}
	pg := s.pg(w)
	if pg == nil {
		return "", false
	}
	hash, _, err := pg.GetSetting(authPassKey)
	if err != nil {
		writeErr(w, http.StatusServiceUnavailable, locale.Text(responseLanguage(w), "Authentication request unavailable"))
		return "", false
	}
	if !verifyCredentialJWT(token, s.jwtKey, hash) {
		writeErr(w, http.StatusUnauthorized, locale.Text(responseLanguage(w), "Invalid or expired token"))
		return "", false
	}
	return hash, true
}

// verifyJWT returns true when tokenStr is a valid, non-expired HS256 token.
func verifyJWT(tokenStr string, key []byte) bool {
	t, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, locale.Errorf("unexpected signing method")
		}
		return key, nil
	})
	return err == nil && t.Valid
}

// extractToken accepts only explicitly supplied credentials. Browser cookies
// cannot authorize the API: cookies cross port boundaries and are ambient on
// same-site requests from unrelated origins. Query tokens support EventSource.
func extractToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return r.URL.Query().Get("token")
}

// requireAuth wraps h with JWT validation.
// /api/auth/* and /api/health are exempt.
func (s *Server) requireAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if strings.HasPrefix(p, "/api/auth/") || p == "/api/health" {
			h.ServeHTTP(w, r)
			return
		}
		tok := extractToken(r)
		if tok == "" {
			writeErr(w, 401, locale.Text(responseLanguage(w), "Unauthorized"))
			return
		}
		if _, ok := s.authenticatedPassword(w, tok); !ok {
			return
		}
		h.ServeHTTP(w, r)
	})
}

// GET /api/auth/status — reports whether the admin password has been initialised.
func (s *Server) authStatus(w http.ResponseWriter, r *http.Request) {
	pg := s.pg(w)
	if pg == nil {
		return
	}
	hash, _, err := pg.GetSetting(authPassKey)
	if err != nil {
		writeErr(w, http.StatusServiceUnavailable, locale.Text(responseLanguage(w), "Authentication request unavailable"))
		return
	}
	writeJSON(w, 200, map[string]any{"initialized": hash != ""})
}

// POST /api/auth/init — sets the password for the first time; rejected if already set.
func (s *Server) authInit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password   string `json:"password"`
		SetupToken string `json:"setup_token"`
	}
	if !decodeAuthRequest(w, r, &req) {
		return
	}
	if len(req.Password) > 72 {
		writeErr(w, http.StatusBadRequest, locale.Text(responseLanguage(w), "Password must be at most 72 bytes"))
		return
	}
	if len(s.jwtKey) < 32 || !hmac.Equal([]byte(req.SetupToken), []byte(bootstrapToken(s.jwtKey))) {
		writeErr(w, http.StatusForbidden, locale.Text(responseLanguage(w), "A valid local setup token is required"))
		return
	}
	pg := s.pg(w)
	if pg == nil {
		return
	}
	existing, _, err := pg.GetSetting(authPassKey)
	if err != nil {
		writeErr(w, http.StatusServiceUnavailable, locale.Text(responseLanguage(w), "Authentication request unavailable"))
		return
	}
	if existing != "" {
		writeErr(w, 403, locale.Text(responseLanguage(w), "Password is already set"))
		return
	}
	if req.Password == "" {
		writeErr(w, 400, locale.Text(responseLanguage(w), "Password cannot be empty"))
		return
	}
	if !s.beginPasswordWork(w) {
		return
	}
	defer s.authGate.endPasswordWork()
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeErr(w, 500, locale.Text(responseLanguage(w), "Password hashing failed"))
		return
	}
	initialized, err := pg.InitializeSetting(authPassKey, string(hash))
	if err != nil {
		writeErr(w, http.StatusServiceUnavailable, locale.Text(responseLanguage(w), "Authentication request unavailable"))
		return
	}
	if !initialized {
		writeErr(w, http.StatusForbidden, locale.Text(responseLanguage(w), "Password is already set"))
		return
	}
	tok, err := signCredentialJWT(s.jwtKey, string(hash))
	if err != nil {
		writeErr(w, 500, locale.Text(responseLanguage(w), "Token generation failed"))
		return
	}
	writeJSON(w, 200, map[string]any{"token": tok})
}

// POST /api/auth/change-password — changes the admin password. Requires a valid
// token (this route is under /api/auth/* which requireAuth exempts, so the token
// is validated here) AND the current password.
func (s *Server) authChangePassword(w http.ResponseWriter, r *http.Request) {
	if !verifyJWT(extractToken(r), s.jwtKey) {
		writeErr(w, 401, locale.Text(responseLanguage(w), "Unauthorized"))
		return
	}
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if !decodeAuthRequest(w, r, &req) {
		return
	}
	if len(req.OldPassword) > 72 || len(req.NewPassword) > 72 {
		writeErr(w, http.StatusBadRequest, locale.Text(responseLanguage(w), "Password must be at most 72 bytes"))
		return
	}
	pg := s.pg(w)
	if pg == nil {
		return
	}
	if req.NewPassword == "" {
		writeErr(w, 400, locale.Text(responseLanguage(w), "New password cannot be empty"))
		return
	}
	hash, authenticated := s.authenticatedPassword(w, extractToken(r))
	if !authenticated {
		return
	}
	if !s.beginPasswordWork(w) {
		return
	}
	defer s.authGate.endPasswordWork()
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.OldPassword)); err != nil {
		writeErr(w, 401, locale.Text(responseLanguage(w), "Current password is incorrect"))
		return
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		writeErr(w, 500, locale.Text(responseLanguage(w), "Password hashing failed"))
		return
	}
	tok, err := signCredentialJWT(s.jwtKey, string(newHash))
	if err != nil {
		writeErr(w, 500, locale.Text(responseLanguage(w), "Token generation failed"))
		return
	}
	changed, err := pg.CompareAndSwapSetting(authPassKey, hash, string(newHash))
	if err != nil {
		writeErr(w, http.StatusServiceUnavailable, locale.Text(responseLanguage(w), "Authentication request unavailable"))
		return
	}
	if !changed {
		writeErr(w, http.StatusUnauthorized, locale.Text(responseLanguage(w), "Invalid or expired token"))
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "token": tok})
}

// POST /api/auth/login — validates username/password and returns a JWT.
func (s *Server) authLogin(w http.ResponseWriter, r *http.Request) {
	if !s.authGate.allowLogin(authPeer(r), time.Now()) {
		authBusy(w)
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !decodeAuthRequest(w, r, &req) {
		return
	}
	if len(req.Password) > 72 {
		writeErr(w, http.StatusBadRequest, locale.Text(responseLanguage(w), "Password must be at most 72 bytes"))
		return
	}
	if req.Username != "SKOPOS" {
		writeErr(w, 401, locale.Text(responseLanguage(w), "Incorrect username or password"))
		return
	}
	pg := s.pg(w)
	if pg == nil {
		return
	}
	hash, ok, err := pg.GetSetting(authPassKey)
	if err != nil {
		writeErr(w, http.StatusServiceUnavailable, locale.Text(responseLanguage(w), "Authentication request unavailable"))
		return
	}
	if !ok || hash == "" {
		writeErr(w, 403, locale.Text(responseLanguage(w), "Password is not initialized; set it first"))
		return
	}
	if !s.beginPasswordWork(w) {
		return
	}
	defer s.authGate.endPasswordWork()
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		writeErr(w, 401, locale.Text(responseLanguage(w), "Incorrect username or password"))
		return
	}
	tok, err := signCredentialJWT(s.jwtKey, hash)
	if err != nil {
		writeErr(w, 500, locale.Text(responseLanguage(w), "Token generation failed"))
		return
	}
	writeJSON(w, 200, map[string]any{"token": tok})
}
