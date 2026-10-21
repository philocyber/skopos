# Product validation, 2026-10-21

**Verdict: the committed application starts and its core UI/API flows work, but
the project does not yet have a clean release gate. Password-based agent-adapter
login is broken, CI contains a stale module import, and the combined server test
suite fails. Real-model autonomous execution remains unverified.**

## Tested source and environment

The checkout was being edited by another process. Tests used isolated copies;
product code and existing tests were not changed by this validation.

- Committed application: `1f7a2429575c1feeebb4da1020b0f464d9b74f13`.
- Initial working-tree snapshot: captured the unfinished English-only backend
  migration. Production compilation and multiple test packages failed on removed
  Korean-language symbols and a removed metadata-translation function.
- A later working-tree snapshot: production compilation passed, but test
  compilation still failed in `locale` and `server` on remaining `Ko` references.
  This later snapshot did not receive another full regression run.
- Windows amd64, Go 1.27.0, local Node.js/npm dependencies and PostgreSQL 16.15.
- Docker's Linux engine was unavailable. A portable PostgreSQL distribution
  downloaded from EDB's official binary-download link supplied a fresh,
  loopback-only test database server. Separate databases were initialized per
  Go package; the real application and adapter E2E used their own databases.
- No production configuration or provider keys were used. Test tasks and
  companies were synthetic. No external targets were scanned.

The initial frontend source matches the committed frontend after normalizing
Windows line endings. The production UI and browser tests used the committed
source. Source hashes, status snapshots, commands, raw logs and screenshots were
retained outside the repository in the local test-artifact directory.

## Checks executed

| Check | Command or method | Observed result |
| --- | --- | --- |
| Frontend regression tests | `npm.cmd test` in `web/` | 25 passed, 1 skipped, 0 failed |
| TypeScript | `npm.cmd exec -- tsc --noEmit` in `web/` | Passed |
| Production frontend export | Set `NEXT_EXPORT=1`, then `npm.cmd run build` in `web/` | Passed using ordinary copied dependencies |
| Adapter unit tests | `npm.cmd test` in `adapters/agent/` | 7 passed |
| Adapter syntax checks | `npm.cmd run check` in `adapters/agent/` | Passed |
| Committed backend | `go build -o <test-binary> ./cmd/skopos` | Passed |
| Embedded application | Copy `web/out/` into `server/webui/dist/`; `CGO_ENABLED=0 go build -tags embedui -o <test-binary> ./cmd/skopos` | Passed |
| Database-backed Go suite | For each of 20 packages: `go test -p 1 <package> -count=1 -timeout=5m -json` with its own initialized database | 19 packages passed; `server` failed. 2,226 passing, 26 failing and 57 skipped tests/subtests |
| Adapter against real HTTP handler and PostgreSQL | `SKOPOS_ADAPTER_E2E=1 go test ./server -run '^TestAgentAdapterE2E$' -count=1 -v -timeout=3m` | Failed; all 3 Node E2E cases failed on authentication |
| Static analysis | `go vet ./...` | Failed with 13 possible `unsafe.Pointer` misuse warnings in `traffic/wal_vfs.go` |
| Latest working snapshot | `go build -o <test-binary> ./cmd/skopos`; `go test ./... -run '^$' -count=1 -json` | Build passed; test compilation failed |

Go counts include parent tests and subtests, not independent scenarios. Skips
are not counted as passes. Live-provider tests remained disabled.

The first isolated frontend build used a dependency junction and failed because
Turbopack rejects dependencies outside its configured root. Repeating the
unmodified production build with ordinary copied dependencies passed. This was
a validation-environment issue, not a product source fix.

## Running product: verified behavior

The compiled embedded binary started against the disposable PostgreSQL database
and served the real UI and APIs on loopback.

