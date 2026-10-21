---
name: Skopos application console
description: A compact desktop console for inspecting agent work, scope and recorded evidence.
colors:
  light-background: "#f3f5f6"
  light-foreground: "#18212b"
  light-card: "#ffffff"
  light-primary: "#006879"
  light-primary-foreground: "#ffffff"
  light-muted: "#edf1f4"
  light-muted-foreground: "#566575"
  light-accent: "#e6f3f5"
  light-accent-foreground: "#005b6c"
  light-border: "#d4dce2"
  light-input: "#becad3"
  light-ring: "#007f91"
  light-destructive: "#b42338"
  dark-background: "#080c11"
  dark-foreground: "#e8eef3"
  dark-card: "#0e141c"
  dark-popover: "#121a24"
  dark-primary: "#ba91ce"
  dark-primary-foreground: "#160d1c"
  dark-muted: "#19232f"
  dark-muted-foreground: "#9aaab9"
  dark-accent: "#261e30"
  dark-accent-foreground: "#d7b7e7"
  dark-border: "#293440"
  dark-input: "#3a4857"
  dark-destructive: "#ff8c97"
typography:
  headline:
    fontFamily: "Inter, system-ui, sans-serif"
    fontSize: "1.5rem"
    fontWeight: 600
    lineHeight: 1.3
    letterSpacing: "-0.025em"
  title:
    fontFamily: "Inter, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 600
    lineHeight: 1.375
  body:
    fontFamily: "Inter, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 400
    lineHeight: 1.5
  label:
    fontFamily: "Inter, system-ui, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 500
    lineHeight: 1.333333
  mono:
    fontFamily: "JetBrains Mono, ui-monospace, monospace"
    fontWeight: 400
rounded:
  sm: "3px"
  md: "5px"
  lg: "7px"
spacing:
  1: "4px"
  2: "8px"
  3: "12px"
  4: "16px"
  6: "24px"
components:
  button-primary:
    rounded: "{rounded.md}"
    height: "36px"
    padding: "0 12px"
  button-outline:
    rounded: "{rounded.md}"
    height: "36px"
    padding: "0 12px"
  button-ghost:
    rounded: "{rounded.md}"
    height: "36px"
    padding: "0 12px"
  input:
    rounded: "{rounded.md}"
    height: "32px"
    padding: "4px 10px"
  card:
    rounded: "{rounded.lg}"
    padding: "16px"
---

# Design System: Skopos application console

## Overview

**Creative North Star: "The pentest console"**

Skopos is an operational research console with PhiloCyber brand details. Compact controls, quiet navigation and readable evidence take priority over editorial or marketing composition. Inter supplies the interface voice; JetBrains Mono distinguishes identifiers and code. The existing Skopos name and scope-grid logo remain.

This document describes the implemented default console theme at the `web/` boundary. The separate public landing page retains its own system in [root DESIGN.md](../DESIGN.md). Existing user-selectable layout, font and theme preferences remain available; they are compatibility features, not new defaults. The approved Chat + Auto composition belongs to the [console surface contract](../.impeccable/surfaces/console.md), rather than a rule that every workspace must become a chat layout.

**Key Characteristics:**

- Dense workspaces with bounded scrolling and persistent task navigation.
- Cool neutral surfaces with cyan light-mode and violet dark-mode interaction signals.
- Semantic severity and review states that remain distinct from brand accents.
- Shared tables, fields and Radix overlays across workspaces.

## Colors

The neutral palette separates application ground, content surfaces and quiet metadata without decorative color fields.

### Primary

- **Deep cyan:** light-mode actions, selected navigation, caret and focus signals.
- **Soft violet:** the corresponding dark-mode interaction signal, with dark text on filled primary controls.
- **Tinted accent surfaces:** selected or contextual states use the paired accent foreground rather than replacing semantic status colors.

### Neutral

- **Cool ground and white surface:** light-mode page background and bordered content containers.
- **Near-black ground and layered slate:** dark-mode page, card and popover surfaces.
- **Muted ink and hairlines:** supporting labels, table headers, input boundaries and dividers remain visible in both themes.

Destructive actions use the destructive token. Status colors come from the existing shared status map: Critical uses solid rose with white text, High soft red, Medium amber and Low slate. Finding review status is independent: Pending amber, In progress blue, Confirmed red, Resolved/Fixed green, False positive slate and Ignored/Duplicate neutral. Preserve the written label alongside color.

**The Semantic Color Rule.** Brand accents identify interaction; severity and review colors retain their domain meaning.

## Typography

**Interface Font:** Inter, with system-ui and sans-serif fallbacks. Regular, Medium and SemiBold are self-hosted.

**Code Font:** JetBrains Mono Regular, with ui-monospace and monospace fallbacks; self-hosted independently of font CDNs.

The hierarchy is intentionally shallow. Standard workspace headings use the headline role; compact cards use the title role. Chat retains its 16px heading and task detail its 14px header. Body text and ordinary controls use the body scale, while supporting labels use the label scale. Code, preformatted evidence and keyboard hints use the mono family. Code and tables use tabular numerals. Route-specific overrides remain authoritative where they carry a denser tool layout.

