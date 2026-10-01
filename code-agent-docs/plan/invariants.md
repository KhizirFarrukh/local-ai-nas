# Project invariants

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](PLAN_INDEX.md)). Moved unchanged from `plan.md` section 2a in version 1.9.0 (ADR-0044).

## 2a. Project invariants

These invariants also live in `RULES.md` ("Project invariants"). **No code or plan change may violate them without explicit user approval.**

- **I1:** The storage root contains exactly two user-data areas: `files/` and `photos/`. They never intersect. An item moves between them only when the user explicitly copies or moves it.
- **I2:** Internal application data (database, search index, caches, thumbnails, trash, configuration) lives outside `files/` and `photos/`.
- **I3:** Photo sidecar JSON files are the source of truth for descriptive photo metadata. Ownership, access, and sharing are authoritative in the database; sidecars keep a read-only mirror of them. The search index is a cache that can always be rebuilt from disk.
- **I4:** Search reads only the index. No model reads user content in a request path. The one exception: when the optional AI worker is running and semantic search is enabled, a small local text model may embed the typed search text, within a strict time limit, falling back to normal search.
- **I5:** A user cannot access another user's files or photos unless they have been explicitly shared. This is enforced server-side on every access path: API, downloads, previews, thumbnails, search, network shares, and background jobs.
- **I6:** Everything runs locally. No telemetry. No network calls at runtime except for features the user has explicitly enabled.
- **I7:** AI is optional and opt-in. The NAS must be fully functional with AI disabled.
- **I8:** AI work is always the last stage of the roadmap. Any stage added in the future is inserted before it, and the AI stage is renumbered.
- **I9:** Only the core server writes sidecar files and the search index. Other processes, including the AI worker, submit results to the core server, which validates and writes them. _(Added in 0.3.0, P003.)_
- **I10:** Destructive bulk operations (deleting duplicates, reducing media resolution or quality, creating or changing drive pools) always show a preview of what will change, require explicit confirmation, and keep an undo window wherever technically possible. Where undo is impossible (e.g. erasing drives to create a pool), the confirmation says so plainly. _(Added in 1.4.0, pre-approved by the user in P005.)_
- **I12:** User-supplied content is data, never code. The server never executes, evaluates, imports, or loads uploaded or transferred content as code, scripts, configuration, or plugins. Every program that parses untrusted content runs with least privilege, resource limits, and time limits. User content is served so that browsers cannot run it with the application's privileges. _(Added in 1.12.0, P010; pre-approved: the user's requirements UR-3 and UR-5 in invariant form; RULES 1.10.0.)_

_I12 added in 1.12.0 (P010). I11 is the number of the invariant proposed for the public release (plan 11c, pending Q66); it is not in force._

_I3 refined and I4 reworded in 1.9.0 by the user's decisions D2 and D3 (P008, S007 E059). Before: I3 "Photo sidecar JSON files are the source of truth for photo metadata. …"; I4 "… It never scans files or runs AI at query time."_

---
