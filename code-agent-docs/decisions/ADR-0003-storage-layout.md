# ADR-0003: Storage layout, per-user namespaces, and internal data location

| Field | Value |
|---|---|
| Number | ADR-0003 |
| Status | **Accepted** |
| Date proposed | 2026-09-24 (session S002) |
| Date of last status change | 2026-09-24 (session S005) |
| Supersedes | none |
| Superseded by | none |

## Context

- **Invariants:** **I1** requires a storage root with exactly two user-data areas, `files/` and `photos/`, that never intersect. **I2** requires internal app data (database, index, caches, thumbnails, trash, configuration) outside both areas.
- **S01.2** must create a layout that allows per-user namespaces for S07.2 without breaking changes (FR-071), and the S01 design notes ask for an owner on every item from the start.
- **Atomic finalize:** S01.4 finalizes uploads by atomic rename, and S08.1 moves items to and from trash. Both need the temp and trash locations on the same filesystem as the areas.

## Options considered

### Namespace layout

| Option | Pros | Cons |
|---|---|---|
| **A. Namespace directories from day one**: `files/<ns>/…`, `photos/<ns>/…`, one namespace until S07 | No data move in S07. The owner is derivable from the path. Every path already goes through a namespace resolver. | One extra directory level for a single user. |
| B. Flat `files/…` now, migrate into `files/<user>/…` in S07 | Simpler at first. | A data migration in S07 that breaks paths, bookmarks, and shares. It violates the forward-compatibility principle. |
| C. User at the top: `<user>/files`, `<user>/photos` | Natural per-user grouping. | Violates I1 ("two folders at the root: files and photos"). |

### Namespace directory name

| Option | Pros | Cons |
|---|---|---|
| **(i) Stable, immutable short ID** (e.g. `u0001`) | Never changes. The S01 namespace can be bound to the S03 admin without a rename. | Less readable when browsing the disk directly. |
| (ii) Username | Readable. | Renaming a user renames directories. The S01 placeholder name will not match the S03 admin's chosen name. |
| (iii) Immutable username chosen at creation | Readable and stable. | Users cannot change their login name, and the S01 placeholder problem remains. |

### Internal data location

| Option | Pros | Cons |
|---|---|---|
| **(a) `<root>/.local-ai-nas/` by default**, with an optional override for database, index, and caches | Same filesystem, so atomic rename works for uploads and trash. One volume to back up. Clearly outside both areas. | Hidden-folder visibility varies by OS. |
| (b) Fully separate path (e.g. `/var/lib/local-ai-nas`) | Can sit on an SSD. | Uploads and trash still need the root's filesystem, so it can't be used for everything. |

## Decision

**Decision (Accepted in S005; this was the recommendation the user accepted):**
- **Namespaces:** Option A with **stable short IDs** (option i). S01 creates `files/u0001/` and `photos/u0001/` for a single default owner. In S03.2 the first-run admin is **bound** to `u0001` (no file move). S07 creates `u0002`, and so on. The API presents each user's namespace as `/`.
- **Internal data:** option (a). Default `<root>/.local-ai-nas/` with `tmp/uploads/`, `trash/`, `db/`, `index/`, `thumbnails/`, `metadata/`, `ai/`, and `logs/`. Database, index, thumbnails, and logs may be relocated by configuration. `tmp/` and `trash/` always stay on the root's filesystem.
- **Configuration file:** lives **outside** the storage root. Its location comes from a CLI flag, then an environment variable, then an OS default path, because it is what tells the app where the root is.
- **Validation at startup:** areas exist and are directories; neither area contains the other or the internal data dir; temp and areas share one filesystem; the root is writable; unknown extra entries at the root are left alone and reported.

## Consequences

- **Easier:** S07.2 (no data move); ownership checks (the owner is derivable from the path); S08 trash and S01.4 uploads (atomic renames).
- **Harder:** directly browsing the disk shows IDs, not names. Mitigation: an admin-visible mapping (S07.1), and a `README.txt` in each area root may be added.
- **Required:** a single namespace resolver per area (S01.6); a health check for the same-filesystem condition (S01.2).

## Approval record

> Accepted by the user in S005 as decision D-01 of audit A001 (the storage layout recommended above: per-user namespace folders `files/u0001/`, `photos/u0001/` from S01; internal data in `<root>/.local-ai-nas/`). User's answer: "Accept all (Recommended)"
> (2026-09-24, session S005, log E007)

## Links

- **Related requirements:** FR-069, FR-070, FR-071, FR-072, FR-074, FR-008, NFR-025, NFR-026
- **Related ADRs:** ADR-0001 (Go), ADR-0002 (API), ADR-0004 (repository layout), ADR-0007 (SQLite database at `<internal>/db/`), ADR-0008 (tus temp uploads at `<internal>/tmp/uploads/`), ADR-0014 (index at `<internal>/index/`)
- **Related stages:** S01.2, S03.2, S07.2, S08.1
- **Plan version:** 0.5.0. _History: Proposed in S002; not covered by P003 (links updated in S003); Accepted in S005._
