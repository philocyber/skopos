package agent

import (
	"encoding/json"
	"github.com/Philocyber/skopos/locale"
	actool "github.com/Autumn-27/norma/tool"
	"strings"
	"testing"
)

func TestBuiltinPromptLanguageAndInterpolation(t *testing.T) {
	old := PromptOverride
	PromptOverride = nil
	defer func() { PromptOverride = old }()
	en := renderSystem("planner", plannerDefaultTmpl, PlannerVars{Goal: "用户目标 / 원문"}, locale.En)
	ko := renderSystem("planner", plannerDefaultTmpl, PlannerVars{Goal: "用户目标 / 원문"}, locale.Ko)
	if !strings.Contains(en, "You are the planner") || !strings.Contains(ko, "당신은") {
		t.Fatal("prompt language selection failed")
	}
	for _, text := range []string{en, ko} {
		if !strings.Contains(text, "用户目标 / 원문") || strings.Contains(text, "{{.Goal}}") {
			t.Fatal("goal altered or not interpolated")
		}
	}
}

func TestEditedPromptIsNotTranslated(t *testing.T) {
	old := PromptOverride
	defer func() { PromptOverride = old }()
	const custom = "用户自己写的模板 {{.Goal}}"
	PromptOverride = func(string) (string, bool) { return custom, true }
	got := renderSystem("planner", plannerDefaultTmpl, PlannerVars{Goal: "original"}, locale.Ko)
	if got != "用户自己写的模板 original" {
		t.Fatalf("edited prompt changed: %q", got)
	}
}

func TestBuiltinPromptTranslationCoverage(t *testing.T) {
	for _, p := range []string{goalsDefaultTmpl, goalsScopeTail, plannerDefaultTmpl, workerDefaultTmpl, mainAgentDefaultTmpl, autoDefaultTmpl, pentestDefaultTmpl, DefaultAssistantPrompt, ReporterDefaultPrompt, RetesterDefaultPrompt} {
		ko, ok := locale.Lookup(locale.Ko, p)
		if !ok || ko == p || strings.TrimSpace(ko) == "" {
			t.Fatalf("missing Korean prompt: %.60s", p)
		}
		for _, contract := range []string{"{{.Goal}}", "set_constraints", "add_task_scope", "report_finding", "update_finding_report", "record_finding_retest_result"} {
			if strings.Contains(p, contract) != strings.Contains(ko, contract) {
				t.Fatalf("translation lost contract %q", contract)
			}
		}
	}
}

func TestBuiltinToolMetadataLanguagePreservesEdits(t *testing.T) {
	en := NewToolSet(nil, "", locale.En).nodeDetail()
	original := en.InputSchema()
	// Simulate a stored English default overriding a runtime Korean tool.
	ko := localizeBuiltinTools([]actool.CoreTool{en}, locale.Ko)[0]
	if ko.Description() == en.Description() {
		t.Fatal("built-in description was not localized")
	}
	if en.Description() != NewToolSet(nil, "", locale.En).nodeDetail().Description() {
		t.Fatal("shared English metadata mutated")
	}
	data, _ := json.Marshal(original)
	var edited map[string]any
	_ = json.Unmarshal(data, &edited)
	props := edited["properties"].(map[string]any)
	props["id"].(map[string]any)["description"] = "用户原文 custom description"
	custom := DecorateTool(en, "用户原文 custom tool", edited)
	localized := localizeBuiltinTools([]actool.CoreTool{custom}, locale.Ko)[0]
	if localized.Description() != "用户原文 custom tool" || localized.InputSchema()["properties"].(map[string]any)["id"].(map[string]any)["description"] != "用户原文 custom description" {
		t.Fatal("edited metadata changed")
	}
	if localized.Name() != en.Name() {
		t.Fatal("machine tool name changed")
	}
}
