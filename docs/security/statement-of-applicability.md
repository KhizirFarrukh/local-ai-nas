# Statement of Applicability (ISO/IEC 27001:2022 alignment matrix)

**Designed in alignment with ISO/IEC 27001:2022 Annex A controls (self-assessed; not certified).** This is not a certificate and not an audit result. It is the project's own account of how each control is handled, with honest statuses. See the [security policy](security-policy.md), section 4.

| Field | Value |
|---|---|
| Version | 1.0 (2026-10-01, plan 1.12.0, plan change request #10) |
| Controls | All 93 controls of ISO/IEC 27001:2022 Annex A: organizational 5.1–5.37 (37), people 6.1–6.8 (8), physical 7.1–7.14 (14), technological 8.1–8.34 (34) |
| Source of the control titles | The publicly listed Annex A titles (checked 2026-10-01 against public lists); the standard's text was not read. Statements about what a control requires are **based on public summaries; verify against the standard**. |
| Reviewed | At the final review of every stage (rule R15) |

## How to read it

**Applies to** (one or more):
- **Software:** the NAS itself must meet it (code, defaults, features).
- **Operator:** whoever deploys and runs the NAS must meet it; the software can only help (the [hardening guide](hardening-guide.md) gives guidance).
- **Project:** the development and release process (repository, CI, releases, the owner and the coding agent).
- **N/A:** not applicable, with the reason.

**Status** (honesty rule 3: a control is never claimed before its test passes):
- **Planned:** designed and placed in a stage, not built yet. Notes may name parts that already exist.
- **Implemented:** built or in force, with evidence named; not yet checked by a test or review that proves it.
- **Verified:** a test or review proves it; the evidence is named.
- For operator-only controls the software's part is the guidance; "Implemented (guidance)" means the guidance is published, not that any operator follows it.

**Where** points to plan requirements (FR/NFR), stages (S01–S17), releases (R01–R12), ADRs, and invariants (I1–I12). The status reflects reality on 2026-10-01: stages S01 and S02 are done, S03 (security) is in progress.

## 5. Organizational controls (37)

| Control | Applies to | How it is handled | Where | Status |
|---|---|---|---|---|
| 5.1 Policies for information security | Project | The [security policy](security-policy.md) and rule R15, approved by the owner | P010, RULES R15 | Implemented (this policy, 2026-10-01) |
| 5.2 Information security roles and responsibilities | Project, Operator | Owner approves; the coding agent develops under R15; operators run the NAS; reporters report privately (policy section 8) | Policy 8 | Implemented |
| 5.3 Segregation of duties | Project, Software | The agent cannot approve its own plans, ADRs, risk acceptances, or releases (owner approval, RULES R3–R5, R13); release signing needs the owner's approval in a protected environment. Software: admin and user roles; admin actions need recent re-authentication | RULES R3–R5, R13; FR-382; S03.3, S03.9 | Planned (approval gates implemented; roles and release protection planned) |
| 5.4 Management responsibilities | Project | The owner requires the developer to follow the policy (RULES is binding on every session) | RULES R10, R15 | Implemented |
| 5.5 Contact with authorities | Operator | The operator knows whom to contact (police, data protection authority) after an incident; the incident guide says when | [Incident response](incident-response.md) | Implemented (guidance) |
| 5.6 Contact with special interest groups | Project | Follow OWASP, the Go security announcements, and upstream security lists of the tools used | NFR-061, [vulnerability management](vulnerability-management.md) | Planned |
| 5.7 Threat intelligence | Project | Watch advisories for Go, Go modules, npm and Python packages, ExifTool, libvips, libheif, FFmpeg, and container base images (Dependabot, govulncheck, upstream announcements) | NFR-061, NFR-079, S13 | Planned (Dependabot and govulncheck already run) |
| 5.8 Information security in project management | Project | Every stage has security in its plan: Tier 1 parts listed, threat model updated, the stage security gate | R15, NFR-059 | Implemented (from plan 1.12.0; S03 is the first stage under it) |
| 5.9 Inventory of information and other associated assets | Project, Software | The dependency register lists every component; the SBOM per release (S13); the threat model lists assets; the software keeps an item registry (stable IDs) and a drive inventory (S15) | `code-agent-docs/dependencies.md`, FR-363, FR-382, FR-346 | Implemented (register); Planned (SBOM) |
| 5.10 Acceptable use of information and other associated assets | Operator | The operator sets rules for the people who use their NAS | Hardening guide | Implemented (guidance) |
| 5.11 Return of assets | Operator | Organizational; the software supports it by disabling accounts and signing out devices | S07, FR-293, FR-365 | Implemented (guidance); software support Planned |
| 5.12 Classification of information | Software, Project | Two separate areas (I1); everything private to its owner by default (I5); a locked folder (R05). Project: public documents versus private security advisories | I1, I5, FR-265, policy 7.4 | Planned (I1 enforced since S01) |
| 5.13 Labelling of information | Software | Shared and public items are clearly marked in the interface; public links show their expiry | S07, R09 (FR-294) | Planned |
| 5.14 Information transfer | Software | HTTPS for every network connection (NFR-067); signed webhooks (FR-221); share links with expiry (R09); WebDAV only over HTTPS (NFR-072); client pinning (NFR-069) | NFR-067, NFR-069, NFR-072 | Planned |
| 5.15 Access control | Software | One central authorization check on every route, default deny (FR-089); per-user isolation on every path (I5) | FR-089, I5, NFR-081 | Planned (S03.5-T01) |
| 5.16 Identity management | Software | Unique, case-insensitive user names; the first admin created at first run or by the command line; user lifecycle in S07 | FR-064, S03.2, S07 | Planned (the user store exists, S03.2-T01) |
| 5.17 Authentication information | Software | Argon2id hashes (ADR-0010); NIST SP 800-63B password rules; API tokens shown once and stored hashed; two-factor secrets encrypted with a key outside the database | FR-085, FR-087, NFR-074, NFR-080 | Planned (Argon2id built, S03.2-T02) |
| 5.18 Access rights | Software | Roles, shares, and revocation take effect on the next request; the admin sees who has access (S07, S10) | FR-086, S07, S10 | Planned |
| 5.19 Information security in supplier relationships | Project | Suppliers are open-source projects; the dependency policy decides which ones are used and checks them | NFR-079, [supply chain](supply-chain.md) | Planned |
| 5.20 Addressing information security within supplier agreements | N/A | There are no supplier agreements; components are used under their open-source licenses | — | Not applicable |
| 5.21 Managing information security in the ICT supply chain | Project | Pinned dependencies and build tools, verified packages, SBOM, signed releases, provenance | NFR-078, NFR-079, FR-382, S13 | Planned (lockfiles and license checks in place; CI actions pinned by commit SHA, S03.5-T06) |
| 5.22 Monitoring, review and change management of supplier services | Project | Dependabot alerts; the dependency security review before every release (S13); GitHub settings reviewed at each stage end | NFR-079, FR-363 | Planned |
| 5.23 Information security for use of cloud services | Software, Project | The NAS needs no cloud service (I6); opt-in features that use one (cloud imports, off-site backup, ACME, push) state the trade-off. Project: GitHub (repository, CI, releases) is secured by NFR-078 | I6, NFR-071, NFR-078 | Planned (I6 holds since S01) |
| 5.24 Information security incident management planning and preparation | Project, Operator | The [incident response](incident-response.md) plan for operators and for the project | NFR-062 | Implemented (documented; not exercised) |
| 5.25 Assessment and decision on information security events | Software, Project | Security events raise alerts (FR-378); reports are triaged with CVSS | FR-378, NFR-061 | Planned |
| 5.26 Response to information security incidents | Project, Operator | Steps in the incident plan: contain, revoke, rotate, restore, notify, fix | NFR-062 | Implemented (documented; not exercised) |
| 5.27 Learning from information security incidents | Project | A post-incident review adds tests and threat-model entries | NFR-062, R15 | Implemented (documented) |
| 5.28 Collection of evidence | Software, Operator | A tamper-evident audit log (FR-379) that can be exported; operators preserve logs before changing anything | FR-090, FR-379, incident response | Planned |
| 5.29 Information security during disruption | Software | Fails closed; crash consistency (NFR-053); degraded pools keep working read-safe (S15); recovery procedures (S08.4) | NFR-053, S08, S15 | Planned |
| 5.30 ICT readiness for business continuity | Software, Operator | Database snapshots (FR-355), backups and restore tests (S08, R04); operators keep an off-site copy | FR-355, S08, R04 | Planned |
| 5.31 Legal, statutory, regulatory and contractual requirements | Project, Operator | Project: license obligations (AGPL-3.0-or-later and dependencies). Operator: data-protection law for their users; the software helps with export and deletion | `docs/dev/licensing.md`, R01, NFR-075 | Implemented (licensing); Planned (export, deletion) |
| 5.32 Intellectual property rights | Project | License policy, the register's license column, license checks in CI | `docs/dev/licensing.md`, register | Implemented |
| 5.33 Protection of records | Software, Project | The audit log is append-only and retained (S03.6) and tamper-evident (FR-379); the plan, logs, and decisions are kept in git | FR-090, FR-379 | Planned (software); Implemented (project records) |
| 5.34 Privacy and protection of PII | Software | Per-user isolation (I5); no telemetry (I6); face data rules (NFR-011); no secrets or file contents in logs (NFR-016 extension); export and deletion (R01, NFR-075) | I5, I6, NFR-011, NFR-016 | Planned (I6 and log redaction already hold) |
| 5.35 Independent review of information security | Project | External reviews so far (CR001, P008); an independent penetration test before the public release (R09); this matrix is self-assessed | R09, 11c | Planned |
| 5.36 Compliance with policies, rules and standards for information security | Project | Documentation audits (R12) check that this matrix matches reality, that each security-relevant task has its R15 record, and that no document claims certification | R12, audit checklist group M | Implemented (from plan 1.12.0) |
| 5.37 Documented operating procedures | Software | Install, admin, recovery, and security guides | S14.4, S08, `docs/guide/` | Planned (first guides exist) |

## 6. People controls (8)

| Control | Applies to | How it is handled | Where | Status |
|---|---|---|---|---|
| 6.1 Screening | N/A | No staff. Contributions from anyone are reviewed like any change (8.32), not by screening people | — | Not applicable |
| 6.2 Terms and conditions of employment | N/A | No employment | — | Not applicable |
| 6.3 Information security awareness, education and training | Project, Operator | The secure coding standard and R15 guide the developer; guides and the admin security checklist (FR-270) teach operators | R15, FR-270 | Implemented (developer); Planned (operator checklist) |
| 6.4 Disciplinary process | N/A | No staff | — | Not applicable |
| 6.5 Responsibilities after termination or change of employment | Operator | Organizational; the software lets the operator disable an account and revoke its sessions, tokens, and devices at once | S07, FR-086, FR-293 | Planned (software support) |
| 6.6 Confidentiality or non-disclosure agreements | N/A | Open-source project, no agreements; private vulnerability reports are handled under the disclosure policy | [Vulnerability management](vulnerability-management.md) | Not applicable |
| 6.7 Remote working | Operator, Software | Remote access only through private access (VPN or relay, R06), never by opening the admin to the internet | R06, I11 (proposed) | Planned |
| 6.8 Information security event reporting | Project, Software | Reporters use [`SECURITY.md`](../../SECURITY.md) and GitHub private vulnerability reporting; users see their sessions and devices and can report suspicious activity to their admin (R05 security page) | FR-377, NFR-061, FR-269 | Implemented (SECURITY.md); Planned (the GitHub setting is an owner task; the in-app parts) |

## 7. Physical controls (14)

The software cannot guard a room. These are the operator's; the [hardening guide](hardening-guide.md) section "Physical security" gives guidance.

| Control | Applies to | How it is handled | Where | Status |
|---|---|---|---|---|
| 7.1 Physical security perimeters | Operator | Keep the NAS in a place strangers cannot reach | Hardening guide | Implemented (guidance) |
| 7.2 Physical entry | Operator | Limit who can get to it | Hardening guide | Implemented (guidance) |
| 7.3 Securing offices, rooms and facilities | Operator | As above | Hardening guide | Implemented (guidance) |
| 7.4 Physical security monitoring | Operator | Optional; the software logs drive removal and power events (S15, R04 UPS) | Hardening guide, S15 | Implemented (guidance) |
| 7.5 Protecting against physical and environmental threats | Operator | Heat, water, fire; temperatures reported by the NAS (FR-360) | Hardening guide, FR-360 | Implemented (guidance) |
| 7.6 Working in secure areas | Operator | Organizational | Hardening guide | Implemented (guidance) |
| 7.7 Clear desk and clear screen | Operator, Software | Software: idle session expiry (S03.3), app lock in the phone apps (FR-284) | FR-086, FR-284 | Planned (software part) |
| 7.8 Equipment siting and protection | Operator | Ventilation, stable power | Hardening guide | Implemented (guidance) |
| 7.9 Security of assets off-premises | Operator | Disk encryption for portable drives and backups (R05, R04) | Hardening guide, FR-266 | Implemented (guidance) |
| 7.10 Storage media | Software, Operator | Drives are never mounted automatically from hot-plug (P007); qualification of new drives; encryption at rest (R05); `noexec,nosuid,nodev` mounts (NFR-073) | S15, FR-266, NFR-073 | Planned |
| 7.11 Supporting utilities | Operator | A UPS is recommended; UPS support in R04 | Hardening guide, R04 | Implemented (guidance) |
| 7.12 Cabling security | Operator | Organizational | Hardening guide | Implemented (guidance) |
| 7.13 Equipment maintenance | Operator, Software | Drive health and predictive warnings (S10.3, FR-332); integrity scans (S08.2) | FR-332, S08.2 | Planned |
| 7.14 Secure disposal or re-use of equipment | Software, Operator | Secure erase of retired drives (S15.11); with disk encryption, destroying the key makes the data unrecoverable (crypto-erase) | S15.11, FR-266, [storage security guide](storage-security-guide.md) | Planned |

## 8. Technological controls (34)

| Control | Applies to | How it is handled | Where | Status |
|---|---|---|---|---|
| 8.1 User end point devices | Operator, Software | Operators secure their computers and phones; the client apps keep tokens in the OS secure store, pin the NAS certificate, and support remote sign-out and wipe | NFR-069, FR-364, FR-293 | Planned |
| 8.2 Privileged access rights | Software | The admin role; recent re-authentication for sensitive actions; the privileged storage helper's allow-list | S03.3-T03, S03.9, NFR-037, ADR-0029 | Planned |
| 8.3 Information access restriction | Software | Central default-deny authorization (FR-089); I5 on every path; cross-user leak tests (S07.7); the authorization matrix test | FR-089, I5, NFR-081 | Planned |
| 8.4 Access to source code | Project | The owner's GitHub account with two-factor authentication; branch protection on `main` and `develop`; every change reviewed through the plan and stage process | NFR-078 | Planned (owner tasks: 2FA, branch protection) |
| 8.5 Secure authentication | Software | Argon2id, lockout and rate limits (S03.2), NIST SP 800-63B rules, TOTP (S03.7), passkeys (R05) | FR-085, FR-091, FR-263, NFR-080 | Planned (password hashing built) |
| 8.6 Capacity management | Software | A free-space reserve (S01.2); per-route body limits; listener limits (NFR-068); quotas (S10); Raspberry Pi budgets (NFR-051) | NFR-068, NFR-051, S10 | Planned (free-space reserve and body limits implemented) |
| 8.7 Protection against malware | Software | User content is never executed (I12); media tools in a sandbox (NFR-064); user content served so browsers cannot run it in the app (NFR-022, NFR-066); malware scanning of uploads (R05) | I12, NFR-063–NFR-066, FR-267 | Planned (safe serving of downloads implemented in S02.6) |
| 8.8 Management of technical vulnerabilities | Project, Software | [Vulnerability management](vulnerability-management.md); dependency, container, and code scanning; the dependency security review (S13); signed updates | NFR-061, NFR-079, FR-363, FR-381 | Planned (govulncheck, `pnpm audit`, Trivy run at stage end since S01/S02) |
| 8.9 Configuration management | Software | Secure defaults; validated configuration; startup checks (permissions, NFR-073); the admin security checklist (R05) | NFR-073, NFR-076, FR-270 | Planned (configuration validation implemented in S01) |
| 8.10 Information deletion | Software | Trash with retention, permanent deletion, account deletion, secure erase of retired drives, crypto-erase with disk encryption; documented retention periods | NFR-075, S08, S15.11 | Planned |
| 8.11 Data masking | Software | Secrets are redacted in logs; tokens are shown once; file names in logs only for admins | NFR-016, NFR-074 | Planned (request-header redaction implemented in S01.5) |
| 8.12 Data leakage prevention | Software | No secrets or personal data in logs; error answers without internals; per-user search filtering; content served from an isolated origin | NFR-016, NFR-022, NFR-066, I5 | Planned (error answers without internals implemented and tested since S01) |
| 8.13 Information backup | Software, Operator | Database snapshots (FR-355); metadata and configuration backup (S08.3); external and off-site backup (S08.6, R04) | FR-355, S08, R04 | Planned |
| 8.14 Redundancy of information processing facilities | Software, Operator | RAID 1 pools (S15); backups; the operator decides on spare hardware | S15, S08 | Planned |
| 8.15 Logging | Software | Structured application logs with rotation (NFR-016); an append-only, tamper-evident audit log (FR-090, FR-379) | NFR-016, FR-090, FR-379 | Planned (application logs implemented in S01) |
| 8.16 Monitoring activities | Software | Security events raise alerts (FR-378); health checks; the admin console overview | FR-378, S10 | Planned |
| 8.17 Clock synchronization | Software | The health check warns when the clock is not synchronized (logs, certificates, and two-factor codes depend on time) | FR-380 | Planned |
| 8.18 Use of privileged utility programs | Software | The privileged storage helper runs only allow-listed, typed operations for an authenticated caller, each audited; the admin commands of the CLI need write access to the database file, which only the service account has | NFR-037, NFR-073, ADR-0029 | Planned |
| 8.19 Installation of software on operational systems | Software, Project | Signed releases and updates (FR-381, FR-382); setup scripts verify every download (NFR-077); plugins are not loaded (I12) | FR-381, FR-382, NFR-077 | Planned |
| 8.20 Networks security | Software, Operator | Localhost only until HTTPS is set up (NFR-020); HTTPS on the LAN (NFR-067); a Host allow-list and Origin check (FR-350); firewall guidance | NFR-020, NFR-067, FR-350 | Planned (localhost-only binding verified by S01 tests; the Host allow-list and Origin check implemented, S03.5-T02) |
| 8.21 Security of network services | Software | One HTTPS port; WebDAV on the same port over HTTPS; no debug endpoints in releases; listener limits | NFR-072, NFR-076, NFR-068 | Planned |
| 8.22 Segregation of networks | Software, Operator | The admin console limited to the LAN or VPN once the NAS is exposed (R09); internal services (AI worker, storage helper) never on a published port (NFR-070); operators may put the NAS on its own network segment | NFR-070, T-56, hardening guide | Planned |
| 8.23 Web filtering | N/A | The NAS does not browse the web for its users; operators filter web access on their own network if they want | — | Not applicable |
| 8.24 Use of cryptography | Software, Project | The [crypto policy](crypto-policy.md): TLS policy, Argon2id, SHA-256, HMAC-SHA-256, AES-256-GCM or XChaCha20-Poly1305, Cosign and Minisign | NFR-080, NFR-067 | Planned (SHA-256 content hashes and Argon2id implemented) |
| 8.25 Secure development life cycle | Project | Rule R15; NIST SSDF practices; the stage security gate | R15, NFR-059 | Implemented (rule in force from plan 1.12.0; first gate at S03's end) |
| 8.26 Application security requirements | Project | [Security requirements](security-requirements.md) from OWASP ASVS 5.0 (and MASVS for phones), mapped to plan requirements and tiers | NFR-059 | Implemented (catalog written; verification per stage) |
| 8.27 Secure system architecture and engineering principles | Project | The principles of the security policy; the threat model updated by every stage that adds a surface | Policy 3, FR-084 | Implemented (threat model v1.4) |
| 8.28 Secure coding | Project | The [secure coding standard](secure-coding-standard.md), enforced by linters (gosec, forbidden-API rules, ESLint and Svelte rules, Ruff security rules) | NFR-059, NFR-023 | Planned (gosec and depguard already run) |
| 8.29 Security testing in development and acceptance | Project | The stage security gate: static analysis, secret scan, dependency and container scans, fuzzing, a dynamic scan (OWASP ZAP), the ASVS checklist; an independent penetration test before R09 | NFR-059, NFR-023, NFR-081 | Planned (traversal and error-format tests exist since S01) |
| 8.30 Outsourced development | Project | Code is written by the coding agent under the owner's supervision (approval gates, R15); no other outside developers. The agent counts as part of the supply chain: it verifies every package it adds and never weakens a check | R15, NFR-079 | Implemented |
| 8.31 Separation of development, test and production environments | Project | Tests never use real data or real disks (NFR-049); development builds report version `dev`; local test data stays in the git-ignored `dev/` folder | NFR-049, `.gitignore` | Implemented |
| 8.32 Change management | Project | Plan changes (R4), git conventions (R7), stage approval (R3), the release process (R13) | RULES R3, R4, R7, R13 | Implemented |
| 8.33 Test information | Project | Synthetic or license-clean fixtures only; never real personal photos in the repository; the owner's Google Takeout sample stays in `dev/` | 12.3 fixture sets, A28 | Implemented |
| 8.34 Protection of information systems during audit testing | Project, Operator | Security tests and scans run against test installations, never against an operator's live data; the R09 penetration test runs on a dedicated instance | NFR-059, R09 | Planned |

## Summary (2026-10-01)

Counted from the tables above (they are authoritative); 93 controls in all.

| Status | Controls |
|---|---|
| Not applicable, with the reason | 6: 5.20, 6.1, 6.2, 6.4, 6.6, 8.23 |
| Implemented (guidance) for operator controls | 13: 5.5, 5.10, 5.11, 7.1–7.6, 7.8, 7.9, 7.11, 7.12 |
| Implemented (parts still planned are named in the row) | 20: 5.1, 5.2, 5.4, 5.8, 5.9, 5.24, 5.26, 5.27, 5.31, 5.32, 5.36, 6.3, 6.8, 8.25, 8.26, 8.27, 8.30, 8.31, 8.32, 8.33 |
| Planned (parts already built are named in the row) | 54: all other controls |
| Verified | 0. The first verification happens at the S03 security gate (rule R15). |
