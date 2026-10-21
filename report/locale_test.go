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
	ko := FindingsMarkdown([]*db.DBFinding{f}, time.Unix(0, 0), locale.Ko)
	if !strings.Contains(en, "# Vulnerability findings report") || !strings.Contains(ko, "# 취약점 발견 보고서") {
		t.Fatalf("missing translated headings: %q / %q", en, ko)
	}
	for _, content := range []string{f.Name, f.Summary, f.Evidence, f.Report} {
		if !strings.Contains(en, content) || !strings.Contains(ko, content) {
			t.Fatalf("stored content changed: %q", content)
		}
	}
	if strings.Contains(en, "%!") || strings.Contains(ko, "%!") {
		t.Fatal("invalid report interpolation")
	}
	if !strings.Contains(string(FindingsCSV([]*db.DBFinding{f}, locale.Ko)), "트래픽 증거 수") {
		t.Fatal("CSV headings not localized")
	}
}

func TestEmptyReportLanguages(t *testing.T) {
	for _, tc := range []struct {
		l    locale.Lang
		want string
	}{{locale.En, "No vulnerabilities were confirmed."}, {locale.Ko, "확인된 취약점이 없습니다."}} {
		text := Markdown(Input{Language: tc.l, GeneratedAt: time.Unix(0, 0), Goal: "保留目标"})
		if !strings.Contains(text, tc.want) || !strings.Contains(text, "保留目标") {
			t.Fatalf("wrong language or altered goal: %q", text)
		}
	}
}
