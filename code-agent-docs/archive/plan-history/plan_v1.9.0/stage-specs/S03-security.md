# S03: Security

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](../PLAN_INDEX.md)). Moved unchanged from `plan.md` section 10.4 in version 1.9.0 (ADR-0044).

### 10.4 S03: Security

- **Origin:** User-defined
- **Goal:** Comprehensive security so the NAS can be safely reached from the local network. There is a single admin account at this stage; multiple users come in S07.
- **User requirements (quoted):**
  > "Stage 3 is security implementation."
- **Status:** **Approved** (2026-09-29, S007 E050; `stages/S03-security.md`); In Progress from S03.1-T01

#### S03.1: Threat model
- **Goal:** Identify what must be protected, from whom, and where, to drive the rest of S03.
- **Scope:**
  - Assets.
  - Attackers: someone on the LAN, a malicious local user, a stolen session.
  - Attack surfaces: API, GUI, uploads, previews, future network shares and AI worker.
  - Documented in `code-agent-docs/security/threat-model.md` (the folder is added to the documentation map, with approval, when created).
- **Deliverables:** threat model with numbered threats (T-01…), each mapped to a mitigation substage or marked as an accepted-risk candidate.
- **Depends on:** S02 (Done).
- **Requirements:** FR-084.
- **Acceptance criteria:**
  1. The document lists assets, attackers, surfaces, and numbered threats.
  2. Every threat maps to an S03 substage or is marked as an accepted-risk candidate.
  3. The user has reviewed it.
- **Risks/notes:** It is revisited in S07 (multi-user), S09 (shares), S14.6, and S17 (AI worker).
- **P005 change (1.4.0):** the threat model lists the future privileged storage helper (S15.2) as an attack surface: a root service on the host, reached over a Unix socket (ADR-0029, NFR-037).
- **P006 addition (1.6.0, [Planner addition], pending Q54):** Threat model: the outbound connections of alert delivery (SMTP, webhook, ntfy) (FR-221).
- **1.7.0 (P007):** the threat model lists the new storage-helper operations of S15 (device events, SMART, self-tests, partitioning and formatting, mounting, degraded RAID 1, member add, hot replace, growth, LED, secure erase) and hot-plug risks: a hostile USB drive with a crafted filesystem is never mounted automatically, only inside a flow the admin started; the admin console's routes and re-authentication (ADR-0039).
- **Status:** Not started

#### S03.2: First-run setup and authentication
- **Goal:** Only the admin can use the NAS, with strong credentials and no defaults.
- **Scope:**
  - First-run creation of the admin account (never default passwords); Argon2id password hashing.
  - Login and logout; password change; login rate limiting and lockout.
  - _Added scope:_ users and sessions tables via goose migrations on the S01 SQLite database (ADR-0007); an Argon2id parameter benchmark on reference hardware (ADR-0010); a CLI admin password reset; _(1.8.0)_ CLI admin creation (`admin create`) for machines without a local browser, such as a headless Raspberry Pi, used by the deployers (S14.2).
- **Deliverables:** user and session migrations; user store; auth endpoints; first-run flow; CLI password reset.
- **Depends on:** S03.1.
- **Requirements:** FR-064, FR-085, FR-068, NFR-010.
- **Acceptance criteria:**
  1. Until the admin exists, only the first-run endpoint is reachable, and only from localhost. On a machine without a local browser the admin is created on the machine with the command line _(1.8.0, S03 decision D-3)_.
  2. Passwords are stored as Argon2id hashes with the parameters from the ADR, and no default credentials exist anywhere.
  3. Repeated failed logins trigger rate limiting and a temporary lockout (tested).
  4. A password change invalidates the user's other sessions.
- **Risks/notes:** The S01 default namespace is bound to the admin account created here, with no file moves (ADR-0003).
- **P008 (1.9.0):** database snapshots (FR-355, task S03.2-T06).
- **Status:** Not started

#### S03.3: Sessions and tokens
- **Goal:** Sessions and tokens that are hard to steal and easy to revoke.
- **Scope:** secure session cookies (HttpOnly, Secure, SameSite) or tokens; expiry; revocation; "log out everywhere"; optional API tokens for scripts.
- **Deliverables:** session store; cookie policy; API token model (scoped, hashed, revocable).
- **Depends on:** S03.2.
- **Requirements:** FR-086, FR-087.
- **Acceptance criteria:**
  1. Session cookies are HttpOnly, Secure (under HTTPS), SameSite, with idle and absolute expiry.
  2. Revoking a session or using "log out everywhere" takes effect on the next request.
  3. API tokens can be created, scoped, listed, and revoked, and are stored only as hashes.
