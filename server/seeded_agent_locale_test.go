package server

import (
	"github.com/Philocyber/skopos/db"
	"github.com/Philocyber/skopos/locale"
	"testing"
)

func TestSeededAgentMetadataLocalePreservesEdits(t *testing.T) {
	for _, tc := range []struct{ key, name, want string }{
		{"reporter", "Report writer", "보고서 작성자"},
		{db.FindingRetestAgentKey, "Finding retest", "취약점 재검증"},
	} {
		original := &db.Agent{Key: tc.key, Name: tc.name, Description: "Custom description 원문"}
		result := agentDTO(original, locale.Ko)
		if result.Name != tc.want || result.Description != original.Description || original.Name != tc.name {
			t.Fatalf("stock metadata localization changed custom/stored content: %+v", result)
		}
		original.Name = "Custom name 원문"
		if result := agentDTO(original, locale.Ko); result.Name != original.Name {
			t.Fatal("custom name translated")
		}
	}
}
