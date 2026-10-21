//go:build !embedui

package server

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebUIStubUsesRequestLanguage(t *testing.T) {
	for _, lang := range []string{"en", "ko"} {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Language", lang)
		rec := httptest.NewRecorder()
		(&Server{}).webuiHandler().ServeHTTP(rec, req)
		if rec.Code != 404 {
			t.Fatalf("status=%d", rec.Code)
		}
		want := "The frontend is not embedded"
		if lang == "ko" {
			want = "프런트엔드가 포함되지 않았습니다"
		}
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("%s response=%q", lang, rec.Body.String())
		}
	}
}
