# ADR-0008: Resumable uploads: the tus protocol (tusd embedded, Uppy in the browser)

| Field | Value |
|---|---|
| Number | ADR-0008 |
| Status | **Accepted** |
| Date proposed | 2026-09-24 (session S003); first drafted as a sub-decision inside ADR-0002 (S002) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none (split out of ADR-0002 before either was Accepted) |
| Superseded by | none |

## Context

Large, resumable uploads are needed in S01.4 (API), S02.4 (GUI), and S04.2 (photos): FR-004 and FR-074. Partial files must never appear in `files/` or `photos/`, and incomplete uploads live in internal app data (I2). Finalizing must be an atomic rename, which requires the same filesystem as the storage root (ADR-0003, A18). Authentication and authorization must run on uploads from S03.

## Options considered

| Option | Pros | Cons |
|---|---|---|
| **tus 1.0, with tusd as an embedded Go library and Uppy + @uppy/tus in the browser** (chosen) | An open, documented protocol. Mature server and clients. Resumability solved | Its own endpoint surface and hooks to integrate |
| Custom chunked-upload API | Full control | Reinvents resumability; no standard clients |
| S3-style multipart | Well known | Needs S3 semantics; overkill |

## Decision

- **Server:** **tusd v2.10.1** (`github.com/tus/tusd/v2`, MIT, verified 2026-09-14 release), **embedded in the core server**.
  - Evidence it can be embedded: `pkg/handler` exposes `handler.Config` with `StoreComposer`, `BasePath`, `PreUploadCreateCallback`, `PreFinishResponseCallback`, and `NotifyCompleteUploads`. `pkg/filestore` exists. `examples/server/main.go` embeds it via `handler.NewHandler`.
  - Mounted at `/api/v1/files/uploads/` (S01.4) and later `/api/v1/photos/uploads/` (S04.2).
- **Browser:** **Uppy** `@uppy/core` **6.0.1** + `@uppy/tus` **6.0.0** (MIT), from S02.4.
- **Storage of incomplete uploads:** tusd `filestore` rooted at `<internal data>/tmp/uploads/` (I2). The startup health check (S01.2) verifies that it is on the same filesystem as the storage root.
- **Finalize:** when an upload completes, the core validates the target (namespace, name rules, conflict policy, free space), fsyncs the data file, and **atomically renames** it into the target area through the FilesService or PhotosService. A partial file is therefore never visible.
  - **Fallback (per P003):** if temp and root cannot share a filesystem (e.g. a deployment mounts them separately), finalize copies to a temp name inside the target directory, fsyncs, and renames. This is documented and the health check warns.
- **Hooks:**
  - `PreUploadCreateCallback` validates upload metadata (target path, declared size vs. limits and free space). From S03 it runs the `authorize()` policy check.
  - The whole tus handler sits behind the same authentication middleware as the rest of the API (from S03).
  - `PreFinishResponseCallback` performs the finalize, so the client learns of finalize errors (e.g. conflict).
- **Integrity:** an optional client-provided SHA-256 in the upload metadata is verified at finalize. (Whether tusd implements the tus checksum extension was not verified, so the project does its own finalize-time check.)
- **Cleanup:** expired incomplete uploads are removed by a cleanup task (a simple scheduler in S01.4, a job in S04.3).

## Consequences

- **Easier:** robust resume with standard clients; a GUI queue with pause and resume through Uppy.
- **Harder:** mapping tus metadata to area and target semantics; making sure finalize is the only path into the areas.
- **Required:** finalize-path tests with fault injection (S01.4-T02); auth hooks in S03; photos-area media validation at finalize (S04.2).

## Approval record

> Decision made by the user in plan change request #3 (`code-agent-docs/prompts/P003-technology-stack.json`, `technology_stack.resumable_uploads`)
> (2026-09-24, session S003). Verified in S003 log E005 (tusd embeddability, hooks, filestore; versions and licenses).

## Links

- **Related requirements:** FR-003, FR-004, FR-074, FR-080, NFR-006, NFR-021
- **Related ADRs:** ADR-0002 (split from), ADR-0003 (temp location), ADR-0009 (Uppy), ADR-0010 (auth), ADR-0011 (cleanup job)
- **Related stages:** S01.4, S02.4, S04.2
- **Plan version:** 0.3.0
