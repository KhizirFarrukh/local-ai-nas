# Security requirements (OWASP ASVS 5.0 and MASVS 2.1)

| Field | Value |
|---|---|
| Version | 1.0 (2026-10-01, plan 1.12.0, plan change request #10) |
| Catalogs | OWASP Application Security Verification Standard **5.0.0** (May 2025; 17 chapters, about 350 requirements, levels L1–L3); OWASP Mobile Application Security Verification Standard **2.1.0** (8 categories) for the phone apps. Versions checked 2026-10-01. |
| Levels (the owner's decision D1) | **Tier 1** parts are verified at **ASVS Level 3**; **Tier 2** parts at **Level 2**. When unsure, a part is Tier 1 ([security policy](security-policy.md), 7.1). |
| ISO/IEC mapping | Control 8.26 (application security requirements) and ISO/IEC 27034 (application security), based on public summaries |

## How this catalog is used

1. Each stage document lists its Tier 1 parts (rule R15).
2. Each security-relevant task records the ASVS requirement IDs it implements, its threat model change, its secure-coding self-review, and its abuse-case tests (R15).
3. The stage's final testing substage runs the ASVS checklist for the stage's scope at the tier's level and records the evidence; this catalog and the [Statement of Applicability](statement-of-applicability.md) are updated then.
4. Statuses: **Planned**, **Implemented** (built, evidence named), **Verified** (a test or review proves it). Nothing is claimed early.

Requirement IDs below (FR, NFR) are the plan's (`code-agent-docs/plan/requirements.md`).

## 1. ASVS 5.0 chapters in scope

| Chapter | In scope | Tier | Plan requirements | Stages | Status |
|---|---|---|---|---|---|
| V1 Encoding and Sanitization | Yes | 1 | NFR-063 (no dynamic code), secure coding standard (parameterized SQL, output encoding, argument lists for tools) | S01 on | Planned (parameterized SQL and argument-free tool use so far) |
| V2 Validation and Business Logic | Yes | 1 | FR-076 (names), FR-092 (photo types), NFR-065 (input limits) | S01, S04, S05 | Planned (name validation verified by S01.6 tests) |
| V3 Web Frontend Security | Yes | 1 (content serving, CSRF, CSP), 2 (cosmetic pages) | NFR-022, NFR-066, FR-350 | S02, S03 | Planned (hash-based CSP and sandboxed downloads implemented in S02; Host and Origin checks in S03.5-T02) |
| V4 API and Web Service | Yes | 1 | FR-089, NFR-068, NFR-072 (WebDAV) | S01, S03, S09 | Planned |
| V5 File Handling | Yes | 1 | Section 2 below | S01, S02, S04, S05, R02, R05, R09 | See section 2 |
| V6 Authentication | Yes | 1 | FR-064, FR-085, FR-091, FR-263, NFR-080 | S03, R05 | Planned (Argon2id built, S03.2-T02) |
| V7 Session Management | Yes | 1 | FR-086, FR-087 | S03.3 | Planned |
| V8 Authorization | Yes | 1 | FR-089, I5, NFR-081 | S03.5, S07 | Planned |
| V9 Self-contained Tokens | Not now | — | The NAS uses opaque random tokens stored as hashes, not self-contained tokens; becomes relevant only if R05 single sign-on handles ID tokens | R05 | Not applicable until R05 |
| V10 OAuth and OIDC | Later | 1 | FR-264 (single sign-on as a relying party) | R05 | Planned |
| V11 Cryptography | Yes | 1 | NFR-080, NFR-074 | S03 on | Planned (SHA-256 and Argon2id implemented) |
| V12 Secure Communication | Yes | 1 | NFR-067, NFR-071, NFR-069 | S03.4, S10, R06, R07 | Planned |
| V13 Configuration | Yes | 1 | NFR-073, NFR-074, NFR-076, NFR-077 | S03, S14 | Planned (configuration validation implemented) |
| V14 Data Protection | Yes | 1 | I5, NFR-011, NFR-016, NFR-075 | S03, S07, S08 | Planned |
| V15 Secure Coding and Architecture | Yes | 1 | NFR-059, NFR-063, NFR-064, NFR-079 | All | Planned |
| V16 Security Logging and Error Handling | Yes | 1 (audit), 2 (diagnostic pages) | FR-090, FR-378, FR-379, NFR-016, NFR-022 | S01, S03.6, S10 | Planned (error answers without internals implemented and tested since S01) |
| V17 WebRTC | No | — | The NAS uses no WebRTC | — | Not applicable |

Individual requirement IDs (for example V8.2.1) are recorded per task in the stage documents and collected here at each stage's final review.

## 2. V5 File Handling, requirement by requirement (Tier 1, Level 3)

Requirement summaries are paraphrased from ASVS 5.0.0 (checked 2026-10-01).

| ASVS | Level | Requirement (short) | How the NAS meets it | Plan | Status |
|---|---|---|---|---|---|
| 5.1.1 | L2 | Document accepted file types, extensions, size limits, and how malicious files are handled | This catalog, the [storage security guide](storage-security-guide.md), and the API documentation | NFR-060, NFR-065 | Planned (S04 adds the photo types) |
| 5.2.1 | L1 | File sizes cannot exceed processing capacity or cause denial of service | Upload size limits and the free-space reserve (S01.2, S01.4); per-route body limits; processing limits (NFR-065) | FR-004, NFR-065, NFR-068 | Implemented for storage (S01); Planned for processing |
| 5.2.2 | L1 | The extension matches the content (magic bytes) | The photos area accepts only allow-listed media types detected from the bytes (FR-092); the files area accepts any type (it is a NAS) and serves every file safely | FR-092, NFR-022 | Planned (S04) |
| 5.2.3 | L2 | Archives: maximum uncompressed size and file count checked before extraction | Archive extraction (R02) checks total size, count, ratio, and paths first | NFR-065, R02 | Planned (R02) |
| 5.2.4 | L3 | Per-user storage quotas and file-count limits | Quotas (S10) | S10 | Planned |
| 5.2.5 | L3 | Archives with symlinks rejected unless allowed explicitly | Extraction never creates links (R02); downloads already never follow links (S02.4) | NFR-065, R02 | Planned |
| 5.2.6 | L3 | Images over a maximum pixel size rejected (pixel floods) | The header is read first; over the limit (default 250 megapixels) the image is stored but not processed, and the owner is told why | NFR-065 | Planned (S04) |
| 5.3.1 | L1 | Uploaded files are never executed as server code | Invariant I12: content is never executed, imported, or loaded; files are stored without execute permission; the NAS has no server-side scripting | I12, NFR-063 | Implemented (no execution path exists); Planned (the permission rule, S03.5-T05) |
| 5.3.2 | L1 | System-generated names, or strict validation of user names, against traversal, inclusion, and SSRF | Names validated for every OS (FR-076); `os.Root` confines every path; media tools receive core-generated names, never user names (NFR-064) | FR-076, NFR-064 | Verified for the files API (S01.6 tests); Planned for the tools |
| 5.3.3 | L3 | User path data ignored during decompression (zip slip) | Extraction resolves every entry through the area resolver and refuses escaping paths | NFR-065, R02 | Planned (R02) |
| 5.4.1 | L2 | User-submitted names validated or ignored; `Content-Disposition` set | Downloads send `attachment` with an encoded name (S02.4, S02.6) | NFR-022 | Implemented |
| 5.4.2 | L2 | Served names encoded (RFC 6266) | The `filename*` form with encoding; control and quote characters never reach the header | NFR-022 | Implemented |
| 5.4.3 | L2 | Files from untrusted sources scanned by an antivirus before serving | Opt-in ClamAV scanning (R05); required for upload-only links before the owner sees the files (R09 quarantine) | FR-267, FR-295 | Planned (R05, R09) |

## 3. MASVS 2.1.0 for the phone apps (R07)

| Category | What the apps do | Plan |
|---|---|---|
| MASVS-STORAGE | Tokens only in the Android Keystore or the iOS Keychain; offline files in app-private storage | FR-364, NFR-069 |
| MASVS-CRYPTO | Platform cryptography only; the crypto policy | NFR-080 |
| MASVS-AUTH | Device-scoped, revocable tokens from pairing; app lock (FR-284) | FR-364, FR-284 |
| MASVS-NETWORK | Certificate or public-key pinning from the pairing QR code; no cleartext traffic (Android network security configuration) | NFR-069 |
| MASVS-PLATFORM | Share targets accept only what they declare; temporary grants copied, never kept | FR-285, FR-372 |
| MASVS-CODE | Dependencies verified and pinned; signed builds; in-app updates verify signatures | NFR-069, NFR-079 |
| MASVS-RESILIENCE | Signed APK and app bundle; no claim of tamper-proofing | NFR-069 |
| MASVS-PRIVACY | No telemetry; data only to the user's own NAS | NFR-058 |

Desktop apps (R08) follow the same rules where they apply (secure store, pinning, signed builds).

## 4. The P010 requirements, mapped to ASVS and ISO controls

Every new requirement of plan change request #10, with its tier, its ASVS chapters, and its ISO/IEC 27001:2022 Annex A controls.

| Requirement | Short title | Tier | ASVS 5.0 | ISO controls | Stage | Status |
|---|---|---|---|---|---|---|
| NFR-059 | Secure development lifecycle and the tier policy (R15) | 1 | V15 | 8.25, 8.26, 8.28, 8.29, 5.8 | Every stage from S03 | Implemented (rule); Planned (first gate) |
| NFR-060 | ISO-aligned documents, published | — | V15 (documentation) | 5.1, 5.36, 5.37 | Now; each stage review | Implemented (first versions) |
| NFR-061 | Vulnerability management | 1 | V15 | 8.8, 5.7, 5.6, 6.8 | Now, S14 | Implemented (documents); Planned (advisory process in use) |
| NFR-062 | Incident response | 1 | V16 | 5.24–5.28 | Now, S14.4 | Implemented (documented) |
| NFR-063 | User content is never executed (I12) | 1 | V1, V5 (5.3.1), V15 | 8.7, 8.19 | S03.5-T05 and every stage | Planned |
| NFR-064 | Media processing sandbox | 1 | V5, V15 | 8.7, 8.2 | Before S04.3/S04.4 process uploads; S05.3 | Planned |
| NFR-065 | Untrusted-input limits, tool restrictions, hostile media corpus | 1 | V2, V5 | 8.7, 8.6, 8.29 | S04, S05, R02 | Planned |
| NFR-066 | User-content origin isolation | 1 | V3 | 8.7, 8.12 | ADR in S03; required before R09 | Planned |
| NFR-067 | TLS policy | 1 | V12 | 8.24, 8.20, 5.14 | S03.4 | Planned |
| NFR-068 | Listener resource limits | 1 | V4, V13 | 8.6, 8.21 | S03.4, S03.5 | Planned (parts implemented) |
| NFR-069 | Client app security (MASVS) | 1 | V12; MASVS | 8.1, 8.24 | R07, R08 | Planned |
| NFR-070 | Internal service isolation | 1 | V13, V15 | 8.22, 8.2 | S15.2, S17 | Planned |
| NFR-071 | Outbound connection safety (TLS verification, SSRF) | 1 | V12, V4 | 8.20, 8.21 | S10 (alerts) and every outbound feature | Planned |
| NFR-072 | WebDAV hardening | 1 | V4, V5, V8 | 8.21, 8.3 | S09 | Planned |
| NFR-073 | Service account, file permissions, mount options | 1 | V13 | 8.9, 8.18, 7.10 | S03.5-T05, S14.2, S15 | Planned |
| NFR-074 | Secrets management | 1 | V11, V13 | 8.24, 5.17 | S03 | Planned |
| NFR-075 | Retention and deletion | 2 | V14 | 8.10, 5.33, 5.34 | S08, S03.6 | Planned |
| NFR-076 | Minimal attack surface | 1 | V13 | 8.9, 8.21 | S03 | Planned |
| NFR-077 | Host and container hardening, safe deployers | 1 | V13 | 8.9, 8.19, 8.20 | S14.1, S14.2 | Planned |
| NFR-078 | Repository and CI integrity, secret scanning (D3) | 1 | V15 | 8.4, 5.21, 8.32 | Now (S03 foundation) | Planned (owner tasks and foundation tasks) |
| NFR-079 | Dependency policy | 1 | V15 | 5.19, 5.21, 5.22, 8.8 | Every stage, S13 | Planned |
| NFR-080 | Cryptography policy | 1 | V11 | 8.24 | Every stage | Planned (parts implemented) |
| NFR-081 | Authorization matrix test | 1 | V8 | 8.3, 8.29 | S03.10, S07.7 | Planned |
| FR-377 | Security contact: `SECURITY.md` and `security.txt` | 1 | V15 | 6.8, 5.5 | Now (SECURITY.md), S03.5 | Implemented (SECURITY.md); Planned (security.txt) |
| FR-378 | Security event alerts | 1 | V16 | 8.16, 5.25 | S03.6, S10 | Planned |
| FR-379 | Tamper-evident audit log | 1 | V16 | 8.15, 5.28, 5.33 | S03.6 | Planned |
| FR-380 | Clock synchronization check | 2 | V16 | 8.17 | S10 | Planned |
| FR-381 | Signed updates (Minisign, D2) | 1 | V15 | 8.19, 8.32 | S14.3 | Planned |
| FR-382 | Release integrity: Cosign, signed checksums, SLSA provenance, SBOM (D2) | 1 | V15 | 8.19, 5.21 | S14.7, S13.1 | Planned |
