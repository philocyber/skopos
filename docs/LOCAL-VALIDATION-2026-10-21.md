# Local Windows validation, 2026-10-21

Product names and command examples in this historical report have been normalized to the current Skopos branding. The rename validation log is `C:/Users/richi/AppData/Local/Temp/skopos-rename-tests.jsonl`; the original run's temporary log filename used the former product name.

Environment: Windows amd64, PowerShell. Existing working-tree changes were preserved; no product code or tests were edited. This report describes this run, not the older results in VERIFICATION.md.

## Setup and results

Installed Go 1.27.0 with:

```powershell
winget install --id GoLang.Go --exact --silent --accept-package-agreements --accept-source-agreements --disable-interactivity
```

Verified `C:/Program Files/Go/bin/go.exe version`: `go version go1.27.0 windows/amd64`. The project declares Go 1.26.3. Existing terminal sessions may need reopening to pick up the installer PATH change; validation shells explicitly added `C:/Program Files/Go/bin` to PATH.

| Command | Working directory | Observed result |
| --- | --- | --- |
| `npm.cmd test` | `web` | 25 passed, 1 skipped, 0 failed |
| `npm.cmd exec -- tsc --noEmit` | `web` | Passed |
| `$env:NEXT_EXPORT='1'; npm.cmd run build` | `web` | Static production export passed, 31 generated pages |
| `npm.cmd test` | `adapters/agent` | 7 passed |
| `npm.cmd run check` | `adapters/agent` | Passed |
| `go test ./... -count=1 -timeout=5m -json` | repository root | Failed: 1,268 passed, 30 failed, 320 skipped tests/subtests; 18 packages passed and 2 failed |
| `$env:CGO_ENABLED='0'; go build -tags embedui -o "$env:TEMP/skopos-check.exe" ./cmd/artex` | repository root | Passed after copying the exported frontend into the documented embed directory |
| `go vet ./...` | repository root | Failed: 13 possible unsafe.Pointer misuses in `traffic/wal_vfs.go` |
| `& "$env:TEMP/skopos-check.exe" -h` | repository root | Passed; CLI flags displayed |

Counts include parent tests and subtests, not unique independent scenarios. Some skipped tests require PostgreSQL, a live provider, browser tooling, or platform-specific capabilities; they are not counted as passes.

The initial embedded build failed with `pattern all:webui/dist: no matching files found`. The documented packaging step resolved this:

```powershell
New-Item -ItemType Directory -Path server/webui/dist -Force | Out-Null
Copy-Item -Path web/out/* -Destination server/webui/dist -Recurse -Force
```

## Blockers and findings

1. All 30 failing tests stopped at missing PostgreSQL configuration (`ARTEX_PG_DSN` unset and no database configuration). They belong to `agent` (1) and `server` (29), covering worker context, archives, evidence, findings, notifications, task guards, and capture startup. Their assertions beyond setup remain untested. No assertions or skip behavior were changed.
2. Docker CLI is present, but Docker Desktop's Linux engine pipe was unavailable. Launching Docker Desktop and attempting `docker desktop start` did not make the daemon available; the pending CLI start attempt was interrupted. No disposable database could be started through Docker during this run.
3. A real binary startup with a deliberately absent temporary config, no DSN, a temporary data directory, loopback address and recording proxy disabled exited with the expected database-configuration error. No HTTP service became usable. A database-backed UI/API/task lifecycle and live model/tool flow remain unverified.
4. `go vet` flagged `traffic/wal_vfs.go` lines 74, 75, 157, 162, 166, 195, 217, 300, 303, 306, 315, 334 and 337. These are pointer conversions in the native SQLite VFS bridge. This is a static-analysis finding, not proof of an observed memory corruption. The traffic package tests passed, but that does not clear these warnings.
5. The frontend test runner emitted module-type warnings. The static export warned that Next.js API routes/middleware are unavailable in an export; the embedded architecture relies on the Go backend for API enforcement. Browser behavior was not checked in this run.

No live provider test or external target scan was performed. No API credentials were configured. The next required step for complete integration validation is a disposable PostgreSQL instance, using separate initialized databases per package as in `.github/workflows/ci.yml`.

## Model support found in this checkout

- `agent/provider.go` accepts a model string and selects Anthropic Messages, OpenAI Chat Completions, or OpenAI Responses adapters. There is no fixed model-ID allowlist in this configuration path. Actual usability depends on provider access and compatibility with the agent's tool calls and configured streaming/reasoning options.
- The fallback model IDs are `claude-opus-4-8`, `gpt-4o`, and `gpt-5`, respectively. These are code defaults, not benchmark-backed recommendations or verified current account availability.
- The LLM profile UI supports entering a model ID, loading the provider's model list, and testing the connection. Listing a model or receiving a text completion does not establish a complete agent/tool workflow.
- `docs/llm-providers.md` documents GLM-5.3 templates and labels the General API endpoint recommended over the Coding Plan reference. This is an endpoint/setup recommendation; no comparative ranking of models was found. The document states that Skopos is not an officially supported Coding Plan tool. This run did not revalidate the vendor's current policies or model availability.
