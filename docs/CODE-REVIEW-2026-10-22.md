# Code review — quality and performance (2026-10-22)

Scope: the `web/` Next.js frontend and the Go backend, reviewed for concrete
performance and quality improvements. Each finding was verified against the
source (and, where noted, against a production build or a running instance),
not inferred. Severity is impact × likelihood; effort is the fix cost.

Two fixes are already applied in this branch (F1, P1); the rest are
recommendations with the evidence to act on them.

## Applied

### F1 — Mock fixtures shipped in the production bundle · perf · **applied**

`web/src/lib/api.ts` imported the demo handler statically
(`import { mockHandle } from "@/lib/mock/handler"`). That handler pulls
`mock/handler.ts` (2.6k lines) and `mock/data.ts` (6.2k lines) of demo fixtures.
A production build (`NEXT_EXPORT=1 next build`, no `NEXT_PUBLIC_MOCK`) still
contained mock-only strings — `DeepSeek V4`, `admin/admin123`, `shop.acme.com
Fastjson` — in a ~206 KB shared chunk on the initial load path, because the
static import kept the subtree in the graph even though `if (MOCK)` is dead code
in production.

Fix: load the handler with a dynamic `await import("@/lib/mock/handler")` inside
the `if (MOCK)` branch. After rebuilding, the mock strings live only in async
chunks that are not referenced by any build manifest or HTML, so production
never downloads them. The demo build (`NEXT_PUBLIC_MOCK=1`) is unaffected — the
import simply resolves on first API call.

### P1 — Always-on sidebar poll ignores tab visibility · perf · **applied (example)**

The interception-pending badge (`web/src/app/(main)/_components/sidebar/nav-main.tsx`)
is mounted on every page and polled `/api/intercept/pending` every 5 s with a
raw `setInterval`, even while the tab is backgrounded. Added a `usePolling` hook
(`web/src/hooks/use-polling.ts`) that pauses on `document.hidden` and resumes
with an immediate fetch, and migrated this badge to it. See F2 for the rest.

## Recommended — frontend

### F2 — 35 remaining polls never pause when the tab is hidden · perf · low effort

`setInterval` appears 36 times across 16 files; none check
`document.hidden`. A task-detail view alone runs several concurrent 2–5 s polls
(`_tabs/sessions-tab.tsx` has seven). A console left open in a background tab
keeps issuing these requests indefinitely, each driving backend DB work and SSE.
Migrate the remaining call sites to the `usePolling` hook added in P1. The call
sites follow one shape (`load(); const t = setInterval(load, N); return () =>
clearInterval(t)`), so most are a direct swap; the few with extra guards
(`findings/page.tsx` early-returns for `view === "asset"`,
`use-side-questions.ts` has an epoch guard) need the guard folded into the
hook's `enabled` argument.

### F3 — Full payload re-fetch + full re-render on every poll · perf · medium effort

The polled lists (Tasks, Findings, task sessions) re-request the entire list and
replace state every few seconds, and the components are marked `"use no memo"`
(279 directives across 93 files), which opts them out of the React Compiler's
automatic memoization. Frequent polling plus unmemoized large tables means the
whole table reconciles every few seconds even when nothing changed. Options, in
increasing effort: (a) skip the state update when the new payload deep-equals the
current one, to avoid the re-render; (b) add `If-None-Match`/`updated_since`
style conditional fetches so unchanged lists return quickly; (c) revisit whether
`"use no memo"` is still needed on the heavy list components. This is the main
frontend perf theme, but it needs care, so it is not a blind fix.

### F4 — Hook-dependency warnings · quality · medium effort

`useExhaustiveDependencies` fires 188 times. Some are intentional and already
carry a `biome-ignore`; the rest are unaudited and can hide stale-closure bugs
in effects that poll or subscribe. Worth auditing rather than blanket-silencing.
Note: the ~211 `noUnusedVariables` warnings are mostly the deliberate
`const { locale: swLocale } = useI18n()` pattern that subscribes a component to
locale changes without reading the value — those must not be "cleaned up", or
components stop re-rendering on language switch.

### F5 — Cosmetic lint noise · quality · low value

`useSortedClasses` (676) is Tailwind class ordering only; the repo does not
enforce it and it has no runtime effect. Not worth a churn-heavy pass unless the
team decides to adopt the rule and format once.

## Recommended — backend

### B1 — N+1 in the coverage-graph builder · perf · low effort

`db/task_scope.go` (~line 538) loops over the set of missing root domains and
runs one `SELECT id, COALESCE(company_id,0) FROM assets WHERE type='root_domain'
AND domain=$1 LIMIT 1` per root. For a task spanning many distinct root domains
this is a query per root on each coverage-graph build (the Assets coverage view
and the task "Asset coverage graph" tab). Replace with a single
`WHERE type='root_domain' AND domain = ANY($1)` and index the result in a map.
Impact is bounded by the number of distinct out-of-scope roots, so medium-low,
but the fix is mechanical and removes a loop of round-trips.

### B2 — Row-closing and single-startup loops look healthy · no action

Checked as part of the review and found clean, recorded so it is not re-flagged:
DB multi-row queries in `db/` balance against `rows.Close()` calls (≈174 vs 201,
the surplus from helpers and `QueryRow`). The `QueryRow`-in-loop in `db/db.go`
(`seedBuiltins`) runs once at startup over two built-in agents, not a hot path.
The 37 apparent "unclosed queries" in `server/server.go` are `r.URL.Query()`
(net/http URL parsing), not database calls.

## Suggested order

1. F2 — migrate the rest of the polls to `usePolling` (low effort, broad win).
2. B1 — collapse the coverage-graph N+1 (low effort, mechanical).
3. F3 — reduce re-render/refetch cost on the polled lists (the real ceiling).
4. F4 — audit the hook-dependency warnings.
