// Package report renders a deliverable Markdown report from the exploration
// graph: confirmed findings (with severity/evidence) (docs §15 P5).
package report

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Philocyber/skopos/db"
	"github.com/Philocyber/skopos/locale"
)

// Input bundles what the report needs.
type Input struct {
	Language       locale.Lang
	Title          string
	Goal           string
	GeneratedAt    time.Time
	AssetCounts    map[string]int
	Findings       []*db.Node
	FindingRecords []*db.DBFinding // Non-nil selects the authoritative findings register.
	Facts          []*db.Node
	Intents        []*db.Node
	Scope          []db.TaskScope
	Coverage       *db.Coverage
	Truncated      bool
}

type findingView struct {
	VulnClass string
	Name      string
	Severity  string
	Summary   string
	PoC       string
}

func parseFinding(n *db.Node) findingView {
	var p struct {
		VulnClass string `json:"vulnclass"`
		Name      string `json:"name"`
		Severity  string `json:"severity"`
		Summary   string `json:"summary"`
		Evidence  struct {
			PoC string `json:"poc"`
		} `json:"evidence"`
	}
	_ = json.Unmarshal(n.Payload, &p)
	return findingView{p.VulnClass, p.Name, p.Severity, p.Summary, p.Evidence.PoC}
}

var sevRank = map[string]int{"critical": 0, "high": 1, "medium": 2, "low": 3, "": 4}

// Markdown renders the report.
func Markdown(in Input) string {
	lang := locale.Resolve(in.Language)
	var b strings.Builder
	fmt.Fprintf(&b, locale.Text(lang, "# Penetration test report — %s\n\n"), nz(in.Title, locale.Text(lang, "Untitled task")))
	fmt.Fprintf(&b, locale.Text(lang, "- **Task objective**: %s\n"), nz(in.Goal, locale.Text(lang, "(not specified)")))
	fmt.Fprintf(&b, locale.Text(lang, "- **Generated at**: %s\n\n"), in.GeneratedAt.Format("2006-01-02 15:04:05"))

	// summary
	b.WriteString(locale.Text(lang, "## Summary\n\n"))
	if in.FindingRecords != nil {
		fmt.Fprintf(&b, "- Registered findings: **%d** (review status is shown per finding)\n", len(in.FindingRecords))
	} else {
		fmt.Fprintf(&b, locale.Text(lang, "- Confirmed findings: **%d**\n"), len(in.Findings))
	}
	b.WriteString(locale.Text(lang, "- Assets: "))
	var types []string
	for t := range in.AssetCounts {
		types = append(types, t)
	}
	sort.Strings(types)
	for i, t := range types {
		if i > 0 {
			b.WriteString("、")
		}
		fmt.Fprintf(&b, "%s %d", t, in.AssetCounts[t])
	}
	b.WriteString("\n\n")
	b.WriteString("## Scope and coverage\n\n")
	if len(in.Scope) == 0 {
		b.WriteString("No scope rows were recorded for this task. Review the authorized scope separately.\n\n")
	}
	for _, row := range in.Scope {
		label := row.Domain
		if label == "" {
			label = row.Net
		}
		if label == "" {
			label = row.CompanyName
		}
		if label == "" {
			label = row.Value
		}
		if label == "" {
			label = row.Kind
		}
		fmt.Fprintf(&b, "- %s: %s (source: %s)\n", row.Kind, label, row.Source)
	}
	if in.Coverage != nil {
		fmt.Fprintf(&b, "\nAssets with recorded results: **%d / %d** within this task and its direct read-only source context. This is approximate asset coverage, not a completed security test suite.\n\n", in.Coverage.Tested, in.Coverage.Denominator)
	}

	// findings
	b.WriteString(locale.Text(lang, "## Findings\n\n"))
	if in.FindingRecords != nil {
		if len(in.FindingRecords) == 0 {
			b.WriteString("No findings are registered for this task. This does not establish that the task completed or that the assets are safe.\n\n")
		}
		for i, finding := range in.FindingRecords {
			fmt.Fprintf(&b, "### %d. [%s] %s\n\n", i+1, severityLabel(lang, finding.Severity), findingTitle(finding, lang))
			fmt.Fprintf(&b, "- Finding ID: %d\n- Review status: %s\n\n%s\n\n", finding.ID, nz(finding.Status, "pending"), finding.Summary)
			if finding.Evidence != "" {
				fmt.Fprintf(&b, "**Recorded evidence:**\n\n%s\n\n", finding.Evidence)
			}
			if finding.Report != "" {
				fmt.Fprintf(&b, "**Detailed report:**\n\n%s\n\n", finding.Report)
			}
		}
	} else if len(in.Findings) == 0 {
		b.WriteString(locale.Text(lang, "_No vulnerabilities were confirmed._\n\n"))
	} else {
		fs := make([]findingView, 0, len(in.Findings))
		for _, n := range in.Findings {
			fs = append(fs, parseFinding(n))
		}
		sort.SliceStable(fs, func(i, j int) bool { return sevRank[fs[i].Severity] < sevRank[fs[j].Severity] })
		for i, f := range fs {
			fmt.Fprintf(&b, "### %d. [%s] %s\n\n", i+1, severityLabel(lang, nz(f.Severity, "info")), nz(f.Name, nz(f.VulnClass, locale.Text(lang, "Unclassified"))))
			fmt.Fprintf(&b, "%s\n\n", nz(f.Summary, ""))
			if f.PoC != "" {
				fmt.Fprintf(&b, locale.Text(lang, "**PoC / evidence:**\n\n```\n%s\n```\n\n"), f.PoC)
			}
		}
	}

	b.WriteString("## Recorded exploration conclusions\n\n")
	b.WriteString("These are recorded observations, including negative or inconclusive results where the agent supplied them. A completed attempt or a linked fact is not proof that an asset is safe.\n\n")
	written := 0
	for _, n := range in.Facts {
		if n.State == db.StateOrigin {
			continue
		}
		writeRecordedNode(&b, n)
		written++
	}
	if written == 0 {
		b.WriteString("No exploration conclusions were recorded in this report.\n\n")
	}
	b.WriteString("## Exploration execution record\n\n")
	b.WriteString("States describe execution, not a vulnerability verdict. Blocked, exhausted, stopped and deleted attempts require separate review.\n\n")
	for _, n := range in.Intents {
		writeRecordedNode(&b, n)
	}
	if len(in.Intents) == 0 {
		b.WriteString("No exploration attempts were recorded in this report.\n\n")
	}
	b.WriteString("## Limitations and remaining review\n\n")
	b.WriteString("- This export is generated from stored task records. It is not an independent verification of findings or a complete raw activity archive.\n- Unrecorded checks cannot be inferred as passed. Review the asset backlog, session traces, traffic and approvals for gaps.\n- Scope and textual constraints alone do not establish an enforced low-impact execution boundary.\n")
	if in.Truncated {
		b.WriteString("- **Record limit reached:** each section includes at most 5,000 newest records. Older records remain in the application; this export is incomplete.\n")
	}
	return b.String()
}

