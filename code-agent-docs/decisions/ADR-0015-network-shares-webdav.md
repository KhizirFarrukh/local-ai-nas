# ADR-0015: Network shares: WebDAV first, via golang.org/x/net/webdav

| Field | Value |
|---|---|
| Number | ADR-0015 |
| Status | **Accepted** (SMB is a separate, **Proposed**, deferred ADR: ADR-0019) |
| Date proposed | 2026-09-24 (session S003) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none |
| Superseded by | none |

## Context

S09 makes the NAS usable as a network drive (FR-009, FR-124–FR-126). Network access must follow exactly the same rules as the web UI:
- Area separation (I1).
- Per-user access (I5).
- Sidecar lifecycle (S05.6) and index updates.
- Audit.

P003 resolves open question Q32 (in part): WebDAV first, SMB later, Linux-only.

## Options considered

| Option | Pros | Cons |
|---|---|---|
| **WebDAV served by the core via `golang.org/x/net/webdav`** (chosen) | Every operation passes through the core (policy, area services, sidecars, index, audit). Cross-platform clients. No extra service | Client quirks (Windows WebClient); slower than SMB |
| SMB via Samba | Native network-drive experience, fast | Bypasses the app (needs the watcher, ADR-0016). Complex permission mapping. Linux-only. Deferred: ADR-0019 |
| Both | Best coverage | Most work. SMB is deferred |

## Decision

- **WebDAV via `golang.org/x/net/webdav`** (in `golang.org/x/net` **v0.59.0**, BSD-3-Clause, verified).
- **Implement `webdav.FileSystem`** so that every WebDAV operation goes through the core's authorization (`authorize()`), area separation, and sidecar logic. Network access then follows exactly the same rules as the web UI (I1, I5).
- **Verified:** the package defines `type FileSystem interface { Mkdir, OpenFile, RemoveAll, Rename, Stat }`, and `webdav.Handler` has a `FileSystem` field, so a custom implementation is supported.
- **SMB via Samba:** deferred, Linux-only, and still **Proposed** (ADR-0019). Mapping per-user permissions onto Samba needs its own design in S09.1.

### Implementation details chosen by agent (to be confirmed in S09.1 with the threat model)
| Detail | Choice | Reason |
|---|---|---|
| Lock system | `webdav.NewMemLS()` (in-memory locks, lost on restart) | Adequate for household use; documented |
| Authentication | HTTP Basic **over HTTPS only**, using NAS accounts or per-user app passwords (S03.3 API tokens) | WebDAV clients support Basic. HTTPS is mandatory (Windows WebClient also requires it for Basic) |
| Mount layout | `/dav/files/` and `/dav/photos/` (photos read-only by default, pending Q32's second half) plus a "Shared with me" view | I1 stays visible to the user. Exposure policy per S09.3 |
| Windows client notes | Document the WebClient service and the 50 MB default file-size limit registry setting in the S09.5 connection instructions | Known client limitations |

## Consequences

- **Easier:** one enforcement path for all access; no watcher needed for WebDAV writes.
- **Harder:** WebDAV client quirks across operating systems (the S09.6 client matrix); performance compared with SMB.
- **Required:** the S09.6 client test matrix (Windows, macOS, Linux); cross-user tests over WebDAV.

## Approval record

> Decision made by the user in plan change request #3 (`code-agent-docs/prompts/P003-technology-stack.json`, `technology_stack.network_shares`)
> (2026-09-24, session S003). The custom FileSystem capability was verified in S003 log E005.

## Links

- **Related requirements:** FR-009, FR-124, FR-125, FR-126, NFR-024, NFR-026
- **Related ADRs:** ADR-0010, ADR-0016, ADR-0019 (SMB, Proposed)
- **Related stages:** S09
- **Plan version:** 0.3.0
