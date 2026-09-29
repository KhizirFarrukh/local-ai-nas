# ADR-0037: Fast internal-data placement on an SSD

| Field | Value |
|---|---|
| Number | ADR-0037 |
| Status | Proposed |
| Date proposed | 2026-09-29 (session S007) |
| Date of last status change | 2026-09-29 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- The biggest speed-up for everyday use is putting the database, search index, thumbnails, and the transcode cache on an SSD (P007 tier 1, FR-338). A19 already allows a separate internal-data location.
- On a Raspberry Pi the SD card must never hold the database or other busy data (wear, S007 E046), so this placement is also a reliability measure there.

## Options considered

### Option A: Move derived data only (index, thumbnails, transcode cache)
- **Pros:** everything moved is rebuildable; an SSD failure loses nothing.
- **Cons:** the database stays on the slower drive.

### Option B: Move derived data and the database
- **Pros:** the largest speed-up.
- **Cons:** the database is not rebuildable; a single-SSD failure needs a restore from the metadata backup (S08.3).

## Decision

**Recommended: Option A by default; Option B allowed** with a recent metadata backup required and a mirrored SSD pair recommended (Q71). Moves use the migration engine's copy, verify, and switch steps (ADR-0030) in a short maintenance window. Trash, temp uploads, and replaced originals stay on the storage root (A18). If the SSD fails, derived data is rebuilt on the storage root and the database is restored from the metadata backup.

## Consequences

- **Required:** the move wizard in the admin console (S16.2, S16.7); failure tests (S16.8).

## Approval record

_Pending: put to the user with the P007 report (S007 E047). Decided when its stage is planned in detail (S16.2)._

## Links

[ADR-0030](ADR-0030-storage-migration-engine.md) · plan A18, A19, FR-338
