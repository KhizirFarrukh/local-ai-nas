# ADR-0030: Storage migration engine

| Field | Value |
|---|---|
| Number | ADR-0030 |
| Status | Proposed |
| Date proposed | 2026-09-29 (session S007) |
| Date of last status change | 2026-09-29 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- The storage root must move to another drive in several flows: a bigger drive (capacity upgrade), a replacement for a failing drive, a new RAID 1 mirror (mirror conversion), and a pool (plan 10.15; P007 `drive_lifecycle.migration_engine`; FR-329, FR-331, FR-333).
- The user runs the NAS on a **Raspberry Pi** (S007 E046): drives on USB 3 or PCIe NVMe, four cores, limited memory. A migration of a large library takes hours, so it must run while the NAS stays usable, and survive power cuts.
- Every file already has a stored SHA-256 content hash (FR-211, ADR-0021), so verification needs no second hashing of the source.
- The whole root moves together: files, photos, trash, temp uploads (A18), and internal data (unless A19 placed it elsewhere). The source is only ever read until the user retires it (NFR-044, I10).

## Options considered

### Option A: A built-in copier in Go
- **Pros:** one implementation on Linux, Windows, and macOS; walks the areas through the storage resolver and `os.Root` (no path escapes, links never followed, as S01); verifies each copied file against its stored content hash while copying; throttling, pause, resume, and progress are ours; the journal is part of the program.
- **Cons:** we write and test the copy, catch-up, and metadata preservation ourselves.
- **License / cost:** AGPL-3.0-or-later (project code).

### Option B: rsync as a separate program
- **Pros:** mature, fast delta transfer, preserves metadata.
- **Cons:** not on Windows without extra software; its verification is by its own checksums, not our stored hashes; progress and pause are harder to integrate; a new runtime dependency on every Linux install.
- **License / cost:** GPL-3.0-or-later (separate program; its `COPYING` is GPL v3, recorded as a candidate).

### Option C: The OS's copy tools (cp, robocopy)
- **Pros:** nothing to write.
- **Cons:** different per platform; no hash verification; no journal.

## Decision

**Recommended: Option A, a built-in copier.**

- **Phases** (as P007): plan (sizes, counts, estimate, confirmation) → bulk copy online (throttled; at low I/O and CPU priority on the Pi) → catch-up passes over what changed (by modification time and size, then the change journal of later stages) → final sync in maintenance mode (writes stopped, database checkpointed) → verify (every file against its content hash; counts, sizes, sidecars, database integrity) → switch (ADR-0031) → post-switch check → rollback window.
- **Journal:** an append-only file of phase records kept on the **target** (so it moves with the data) and mirrored in the source's internal data; a restart resumes from the last completed step.
- **Throttling:** a bandwidth and concurrency limit (default sized for a Raspberry Pi with USB 3 drives: one copy stream, low priority), adjustable in the admin console.
- **Files without a stored hash** (added outside the app before hashing) are hashed on the source once during the copy.

## Consequences

- **Easier:** one engine for the command line in the MVP (S08.4) and the console wizards later (S14.6, S14.10); verification is exact.
- **Harder:** metadata preservation (times, and on Linux ownership and modes the app relies on) must be tested per platform.
- **Required:** crash-injection tests at every phase (S08.7, S14.13); the maintenance-mode page (S08.4); the downtime target in the stage document (NFR-045).

## Approval record

_Pending: put to the user with the P007 report (S007 E047). Decided when its stage is planned in detail (S08.4 (command line, MVP-B) and S14.6 (GUI and online))._

## Links

[P007](../prompts/P007-drive-lifecycle-and-ssd-cache.json) · [ADR-0021](ADR-0021-content-hash-algorithm.md) · [ADR-0031](ADR-0031-switching-the-storage-root.md) · plan FR-329, FR-331, FR-333, NFR-044, NFR-045
