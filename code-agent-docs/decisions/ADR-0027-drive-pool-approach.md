# ADR-0027: Drive pool approach (RAID 0 and RAID 1)

| Field | Value |
|---|---|
| Number | ADR-0027 |
| Status | Proposed |
| Date proposed | 2026-09-28 (session S007) |
| Date of last status change | 2026-09-28 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- **The user's requirements (P005):** "Handle multiple drives, RAID style." Custom layouts, RAID 0, RAID 1, parity, and combining smaller drives into a larger virtual drive.
- **The user's decision (S007, E008):** "if working with raids is a complex problem, just do raid 0 and 1 implementation and that too in the end, and leave complex raid for later as planned non implemented work". So S15 builds **RAID 0 and RAID 1 only**; parity, virtual drives, and nesting are deferred (plan 11a).
- The NAS must never implement striping or mirroring itself (NG8, revised): it orchestrates mature tools. Pools are Linux-only (A22, Q47); the core never runs as root (ADR-0029).
- **Capacity rule (the user's):** every member contributes the size of the smallest member. mdadm RAID 0 would use unequal members fully (verified in S007, E006/E007), so the NAS sizes members itself to keep the user's rule.
- The design must leave room for the deferred parity layouts without migrating existing pools.

## Options considered

### Option A: Linux md RAID with mdadm
- **Pros:** the standard Linux software RAID; levels 0, 1, 4, 5, 6, 10, and linear, so the deferred layouts need no new tool (the level list is from the mdadm man page, verified in S007); array metadata lives on the drives, so arrays are found after a reinstall (`--assemble --scan`); monitoring (`--monitor`), member replacement, and a write-intent bitmap for faster RAID 1 resync; independent of the filesystem.
- **Cons:** the RAID 5/6 write hole needs a journal or the partial parity log (plan 8.25, relevant only for the deferred layouts); nesting md arrays has boot-assembly pitfalls (only partly verified in S007, deferred anyway).
- **License / cost:** GPL-2.0 (mdadm), a separate program run by the storage helper.

### Option B: Btrfs RAID profiles (raid0, raid1)
- **Pros:** checksums with self-healing on raid1; mixed drive sizes.
- **Cons:** the filesystem and the RAID are one choice (ADR-0028 would be forced); Btrfs RAID5/6 is documented upstream as not recommended for production (to re-verify in S15.1), which blocks the deferred parity path; the capacity rules differ from the user's.
- **License / cost:** GPL-2.0 (kernel, btrfs-progs).

### Option C: ZFS (mirror, raidz)
- **Pros:** strong integrity; snapshots.
- **Cons:** out-of-tree kernel module on Debian and Raspberry Pi OS (DKMS builds); high memory use; CDDL licence incompatible with the kernel's GPL (a packaging concern); snapshots are out of scope (NG8).
- **License / cost:** CDDL-1.0.

### Option D: LVM RAID
- **Pros:** flexible volumes.
- **Cons:** uses md underneath with an extra layer; more to explain and repair.
- **License / cost:** GPL-2.0.

### Option E: SnapRAID with mergerfs
- **Pros:** made for mixed-size media drives; parity is computed on a schedule.
- **Cons:** not real-time protection; no striping; fits the deferred parity layouts, not RAID 0/1.
- **License / cost:** SnapRAID GPL-3.0; mergerfs ISC (verified in S007). **Deferred** with the complex RAID (11a).

## Decision

**Recommended: Option A, mdadm**, for RAID 0 and RAID 1, driven only through the storage helper (ADR-0029).

- **RAID 0:** capacity = members × the smallest member; no redundancy, stated plainly in the GUI.
- **RAID 1:** capacity = the smallest member; survives all but one member failing; write-intent bitmap on.
- **Member sizing:** the helper creates one partition per drive sized to the smallest member (rounded down to a MiB boundary), so the extra space on larger drives stays unused and visible (FR-196), matching the user's rule and leaving room for FR-201 later.
- **Import:** pools are identified by their md UUID; the configuration is on the drives.
- **Deferred:** RAID 4/5/6, linear (virtual drives), nesting, SnapRAID with mergerfs (11a).

## Consequences

- **Easier:** mature, documented tooling; the deferred layouts can later use the same tool.
- **Harder:** Linux-only; loop-device tests need root on the test machine (S15.13).
- **Required (follow-up work, constraints this imposes):** S15.1 re-verifies every mdadm capability used against the current man page and records it here; `dependencies.md` rows for mdadm, util-linux, smartmontools, e2fsprogs/xfsprogs (Linux section 12, optional component).

## Approval record

_Pending: put to the user with the P005 report (S007). Q47 is asked at the same time._

## Links

- **Related requirements:** FR-195–FR-210, NFR-036–NFR-038; deferred FR-197, FR-198, FR-199
- **Related ADRs:** ADR-0028 (pool filesystem), ADR-0029 (storage helper), ADR-0003 (storage layout)
- **Related stages:** S10.3, S15
- **Plan version:** 1.4.0 (concerns 8.24, 8.25; 11a)
