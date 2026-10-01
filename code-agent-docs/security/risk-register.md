# Security risk register

| Field | Value |
|---|---|
| Version | 1.0 (2026-10-01, session S007, plan 1.12.0, plan change request #10) |
| Method | ISO/IEC 27005:2022 style (based on public summaries): asset, threat, vulnerability, likelihood, impact, risk level, treatment (reduce, avoid, transfer, accept), owner, status |
| Links instead of copies | Threats are detailed in the [threat model](threat-model.md) (T-xx); project risks in the plan's risk table (`plan/risks.md`, RK-xx). This register records the security risk decisions, above all **accepted risks with the owner's acceptance**. |
| Maintained | By every security-relevant task (rule R15) and at every stage's final review |

## Scales

- **Likelihood:** Low (needs unusual access or skill), Medium (plausible for a home NAS on a shared network), High (expected).
- **Impact:** Low (inconvenience), Medium (one user's data or availability for a while), High (all data, the host, or every installation).
- **Level:** High if either is High and the other at least Medium; Medium if both are Medium or one is High and the other Low; otherwise Low.

## Register

| ID | Asset | Threat (threat model) | Vulnerability | Likelihood | Impact | Level | Treatment | Owner | Status |
|---|---|---|---|---|---|---|---|---|---|
| SR-01 | The host (A-08) | A crafted media file exploits a parser (T-50, T-60, T-64, T-65, T-68) | ExifTool, libvips, libheif, and FFmpeg parse attacker-supplied files | Medium | High | High | **Reduce:** sandbox and limits (NFR-064, NFR-065, ADR-0045), core-generated names, patched tools, hostile corpus tests | Project | Open → S04, S05 (before any tool processes uploads) |
| SR-02 | The host, every installation | A compromised dependency, build tool, or CI action (T-44, T-61, T-63) | Many third-party components; actions referenced by tag today | Medium | High | High | **Reduce:** pinning by SHA and checksum, verified packages, scanning, the dependency review stage, signed releases (NFR-078, NFR-079, FR-382, S13) | Project | Partly treated (actions pinned by SHA, S03.5-T06); Open → S13, S14.7 |
| SR-03 | Credentials and secrets (A-02, A-07) | A secret committed to the public repository (T-62) | The repository is public; local hooks can be skipped | Medium | High | High | **Reduce:** gitleaks pre-commit and pre-push (D3), GitHub push protection, a full-history scan; anything found is rotated first | Project (owner for the GitHub settings) | Open → S03.5-T07; owner task |
| SR-04 | Users' trust | A false sense of compliance (claims beyond what is verified) | Security documents can drift from reality | Medium | Medium | Medium | **Avoid:** honesty rules; statuses with evidence in the Statement of Applicability; audit checklist group M checks them | Project | Treated (plan 1.12.0) |
| SR-05 | User files (A-01) | One user reads another's data (T-55, T-47) | Many access paths (API, thumbnails, search, shares, jobs) | Medium | High | High | **Reduce:** central default-deny authorization, I5, the authorization matrix test, cross-user leak tests (FR-089, NFR-081, S07.7) | Project | Open → S03.5, S07 |
| SR-06 | User files (A-01) | User content runs scripts in the app's origin (T-21, T-22) | Browsers render HTML, SVG, PDF | Low | High | Medium | **Reduce:** sandbox CSP, `nosniff`, attachment for active types (S02.6, done); a separate content origin (NFR-066, ADR-0046) | Project | Partly treated (S02.6); Open → ADR-0046 |
| SR-07 | All stored data (A-01–A-05) | A stolen disk or SD card is read (T-45) | The NAS does not encrypt at rest before R05 | Low | High | Medium | **Transfer** to the operator (disk encryption in the hardening guide) until R05 builds it in; **acceptance requested** for the MVP | Owner | Accepted-risk candidate (S03.10-T06) |
| SR-08 | Every installation | A tampered release or update (T-57) | Releases and the updater do not exist yet | Low | High | Medium | **Reduce:** Cosign keyless, signed checksums, Minisign-verified updates, SLSA provenance (FR-381, FR-382, ADR-0049) | Project | Open → S14.3, S14.7 |
| SR-09 | Availability (A-09) | HSTS with a self-signed certificate locks users out after renewal (T-66) | The approved S03.4 design sends HSTS on every HTTPS answer | Medium | Medium | Medium | **Avoid:** HSTS only with a trusted certificate (P010 change to S03.4-T01) | Project | Open → S03.4-T01 |
| SR-10 | The host (A-08) | Weaker sandbox where Landlock is unavailable (some Raspberry Pi kernels; Windows) | Kernel or platform support | Medium | Medium | Medium | **Reduce:** process limits, no network for tools (namespace or seccomp), Docker restrictions; the health page reports the sandbox level; documented in the hardening guide | Project | Open → ADR-0045 |
| SR-11 | Audit log (A-06) | The audit log is changed by someone with write access to the database file (T-37, T-70) | Database triggers cannot stop direct file edits | Low | Medium | Low | **Reduce:** a hash chain verified on schedule and on export (FR-379); file permissions (NFR-073) | Project | Open → S03.6 |

## Accepted risks

_None yet._ A risk is accepted only with the owner's written acceptance, quoted here with the date, a compensating control, and (for a Tier 1 Medium finding, decision D1) a fix date. Candidates: T-33, T-37, T-44, T-45 (threat model section 8), SR-07.

| ID | Risk | Owner's acceptance (quote, date) | Compensating control | Fix date or review date |
|---|---|---|---|---|
