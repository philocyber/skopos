# Restricted execution boundary

Status: implementation requirements, verified preflight and executor credential masks; no implemented OS sandbox. REC-003, REC-022 and REC-034 remain partial. The existing 36 closed families are unchanged.

## Required boundary

The original REC-034 records 46/78 require writable filesystem/process access and target egress to be constrained outside the model. REC-022 records 29/79/104 require independently checked effective destinations across command, HTTP, custom and MCP paths. REC-003's literal insertion and operation review must coexist with that execution boundary; parser checks cannot generally prove arbitrary interpreted programs harmless.

Restricting only the Bash tool name is insufficient. Foreground commands, background jobs, interactive sessions, custom command/script execution, child processes and process-based MCP/skill adapters must inherit the same task boundary. Required enforcement must fail closed when unavailable. Privileged server/database/provider credentials must stay outside the worker filesystem and environment.

## Verified preflight, 2026-10-11

- The existing skopos-security-tests-20261010 container runs Linux 6.18.40.1-microsoft-standard-WSL2. No new container/service was launched.
- A cached-dependency Linux/amd64 Go probe returned landlock_abi=7 and no_new_privs=0. This queries kernel capability/current process state; it does not apply restrictions or prove a prohibited effect is blocked. Source, binary and output are in C:/Users/richi/AppData/Local/SkoposSecurityTests/isolation-capability-probe*.
- docker exec skopos-security-tests-20261010 unshare --user --map-root-user --mount --net true returned Operation not permitted, exit 1. Output is retained in isolation-namespace-probe.log. Do not assume a process can create its own isolation namespaces inside the current container policy, or silently request broader privileges as a fallback.
- norma v0.4.3 agentcore.Options exposes Tools, WorkingDir and BashEnv. BashEnv overrides inherited child environment entries; tool/task.go withEnv begins with os.Environ(). A working directory and injected proxy variables do not remove server credentials or prevent another destination.
- SDK foreground Bash, background Manager and ShellSessionManager call exec.CommandContext directly (tool/bash.go, tool/task.go and tool/shell_session.go). Session options inspected provide no process sandbox callback. The module cache was not modified.
- Dockerfile currently places the application, research tools and skill files in one runtime image. docker-compose.yml runs that application with provider/database environment and shared data. This is application packaging, not a per-task worker boundary.

## Integration points

Implemented prerequisite: agent session setup and native custom command/Python executors mask the finite server-credential key set documented in SECURITY-REMEDIATION.md. Native Windows foreground/background/custom tests and Linux foreground/background/PTY tests observe controlled fake credentials becoming empty while ordinary task variables and the parent provider environment remain intact. This does not remove unknown secrets or stop filesystem reads, sockets or unbound MCP processes. It must not be used as evidence that the boundary below is implemented.

| Path | Required integration |
| --- | --- |
| agent worker/planner/main/chat session assembly | Attach a code-owned task execution capability to every executable role; default SDK registrations must not restore a local executor. |
| Foreground/background Bash | Delegate execution to the confined worker; preserve timeout, output spill, cancellation, task output and process-tree cleanup. |
| Interactive shell | Keep session identity and stdin/stdout lifecycle inside the same worker boundary; preserve existing shell_open/send/close workflows. |
| server custom command/script tools | Keep frozen schemas, literal interpolation and effective-operation review; execute the reviewed program/data through the same worker boundary. |
| Read/Write/Edit and process-based skills/MCP | Resolve paths and run processes inside the worker's permitted filesystem, rather than the credential-bearing server. |
| Native HTTP/WebFetch | Retain independently checked URL/redirect/peer/Host/SNI behavior; required worker networking cannot bypass it by a direct socket. |
| Remote MCP/extensions | Treat remote execution authority separately; a request wrapper cannot attest a remote server's actual egress or filesystem effects. |
| Task lifecycle | Bound capabilities, sessions and processes must stop on cancellation/deletion and policy revocation; retries/resume must not inherit stale authority. |

## Acceptance checks

1. An allowed job can read/write its task workspace and use supported programs, but cannot read a controlled credential fixture or alter a marker outside that workspace, even through children, symlinks, background jobs or interactive sessions.
2. A prohibited TCP/UDP destination cannot receive a marker through Bash, Python, custom scripts, MCP or children when lexical/model checks miss the operation. Validate actual connection/packet delivery, not just a denied dispatch stub.
3. URL-only grants cannot silently become arbitrary raw-socket/IP authority. Redirects, DNS changes, Host/SNI changes, proxies and alternate transports must obey the same original operator policy. Mutable discovery scope must not expand that authority.
4. Allowed target workflows remain usable: HTTP/TLS capture, structured target tools, research programs, file uploads/artifacts, interactive stdin, output persistence, bilingual errors and task resume/cancel. Any necessary capability distinction must be explicit and reviewable.
5. A missing/failed backend must not invoke the local SDK executor. Environment sanitization is a server-controlled allowlist; model parameters cannot restore provider/database credentials or gain process/container-management authority.
6. Native Windows support must be proven independently if required. A successful Linux/container fixture does not prove Windows process, filesystem or network confinement.

The requested runtime preference is Docker/Linux restricted workers versus native Windows restricted workers. This decision changes the platform backend and migration path. Neither answer narrows the finding requirements above, and neither authorizes changing host security policy or launching privileged services implicitly.
