# Secure coding standard

| Field | Value |
|---|---|
| Version | 1.0 (2026-10-01, plan change request #10) |
| Applies to | All code in this repository: Go (core), TypeScript and Svelte (web interface), Python (AI worker, S17), shell and PowerShell (setup scripts), and the client apps (R07, R08) |
| Enforced by | Linters wherever possible (section 6); the self-review of rule R15 for everything else |
| Based on | CWE Top 25, the OWASP Cheat Sheet Series (File Upload, Input Validation, Session Management, Transport Layer Security), OWASP ASVS 5.0 chapters V1, V2, V5, V15; ISO/IEC 27001:2022 control 8.28 (based on public summaries) |

**Rule zero (invariant I12): user-supplied content is data, never code.** Nothing in this project executes, evaluates, imports, or loads uploaded or transferred content as code, scripts, configuration, or plugins.

## 1. Go (the core)

- [ ] **Files:** every access to the user areas goes through `os.Root` and the area resolver. Never `filepath.Join` with user input directly. (`os.Root` refuses `..` and links that escape, and Windows device names such as `NUL`; it does not stop bind mounts, which need root to create. Verified 2026-10-01, Go documentation and the Go blog "Traversal-resistant file APIs".)
- [ ] **External programs only through the one wrapper package** (planned `internal/extexec`, NFR-064): absolute binary path resolved at startup with its version checked, an argument list (never a command string, never a shell), `--` before file arguments where the tool supports it, core-generated file names (never the user's name), a timeout, resource limits, and the sandbox. `os/exec` anywhere else is forbidden by lint.
- [ ] **SQL only with parameters.** No query is built from strings that contain input.
- [ ] **HTML** only with `html/template`; never `template.HTML` with user data.
- [ ] **Size limits on every reader of client input:** `http.MaxBytesReader` (per-route limits exist since S01), limited JSON decoding, and unknown fields refused where the API says so.
- [ ] **Integer overflow:** check sizes, offsets, and `Range` arithmetic before use; conversions between integer types are bounded (gosec G115).
- [ ] **Randomness:** `crypto/rand` only for anything security-related; `math/rand` in security code is forbidden by lint. Secrets and tokens are compared in constant time (`crypto/subtle`).
- [ ] **Errors to clients** carry stable codes (RFC 9457 problems through `internal/apperr`), never internal paths, stack traces, or internal IDs.
- [ ] **Regular expressions:** Go's `regexp` (linear time) only; no other regex engine on untrusted input.
- [ ] **Logging:** never passwords, tokens, cookies, session IDs, secrets, or file contents (NFR-016); request headers are redacted (S01.5).
- [ ] **TLS:** never `InsecureSkipVerify`; outbound connections verify certificates (NFR-071).
- [ ] **Deserialization:** only JSON (and the formats the plan names) into typed structures; nothing that creates arbitrary objects.

## 2. TypeScript and Svelte (the web interface)

- [ ] No `{@html}` and no `innerHTML` with user-controlled data; no `eval`, no `new Function`. File names, descriptions, tags, and EXIF fields are always rendered as text (lint: `svelte/no-at-html-tags`).
- [ ] URLs are never built from user data without encoding; user-supplied links open with `rel="noopener noreferrer"`.
- [ ] No regular expressions that can backtrack catastrophically on untrusted input (ReDoS); search parsing stays on the server.
- [ ] No secrets in the front-end bundle; credentials only in `HttpOnly` cookies or the documented header.
- [ ] The app keeps its hash-based Content Security Policy (S02); a change that needs `unsafe-inline` or `unsafe-eval` is refused.

## 3. Python (the AI worker, S17)

- [ ] No `pickle`, `joblib`, or other object deserialization of anything not created by the worker in the same run; models only as ONNX files verified by checksum.
- [ ] No network access at runtime (I6); never `subprocess` with `shell=True`.
- [ ] Job payloads from the core are validated strictly; results are validated again by the core (I9).

## 4. Shell and PowerShell (setup scripts and deployers)

- [ ] Strict modes: `set -euo pipefail` in shell; `Set-StrictMode -Version Latest` and `$ErrorActionPreference = 'Stop'` in PowerShell. Every variable is quoted.
- [ ] Every download is verified by checksum or signature before use; **never `curl | sh`** or any download piped into a shell.
- [ ] No secrets on command lines (other users can see them in process lists); pass them through files with tight permissions or standard input.
- [ ] Never disable an operating system security feature (SELinux, AppArmor, Windows Defender, the firewall); every change the script makes is logged (NFR-077).

## 5. Client apps (R07, R08)

- [ ] Tokens only in the platform's secure store; certificate or public-key pinning from the pairing QR code; no cleartext traffic; never "ignore certificate errors" (NFR-069).
- [ ] Shared items from other apps are copied while the temporary grant lasts; nothing outside what the user picked is read.
- [ ] OWASP MASVS 2.1 for the phone apps ([security requirements](security-requirements.md), section 3).

## 6. Enforcement by linters

| Check | Tool | Status |
|---|---|---|
| Go security rules | gosec (in golangci-lint) | Implemented (since S01) |
| Area separation and the API's disk access | depguard (in golangci-lint) | Implemented (since S01) |
| Forbidden APIs: `os/exec` outside the wrapper, `math/rand` in security packages, `template.HTML`, `InsecureSkipVerify` | forbidigo (in golangci-lint), configured in `.golangci.yml` | Planned (foundation task S03.5-T08) |
| Svelte and TypeScript: `{@html}`, `eval`, unsafe links | ESLint with the Svelte plugin's `svelte/no-at-html-tags` and security rules | Planned (S03.5-T08) |
| Python security rules (Bandit-equivalent) | Ruff, `S` rules | Planned (S17, when the worker exists) |
| Secrets in commits | gitleaks pre-commit and pre-push hooks (decision D3) | Planned (S03.5-T07) |
| Code scanning | GitHub CodeQL at stage completion | Planned (S03.10) |

## 7. Review checklist (rule R15)

Every security-relevant change (input parsing, file paths, uploads or downloads, external programs, authentication, authorization, cryptography, network exposure, the storage helper, dependencies) is self-reviewed against this standard and the matching OWASP cheat sheets. The stage document notes the review, the ASVS IDs, the threat model change, and the abuse-case tests.
