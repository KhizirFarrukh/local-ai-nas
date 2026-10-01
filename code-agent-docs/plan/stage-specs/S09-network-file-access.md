# S09: Network file access and external change sync

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](../PLAN_INDEX.md)). Moved unchanged from `plan.md` section 10.10 in version 1.9.0 (ADR-0044).

### 10.10 S09: Network file access and external change sync

- **Origin:** Planner-proposed
- **Reason added (quoted from P002):** "Most people expect a NAS to show up as a network drive on their computers. This depends on S03 (security) and S07 (per-user permissions). Changes made through network shares bypass the app, so live sync of sidecars and the index is needed."
- **Goal:** The NAS works as a network drive with per-user permissions, and external changes stay in sync.
- **User requirements:** none (planner-proposed).
- **Status:** Not started
- **Security (1.12.0, P010; rule R15):** WebDAV hardening (NFR-072) is **Tier 1**: HTTPS only, NAS accounts or app passwords, limited `PROPFIND` depth, XML without external entities, `Destination` through the resolver and authorization.

#### S09.1: Protocol selection
- **Goal:** Choose the network-share protocol(s).
- **Scope:** WebDAV decided in 0.3.0 (ADR-0015). S09.1 designs its authentication, locks, and mounts, and the SMB-via-Samba design for ADR-0019 (Proposed), for the user to adopt or keep deferred; platform constraints (Q32).
- **Deliverables:** ADR-0015 details confirmed; ADR-0019 decision or deferral; a prototype connection from each client OS.
- **Depends on:** S07 (Done).
- **Requirements:** FR-009.
- **Acceptance criteria:**
  1. The WebDAV design (ADR-0015 details) and the ADR-0019 SMB decision cover Windows, macOS, and Linux clients, host constraints, authentication integration, and how changes reach the index.
  2. A prototype connection works from each target client OS.
- **Risks/notes:** Samba is not available on Windows hosts (RK-22).
- **Status:** Not started

#### S09.2: Network shares with permissions
- **Goal:** Network access obeys the same permissions as the app.
- **Scope:** shares mapped to users and permissions: each user sees only their own items plus items shared with them.
- **Deliverables:** share gateway integrated with auth and `authorize()`.
- **Depends on:** S09.1, S07.4.
- **Requirements:** FR-124, NFR-024.
- **Acceptance criteria:**
  1. A user connected over the network sees only their own items and items shared with them.
  2. Read-only shares cannot be written over the network.
  3. Network logins use NAS accounts and are audit-logged.
- **Risks/notes:** None.
- **P005 change (1.4.0):** the WebDAV FileSystem design leaves room for representing app-level shortcuts (S11.6, ADR-0024, Q46).
- **P006 addition (1.6.0, [Planner addition], pending Q54):** The camera-upload endpoint, through the same policy and services as the shares (FR-219).
- **Status:** Not started

#### S09.3: Photos area exposure policy
- **Goal:** Network shares cannot break the area rules.
- **Scope:** whether `photos/` is exposed over network shares, and if so, whether read-only (Q32). Invariant I1 must hold.
- **Deliverables:** enforced exposure policy; tests.
- **Depends on:** S09.2.
- **Requirements:** FR-125, NFR-026.
- **Acceptance criteria:**
  1. The Q32 policy is enforced.
  2. No network operation can place an item in the other area.
  3. If `photos/` is writable over shares, media added this way gets sidecars through S09.4, and non-media is reported, not ingested.
- **Risks/notes:** None.
- **P006 addition (1.6.0, [Planner addition], pending Q54):** The Q32 policy plus this ingest-only exception: a client sees only its own uploads (FR-219).
- **Status:** Not started

#### S09.4: Real-time watcher and reconciliation
- **Goal:** Changes made outside the app appear quickly and correctly.
- **Scope:** detect external creates, updates, renames, and deletes; update sidecars and the index; debouncing; rename detection.
- **Deliverables:** watcher service; debounce queue; rename detector; integration with the S05.7 reconciler.
- **Depends on:** S09.2, S05.7, S06.2.
- **Requirements:** FR-027, FR-028.
- **Acceptance criteria:**
  1. Changes made over the network or directly on disk appear in the app and search within the target delay.
  2. A rename is handled as a rename (the sidecar follows), not as a delete plus create.
  3. Bursts of changes are debounced into bounded work.
  4. Periodic reconciliation catches anything the watcher missed.
- **Risks/notes:** Sync conflicts between app and share edits (RK-18).
- **P008 (1.9.0):** moves made outside the app keep their item IDs (FR-347).
- **Status:** Not started

#### S09.5: Share settings GUI
- **Goal:** Shares are easy to enable and connect to.
- **Scope:** enabling and configuring shares, and connection instructions for Windows, macOS, and Linux.
- **Deliverables:** admin share settings page; per-user connection help.
- **Depends on:** S09.2.
- **Requirements:** FR-126.
- **Acceptance criteria:**
  1. The admin enables and configures shares from the GUI.
  2. Each user sees connection instructions for their OS.
- **Risks/notes:** None.
- **P006 addition (1.6.0, [Planner addition], pending Q54):** Per-device setup with a QR code for the server address (FR-220).
- **1.7.0:** the share settings are the console's Network shares section (6.6).
- **Status:** Not started

#### S09.6: Testing and stage review
- **Goal:** Prove network access and permissions from real clients, then close the stage.
- **Scope:** tests from Windows, macOS, and Linux clients; permission tests over the network; completion record; user sign-off.
- **Deliverables:** client test matrix and results; updated threat model; documentation; completion record.
- **Depends on:** S09.1–S09.5.
- **Requirements:** FR-124, NFR-024.
- **Acceptance criteria:**
  1. Connect, read, write, and permission tests pass from each client OS.
  2. Cross-user tests over the network show no leaks.
  3. The completion record is written and the user's sign-off is recorded.
- **Risks/notes:** macOS client testing needs a Mac.
- **P006 addition (1.6.0, [Planner addition], pending Q54):** Tests with at least one Android and one iOS auto-upload app (FR-220).
- **Status:** Not started

**Design notes (S09):** With in-app WebDAV, every change goes through the service layer (policy, sidecars, index, audit). The watcher is essential only for Samba and for direct disk edits. The watcher reuses the reconciler and the job system.

**Exit criteria (planner-proposed):** Users can mount the NAS from Windows, macOS, and Linux, see exactly their own and shared items, and external changes appear in the app and search with sidecars intact.

---
