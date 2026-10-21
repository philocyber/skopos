---
version: 1
slug: "console"
primary_target: "web/src/app/layout.tsx"
related_targets: ["web/src/app/globals.css", "web/src/styles/console.css", "web/src/components/ui", "web/src/app/(main)", "web/src/app/(auth)"]
---

# Skopos application console

The recorded design system is [web/DESIGN.md](../../web/DESIGN.md), with its
extension sidecar at `web/.impeccable/design.json`. Root DESIGN.md continues to
describe the separate public landing page.

Mode: Operate. Users inspect agent execution, scope, findings and evidence and
manage a self-hosted security-research tool. All workspaces have equal priority.

The owner's explicit direction is a pentest tool with PhiloCyber brand details.
The personal website's newspaper layout, display headlines and marketing spacing
do not transfer here. Preserve task density, semantic severity colors, Radix
interactions and all existing task/scope/approval/evidence behavior.

Use Inter for compact UI text, JetBrains Mono for code and identifiers, cool
neutral surfaces, cyan interaction signals in light mode and violet signals in
dark mode. Navigation and toolbars remain quiet. Tables carry clear headers and
selected states. Dialogs, confirmations, sheets, menus and nested selects share
the same form vocabulary and bounded mobile dimensions. English remains the
application language.

Owner-approved composition (21 October 2026): desktop Chat + Auto with the full
existing feature set. Reference: .impeccable/review/proposals/chat-auto-workspace.jpg,
view=e in the local proposal. Preserve all original workspaces and task tabs.
History is a toggleable rail; conversation and composer remain central; task
context sits alongside with scope, constraints, counts and workspace links.
Only successful spawn_task tool results may suggest a created task. Selecting
context never changes chat instructions, active execution or authorization.

This owner-selected refinement supersedes the earlier three-proposal surface
roll (seed `598de6b0`). There is no new concept seed or fabricated QUALITY BAR
for the final direction. Existing Chat widgets are a justified adaptation of
the selected mock: Auto, history management, model selection, attachments,
side questions, full transcript/tool activity, composer and stop behavior remain.
The planning helper fills an editable draft without sending it. The shipped Chat
root retains `data-design-direction="owner-approved-chat-auto-task-context"`
in production JavaScript and the runtime DOM; this is not an emitted HTML comment.

Task context shows the authoritative registered-findings count, including
standalone/pending findings, and links to the corresponding task records. Missing
API data is unavailable rather than zero. It exposes recorded scope, constraints,
asset rules, pending approvals, qualified coverage and all eleven workspace
links. Task-to-Chat links retain conversation context. The ten original task tabs
remain alongside Results ledger. Fixed task controls and wrapping tabs sit above
a keyed scrolling content pane, resetting scroll when workspaces change.

Results ledger adds original conclusions and incomplete exploration without
inferring a passed check from a completed intent. Reporting includes recorded
conclusions, execution records, scoped asset counts and honest export limits.
Enforced low-impact profiles remain outside this implementation; no decorative
policy toggle may imply enforcement the runtime does not provide.

The searchable, paginated ledger includes a sticky selected-record inspector.
Findings shows the authoritative register inline first and retains the separately
labeled original exploration findings view alongside it. Coverage uses a
transparent canvas over the themed surface, with theme-aware node labels, edges
and legend in light and dark modes. Report preview supports refresh,
copy, Markdown download and failure/retry. Export preserves independent finding
review states and original negative/incomplete records, caps sections at the
5,000 newest records with declared truncation, and is not a frozen audit snapshot.
No new PDF generator is included.

Retain the existing component system. Desktop and laptop light/dark, keyboard
focus, collapsed navigation, bounded scrolling and reduced motion govern QA.
Mobile-specific work is explicitly excluded by the owner. No new UI framework.

## Verification boundary

The finish reviewer recorded `disposition: ship` in
`.impeccable/review/approved/finish-review.md` after registered-findings navigation
and dark coverage canvas corrections. That verdict covers the scored corrections.
Current primary captures are `chat-light-1440.jpg`, `chat-dark-1280.jpg` and
`draft-dark-1440.jpg`; corrected task captures are `task-tabs/light-findings.jpg`,
`task-tabs/dark-findings.jpg`, `task-tabs/light-coverage.jpg` and
`task-tabs/dark-coverage.jpg`, alongside ledger and report captures in the same
approved review directory.

[UI verification](../../docs/UI-VERIFICATION-2026-10-21.md) records 24 routes in
light desktop/dark laptop, all eleven task tabs in both variants (22 captures),
tab/URL/scroll checks, focused source tests and Markdown download inspection.
Contact sheets and `task-tab-checks.json` are local review evidence. Fixtures were
explicitly synthetic: no real model-driven task creation, scan or paid model call
was exercised. Complete backend-suite success and mobile acceptance are not
claimed. Shared scope/interception/approval controls do not establish universally
enforced low-impact execution across arbitrary shell, custom-tool or MCP actions.