First-owner API enrollment succeeded with the local setup token, and enrollment
without that token was rejected. A subsequent set of **17 API checks passed**:
login as `SKOPOS`, rejection of the obsolete adapter username, authenticated
stats/tasks/companies/model configuration/findings/settings reads, synthetic
company and task creation, task detail/coverage/scope reads, rename and pin,
pause, CSV export, and rejection of unauthenticated task-list access.

Two exploratory smoke requests used an incorrect findings path and omitted the
export format. They returned 404 and 400 respectively. The checks were corrected
against the registered routes, using `/api/exploration/findings` and
`/api/exploration/findings/export?format=csv`; both passed. These request mistakes
are not product defects.

Browser observations against this real backend:

- First-run setup rendered with the local setup-token instructions.
- Login succeeded using the disposable account after the user authorized
  accepting the displayed terms for this local test.
- The task created through the API appeared with its persisted name, pinned
  state and paused status.
- A second synthetic task was created through the browser. It appeared in the
  task list, opened in the detail view, was independently read back through the
  API, and appeared on the loaded dashboard.
- System settings and the Notifications page rendered. No notification was sent
  and no settings were changed.
- No browser console errors or warnings were captured during these checks.

No model was configured. Task persistence and navigation therefore do not prove
goal decomposition, worker tool execution or useful vulnerability detection.
The UI showed the model as unavailable, and the attempted planner work ended
with a model error. Both synthetic tasks were paused after testing.

## Findings

### 1. Password-based adapter authentication fails

`adapters/agent/src/client.js:45` submits the fixed username `ARTEX`, while
`server/auth.go` accepts `SKOPOS`. The opt-in MCP/Pi E2E test reproduced HTTP 401
in all three cases. A separate request against the running binary confirmed
that the same disposable password succeeds with `SKOPOS` and fails with `ARTEX`.

This finding applies to the adapter's password-based login path. Supplying an
existing bearer token was not independently validated as a workaround.

### 2. Backend CI initialization uses the old Go module

`.github/workflows/ci.yml:95` imports `github.com/Autumn-27/artex/db`, but this
repository's module is `github.com/Philocyber/skopos`. An isolated compile probe
using that exact import failed with `no required module provides package`.

The database helper used for this manual run imported the current module. The
manual results therefore do not mean the checked-in CI workflow passes.

### 3. Combined server tests have state/lifecycle contamination

The two failing parent tests were:

- `TestAssembleVisibleSkill`.
- `TestCommandToolDeclaredParametersThroughNativeSDK`, including its 24 subtests.

The combined package run logged `sql: database is closed` while resolving tools;
the expected skill/custom tools disappeared from the resulting catalog. Both
parent tests passed when run separately against new disposable databases:

```sh
go test ./server -run '^TestAssembleVisibleSkill$' -count=1 -v -timeout=2m
go test ./server -run '^TestCommandToolDeclaredParametersThroughNativeSDK$' -count=1 -v -timeout=2m
```

This supports test-order/global-state contamination as an explanation; it does
not establish the exact offending fixture or prove that all application
lifecycle paths are unaffected. The global tool resolver is wired to a database
closure. Assertions were not weakened and the failing suite was not marked green.

### 4. Static-analysis and migration checks remain open

The 13 `go vet` warnings concern pointer conversions in the native SQLite VFS
bridge. Traffic tests passed, but that does not resolve the warnings or prove a
memory-safety defect. The English-only migration was still in progress at the
last snapshot; remaining locale references prevent test compilation.

## Not established by this run

- Real-provider authentication, tool-call compatibility, agent behavior or task
  completion with an actual model. No paid model calls were made.
- Per-task OS filesystem/network/process confinement. Existing documented
  isolation limitations remain relevant.
- Docker image execution, Linux runtime behavior, release archives, updater
  installation, mobile layout, every UI route, actual traffic capture or real
  notification delivery.
- Correctness of edits made after the recorded snapshots.

The next release gate is to finish the ongoing migration, repair the adapter and
CI mismatches and the suite-state failure, then rerun against one stable candidate.