**The Data Type Rule.** Keep code and identifiers typographically distinct from interface prose; do not turn workspace headings into display typography.

## Layout

The application shell combines collapsible side navigation with a toolbar and workspace content. Repeated spacing follows a four-pixel rhythm, usually 12 or 16 pixels inside compact controls and cards, with 24-pixel desktop workspace padding where the route calls for it. Card small variants use 12-pixel spacing.

The approved Chat workspace uses a flexible conversation column and an optional task-context column bounded from 20 to 24rem. History is a toggleable 14rem rail; history and task context are mutually exclusive. The transcript scrolls within its panel and the composer remains available below it; the context panel scrolls independently. Existing Auto, model, attachment and side-question controls remain part of that working layout.

Task detail fixes its controls and wrapping workspace tabs above a separate scrolling content pane. Each selected tab mounts a fresh pane, so a long report does not leave the next tab scrolled down. Results ledger uses a flexible table and a 20rem inspector at the extra-large breakpoint; the inspector is sticky within the workspace. Preserve the original ten task tabs alongside Results ledger.

Existing responsive fallbacks and coarse-pointer control sizing remain in the implementation. Owner acceptance for this work is desktop/laptop only: 1440px light and 1280px dark captures, not a mobile acceptance claim. Breakpoints and motion are recorded in the sidecar.

## Elevation & Depth

Borders and tonal layering provide normal workspace separation. Cards have no decorative lift. Active sidebar items combine a tinted surface with a one-pixel inset primary marker. The toolbar has a faint bottom shadow. Popovers, menus, sheets and dialogs use the existing shared elevation utilities; dialogs pair that elevation with a dark translucent backdrop. Preserve these layers and Radix focus behavior rather than adding separate route-specific overlay treatments.

**The Quiet Surface Rule.** Use the existing border and surface hierarchy for workspaces; reserve stronger elevation for overlays and existing active-control states.

## Shapes

Controls use small curved corners, cards slightly larger ones, and badges a pill silhouette. The corner scale comes from the seven-pixel base radius and its derived smaller values. One-pixel boundaries establish tables and containers; a table already inside a card does not add another outer border. Lucide SVG icons remain compact and visually secondary to labels.

## Components

### Buttons

Compact, labeled controls use medium-weight interface text. Primary buttons pair the theme's primary and primary-foreground tokens. Outline buttons retain a visible boundary, secondary buttons use the muted surface family, and ghost buttons supply quiet toolbar actions. Destructive buttons use a tinted destructive surface. Default buttons are 36px tall; small, extra-small and icon variants serve dense controls. Hover changes color or tone without moving the control. Focus uses the theme ring; disabled state reduces opacity and prevents interaction.

### Inputs / Fields

Fields use the input border, compact horizontal padding and a theme-aware caret. Default inputs are 32px tall, with an additional translucent input surface in dark mode. Focus changes the border and adds a ring; invalid fields use destructive color, and disabled fields reduce contrast and block interaction. Selects and textareas share this vocabulary.

### Chips

Badges are compact pills with explicit text. Default and secondary variants carry generic emphasis; status badges use the shared severity or lifecycle mapping. Critical remains visibly stronger than High. A completed agent intent is not a passed security test.

### Cards / Containers

Cards use the card surface, a one-pixel border, the larger corner step and 16px internal spacing, or 12px for the small variant. Headers, descriptions and footer boundaries establish structure without decorative panels nested inside panels. Tables preserve clear headers, selected rows and readable wrapping for evidence summaries.

### Navigation

Sidebar rows use compact interface text, a quiet hover surface and the tinted active state with its inset marker. The icon-only collapsed state retains accessible controls. Task workspace tabs use the existing line variant with primary active underline; the task header wraps them instead of hiding original workspaces. Shared dialogs, confirmations, sheets, menus and nested selects retain keyboard behavior and visible focus.

### Evidence workspaces

Task context is explicitly an inspection selection. A successful `spawn_task` tool result may suggest the created task; free-form conversation text does not. Findings counts lead to the authoritative task-filtered registered records, while the original exploration findings remain separately labeled. Context links expose all eleven task workspaces.

Results ledger preserves original conclusions and incomplete attempts, with a searchable list and selected-record inspector. Report preview provides refresh, copy, download and retry. Markdown export retains finding review states, task-scoped counts, negative/incomplete records and explicit caps/limitations. These views must not imply exhaustive testing, a new PDF generator or universal low-impact execution enforcement.

## Do's and Don'ts

### Do:

- **Do** reuse the shared theme tokens, components and focus treatment in every workspace.
- **Do** preserve full task navigation, evidence provenance and independent review labels.
- **Do** keep scroll ownership explicit so task controls and the chat composer stay available.
- **Do** respect reduced-motion preferences, including portalled overlays.

### Don't:

- **Don't** transplant the public landing page's display typography or marketing spacing into the console.
- **Don't** replace semantic severity with the cyan or violet brand accent.
- **Don't** present missing data as zero, a completed intent as a passed check, or task inspection as authorization.
- **Don't** infer mobile acceptance, live agent validation or enforcement guarantees from the desktop synthetic fixtures.
