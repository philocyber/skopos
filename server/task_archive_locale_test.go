package server

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/Philocyber/skopos/db"
	"github.com/Philocyber/skopos/locale"
)

func TestTaskTrancheBilingualValidation(t *testing.T) {
	long := strings.Repeat("x", db.MaxTaskTemplateNameRunes+1)
	tests := []struct {
		name   string
		err    error
		en, ko string
	}{
		{"archive empty", func() error { _, err := normalizeArchiveIDs(nil); return err }(), "archive_ids must not be empty", "archive_ids는 비워 둘 수 없습니다"},
		{"archive id", func() error { _, err := normalizeArchiveIDs([]int64{-42}); return err }(), "Invalid archive ID -42", "유효하지 않은 보관 ID -42"},
		{"template", validateTaskTemplateRequest(taskTemplateRequest{Name: &long}), "name must be at most 120 characters", "name은(는) 120자 이하여야 합니다"},
		{"archive path", validateArchivePath(t.TempDir(), ""), "Archive package path is empty", "보관 패키지 경로가 비어 있습니다"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatal("expected validation failure")
			}
			if got := locale.ErrorMessage(locale.En, tt.err); got != tt.en {
				t.Fatalf("English=%q, want %q", got, tt.en)
			}
			if got := locale.ErrorMessage(locale.Ko, tt.err); got != tt.ko {
				t.Fatalf("Korean=%q, want %q", got, tt.ko)
			}
		})
	}
}

func TestTaskTrancheLocalizedBatchAndRequestErrors(t *testing.T) {
	for _, lang := range []locale.Lang{locale.En, locale.Ko} {
		rec := httptest.NewRecorder()
		w := &localeWriter{ResponseWriter: rec, lang: lang}
		req := httptest.NewRequest(http.MethodPost, "/api/task-categories", strings.NewReader(`{"name":""}`))
		if _, ok := decodeTaskCategoryRequest(w, req); ok {
			t.Fatal("empty category name accepted")
		}
		if rec.Code != 400 || !strings.Contains(rec.Body.String(), locale.Text(lang, "Category name must not be empty")) {
			t.Fatalf("%s response=%d %s", lang, rec.Code, rec.Body.String())
		}
	}
}

func TestTaskTrancheResolutionPreservesProfileData(t *testing.T) {
	s := &Server{}
	p := &db.LLMProfile{ID: 7, Name: "Raw profile 원본", Format: "openai", Model: "user-model"}
	got := s.resolutionFromProfileForLanguage(p, "agent_binding", locale.Ko)
	if got.Name != p.Name || got.Model != p.Model || got.Source != "agent_binding" || got.ProfileID == nil || *got.ProfileID != p.ID || got.Available {
		t.Fatalf("profile data changed: %+v", got)
	}
	if got.Reason != "LLM 프로필에 API 키가 설정되지 않았습니다" {
		t.Fatalf("reason=%q", got.Reason)
	}
}

func TestTaskTrancheArchiveErrorsPreserveCauses(t *testing.T) {
	cause := errors.New("driver-owned detail / 원본")
	err := locale.Errorf("stage task archive path %s: %w", "user/path 원본", cause)
	joined := errors.Join(err, locale.NewError("task archive checksum mismatch"))
	if !errors.Is(joined, cause) {
		t.Fatal("lost wrapped error identity")
	}
	ko := locale.ErrorMessage(locale.Ko, joined)
	for _, part := range []string{"작업 보관 경로", "user/path 원본", cause.Error(), "체크섬이 일치하지 않습니다"} {
		if !strings.Contains(ko, part) {
			t.Fatalf("missing %q in %q", part, ko)
		}
	}
}

// Source coverage guards against untranslated literal HTTP and typed-error messages
// in this tranche; driver errors, user strings, SQL identifiers, and archive bytes are excluded.
func TestTaskTrancheMessageCatalogCoverage(t *testing.T) {
	names := []string{"archives", "archive_package", "templates", "categories", "assets", "metadata", "resolution"}
	count := 0
	for _, name := range names {
		path := filepath.Join("task_" + name + ".go")
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			index := -1
			switch fun := call.Fun.(type) {
			case *ast.Ident:
				if fun.Name == "writeErr" {
					index = 2
				}
			case *ast.SelectorExpr:
				if pkg, ok := fun.X.(*ast.Ident); ok && pkg.Name == "locale" {
					switch fun.Sel.Name {
					case "NewError", "Errorf":
						index = 0
					case "Text":
						index = 1
					}
				}
			}
			if index < 0 || len(call.Args) <= index {
				return true
			}
			lit, ok := call.Args[index].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			key, err := strconv.Unquote(lit.Value)
			if err != nil {
				t.Fatal(err)
			}
			en, found := locale.Lookup(locale.En, key)
			ko, kfound := locale.Lookup(locale.Ko, key)
			if !found || !kfound || en == ko {
				t.Errorf("%s: missing Korean message %q", path, key)
			}
			count++
			return true
		})
	}
	if count < 60 {
		t.Fatalf("only %d messages checked", count)
	}
}