- **Risks/notes:** None.
- **P006 addition (1.6.0, [Planner addition], pending Q54):** Upload-only app passwords per device for the camera-upload endpoint (FR-219).
- **1.7.0:** recent re-authentication for sensitive admin actions (the admin console, S03.9); drive flows require it (P007).
- **Status:** Not started

#### S03.4: Transport security
- **Goal:** Encrypted connections on the LAN, and LAN exposure only when it is safe.
- **Scope:** HTTPS; user-provided certificates and generated self-signed certificates; guidance for certificates on a LAN; the gate that allows LAN binding (NFR-020). _(1.8.0)_ Plain HTTP stays on loopback (`server.bind`); LAN access is a separate HTTPS-only listener (`server.lan_bind`).
- **Deliverables:** TLS configuration; certificate generation command; LAN certificate guide; bind-address gate.
- **Depends on:** S03.2.
- **Requirements:** FR-088, NFR-020.
- **Acceptance criteria:**
  1. The server serves HTTPS with a user-provided or generated self-signed certificate.
  2. Binding to a non-loopback address requires HTTPS and an existing admin, and is allowed only once S03 is Done and the user configures it: the LAN listener (`server.lan_bind`) serves HTTPS only and starts only when an admin and a certificate exist, and plain HTTP stays on loopback _(made concrete in 1.8.0)_.
  3. The guide explains trusting the certificate on Windows, macOS, Linux, Android, and iOS.
- **Risks/notes:** Self-signed certificates cause browser warnings. The guide mitigates this; a local CA option can come later.
- **Status:** Not started

#### S03.5: Application hardening
- **Goal:** Close common web-application attack classes, with a default-deny authorization core that S07 extends.
- **Scope:**
  - A central authorization check on every route with default deny, built so S07 can extend it.
  - CSRF protection; CORS policy; security headers including a Content Security Policy.
  - Upload validation: size, content sniffing, nothing is ever executed.
  - Safe previews (no script execution through SVG or HTML).
  - General rate limiting; error messages that do not leak internals.
- **Deliverables:** policy module (`authorize(subject, action, resource)`); middleware; header configuration; route inventory test.
- **Depends on:** S03.2, S03.3.
- **Requirements:** FR-089, NFR-022.
- **Acceptance criteria:**
  1. A route inventory test fails if any route lacks an authorization decision.
  2. State-changing requests without a valid CSRF token are rejected.
  3. Responses carry a strict CSP and the agreed security headers (tested).
  4. Uploaded content is never executed or rendered as active content in the app's origin.
  5. Error responses never contain stack traces, filesystem paths, or internal identifiers.
- **Risks/notes:** A strict CSP can break GUI libraries. Check early in S02 choices.
- **P008 (1.9.0):** FR-350 (Host allow-list, Origin check, no CORS) is built first, before the rest of S03 (bug S03-B01); P008's per-install token is not built, because sessions and CSRF tokens arrive in this stage.
- **Status:** Not started

#### S03.6: Security logging and audit trail
- **Goal:** A trustworthy record of security-relevant events.
- **Scope:** logging of logins, failed attempts, and security-relevant changes; retention policy; an audit log structure that S07 extends with sharing events (and S04.6 with transfers).
- **Deliverables:** audit event schema; append-only audit store; retention job.
- **Depends on:** S03.2.
- **Requirements:** FR-090, NFR-016.
- **Acceptance criteria:**
  1. Logins, logouts, failures, and password and token changes produce events with time, actor, source address, and outcome.
  2. Audit events cannot be modified or deleted through the API.
  3. Retention removes events older than the configured period.
- **Risks/notes:** Retention runs on a simple scheduler until S04.3.
- **Status:** Not started

#### S03.7: Two-factor authentication (optional)
- **Goal:** Optional TOTP 2FA for accounts.
- **Scope:** TOTP-based 2FA with recovery codes. Priority "Could"; **approved**: the user answered Q33 with yes (S007 E050).
- **Deliverables:** TOTP enrolment and verification; recovery codes.
- **Depends on:** S03.2, S03.3.
- **Requirements:** FR-091.
- **Acceptance criteria (approved in 1.8.0):**
  1. Users can enrol a TOTP authenticator, after which login requires a code.
  2. Each recovery code works once.
  3. Disabling 2FA requires the password and a current code.
