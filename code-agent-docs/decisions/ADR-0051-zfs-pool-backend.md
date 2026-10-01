# ADR-0051: ZFS as an optional pool backend

| Field | Value |
|---|---|
| Number | ADR-0051 |
| Status | Proposed |
| Date proposed | 2026-10-01 (session S007, E076) |
| Date of last status change | 2026-10-01 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

The user asked (S007 E076): "add support for: zfs, NAS OS, and everything related to this and hosting a NAS", and delegated open questions: "if you have questions, do what you deem best for that". ADR-0027 (Proposed) recommended mdadm for RAID 0 and RAID 1 and listed ZFS as option C. The project is Raspberry Pi first (NFR-051), its pools are Linux-only (A22, Q47), and the core never runs as root (ADR-0029).

**Verified on 2026-10-01** (research R004): OpenZFS 2.3 added RAIDZ expansion; 2.4 supports Linux up to 6.18; since 2.3 the default ARC limit on Linux is max(RAM − 1 GB, 5/8 of RAM); Debian and Raspberry Pi OS ship ZFS as `zfs-dkms` and `zfsutils-linux` built with DKMS, and Raspberry Pi forum threads report ZFS failing after kernel updates until the module is rebuilt; Arch Linux has no official package (the independent ArchZFS repository or the AUR); the Windows port is in release candidates.

## Options considered

### Option A: mdadm only (ADR-0027 as proposed)
- **Cons:** does not meet the user's request.

### Option B: ZFS only
- **Pros:** one backend; checksums, self-healing mirrors, snapshots, replication, compression, encryption.
- **Cons:** on a Raspberry Pi: memory, and DKMS rebuilds after every kernel update; no official Arch package; makes every installation depend on an out-of-tree module.

### Option C: ZFS as an optional second backend, mdadm the default (recommended)
- **Pros:** meets the request; users with suitable hardware get ZFS's integrity features; the default stays simple and in-kernel.
- **Cons:** two backends to build and test; the GUI must explain the choice.

## Decision

**Option C (recommendation while Proposed; the choice of backend default was decided by the agent under the user's delegation, E076).**

- **Where:** Linux with the storage helper, ZFS installed by the operator's choice from the distribution (FR-384).
- **Layouts:** stripe (RAID 0) and mirror (RAID 1) under the user's capacity rule (each member partitioned to the smallest); RAIDZ waits with the other parity layouts (11a), now with a growth path (RAIDZ expansion).
- **Datasets:** the storage root as one dataset (`compression=lz4`, `atime=off`, `snapdir=hidden`, `xattr=sa`, `acltype=posixacl` where needed); a separate dataset for rebuildable cache data (no snapshots); internal data in the same pool. Encryption at rest (ZFS native encryption) is an option when R05 builds encryption (FR-266).
- **Memory:** the NAS sets `zfs_arc_max` per hardware profile (for the Raspberry Pi profile a fraction of RAM measured in S15.13; never the 2.3 default); dedup stays off.
- **Helper operations** (typed, allow-listed, audited): create a stripe or mirror, attach, detach, replace, online and offline, scrub, import, export, set the allowed properties, list status, and destroy (destructive, with the typed confirmation of FR-203).
- **Health:** after boot the health page checks that the ZFS module loads for the running kernel and that the pools are imported and healthy; `zpool status` errors raise alerts (FR-206).
- **Features used later:** snapshots with holds (R04, FR-256), `send` and `receive` replication (FR-257), scrub results in integrity reports (S08.2), L2ARC and special vdevs as cache options (S16.1).
- **Tests:** file-backed vdevs on a runner where the module loads (no real disks, NFR-049).

## Consequences

- **Easier:** strong integrity and snapshots for users who choose ZFS.
- **Harder:** a second backend in S15; ZFS module availability on each platform.
- **Required:** S15.1 verifies every ZFS command and property used against the OpenZFS documentation for the packaged version; register entries (OpenZFS packages, CDDL-1.0, installed from the distribution) when adopted.

## Approval record

_Not yet approved. The default (mdadm) and ZFS as optional were decided by the agent under the user's delegation (E076); the ADR itself needs the user's approval (R5)._

## Links

- **Related requirements:** FR-383, FR-384, FR-195, FR-196, FR-203–FR-209, FR-256, FR-257, FR-266, NFR-049, NFR-051
- **Related ADRs:** ADR-0027 (pool approach, revision 1), ADR-0028 (pool filesystem), ADR-0029 (storage helper), ADR-0038 (block-level SSD cache)
- **Related stages:** S15, S16, S08, R04, R05
- **Plan version:** 1.13.0
