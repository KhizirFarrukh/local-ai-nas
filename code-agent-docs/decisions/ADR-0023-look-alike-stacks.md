# ADR-0023: Look-alike stacks and bursts

| Field | Value |
|---|---|
| Number | ADR-0023 |
| Status | Proposed |
| Date proposed | 2026-09-28 (session S007) |
| Date of last status change | 2026-09-28 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- **The user's requirements (P005):** "Group photos that are almost the same but have minor differences." "Show the user that it is a grouped photo." "The user chooses which photo in the group is displayed on the grid. By default, the first photo is displayed." "The user can choose to keep one photo and delete the rest."
- **The user in S007 (E004):** "you also need to add auto grouping of burst photos" (FR-165).
- Sidecars are the source of truth for photo metadata (I3); only the core writes them (I9). Stacks are private to their owner (FR-167, I5).
- **Burst identifiers (verified in S007, E007):** Apple maker notes `BurstUUID` ("unique ID for all images in a burst"); Google `XMP-GCamera:BurstID` and `BurstPrimary`. No verified tag for other makes.

## Options considered

### Where stack data lives

| Option | Pros | Cons |
|---|---|---|
| **Each member's sidecar** (`stack` section: stack ID, cover flag, how it was formed, user-locked decisions), cached in the database and index | Survives an index or database rebuild (I3); moves with the photo; the backup of sidecars covers it | A stack change writes several sidecars (done as one job, idempotent) |
| Database only | One write per change | Lost if internal data is lost; breaks I3 for user decisions |
| A special album | Reuses album code | Albums are shareable and many-to-many; stacks are private and one-per-photo |

### How stacks are formed (in order)
1. **Camera burst identifiers:** photos of one user with the same `BurstUUID` or `GCamera:BurstID`; the cover is the `BurstPrimary` photo when marked.
2. **Shot sequence** (no identifier): same camera (make, model, serial where present), consecutive capture times within a short gap (starting point 1 s, using sub-second times where present), and a perceptual match.
3. **Look-alikes:** same camera, within the time window (setting, starting point 2 minutes), and a perceptual distance under the look-alike threshold (ADR-0022).
4. **AI** (S17.11, optional): embedding similarity adds look-alikes not shot together.

## Decision

**Recommended:**

- **Stack data in each member's sidecar** (`stack` section, reserved in S05.1), with a database table and index fields (`stack_id`, `stack_cover`, S06.1) as rebuildable caches.
- **Formation in the order above**; automatic stacking never deletes anything.
- **Cover:** the user's choice; otherwise the camera's burst primary; otherwise the **first photo, meaning the earliest date taken, then the earliest upload** (Q42 to confirm).
- **User decisions win:** a chosen cover, a removal from a stack, an unstack, a merge, and a manual stack are marked `locked` and are never changed by a rescan or by AI (FR-166). A photo removed by the user is not re-added to that stack.
- **Visibility:** stacks exist only for the owner. People a photo is shared with see individual photos (FR-167).
- **Views:** grid, timeline, and albums show the cover with a count badge; search matches any member and shows the stack's cover with a count ("3 of 5 match").
- **"Keep one, delete the rest":** moves the others to the trash after a confirmation (I10, FR-163).

## Consequences

- **Easier:** stacks survive index rebuilds and restores; AI can add members later without a schema change.
- **Harder:** a stack change is a multi-sidecar write (a job with per-sidecar atomic writes; a partial failure is repaired by reconciliation, S05.7).
- **Required (follow-up work, constraints this imposes):** S05.1 reserves the `stack` section; S05.3 extracts burst identifiers and sub-second times; S06.1 reserves the index fields; S11.4 builds it; S17.11 adds AI grouping without overriding locked decisions.

## Approval record

_Pending: put to the user with the P005 report (S007). Q42 (cover definition) is asked at the same time._

## Links

- **Related requirements:** FR-160–FR-168, FR-165 (bursts), FR-164 (AI)
- **Related ADRs:** ADR-0022 (perceptual hash), ADR-0012 (ExifTool extraction), ADR-0014 (index)
- **Related stages:** S05.1, S05.3, S06.1, S11.4, S11.7, S17.11
- **Plan version:** 1.4.0