- **Risks/notes:** If Q33 is "no", this substage is marked Done as "not required" with a deviation note.
- **Status:** Not started

#### S03.8: Security GUI
- **Goal:** Every security feature is usable from the GUI.
- **Scope:** login page; first-run setup wizard; active sessions page; password change; 2FA setup if S03.7 is approved.
- **Deliverables:** GUI pages and flows.
- **Depends on:** S03.2–S03.7, S02.
- **Requirements:** FR-064, FR-086.
- **Acceptance criteria:**
  1. A fresh install opens the setup wizard, which creates the admin and then shows the login page.
  2. Users can view and revoke sessions and change their password in the GUI.
  3. No GUI page is reachable without login.
- **Risks/notes:** None.
- **Status:** Not started

#### S03.9: Admin console foundation (new in 1.7.0, the user's requirement)
- **Goal:** One admin console where every administrative function will live, safe from the first page.
- **User requirement (quoted, S007):** "and this is part of the admin console (gui based) app. if such stage/section does not exist (for an admin console where sysadmin can manage everything related to storage management and system settings and drives management and all the admin stuff) then add it. it is very crucial."
- **Scope:** the console shell (ADR-0039): the `/admin` area of the GUI with its own navigation for every section of the console map (6.6), shown only to admins; the admin API prefix `/api/v1/admin`; server-side role checks on every admin route (default deny); recent re-authentication for sensitive actions; audit logging of every admin action; the shared patterns every later admin page uses (lists, detail panels, wizards, progress, confirmation with a typed phrase for destructive steps, empty states); the first sections: Overview (system status), the admin's own account, System settings (bind address, HTTPS, the Raspberry Pi performance limits), and About and diagnostics; sections of later stages appear when they are built, never as dead links.
- **Deliverables:** console shell and navigation; admin API middleware; re-authentication flow; the shared admin components; the first sections; the console part of the route inventory.
- **Depends on:** S03.2, S03.3, S03.6, S03.8, S02.
- **Requirements:** FR-342, FR-344, NFR-050, NFR-051.
- **Acceptance criteria:**
  1. A non-admin never sees the console, and every admin API route refuses non-admins and anonymous callers (route inventory test).
  2. Sensitive actions ask for re-authentication, and every admin action appears in the audit log.
  3. The first sections work on a phone and a desktop, and the shared patterns are documented for later stages.
- **Risks/notes:** Built once and reused by every later admin page (principle 4). Pi budgets for the console (bundle size, memory) are part of NFR-051.
- **Status:** Not started

#### S03.10: Security testing and stage review
- **Goal:** Evidence that S03 holds, then close the stage.
- **Scope:** automated tests for auth bypass, CSRF, and traversal regressions; dependency vulnerability scanning in CI; static analysis; review of every threat model item (mitigated or documented as an accepted risk); completion record; user sign-off.
- **Deliverables:** security test suite; CI scanning jobs; threat model review record; completion record.
- **Depends on:** S03.1–S03.8.
- **Requirements:** NFR-023, NFR-010.
- **Acceptance criteria:**
  1. Tests show that no endpoint is reachable without authentication, except documented public ones (login, first run on localhost, a minimal health endpoint).
  2. CI runs dependency vulnerability scans and static security analysis, and fails on high-severity findings.
  3. Every threat model item is mitigated or recorded as an accepted risk with the user's approval.
  4. The completion record is written and the user's sign-off is recorded.
- **Risks/notes:** None.
- **1.7.0:** tests of the admin console's route inventory and re-authentication (NFR-050); an **ARM64 CI job** (GitHub `ubuntu-24.04-arm`) and a **Raspberry Pi resource profile** (Docker CPU and memory limits) join the stage-end CI from here on (NFR-051).
- **Status:** Not started

**Design notes (S03):**
- The network binding may change from localhost to LAN only after this stage is Done, and only when the user configures it.
- The policy check is `authorize(subject, action, resource)`, so S07 adds ownership and ACL rules without touching routes.
- Audit events use an extensible type field.
- The internal database introduced here is reused by S04.3 (jobs), S07 (users, shares), S08 (backups), and S10 (settings).

**Exit criteria (quoted):** "No endpoint is reachable without authentication, and every threat model item is either mitigated or documented as an accepted risk."

---
