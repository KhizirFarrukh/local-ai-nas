# ADR-0040: Item identity: UUIDv7 IDs in an items table

| Field | Value |
|---|---|
| Number | ADR-0040 |
| Status | **Accepted** (2026-09-30, session S007) |
| Date proposed | 2026-09-30 (session S007) |
| Date of last status change | 2026-09-30 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- Items are identified by their path today (S01): the files API, content hashes by path. Albums, shares, jobs, trash, faces, shortcuts, the search index, and audit events need references that survive renames, moves, external changes, and moves to another drive (plan 8.35, RK-50).
- Filesystem file IDs (inodes, Windows file IDs) do not survive a copy, a restore from backup, or the storage migration engine (ADR-0030).
- External review #1 (CR001) and plan change request #8 (P008 F1) ask for an immutable ID for every stored object (FR-346–FR-349).

## Options considered

| Option | Pros | Cons |
|---|---|---|
| **(a) UUIDv7 (RFC 9562) in an items table** | A standard; time-ordered, so database indexes stay compact; generated without coordination | 36 characters as text; reveals the creation time (so never used as a secret) |
| (b) ULID | Time-ordered, shorter text | Not an IETF standard; another library |
| (c) Filesystem file ID | Free | Changes on copy, restore, and drive migration; differs per OS |
| (d) Content hash | Detects identical content | Changes on every edit; identical copies would share an ID |

## Decision

**(a)**, with **`github.com/google/uuid` v1.6.0** (`NewV7`, BSD-3-Clause; verified 2026-09-30 in the module cache and on proxy.golang.org; already an indirect dependency, becomes direct).

- **The items table:** item ID, namespace, area (`files` or `photos`), relative path (NFC), kind, size, modification time, file identity (inode or Windows file ID, for re-association only), content hash (FR-211), status (`present`, `missing`, `trashed`, `retired`), timestamps; unique (namespace, area, path) for present items.
- **Life cycle:** create (upload, folder, copy) → a new ID; rename and move → the same ID (a folder's descendants move in the same transaction); copy → new IDs; cross-area move (S04.6) → the same ID; cross-area copy → a new ID; delete → `trashed` (FR-354), restore → the same ID, purge → `retired`.
- **Items created outside the app:** an ID when first seen: on demand in a listing (an idempotent insert guarded by the unique constraint) and in bulk by the reconciliation scan (S05.7).
- **Moves made outside the app:** the reconciler (S05.7, S09.4) matches the new path to the missing old one by file identity with size and time, or by content hash; unmatched items stay `missing` for a grace period, then `retired`; their references are reported, never dropped silently (FR-347).
- **Photos:** the ID is also the sidecar's `mediaId` (FR-348), so a rebuild restores photo IDs. **Files area:** IDs live only in the database, protected by snapshots (FR-355) and backups (S08.3). This trade-off is accepted: a files-area sidecar per file would clutter the user's folders (plan 8.8).
- **References:** search documents, jobs (with the expected content hash as a safety check), trash records, audit events (ID plus the path at the time), shortcuts (ADR-0024), albums and faces, shares.
- **API:** `id` in every listing and details answer; a lookup by ID; path endpoints stay (FR-349).
- **Security:** an ID is an identifier, not a permission (I5); public links use separate random tokens (R09).
- **Existing data:** a one-time, resumable backfill job (ADR-0011 amendment).

### Implementation notes (S01.3-T11, S007 E067)

- **Overwrite:** writing new content to an existing path (an upload or copy with `on_conflict=overwrite`) keeps that item's ID: it is a new version of the same file, so a share or an album entry survives an update. A new path (including an automatic rename) gets a new ID; a move that replaces an item retires the replaced item's ID.
- **Retired rows stay** with their last path, so references and the audit trail can still name the item.
- **Case-insensitive disks:** the registry keys items by the path as the request spelled it; a case variant of an existing name registers as another path. The listing-based assignment (S01.3-T12) and the reconciler (S05.7) retire such rows, as they must for content hashes.
- **Best effort,** as the content hashes: a failed registry write never undoes the operation; the operation journal (S01.4-T09) closes the gap after crashes.

## Consequences

- **Easier:** every later feature refers to items safely; renames and moves stop breaking references.
- **Harder:** every write path updates the items table (one row per changed item, in the same transaction as the operation journal, ADR-0041); listings need IDs (assigned on demand).
- **Required:** S01.3-T11, S01.3-T12 (P008 follow-ups, built in S03); S05.1 reserves `mediaId`; S05.7 and S09.4 re-associate; ADR-0024 uses item IDs.

## Approval record

> "Accept all three (Recommended)"
> (2026-09-30, session S007; the user's answer to: "Accept the three ADRs the approved follow-ups are built on?", naming ADR-0040, ADR-0041, and ADR-0042 with a one-line summary of each)

Proposed with plan 1.9.0 (P008), whose follow-up tasks the user had approved ("Approve all, F2 first (Recommended)", S007 E059).

## Links

plan FR-346–FR-349, 8.35, RK-50, A26 · [ADR-0041](ADR-0041-operation-journal-and-crash-consistency.md) · [ADR-0024](ADR-0024-file-shortcuts.md) · [ADR-0030](ADR-0030-storage-migration-engine.md) · P008 F1
