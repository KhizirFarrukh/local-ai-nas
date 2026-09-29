# ADR-0028: Filesystem on a drive pool

| Field | Value |
|---|---|
| Number | ADR-0028 |
| Status | Proposed |
| Date proposed | 2026-09-28 (session S007) |
| Date of last status change | 2026-09-28 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- A pool made with mdadm (ADR-0027) is a block device; it needs a filesystem before it can hold the storage root (FR-204).
- The storage root, including upload temp and trash, must be **one filesystem** so atomic renames work (A18, NFR-036).
- Libraries reach 100,000 photos plus 100,000 files (Q18), with files of many gigabytes (S01.4).
- The tools run on Debian/Ubuntu and Raspberry Pi OS (Q1). Snapshots are out of scope (NG8).

## Options considered

### Option A: ext4
- **Pros:** the default Linux filesystem; mature repair tools (`e2fsck`); can be grown online and shrunk offline; well understood by users and recovery tools.
- **Cons:** no data checksums; fixed inode count (set at creation; sized generously for many small sidecars).
- **License / cost:** GPL-2.0 (kernel; e2fsprogs).

### Option B: XFS
- **Pros:** strong with large files and parallel I/O; grows online.
- **Cons:** cannot be shrunk; fewer home users know its repair tools.
- **License / cost:** GPL-2.0 (kernel; xfsprogs).

### Option C: Btrfs on top of md
- **Pros:** data checksums detect corruption.
- **Cons:** on md RAID 1, Btrfs sees one device, so it detects corruption but cannot repair it from the mirror; its main advantages (own RAID, snapshots) are unused or out of scope.
- **License / cost:** GPL-2.0.

## Decision

**Recommended: ext4 by default**, created with a label, a generous inode ratio for sidecars, and reserved blocks reduced for a data volume (e.g. 1%). **XFS offered as an advanced choice** at pool creation. Integrity checks stay at the application level (S08.2), which works on any filesystem.

## Consequences

- **Easier:** familiar, repairable filesystem; nothing new for the setup scripts beyond e2fsprogs (usually present).
- **Harder:** no filesystem checksums; corruption is detected by S08.2 integrity checks and md consistency checks (FR-207).
- **Required (follow-up work, constraints this imposes):** S15.5 formats and mounts through the helper; `dependencies.md` rows for e2fsprogs and xfsprogs (Linux, optional pool component).

## Approval record

_Pending: put to the user with the P005 report (S007)._

## Links

- **Related requirements:** FR-204, NFR-036
- **Related ADRs:** ADR-0027 (pool approach), ADR-0029 (storage helper)
- **Related stages:** S15.1, S15.5
- **Plan version:** 1.4.0
