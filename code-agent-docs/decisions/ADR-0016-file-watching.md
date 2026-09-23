# ADR-0016: File watching: fsnotify plus a periodic reconciliation scan

| Field | Value |
|---|---|
| Number | ADR-0016 |
| Status | **Accepted** |
| Date proposed | 2026-09-24 (session S003) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none |
| Superseded by | none |

## Context

Changes can happen outside the app (FR-027, FR-028):
- Direct edits on the host (any stage).
- Network shares that bypass the core, such as Samba (ADR-0019).

S05.7 builds the reconciliation scan, and S09.4 adds real-time watching. WebDAV writes go through the core, so they do not need the watcher (ADR-0015).

## Options considered

| Option | Pros | Cons |
|---|---|---|
| **fsnotify for real-time events + periodic reconciliation** (chosen) | Fast reaction plus guaranteed correctness. fsnotify is cross-platform (inotify, ReadDirectoryChangesW, kqueue/FSEvents) | Linux inotify is not recursive and has watch limits. Events are lost on some filesystems and while the app is down |
| Reconciliation only | Simple, always correct eventually | Slow to notice changes |
| OS-specific watchers (e.g. fanotify) | Recursive, efficient | Needs privileges; platform-specific |

## Decision

- **fsnotify v1.10.1** (`github.com/fsnotify/fsnotify`, BSD-3-Clause, verified) for real-time events.
- **Periodic reconciliation scan** (S05.7) as the safety net: on startup, on a schedule, and on demand. It catches anything the watcher misses.
- **Linux note:** inotify is **not recursive** and limits the number of watched folders. Large libraries need a higher `fs.inotify.max_user_watches`, which the **install guide must document** (S11.4), together with how to check the current value. If watch registration fails, the watcher degrades gracefully to reconciliation-only for the affected subtree and reports it (health and admin UI).

### Implementation details chosen by agent
| Detail | Choice | Reason |
|---|---|---|
| Recursion | Walk the areas and add a watch per directory. Add watches for new directories on create events | fsnotify does not provide portable recursive watches |
| Scope | Watch `files/` and `photos/` only, never internal app data. Ignore the app's own temp and sidecar temp files. Suppress self-generated events using a recent-writes registry | Avoids feedback loops (plan 8.5) |
| Debounce | Coalesce events per path over a short window (e.g. 1–2 s) and enqueue reconcile jobs (ADR-0011) | Bounded work under bursts |
| Rename detection | Pair remove and create events by file identity (inode or file ID where available), then by content hash, so the sidecar follows | FR-028 |

## Consequences

- **Easier:** near-real-time updates with a correctness backstop.
- **Harder:** watch-limit tuning on Linux; platform-specific event quirks (handled by the reconciler).
- **Required:** install guide section (S11.4); health reporting when watching is degraded; tests for rename detection and burst debounce (S09.4).

## Approval record

> Decision made by the user in plan change request #3 (`code-agent-docs/prompts/P003-technology-stack.json`, `technology_stack.file_watching`)
> (2026-09-24, session S003). Version and license verified in S003 log E005.

## Links

- **Related requirements:** FR-027, FR-028, FR-102
- **Related ADRs:** ADR-0011, ADR-0015, ADR-0019
- **Related stages:** S05.7, S09.4, S11.4
- **Plan version:** 0.3.0
