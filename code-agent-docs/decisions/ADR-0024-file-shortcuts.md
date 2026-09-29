# ADR-0024: File shortcuts after duplicate resolution

| Field | Value |
|---|---|
| Number | ADR-0024 |
| Status | Proposed |
| Date proposed | 2026-09-28 (session S007) |
| Date of last status change | 2026-09-28 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- **The user's requirement (P005):** "The user chooses whether to put a shortcut to the kept file in the deleted file's place." (FR-171)
- The shortcut must keep working when the target is moved or renamed, must never grant access the user does not have (I5), must not count toward quotas, and must behave clearly when the target is trashed or deleted (FR-175).
- The files area is a real folder tree, also reachable over WebDAV (in-app, ADR-0015), optionally over SMB (Samba, ADR-0019), and changed from outside (watcher, S09.4). The S01.6 safety baseline refuses symlinks inside the areas.
- The NAS runs on Linux and Windows 11 (Q1).

## Options considered

### Option A: App-level shortcut records
A database row: owner, the shortcut's path and name, and the **target's item ID** (ADR-0040, since plan 1.9.0; before: "stable file ID"). Listings show it at its path with a shortcut badge; opening or downloading serves the target.
- **Pros:** follows moves and renames (by ID); the same on every platform; access checked on the target every time; zero quota; easy to find all shortcuts to a target.
- **Cons:** nothing exists on disk, so tools that read the disk directly (SMB, a backup of `files/`) do not see it; the WebDAV layer must present it.
- **License / cost:** none.

### Option B: Symbolic links
- **Pros:** visible to every program on the host.
- **Cons:** break when the target moves; creating them on Windows needs Developer Mode or an administrator right; the S01.6 baseline refuses symlinks because they can point outside the area (path traversal); a link over SMB can expose paths.
- **License / cost:** none.

### Option C: Hard links
- **Pros:** transparent to every program.
- **Cons:** free no space (the point of removing duplicates); editing one copy edits "both"; same filesystem only; the file is not really deleted until every link is.
- **License / cost:** none.

### Option D: Shortcut files (`.url`, `.lnk`, `.desktop`)
- **Pros:** visible on disk.
- **Cons:** each format is platform-specific; they hold a path, so they break when the target moves; they could be edited to point anywhere.
- **License / cost:** none.

## Decision

**Recommended: Option A, app-level shortcut records.**

- A shortcut refers to its target by item ID (ADR-0040) and is re-resolved on every access, with the full authorization check on the target (I5). A shortcut can never be shared.
- **Quota:** zero (FR-175, S10.2).
- **Target moved or renamed:** the shortcut follows it and shows the target's current path.
- **Target trashed:** the shortcut shows "target missing" with a restore option; restoring the target brings the shortcut back. **Target permanently deleted:** the user is warned first how many shortcuts point to it; afterwards they show "target deleted" and can be removed.
- **Deleting a shortcut** never touches the target.
- **Over WebDAV** (Q46, proposed): a shortcut appears as a **read-only file** with the target's content; writing to it is refused with an explanation. Over SMB (not in-app) shortcuts do not appear, which is documented.
- **A real file created at a shortcut's path** from outside the app (found by the watcher) wins: the shortcut is removed and the event is logged.

## Consequences

- **Easier:** correct on every platform; no traversal risk; no quota tricks.
- **Harder:** shortcuts are invisible to programs that read the disk directly; the S14 backup of internal data must include the shortcut table (it does: it is in the database).
- **Required (follow-up work, constraints this imposes):** S09.2's WebDAV FileSystem leaves room for virtual entries; S10.2 counts shortcuts as zero; S11.6 builds it; search (S06) indexes shortcuts with the shortcut flag.

## Approval record

_Pending: put to the user with the P005 report (S007). Q46 is asked at the same time._

## Links

- **Related requirements:** FR-171, FR-173, FR-175, NFR-033
- **Related ADRs:** ADR-0003 (storage layout), ADR-0015 (WebDAV), ADR-0019 (SMB)
- **Related stages:** S09.2, S10.2, S11.6
- **Plan version:** 1.4.0
