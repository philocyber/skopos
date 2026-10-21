package main

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestNativeResetRejectsInvalidPasswordAndRedactsConnection(t *testing.T) {
	t.Setenv("SKOPOS_RESET_DSN", "postgres://SECRET_USER:SECRET_PASSWORD@fixture.invalid/db?SECRET=query")
	for _, pw := range []string{"", strings.Repeat("a", 73), "valid fixture"} {
		t.Setenv("SKOPOS_RESET_NEWPASS", pw)
		err := resetAdministratorPassword()
		if err == nil || strings.Contains(err.Error(), "SECRET") {
			t.Fatalf("unsafe reset result %v", err)
		}
	}
}

func TestNativeResetPreservesURIAndKeywordDSNs(t *testing.T) {
	if os.Getenv("SKOPOS_NATIVE_RESET_E2E") != "1" {
		t.Skip("requires explicitly selected disposable PostgreSQL")
	}
	dsn := os.Getenv("SKOPOS_PG_DSN")
	connection, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err := connection.Exec(`CREATE TABLE IF NOT EXISTS settings(key TEXT PRIMARY KEY,value TEXT NOT NULL,updated_at TIMESTAMPTZ DEFAULT now())`); err != nil {
		t.Fatal(err)
	}
	keyword := os.Getenv("SKOPOS_NATIVE_RESET_KEYWORD_DSN")
	if keyword == "" {
		t.Fatal("missing keyword DSN fixture")
	}
	for index, candidate := range []string{dsn, keyword, ""} {
		if index == 2 {
			t.Setenv("SKOPOS_RESET_DBHOST", "127.0.0.1")
			t.Setenv("SKOPOS_RESET_DBPORT", "51297")
			t.Setenv("SKOPOS_RESET_DBUSER", "skopos_test")
			t.Setenv("SKOPOS_RESET_DBPASS", "test_only_password")
			t.Setenv("SKOPOS_RESET_DBNAME", "native_reset_20261010")
			t.Setenv("SKOPOS_RESET_SSLMODE", "disable")
		} else if candidate == "" {
			t.Fatal("missing keyword DSN fixture")
		}
		t.Setenv("SKOPOS_RESET_DSN", candidate)
		t.Setenv("SKOPOS_RESET_NEWPASS", "local reset fixture")
		if err := resetAdministratorPassword(); err != nil {
			t.Fatal(err)
		}
		var hash string
		if err := connection.QueryRow(`SELECT value FROM settings WHERE key='auth.password_hash'`).Scan(&hash); err != nil {
			t.Fatal(err)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("local reset fixture")); err != nil {
			t.Fatal(err)
		}
	}
}
