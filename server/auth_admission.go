package server

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Philocyber/skopos/locale"
)

const maxAuthPeers = 1024

type authBucket struct {
	tokens float64
	last   time.Time
}

func (b *authBucket) refill(now time.Time, burst float64, interval time.Duration) {
	if b.last.IsZero() {
		b.tokens = burst
	} else {
		b.tokens = min(burst, b.tokens+max(0, now.Sub(b.last).Seconds())/interval.Seconds())
	}
	b.last = now
}

// authAdmission bounds login guesses, peer bookkeeping, and concurrent bcrypt
// work. The global bucket also bounds attackers rotating source addresses.
type authAdmission struct {
	mu       sync.Mutex
	global   authBucket
	peers    map[string]authBucket
	password int
}

func (a *authAdmission) allowLogin(peer string, now time.Time) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.global.refill(now, 20, 3*time.Second)
	if a.global.tokens < 1 {
		return false
	}
	if a.peers == nil {
		a.peers = make(map[string]authBucket)
	}
	bucket, exists := a.peers[peer]
	if !exists && len(a.peers) >= maxAuthPeers {
		for key, previous := range a.peers {
			if now.Sub(previous.last) >= 10*time.Minute {
				delete(a.peers, key)
			}
		}
		if len(a.peers) >= maxAuthPeers {
			return false
		}
	}
	bucket.refill(now, 5, 6*time.Second)
	a.peers[peer] = bucket
	if bucket.tokens < 1 {
		return false
	}
	bucket.tokens--
	a.peers[peer] = bucket
	a.global.tokens--
	return true
}

func (a *authAdmission) beginPasswordWork() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.password >= 2 {
		return false
	}
	a.password++
	return true
}

func (a *authAdmission) endPasswordWork() {
	a.mu.Lock()
	a.password--
	a.mu.Unlock()
}

func authPeer(r *http.Request) string {
	// Proxy headers are attacker-controlled without an explicit trusted-proxy
	// configuration. Use the direct TCP peer, independent of source port.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.String()
	}
	return "unknown"
}

func authBusy(w http.ResponseWriter) {
	w.Header().Set("Retry-After", "6")
	writeErr(w, http.StatusTooManyRequests, locale.Text(responseLanguage(w), "Too many authentication requests; try again shortly"))
}

func (s *Server) beginPasswordWork(w http.ResponseWriter) bool {
	if !s.authGate.beginPasswordWork() {
		authBusy(w)
		return false
	}
	return true
}
