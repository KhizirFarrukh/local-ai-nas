# S10: Admin console: monitoring, quotas, and system settings

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](../PLAN_INDEX.md)). Moved unchanged from `plan.md` section 10.11 in version 1.9.0 (ADR-0044).

### 10.11 S10: Admin console: monitoring, quotas, and system settings

- **Origin:** Planner-proposed
- **Reason added (quoted from P002):** "A multi-user system needs admin visibility and control over storage, health, and background work."
- **Goal:** Give the admin visibility and control over storage, health, and background work.
- **User requirements:** none (planner-proposed).
- **1.7.0 (the user's requirement):** the stage completes the **admin console** (6.6, ADR-0039): its pages are console sections, and S10.6 checks that every admin function is there. The user: "and this is part of the admin console (gui based) app. if such stage/section does not exist (for an admin console where sysadmin can manage everything related to storage management and system settings and drives management and all the admin stuff) then add it. it is very crucial."
- **Status:** Not started
- **Security (1.12.0, P010; rule R15):** security event alerts through the alert channels (FR-378); outbound safety for alert delivery (NFR-071, **Tier 1**); the clock synchronization check (FR-380); purely cosmetic console pages are Tier 2.

#### S10.1: Admin dashboard
- **Goal:** One place to see how the NAS is doing.
- **Scope:** storage usage overall and per user; system status.
- **Deliverables:** usage accounting; dashboard page.
- **Depends on:** S07 (Done).
- **Requirements:** FR-127.
- **Acceptance criteria:**
  1. The dashboard shows total and per-user usage, free space, and system status.
  2. Usage figures match the filesystem within a documented tolerance.
- **Risks/notes:** None.
- **P005 change (1.4.0):** the dashboard reserves a panel for reclaimable duplicate space and optimization savings (FR-194).
- **1.7.0:** the dashboard is the console's **Overview** (FR-345): each problem links to where it is fixed.
- **P008 (1.9.0):** system resource metrics with a small history, and optionally a LAN-only Prometheus endpoint (FR-360).
- **Status:** Not started

#### S10.2: Storage quotas
- **Goal:** No single user can fill the disk.
- **Scope:** per-user quotas enforced on upload, copy, and copying shared items.
- **Deliverables:** quota model; enforcement hook in the S01.3 service interface; admin quota settings.
- **Depends on:** S10.1.
- **Requirements:** FR-128.
- **Acceptance criteria:**
  1. Writes that would exceed a quota (upload, copy, copying shared items, transfers) are refused with a clear message.
  2. Usage updates immediately after each operation.
  3. The admin can set, change, and remove quotas.
- **Risks/notes:** Trash and versions count toward quotas (documented).
- **P005 change (1.4.0):** shortcuts count as zero toward quotas, and quotas update when optimization originals are removed (S11.6, S12.5).
- **Status:** Not started

#### S10.3: Disk health and alerts
- **Goal:** Problems are visible before data is lost.
- **Scope:** disk health (SMART where available); low-space and failure warnings.
- **Deliverables:** health collector; alert rules; GUI alerts.
- **Depends on:** S10.1.
- **Requirements:** FR-129.
- **Acceptance criteria:**
  1. SMART data is shown where available, and "not available" otherwise.
  2. Low-space and disk-failure warnings appear in the GUI and the logs.
- **Risks/notes:** SMART access needs privileges and differs by OS and container (RK-24).
- **P005 change (1.4.0):** disk health monitoring is designed to be reused by the drive pools (S15.3, S15.7).
- **P006 addition (1.6.0, [Planner addition], pending Q54):** Alert rules and delivery channels: email, webhook, ntfy (FR-221).
- **P007 addition (1.7.0, [Planner addition] MVP-A, pending Q67):** **predictive drive health** (FR-332, ADR-0032): SMART readings and error counts turned into Healthy, Watch, Replace soon, and Replace now, with reasons and raw values; scheduled self-tests (short weekly, extended monthly, configurable); alerts on status changes (FR-221); "Health data not available" when SMART cannot be read (USB bridges on a Raspberry Pi, containers, virtual disks). Shown in the console's Storage and drives section.
- **Status:** Not started

#### S10.4: Background jobs monitor
- **Goal:** The admin can see and manage background work.
- **Scope:** queue view, progress, failures, and retries for the S04.3 job system.
- **Deliverables:** jobs page; retry and cancel actions.
- **Depends on:** S04.3.
- **Requirements:** FR-067.
- **Acceptance criteria:**
  1. The admin sees queued, running, and failed jobs with progress.
  2. Failed jobs can be retried or cancelled.
- **Risks/notes:** None.
- **Status:** Not started

#### S10.5: Settings and log viewer
- **Goal:** System settings and logs are manageable from the GUI.
- **Scope:** system settings; a viewer for application logs and the audit log.
- **Deliverables:** settings page (replacing the S02.2 placeholder); log viewer.
- **Depends on:** S03.6.
- **Requirements:** FR-066, FR-130, NFR-016.
- **Acceptance criteria:**
  1. Settings can be edited with validation, and invalid values are refused.
  2. Application and audit logs are viewable and filterable by the admin only.
- **Risks/notes:** None.
- **P006 addition (1.6.0, [Planner addition], pending Q54):** Channel settings and the "send test alert" button (FR-221).
- **1.7.0:** settings and logs are the console's System settings and Logs and alerts sections, including the Raspberry Pi performance limits (NFR-051).
- **Status:** Not started

#### S10.6: Admin console completeness and admin guide (new in 1.7.0, the user's requirement)
- **Goal:** Every administrative function is in the console, and the admin can find it.
- **Scope:** a review of every admin function built so far (S03–S10) against the console map (6.6); the missing ones added as console pages; search across console settings; the About and diagnostics section (health check, a support bundle without personal data); the admin guide organized by console section.
- **Deliverables:** the completed console; the completeness checklist; the admin guide.
- **Depends on:** S03.9, S10.1–S10.5.
- **Requirements:** FR-342, FR-343, FR-345.
- **Acceptance criteria:**
  1. Every admin function of S03–S10 is reachable in the console; the command line is needed only for recovery, setup, and scripting.
  2. The admin guide covers every console section.
- **Risks/notes:** Later stages (S11–S17, releases) add their own console pages; each stage's audit checks it (audit checklist).
- **Status:** Not started

#### S10.7: Testing and stage review
- **Goal:** Prove admin features and quota enforcement, then close the stage.
- **Scope:** quota enforcement tests; documentation; completion record; user sign-off.
- **Deliverables:** quota test suite; admin guide; completion record.
- **Depends on:** S10.1–S10.5.
- **Requirements:** FR-128.
- **Acceptance criteria:**
  1. Quota tests cover every write path, including network shares.
  2. The completion record is written and the user's sign-off is recorded.
- **Risks/notes:** None.
- **P006 addition (1.6.0, [Planner addition], pending Q54):** Tests with a local SMTP test server and a webhook receiver (FR-221).
- **P007 addition (1.7.0):** health status rules tested against recorded SMART fixtures of healthy, degrading, and failed drives (NFR-049); the console route inventory (NFR-050).
- **Status:** Not started

**Design notes (S10):** Quotas hook into the same service interface as trash and sharing. The dashboard reads accounting maintained by operation hooks, not directory scans.

**Exit criteria (planner-proposed):** The admin can see usage and health, manage jobs, settings, and logs, and quotas are enforced on every write path.

---
