# Threat model (overview)

| Field | Value |
|---|---|
| Version | 1.0 (2026-10-01; summarizes the working threat model v1.4) |
| Full model | [`code-agent-docs/security/threat-model.md`](../../code-agent-docs/security/threat-model.md): every asset, attacker, surface, and numbered threat (T-01 …) with its mitigation, its task or stage, and its status. It is the working document the developer updates with every stage. |
| Requirement | FR-084 |

This page is the short version for people who deploy or evaluate the NAS. It describes threats and defenses, not unfixed weaknesses of released versions (those stay private until fixed; see [vulnerability management](vulnerability-management.md)). There is no released version yet.

## What is protected

Your files and photos; passwords, sessions, tokens, and two-factor secrets; the TLS private key; the database; the audit log; the configuration and stored secrets; the computer the NAS runs on; and the NAS's availability (especially on a small Raspberry Pi).

## Who might attack

| Attacker | Example |
|---|---|
| Someone on your network without an account | A guest's phone or a compromised smart-home device |
| A malicious web page open in your browser | Tries to make your browser act on the NAS (CSRF, DNS rebinding, clickjacking) |
| A malicious file | A crafted photo, video, PDF, SVG, or archive, uploaded knowingly or not |
| Someone with a stolen session or token | Uses it until it expires or is revoked |
| Another user or program on the NAS computer | Reads files its permissions allow |
| Someone with your disk, SD card, or a backup | Reads whatever is not encrypted |
| A compromised dependency, build tool, or release | Runs code inside the build or the NAS |
| Later: other NAS users (S07), and attackers on the internet if you expose the NAS (R09) | |

## The main defenses

| Threat | Defense | Plan |
|---|---|---|
| Uploaded content runs on the server | Never: content is data, never code (invariant I12); stored without execute permission; the NAS has no server-side scripting | I12, NFR-063 |
| A crafted media file exploits a parser (ExifTool, libvips, libheif, FFmpeg; for example ExifTool CVE-2021-22204) | Each tool runs in a sandbox: its own process, no shell, no network, access only to its one file and a scratch folder, time and memory limits; core-generated file names; tool-specific restrictions; tests with deliberately hostile files; tools kept patched | NFR-064, NFR-065, ADR-0045 |
| A file runs scripts in your browser inside the NAS's web app | Files are served with `nosniff` and a sandboxing Content Security Policy, active types only as downloads; later from a separate origin | NFR-022, NFR-066 |
| Escaping a folder (`../`, links, Windows device names) | Every path through Go's `os.Root` plus strict name rules | FR-076 |
| Guessing or stealing passwords | Argon2id, rate limits and lockout, optional two-factor, passkeys later | S03, R05 |
| Another web page uses your signed-in browser | CSRF tokens, an Origin check, a Host allow-list (against DNS rebinding), no CORS | FR-350, S03.5 |
| Eavesdropping on your network | HTTPS only on the network (TLS 1.3 preferred, 1.2 minimum); apps pin the NAS's certificate at pairing | NFR-067, NFR-069 |
| One user reads another's data | One central default-deny authorization check; isolation on every path; cross-user and authorization-matrix tests | FR-089, I5, NFR-081 |
| The NAS is tricked into calling internal machines (SSRF through webhooks or imports) | Destinations checked after DNS resolution; internal addresses refused unless the admin allows one | NFR-071 |
| A tampered release or update | Signed releases (Cosign) and signed checksums (Minisign); the updater refuses anything unsigned | FR-381, FR-382 |
| A backdoor through a dependency or the CI | Pinned and verified dependencies, actions pinned by commit, secret scanning, the dependency review stage | NFR-078, NFR-079, S13 |
| Someone covers their tracks | An append-only, hash-chained audit log; security events raise alerts | FR-379, FR-378 |
| A stolen disk | Disk encryption (the operating system's now; built in later); secrets in the database encrypted with a key kept outside it | Hardening guide, FR-266, NFR-074 |
| Running out of memory or disk (denial of service) | Size and rate limits, timeouts, a free-space reserve, processing limits against decompression bombs | NFR-068, NFR-065 |

The status of each defense (Planned, Implemented, Verified) is in the [Statement of Applicability](statement-of-applicability.md) and the [security requirements](security-requirements.md).
