# Public-repository privacy audit

## Publication refresh, 2026-10-21

The current source snapshot passes the publication checker and Gitleaks v8.30.1:
1,048 prospective files, approximately 10 MB, no unresolved source findings.
Gitleaks also scanned the 12 commits reachable from the pre-publication branch
HEAD with no secret findings. These are bounded automated checks, not a proof
that every possible confidential datum is absent.

Personal home paths in the three documents listed below now use portable artifact
references, and the internal scan identifier has been removed. Technical evidence
and unresolved execution-isolation limits are retained. Root Codex Security
package manifests are untracked and ignored; application manifests remain tracked.
Local visual QA captures, decision-server payloads and runtime files are excluded.
The ARTEX attribution and license remain in place.

**Historical metadata remains:** the existing remote `main` commit already contains
some of the old machine paths and internal scan identifier. This publication is a
normal update; it does not rewrite that existing public history. A full metadata
scan of all local refs will therefore continue to report those historical values.
The older snapshot below records the original findings and verification limits.

Commands executed for this refresh:

- `node tools/audit-publication.mjs --secrets`: passed.
- `gitleaks git . --log-opts=HEAD --redact --no-banner --no-color`: passed, 12 commits.
- `npm.cmd test` in `web`: 28 passed, 1 skipped, 0 failed.
- `git -c core.safecrlf=false diff --check`: passed.

## Earlier audit snapshot


Audited on **2026-10-21**, with HEAD at `3bf3fb6`. Another process was actively
editing this checkout. This is a snapshot of the privacy review, not approval
to publish the changing working tree.

**Result: publication remains blocked by local metadata and development artifacts.**
The reviewed secret scans pass. Application security, integration readiness and
the product rebrand are separate checks.

## Findings requiring cleanup

| Location | Finding | Required action |
| --- | --- | --- |
| `docs/SECURITY-REMEDIATION.md` | Personal home-directory paths throughout the ledger, plus an internal scan identifier at line 3 | Keep the detailed evidence privately or replace machine-specific references with portable descriptions. Preserve the technical findings and unresolved execution-isolation limits. |
| `docs/EXECUTION-ISOLATION.md:14` | Personal home-directory path | Replace it with an artifact-relative reference without implying that the local artifact is included in the repository. |
| `docs/LOCAL-VALIDATION-2026-10-21.md:3` | Personal home-directory path | Use a portable description of the temporary validation output. |
| Commit `3bf3fb641d4a8e0d397871367bcc6c86ceb16864` | The first two documents above are already in reachable history, including their local paths and internal scan identifier | Before publishing, coordinate a correction to this unpublished commit with the agent editing the branch. A later cleanup commit alone retains the metadata. No history rewrite was performed by this audit. |
| Root `node_modules/`, `package.json`, `package-lock.json` | Local Codex Security tooling is eligible for Git; the application uses manifests under `web/`, `adapters/agent/` and `skills/api-recon/scripts/` | Exclude or relocate the root development tooling. Preserve the application's manifests and lockfiles. |

Line numbers describe the audit snapshot and can move during cleanup. The
publication checker reports current locations without printing source contents.

## Secret scanning and reviewed exceptions

Gitleaks **v8.30.1** was downloaded from its official GitHub release. The Windows
archive's SHA-256 matched the published checksum before execution.

- Initial reachable-history scan: **450 commits**, 8 matches.
- Initial prospective-source snapshot: **1,027 files**, 7 matches.
- Final reviewed scan: **451 commits**, **1,033 prospective files** and approximately **10.16 MB** of prospective
  source, with **no unreviewed Gitleaks matches**.

All initial matches were inspected:

- Notification tests use fixed masking/redaction sentinels. Exceptions require
  both an exact test path and an exact synthetic value, and apply only to the
  generic API-key rule.
- A proxy test uses the RFC 6455 sample WebSocket nonce. Its exception is limited
  to that test path, exact nonce and rule.
- The loopback TLS private key is an intentionally public fixture, documented in
  `skills/api-recon/scripts/test_fixtures/tls/README.md`. The publication checker
  pins its SHA-256. It must never be deployed as a service credential.
- Three historical curl-authentication matches span separate fictional mock
  records and misinterpret sqlmap's `-u` target URL. `.gitleaksignore` suppresses
  only those three commit/path/rule/line fingerprints.

No whole test directories or commits are exempted from secret scanning. A
temporary-repository control verified that a newly introduced synthetic GitHub
token in an exception-bearing test path still fails the scan. An initial global
allowlist configuration failed that control; it was corrected to rule-specific
exceptions before the final scans.

`private-credential.ps1` contains permission-management code, not saved credentials.
The installer, package builder and regression tests depend on it; keep it in the
public source. No private mailbox addresses were found in the inspected source.

The root `LICENSE` has the same Git blob hash as `160fe13:LICENSE`, the retained
ARTEX base. Existing author metadata and upstream attribution were not changed.

## Reproduce before publication

Requires Node.js and Gitleaks v8.30.1 or later. Put `gitleaks` on PATH, or set
`GITLEAKS_BIN` to its executable path. From the repository root:

```sh
node --check tools/audit-publication.mjs
node tools/audit-publication.mjs --secrets --history
```

The second command currently exits **1** because the metadata and artifact
findings above remain open. The Gitleaks source and history subprocesses both
exit **0**. The checker includes tracked and non-ignored untracked source without
staging anything, excludes dependency/build directories from the secret snapshot
while flagging their eligibility for Git, and scans added history lines for
machine paths and internal scan identities. Its source snapshot remains in the
OS temporary directory, outside the repository, for local inspection.

Actual additional checks:

- `node --check tools/audit-publication.mjs`: passed.
- Seven isolated temporary-repository controls passed: clean-tree acceptance,
  runtime-config rejection, home-path rejection, changed TLS-key rejection,
  new-token detection inside an exception-bearing test path, and clean-source
  secret-scan acceptance, and historical-path rejection after current-file cleanup.
- `git -c core.safecrlf=false diff --check -- .gitleaks.toml .gitleaksignore tools/audit-publication.mjs`:
  passed for tracked diffs; these new files were not staged by this audit.
- `npm.cmd test` in `web/`: 25 passed, 1 skipped, 0 failed before the concurrent
  rebrand updates. This does not validate the final rebrand.

## Limits

Regex-based scanning cannot establish that all confidential material is absent.
No production service, paid model call, external target scan or remote GitHub
change was performed. No repository was published, staged, pushed or rewritten
by this audit.

The settings and dashboard English screenshots were visually inspected: no
credentials were visible, but both still show the previous product branding.
Other binary assets were not exhaustively reviewed, and screenshot captures
contain fictional findings and example credentials that must remain clearly
labeled as demo data. Remote GitHub state, unreachable objects, caches, existing
clones, attachment metadata, and every image's visible content are outside this
audit. Re-run the checks against the final candidate commit before publication.
