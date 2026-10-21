package report

import (
	"github.com/Philocyber/skopos/db"
	"github.com/Philocyber/skopos/locale"
	"strings"
	"testing"
	"time"
)

func TestReportLanguagesPreserveUserContent(t *testing.T) {
	f := &db.DBFinding{ID: 1, Name: "用户原文 한국어", Summary: "<raw> 用户摘要", Evidence: "GET /中文 HTTP/1.1", Report: "# 用户报告", Severity: "high", Status: "pending"}
	en := FindingsMarkdown([]*db.DBFinding{f}, time.Unix(0, 0), locale.En)
	if !strings.Contains(en, "# Vulnerability findings report") {
		t.Fatalf("missing heading: %q", en)
	}
	for _, content := range []string{f.Name, f.Summary, f.Evidence, f.Report} {
		if !strings.Contains(en, content) {
			t.Fatalf("stored content changed: %q", content)
		}
	}
	if strings.Contains(en, "%!") {
		t.Fatal("invalid report interpolation")
	}
	if !strings.Contains(string(FindingsCSV([]*db.DBFinding{f}, locale.En)), "Traffic evidence count") {
		t.Fatal("CSV headings not localized")
	}
}

func TestEmptyReportLanguages(t *testing.T) {
	text := Markdown(Input{Language: locale.En, GeneratedAt: time.Unix(0, 0), Goal: "保留目标"})
	if !strings.Contains(text, "No vulnerabilities were confirmed.") || !strings.Contains(text, "保留目标") {
		t.Fatalf("wrong language or altered goal: %q", text)
	}
}
