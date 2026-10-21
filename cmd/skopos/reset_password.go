package main

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/Philocyber/skopos/config"
	"golang.org/x/crypto/bcrypt"
)

// Local recovery uses the existing pgx driver, with secrets in the environment
// rather than a psql child command line. It does not start application services.
func resetAdministratorPassword() error {
	password := os.Getenv("SKOPOS_RESET_NEWPASS")
	if len(password) == 0 || len(password) > 72 {
		return errors.New("reset password must contain 1 to 72 bytes")
	}
	dsn := os.Getenv("SKOPOS_RESET_DSN")
	if dsn == "" && os.Getenv("SKOPOS_RESET_DBUSER") != "" {
		host, port := os.Getenv("SKOPOS_RESET_DBHOST"), os.Getenv("SKOPOS_RESET_DBPORT")
		if host == "" {
			host = "127.0.0.1"
		}
		if port == "" {
			port = "5432"
		}
		u := url.URL{Scheme: "postgres", Host: net.JoinHostPort(host, port), Path: "/" + os.Getenv("SKOPOS_RESET_DBNAME"), User: url.UserPassword(os.Getenv("SKOPOS_RESET_DBUSER"), os.Getenv("SKOPOS_RESET_DBPASS"))}
		ssl := os.Getenv("SKOPOS_RESET_SSLMODE")
		if ssl == "" {
			ssl = "disable"
		}
		u.RawQuery = url.Values{"sslmode": {ssl}}.Encode()
		dsn = u.String()
	}
	if dsn == "" {
		var err error
		dsn, _, err = config.PostgresDSN()
		if err != nil {
			return errors.New("database configuration is unavailable for password reset")
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return errors.New("password reset hashing failed")
	}
	connection, err := sql.Open("pgx", dsn)
	if err != nil {
		return errors.New("password reset database configuration is invalid")
	}
	defer connection.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err = connection.ExecContext(ctx, `INSERT INTO settings(key,value) VALUES ('auth.password_hash',$1) ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value,updated_at=now()`, string(hash))
	if err != nil {
		return errors.New("password reset database write failed; check connection and settings permissions")
	}
	return nil
}
