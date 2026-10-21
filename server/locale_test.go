package server

import (
	"encoding/json"
	"github.com/Philocyber/skopos/locale"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLocaleMiddlewareErrorAndSSE(t *testing.T) {
	locale.Register("language must be en or ko", "language must be en or ko", "언어는 en 또는 ko여야 합니다")
	h := withLocale(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if locale.FromContext(r.Context()) != locale.Ko {
			t.Error("request language missing")
		}
		if _, ok := w.(http.Flusher); !ok {
			t.Error("SSE flushing unavailable")
		}
		writeErr(w, 400, "language must be en or ko")
	}))
	r := httptest.NewRequest("GET", "/api/settings?lang=ko", nil)
	r.Header.Set("Accept-Language", "en")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 400 || !strings.Contains(w.Body.String(), "언어는") || w.Header().Get("Content-Language") != "ko" {
		t.Fatalf("bad localized response: %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(strings.Join(w.Header().Values("Vary"), ","), "Cookie") {
		t.Fatal("cookie language missing cache variation")
	}
}

func TestLocaleMiddlewarePreservesUnknownUserError(t *testing.T) {
	w := httptest.NewRecorder()
	lw := &localeWriter{w, locale.Ko}
	writeErr(lw, 400, "用户提供的数据 원문")
	if !strings.Contains(w.Body.String(), "用户提供的数据 원문") {
		t.Fatal("unknown content changed")
	}
}

func TestDynamicErrorLocalizesBeforeInterpolation(t *testing.T) {
	const template = "Asset %s is outside the authorized scope (task %d)"
	locale.Register(template, template, "자산 %s은(는) 승인 범위 밖입니다(작업 %d)")
	raw := "用户原文/한국어/%s?q=<raw>"
	err := locale.Errorf(template, raw, 17)
	h := withLocale(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { writeError(w, 400, err) }))
	r := httptest.NewRequest("GET", "/api/test", nil)
	r.Header.Set("Accept-Language", "ko")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var response struct {
		Error string `json:"error"`
	}
	if decodeErr := json.Unmarshal(w.Body.Bytes(), &response); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if response.Error != "자산 "+raw+"은(는) 승인 범위 밖입니다(작업 17)" {
		t.Fatalf("wrong localized error or modified raw argument: %q", response.Error)
	}
	if err.Error() != "Asset "+raw+" is outside the authorized scope (task 17)" {
		t.Fatal("error identity changed with request language")
	}
}

func TestUpdateProgressRendersPerSubscriber(t *testing.T) {
	p := updateProgress{message: locale.M("Downloading %s (%s)…", "raw-한글.zip", "12 MB"), cause: locale.Errorf("Release package does not contain %s", "raw-한글.exe")}
	ko, en := p.inLanguage(locale.Ko), p.inLanguage(locale.En)
	if ko.Message != "raw-한글.zip(12 MB) 다운로드 중…" || en.Message != "Downloading raw-한글.zip (12 MB)…" {
		t.Fatalf("progress locale mismatch: %q / %q", ko.Message, en.Message)
	}
	if !strings.Contains(ko.Error, "릴리스 패키지") || !strings.Contains(ko.Error, "raw-한글.exe") {
		t.Fatalf("error locale/raw argument mismatch: %q", ko.Error)
	}
	if p.Message != "" || p.Error != "" {
		t.Fatal("subscriber mutated shared event")
	}
}
