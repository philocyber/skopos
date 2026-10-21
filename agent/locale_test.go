package agent

import (
	"strings"
	"testing"

	"github.com/Philocyber/skopos/locale"
)

func TestBuiltinPromptLanguageAndInterpolation(t *testing.T) {
	old := PromptOverride
	PromptOverride = nil
	defer func() { PromptOverride = old }()
	en := renderSystem("planner", plannerDefaultTmpl, PlannerVars{Goal: "用户目标 / 원문"}, locale.En)
	if !strings.Contains(en, "You are the planner") {
		t.Fatal("prompt language selection failed")
	}
	if !strings.Contains(en, "用户目标 / 원문") || strings.Contains(en, "{{.Goal}}") {
		t.Fatal("goal altered or not interpolated")
	}
}

func TestEditedPromptIsNotTranslated(t *testing.T) {
	old := PromptOverride
	defer func() { PromptOverride = old }()
	const custom = "用户自己写的模板 {{.Goal}}"
	PromptOverride = func(string) (string, bool) { return custom, true }
	got := renderSystem("planner", plannerDefaultTmpl, PlannerVars{Goal: "original"}, locale.En)
	if got != "用户自己写的模板 original" {
		t.Fatalf("edited prompt changed: %q", got)
	}
}
