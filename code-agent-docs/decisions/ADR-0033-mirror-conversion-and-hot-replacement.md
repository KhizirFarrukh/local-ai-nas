# ADR-0033: Mirror conversion, hot replacement, and pool growth with mdadm

| Field | Value |
|---|---|
| Number | ADR-0033 |
| Status | Proposed |
| Date proposed | 2026-09-29 (session S007) |
| Date of last status change | 2026-09-29 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- The user asked for adding a second drive as a RAID 1 mirror of the current one, replacing a failing drive, and growing capacity (P007; FR-330, FR-331, FR-335). ADR-0027 (Proposed) chose mdadm for RAID 0 and RAID 1.

**Verified 2026-09-29 (mdadm(8), man7.org):**
- "For a RAID1 array, only one real device needs to be given. All of the others can be 'missing'." (degraded creation)
- A member added later is recovered (resynced); with a **write-intent bitmap**, a re-added device "avoids a full reconstruction but instead just updates the blocks that have changed".
- `--replace`: the marked device "remains in service during the recovery process to increase resilience against multiple failures" (hot replacement keeps redundancy).
- `--grow --size=max`: "the largest size that fits on all current drives".
- RAID 0: "RAID 0 array size cannot be changed" with `--size`; adding devices to RAID 0 is a reshape. **Unverified** how reliable a RAID 0 reshape is; the default path does not use it.

## Options considered

### Option A: mdadm procedures as verified above
- **Mirror conversion:** create RAID 1 on the new drive with the second member `missing`, internal bitmap, format (ADR-0028), migrate (ADR-0030), switch (ADR-0031), then, after a separate typed confirmation, add the old drive; the pool shows "Protected" only when the resync finishes.
- **Hot replacement of a RAID 1 member:** add the new drive as a spare, `--replace <failing> --with <new>`; redundancy is kept throughout.
- **Growth:** replace RAID 1 members one at a time (hot replacement), then `--grow --size=max`, then grow the filesystem online (ext4 `resize2fs` on a mounted filesystem; XFS `xfs_growfs`, which needs it mounted). RAID 0 grows by migrating to a new, larger pool.
- **Pros:** uses the tools' own tested paths; redundancy is never reduced in hot replacement.
- **Cons:** long resyncs on large drives (hours; slower over USB on a Pi).

### Option B: A fresh pool from two new drives, then migration
- **Pros:** the old drive is never erased.
- **Cons:** needs two new drives. Offered as the alternative in the wizard.

## Decision

**Recommended: Option A, with Option B offered.** Shrinking is never offered (verified: ext4 cannot shrink while mounted; XFS shrinking is only partly implemented).

## Consequences

- **Required:** simulated-disk tests (loop devices) for degraded creation, add and resync (also interrupted), replace mode, and growth (S14.13, NFR-038); the RAID 0 reshape stays unused unless a later check shows it is reliable.

## Approval record

_Pending: put to the user with the P007 report (S007 E047). Decided when its stage is planned in detail (S14.7, S14.8, S14.10 (extends ADR-0027))._

## Links

[ADR-0027](ADR-0027-drive-pool-approach.md) · [ADR-0028](ADR-0028-pool-filesystem.md) · [ADR-0030](ADR-0030-storage-migration-engine.md) · plan FR-195, FR-206, FR-208, FR-330, FR-331, FR-335
