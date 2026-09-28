# ADR-0026: Retention and revert of originals replaced by optimization

| Field | Value |
|---|---|
| Number | ADR-0026 |
| Status | Proposed |
| Date proposed | 2026-09-28 (session S007) |
| Date of last status change | 2026-09-28 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- Storage optimization (S12) replaces originals with smaller files. Invariant **I10** requires an undo window wherever possible; FR-189 asks for a retention period (default 30 days, Q44) and revert per item or per job, with an option to delete originals at once.
- **P005** tied this to trash and versioning (S08.1, S08.5). Versioning is optional (S08.5, "Could"), so S12 must not depend on it (a fix from S007, E006).
- Internal data lives outside the areas (I2). Space is freed only when originals are removed; quotas follow (S10.2).

## Options considered

### Option A: A dedicated originals store in internal data
`originals/<user-namespace>/…` (plan 6.3), one entry per replaced file with its sidecar snapshot; an expiry per entry; revert per item or per job.
- **Pros:** independent of optional versioning; clear semantics ("kept for N days"); purge is a simple job; revert restores byte-identical originals.
- **Cons:** a second "holding area" beside the trash, with similar code.

### Option B: Put replaced originals in the trash
- **Pros:** reuses S08.1.
- **Cons:** the trash holds deleted items, while an optimized item still exists; restoring from the trash would create a second copy instead of reverting; trash emptying would silently remove the undo window.

### Option C: Use versioning (S08.5)
- **Pros:** "revert to the previous version" is the natural model.
- **Cons:** S08.5 is optional and may never be approved.

### Option D: No retention
- **Pros:** space is freed at once.
- **Cons:** violates I10 when undo is possible; one bad setting could degrade a whole library.

## Decision

**Recommended: Option A, a dedicated originals store**, sharing its restore and purge code with the trash (S08.1).

- Retention default **30 days**, configurable per installation (Q44); each entry records the job, the original's hash, size, dimensions, and the sidecar state before the change.
- **Revert** per item or per job restores the original bytes and its sidecar data (the `optimizationHistory` entry records the revert).
- **Delete at once** is an explicit option with a warning that it cannot be undone (FR-189).
- **Quotas:** originals count toward the owner's quota until they are purged, and the GUI says so (space is freed only then).
- If versioning (S08.5) is approved later, it may use the same store; S12 does not depend on it.

## Consequences

- **Easier:** safe optimization in every configuration; predictable space accounting.
- **Harder:** during the retention period an optimization uses more space, not less; the GUI must explain this.
- **Required (follow-up work, constraints this imposes):** S08.1 designs its restore and purge code for reuse; S10.2 counts the store; S12.5 builds it; S08.3 backups include the store's index.

## Approval record

_Pending: put to the user with the P005 report (S007). Q44 is asked at the same time._

## Links

- **Related requirements:** FR-189, FR-191, NFR-006, NFR-033
- **Related ADRs:** ADR-0003 (storage layout), ADR-0025 (video codec)
- **Related stages:** S08.1, S08.5, S10.2, S12.5
- **Plan version:** 1.4.0