// Preserve authored observations and their execution state without classifying
// free-form model prose as a clean bill of health.
func writeRecordedNode(b *strings.Builder, n *db.Node) {
	var payload map[string]json.RawMessage
	_ = json.Unmarshal(n.Payload, &payload)
	summary := ""
	for _, key := range []string{"summary", "text", "name", "description", "conclusion"} {
		if json.Unmarshal(payload[key], &summary) == nil && strings.TrimSpace(summary) != "" {
			break
		}
	}
	if strings.TrimSpace(summary) == "" {
		summary = string(n.Payload)
	}
	fmt.Fprintf(b, "### Record %d — %s\n\n", n.ID, n.State)
	fmt.Fprintf(b, "- Recorded at: %s\n- Origin: %s\n\n%s\n\n", n.CreatedAt.Format(time.RFC3339), nz(n.Origin, "unspecified"), summary)
	if n.BlockedReason != "" {
		fmt.Fprintf(b, "Execution issue: %s\n\n", n.BlockedReason)
	}
	if n.DeleteReason != "" {
		fmt.Fprintf(b, "Removal reason: %s\n\n", n.DeleteReason)
	}
}

func nz(s, d string) string {
	if strings.TrimSpace(s) == "" {
		return d
	}
	return s
}

// severityLabel and statusLabel are display-only; export machine fields remain unchanged.
// lang is accepted for API stability but has no effect: English is the only
// supported language.
func severityLabel(lang locale.Lang, value string) string {
	return strings.ToUpper(value)
}
func statusLabel(lang locale.Lang, value string) string {
	return value
}
