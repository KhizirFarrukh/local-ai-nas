# S15: Drives, pools, and drive lifecycle (RAID 0 and RAID 1)

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](../PLAN_INDEX.md)). Moved unchanged from `plan.md` section 10.15 in version 1.9.0 (ADR-0044).

### 10.15 S15: Drives, pools, and drive lifecycle (RAID 0 and RAID 1)

- **Origin:** User-defined feature (P005), placed by the planner; **scope and position set by the user in S007** (E008): "just do raid 0 and 1 implementation and that too in the end, and leave complex raid for later as planned non implemented work".
- **Goal:** Let users combine several drives into RAID 0 (striped) or RAID 1 (mirrored) pools from the GUI, see exact capacity and fault tolerance first, and handle failures and rebuilds safely.
- **User requirements (quoted from P005):**
  > "Handle multiple drives, RAID style." "The user can configure custom layouts: RAID 0, RAID 1, or parity style …" "Smaller drives can be combined into one larger virtual drive." "When data is split across drives with a parity drive, the smallest member's size becomes the size used on every member."
  > _Parity style and combined drives are deferred to 11a by the user's decision in S007._
- **Placement:** the last stage before AI (I8), after the pre-AI release (S14), as the user asked ("in the end"). It uses the security model (S03), data protection (S08), and disk health (S10.3), and extends the S14 setup scripts with the optional helper.
- **Extended in 1.7.0 (P007, the user's requirements):** the drive lifecycle: new-drive detection and a wizard, drive qualification, single-drive capacity upgrade, mirror conversion, proactive replacement, pool growth, and drive retirement, all from the admin console's Storage and drives section. The user (P007): "System: 1. Ability to install new hard drive into system and let the software adapt it or do the data migration to it for cases of upgrading storage capacity, introducing a raid 1 drive (a backup drive) or a replacement drive for if the old drive is failing or seeming to fail."
- **Platform matrix (1.7.0):** Linux with the helper: everything (kernel-event detection, preparation, migration, mirror conversion, growth, hot replacement, secure erase, LED location). Linux without the helper, Windows, macOS: predictive health where SMART is readable, detection by re-scan, migration to a drive the admin prepared, guides for preparing a drive, backup-drive setup; no pools, mirror conversion, or secure erase from the app. Docker: the helper runs on the host (ADR-0029), and the stable mount point keeps bind mounts valid (ADR-0031). Raspberry Pi: the reference machine (A24); USB bridge warnings (8.33).
- **Status:** Not started
- **ZFS (1.13.0, the user's requirement, S007 E076):** ZFS is an optional second pool backend next to mdadm, which stays the default (FR-383, FR-384; ADR-0027 revision 1, ADR-0051; decided by the agent under the user's delegation (S007 E076)): S15.1 decides both ADRs; S15.2 adds typed ZFS operations to the helper and an optional setup-script component that installs ZFS from the distribution with consent; S15.5 creates stripe or mirror pools with the storage-root and cache datasets; S15.7 shows `zpool status` and runs scrubs; S15.8 imports ZFS pools; S15.10 uses `zpool attach` and `replace`; S15.11 retires members; S15.12 shows the backend in the GUI; S15.13 tests ZFS on file-backed vdevs and checks the ARC limit in the Pi profile. RAIDZ waits with the parity layouts (11a).
- **Security (1.12.0, P010; rule R15):** S15.2: every new storage-helper operation needs a threat-model entry and abuse tests; the socket checks peer credentials and a token (NFR-037 note, NFR-070); S15.5: `noexec,nosuid,nodev` mounts (NFR-073); S15.11: secure erase (ISO 7.14). The helper and every privileged operation are **Tier 1**.

#### S15.1: Approach, platform scope, and filesystem
- **Goal:** Decide how pools are built, where, and on which filesystem.
- **Scope:** the pool approach ADR (ADR-0027: mdadm for RAID 0 and 1, recommended; ZFS or Btrfs mirrors considered; SnapRAID with mergerfs deferred with parity), its verified capabilities; the Linux-only scope (A22, Q47); the filesystem ADR (ADR-0028: ext4 or XFS).
- **Deliverables:** ADR-0027 and ADR-0028 decided; capability notes from the official documentation.
- **Depends on:** S14, S10.3.
- **Requirements:** FR-195, FR-204.
- **Acceptance criteria:**
  1. Both ADRs are Accepted by the user, with every capability they rely on verified against official documentation.
  2. The design leaves room for the deferred parity layouts (11a) without a data migration.
- **Risks/notes:** RK-34.
- **Status:** Not started

#### S15.2: Privileged storage helper
- **Goal:** Disk operations run with the least privilege possible.
- **Scope:** design (ADR-0029); allow-listed operations (discover, create RAID 0/1, format, mount, set up monitoring, check, replace a member, stop); Unix-socket authentication; audit logging; systemd installation, added to the Linux setup scripts of S14.2 as an optional component; Docker deployments run the helper on the host; the threat model update (S03.1).
- **Deliverables:** helper service; allow list; setup-script component; threat model entries.
- **Depends on:** S15.1, S03.1, S03.6, S14.2.
- **Requirements:** FR-210, NFR-037.
- **Acceptance criteria:**
  1. The core runs without root; every disk operation goes through the helper and is audit-logged.
  2. Requests outside the allow list, or from an unauthenticated client, are refused (tested).
  3. The helper refuses the OS drive and drives holding NAS data outside the migration flow.
- **Risks/notes:** RK-35.
- **P007 addition (1.7.0):** new allow-listed operations: subscribe to device events, read drive details and SMART, run self-tests, partition and format a qualified blank drive, mount and unmount at NAS-owned paths, create a degraded RAID 1, add a member, hot-replace a member, grow an array and a filesystem, blink an LED, securely erase a retired drive. Each validates the drive against the inventory and is audit-logged (NFR-037).
- **Status:** Not started

#### S15.3: Drive discovery and health
- **Goal:** The user sees every drive and its state before designing a pool.
- **Scope:** drive listing with model, serial, size, type, SMART health, partitions, mount status, and whether it holds the OS or NAS data; integrated with the S10.3 health collector.
- **Deliverables:** discovery through the helper (lsblk, smartctl); drive model in the API.
- **Depends on:** S15.2, S10.3.
- **Requirements:** FR-202.
- **Acceptance criteria:**
  1. Loop-device and virtual-disk fixtures are listed with every field, and the OS and NAS-data drives are flagged.
  2. SMART data appears where available and "not available" otherwise.
- **Risks/notes:** SMART in containers (RK-24).
- **P007 addition (1.7.0):** new-drive detection by kernel events through the helper, with re-scan as the fallback (ADR-0034); the known-drive registry by serial and WWN; the S10.3 predictive health reused.
- **Status:** Not started

#### S15.4: Layout and capacity engine
- **Goal:** Correct capacity and fault-tolerance figures before anything is erased.
- **Scope:** the layout model (physical drives, members, roles, layout), built so the deferred virtual drives and parity layouts fit later; capacity rules (RAID 0: members × the smallest, RAID 1: the smallest; the rest shown as unused); fault tolerance; validation; suggestions; the live calculator's data. Pure logic, heavily unit-tested.
- **Deliverables:** capacity engine with tests.
- **Depends on:** S15.1.
- **Requirements:** FR-195, FR-196, FR-200, FR-201 (design only).
- **Acceptance criteria:**
  1. For RAID 0 and RAID 1 with equal and unequal drives, capacity, unused space, and fault tolerance match hand-worked examples.
  2. The model can express the deferred layouts; the user's 2+1+1+2+2+2 TB example is kept as a pending test for the 11a candidate.
- **Risks/notes:** mdadm RAID 0 would use unequal members fully; the helper sizes members to the smallest so the user's rule holds.
- **Status:** Not started

#### S15.5: Pool creation and mounting
- **Goal:** A pool is created only when the user clearly means it.
- **Scope:** creation safety (FR-203, I10: model and serial list, typed phrase, OS and NAS-data drives refused, SMART check first); creation through the helper; filesystem creation; mounting; registering the pool as the storage root.
- **Deliverables:** creation wizard API; mount and registration.
- **Depends on:** S15.2, S15.3, S15.4.
- **Requirements:** FR-203, FR-204, NFR-033.
- **Acceptance criteria:**
  1. On loop devices, RAID 0 and RAID 1 pools are created, formatted, mounted, and used as the storage root.
  2. Without the typed phrase, or with a refused drive, nothing is erased.
- **Risks/notes:** RK-34.
- **Status:** Not started

#### S15.6: Storage migration engine (GUI and online)
- **Goal:** An existing library moves onto a new pool without loss.
- **Scope:** maintenance mode; copying with checksum verification; switching paths only after the verification succeeds; rollback if it fails.
- **Deliverables:** migration job and wizard API.
- **Depends on:** S15.5, S08.2.
- **Requirements:** FR-205, NFR-036.
- **Acceptance criteria:**
  1. A fixture library migrates with every file, sidecar, and internal data intact (checksums).
  2. An injected failure during migration leaves the old root in use and unchanged.
- **Risks/notes:** None beyond RK-34.
- **P007 change (1.7.0):** extends the MVP engine (S08.4, ADR-0030) with online bulk copy and catch-up, drive preparation through the helper, the stable-mount-point switch (ADR-0031), and the console wizard. Requirements also FR-329, NFR-044, NFR-045.
- **Status:** Not started

#### S15.7: Monitoring, checks, failures, and rebuilds
- **Goal:** Problems are seen early, and a mirror survives a failed drive.
- **Scope:** pool states (healthy, degraded, rebuilding, failed) and per-drive SMART with alerts; scheduled consistency checks (scrubs) with mismatch reports; a degraded RAID 1 keeps serving data; the replace-drive wizard; rebuild with progress; clear guidance when a RAID 0 member or both mirrors fail; "RAID is not a backup" (S08.6).
- **Deliverables:** monitor; check schedule; replace and rebuild flow.
- **Depends on:** S15.5, S10.3.
- **Requirements:** FR-206, FR-207.
- **Acceptance criteria:**
  1. On loop devices, failing one RAID 1 member keeps the data readable; replacing it rebuilds with progress and ends healthy.
  2. A RAID 0 member failure is reported plainly as data loss, with backup guidance.
- **Risks/notes:** RK-34.
- **P007 addition (1.7.0):** proactive hot replacement of a failing RAID 1 member with mdadm replace mode, keeping redundancy (ADR-0033); the RAID 0 failing-member flow (migrate the whole pool while it still reads; otherwise disaster recovery). Requirement FR-331.
- **Status:** Not started

#### S15.8: Expansion and import
- **Goal:** Pools can grow and survive a reinstall.
- **Scope:** adding drives or replacing them with larger ones and growing, where mdadm supports it for RAID 0 and 1; detecting existing pools after a reinstall or on a new machine.
- **Deliverables:** grow and import flows.
- **Depends on:** S15.7.
- **Requirements:** FR-208, FR-209.
- **Acceptance criteria:**
  1. A pool created on loop devices is found and imported after the helper and core are reinstalled.
  2. Supported growth operations work on loop devices; unsupported ones are explained.
- **Risks/notes:** None.
- **P007 addition (1.7.0):** pool growth (FR-335; FR-208 promoted to Should): RAID 1 members replaced one at a time, then the array and the filesystem grown online; RAID 0 grown by migrating to a new, larger pool; a recent backup checked first.
- **Status:** Not started

#### S15.9: New-drive wizard and drive qualification (new in 1.7.0, P007)
- **Goal:** A newly installed drive is noticed and put to the right use safely.
- **Scope:** new-drive notifications in the console; the wizard with its recommended choice (upgrade, mirror, replace, grow, backup drive, SSD cache once S16 exists, ignore) and a plain summary before anything happens; drives holding unknown data shown read-only first; drive qualification (SMART, short self-test, optional burn-in per Q69, capacity and compatibility warnings); pause and resume with a journal.
- **Deliverables:** wizard API and console pages; qualification jobs and results stored with the drive's identity.
- **Depends on:** S15.3, S15.6.
- **Requirements:** FR-328, FR-334.
- **Acceptance criteria:**
  1. A loop device attached during a test run is detected, identified by its serial, and offered in the wizard; the OS drive and NAS-data drives are never offered for erasing.
  2. A fixture drive with growing reallocated sectors gets a clear warning and a recommendation to return it.
- **Risks/notes:** SMART behind USB bridges may be missing (8.33).
- **Status:** Not started

#### S15.10: Single-drive upgrade, replacement, and mirror conversion (new in 1.7.0, P007)
- **Goal:** The user's three cases work: a bigger drive, a second drive as a mirror, and a failing drive replaced.
- **Scope:** capacity upgrade (qualify, prepare, migrate, retire); the single-drive replacement cases (still readable: prioritized copy with unreadable files listed by name and restore offered; already dead: disaster recovery from backup); mirror conversion (degraded RAID 1 on the new drive, migration, then the old drive added after a separate typed confirmation, "Protected" only after the resync; refused before erasing if the old drive is too small), with "use as a backup drive" offered as the alternative (FR-206).
- **Deliverables:** the three flows in the wizard; tests on loop devices.
- **Depends on:** S15.6, S15.9.
- **Requirements:** FR-329, FR-330, FR-331.
- **Acceptance criteria:**
  1. On loop devices, a capacity upgrade, a mirror conversion (also with an interrupted resync), and a replacement end with all data verified by content hash.
  2. At every moment of a mirror conversion there are two complete copies until the user confirms erasing the old drive, and the GUI says when only one up-to-date copy exists.
- **Risks/notes:** RK-42, RK-43.
- **Status:** Not started

#### S15.11: Drive retirement (new in 1.7.0, P007)
- **Goal:** The old drive ends up where the user wants, safely.
- **Scope:** keep as a rollback copy (default, labelled with the date), reuse as a backup drive, add as a mirror member, securely erase (ADR-0035) with an estimate, a typed confirmation, and a completion record, or forget (with a warning that it still holds readable data); a drive in "Replace now" is not offered as a backup drive or mirror member; LED location (Could).
- **Deliverables:** retirement flows; secure-erase operations in the helper.
- **Depends on:** S15.10.
- **Requirements:** FR-336.
- **Acceptance criteria:**
  1. Each choice works on a virtual disk; secure erase refuses the OS drive and any drive with NAS data.
  2. The completion record names the drive's serial and the method.
- **Risks/notes:** RK-43.
- **Status:** Not started

#### S15.12: Storage GUI
- **Goal:** Everything above from the GUI.
- **Scope:** a storage page; the pool designer with the live calculator; pool status; the creation, migration, and replacement wizards.
- **Deliverables:** storage pages and wizards.
- **Depends on:** S15.4–S15.8.
- **Requirements:** FR-200, FR-203, FR-206.
- **Acceptance criteria:**
  1. A user builds, migrates to, and repairs a RAID 1 pool from the GUI on a test machine.
  2. Every erasing step shows the drives and asks for the typed phrase.
- **Risks/notes:** The page is hidden where the helper is not available (A22).
- **1.7.0:** the storage GUI is the console's **Storage and drives** section: the drive list with health statuses, the new-drive wizard, migration progress, rollback, and retirement (P007).
- **Status:** Not started

#### S15.13: Testing, audit, and stage review
- **Goal:** Prove safe pool handling without real disks in CI, then close the stage.
- **Scope:** simulated-disk tests (loop devices or VM disks): creation, failure, degraded mode, rebuild, expansion, import, and re-assembly after a reboot; the manual real-hardware test plan; a documented recovery drill; "RAID is not a backup" documentation; the stage's unit, integration, and system tests; the R12 documentation audit; completion record; user sign-off.
- **Deliverables:** test suites; manual test plan and recovery drill; documentation; audit; completion record.
- **Depends on:** S15.1–S15.12.
- **Requirements:** NFR-038; verification of every S15 requirement.
- **Acceptance criteria:**
  1. The simulated-disk suite passes; no test touches a real disk.
  2. The manual real-hardware test and the recovery drill are documented and run once before release.
  3. Coverage and the audit meet the stage-end rules; the completion record is written and the user's sign-off is recorded.
- **Risks/notes:** Loop devices need root on the test machine (CI runs at stage completion only, per the user's CI preference).
- **P007 addition (1.7.0):** simulated-disk tests (NFR-049) for hot-plug detection, capacity upgrade, mirror conversion including an interrupted resync, hot replacement, RAID 1 growth by replacing both members, the RAID 0 failing-member migration, rollback, retirement and secure erase on a virtual disk, and crash injection at every migration phase.
- **Status:** Not started

**Design notes (S15):** The app orchestrates mdadm through the storage helper and never implements striping or mirroring itself. The layout model already knows members and roles, so the deferred parity layouts and virtual drives (11a) can be added without migrating existing pools.

**Exit criteria (1.7.0, P007):** a newly installed drive is detected and, through the console wizard, used to upgrade capacity, become a RAID 1 mirror, or replace a failing drive, with every file verified and the old drive kept for rollback until retired. **Exit criteria (adapted from P005 to the user's S007 scope):** A user can build a RAID 0 or RAID 1 pool from their own drives through the GUI, see exact capacity and fault tolerance first, survive a simulated drive failure in a mirror, and rebuild without data loss.

---
