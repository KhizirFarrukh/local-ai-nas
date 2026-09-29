# S08: Data protection and recovery

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](../PLAN_INDEX.md)). Moved unchanged from `plan.md` section 10.9 in version 1.9.0 (ADR-0044).

### 10.9 S08: Data protection and recovery

- **Origin:** Planner-proposed
- **Reason added (quoted from P002):** "A NAS often holds people's only copy of their files. Accidental deletion and corrupted metadata need recovery paths. It is placed after S07 so that trash and restore are per-user from the start."
- **Goal:** Recovery paths for accidental deletion, corruption, and disaster, per user.
- **User requirements:** none. This is a planner-proposed stage that the user may remove or reorder.
- **Status:** Not started

#### S08.1: Trash
- **Goal:** Deleted items can be recovered.
- **Scope:** a per-user trash for both areas; retention period of 30 days (FR-008, the user's requirement in S007); restore to the original location with sidecar and metadata intact.
- **Deliverables:** trash store in internal data (`.local-ai-nas/trash/<ns>/`); delete-to-trash hook; restore; purge job.
- **Depends on:** S07 (Done), S05.6, S04.3.
- **Requirements:** FR-008, FR-026.
- **Acceptance criteria:**
  1. Deleting in either area moves the item (and sidecar) to its owner's trash in internal data (I2).
  2. Restore brings back the item, sidecar, album memberships, and shares, or explains what could not be restored.
  3. Items older than 30 days in the trash are deleted automatically by a job (FR-008).
  4. Trash contents are visible only to their owner.
- **Risks/notes:** Trash must be on the storage root's filesystem so delete and restore are atomic renames (A18).
- **P005 change (1.4.0):** the trash restores items removed through duplicate resolution (S11.3, S11.6) with their sidecars and memberships, and can hold originals replaced by optimization (S12.5) so they can be reverted within the retention period (ADR-0026).
- **P006 addition (1.6.0, [Planner addition], pending Q54):** A pair goes to the trash and comes back as a unit (FR-218).
- **P008 (1.9.0):** extends the minimal trash built in S03 (FR-354) instead of creating it: photos with sidecars, album and share restore, per-user views, removed duplicates, replaced originals.
- **Status:** Not started

#### S08.2: Integrity verification
- **Goal:** Silent corruption is detected.
- **Scope:** checksums; detection of corrupted files and sidecars; scheduled integrity scans.
- **Deliverables:** checksum store; scan job; report.
- **Depends on:** S04.3, S05.7.
- **Requirements:** FR-119.
- **Acceptance criteria:**
  1. A deliberately corrupted file and sidecar are detected in a test.
  2. Scheduled scans run as low-priority jobs and report results.
  3. Corruption is reported to the owner and admin, and nothing is deleted automatically.
- **Risks/notes:** Full scans are I/O heavy, so they are throttled.
- **Status:** Not started

#### S08.3: Metadata and configuration backup
- **Goal:** App state that is not on disk as files can be restored.
- **Scope:** backup and restore of the database, configuration, users, and sharing data (and albums and face registries per Q13).
- **Deliverables:** backup format; backup and restore commands; scheduled backups.
- **Depends on:** S07 (Done).
- **Requirements:** FR-120.
- **Acceptance criteria:**
  1. Backups can be created on demand and on a schedule.
  2. Restoring onto a fresh install restores users, settings, and shares.
  3. Backups contain no plaintext secrets, and the format is documented.
- **Risks/notes:** None.
- **Status:** Not started

#### S08.4: Disaster recovery
- **Goal:** The system can be rebuilt from what is on disk.
- **Scope:** a documented and tested procedure to rebuild the index and internal data from disk and sidecars.
- **Deliverables:** disaster-recovery runbook; automated disaster-recovery test.
- **Depends on:** S08.3, S06.2, S05.7.
- **Requirements:** FR-121, FR-025, FR-068.
- **Acceptance criteria:**
  1. From the storage root plus a metadata backup, the documented procedure restores a working system.
  2. The procedure is tested automatically in CI on a sample library.
  3. What cannot be recovered without a backup is documented.
- **Risks/notes:** None.
- **P007 addition (1.7.0, [Planner addition] MVP-B, pending Q67):** the **storage migration engine** (ADR-0030) with a command line (e.g. `local-ai-nas storage migrate --to <path>`, exact name in the stage document) **and a page in the admin console** (Storage and drives, the user's requirement) on every platform, onto a drive the admin prepared: plan, copy, catch-up, maintenance-mode final sync, content-hash verification, switch by configuration change, rollback window (Q68), journal. The drive-pools stage adds drive preparation, online migration, and the stable-mount-point switch on top of the same engine. Also the recovery path when a single drive has died (restore from backup).
- **Status:** Not started

#### S08.5: File versioning (optional)
- **Goal:** Overwrites can be undone.
- **Scope:** keeping previous versions on overwrite. Priority "Could", pending the user's decision (Q34).
- **Deliverables (if approved):** version store; list and restore API and GUI.
- **Depends on:** S08.1.
- **Requirements:** FR-122.
- **Acceptance criteria (if approved):**
  1. Overwriting keeps the previous version, up to a configured count or age.
  2. Users can list and restore versions.
  3. Versions count toward quotas (S10.2).
- **Risks/notes:** If Q34 is "no", this is marked Done as "not required".
- **P005 change (1.4.0):** if approved, versioning can also hold originals replaced in S12 and revert them. S12 does not depend on it: without versioning, the trash or the originals store of ADR-0026 is used.
- **Status:** Not started

#### S08.6: External backup
- **Goal:** An off-disk copy of user data.
- **Scope:** a backup job to an external drive or another local location.
- **Deliverables:** backup target configuration; incremental backup job; restore procedure.
- **Depends on:** S04.3, S08.3.
- **Requirements:** FR-123.
- **Acceptance criteria:**
  1. A configured job copies the selected areas and users incrementally to the target.
  2. A restore from the target is tested.
  3. Progress and failures are visible.
- **Risks/notes:** Local targets only (I6). Remote targets would need explicit opt-in.
- **Status:** Not started

#### S08.7: GUI, testing, and stage review
- **Goal:** Protection features are usable and proven, then the stage closes.
- **Scope:** trash and backup UI; recovery tests; documentation; completion record; user sign-off.
- **Deliverables:** trash UI (per user); backup UI (admin); recovery test suite; documentation; completion record.
- **Depends on:** S08.1–S08.6.
- **Requirements:** FR-008, FR-120, FR-123.
- **Acceptance criteria:**
  1. Users manage their trash in the GUI, and the admin manages backups.
  2. The recovery tests pass.
  3. The completion record is written and the user's sign-off is recorded.
- **Risks/notes:** None.
- **P007 addition (1.7.0):** migration tests with crash injection at every phase, a verification failure that stops the switch, and rollback. **1.7.0:** the backup UI is the console's Backups and recovery section.
- **Status:** Not started

**Design notes (S08):**
- The trash hooks into the S01.3 service interface ("before delete").
- The integrity scan reuses the S05.7 reconciler.
- Backups use the job system.

**Exit criteria (planner-proposed):** Deleted items can be restored per user, corruption is detected and reported, and the system can be rebuilt from disk plus a metadata backup, all proven by automated tests.

---
