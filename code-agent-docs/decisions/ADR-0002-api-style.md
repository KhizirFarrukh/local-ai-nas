# ADR-0002: API style and resumable upload protocol

| Field | Value |
|---|---|
| Number | ADR-0002 |
| Status | Proposed |
| Date proposed | 2026-09-24 (session S002) |
| Date of last status change | 2026-09-24 (session S002) |
| Supersedes | none |
| Superseded by | none |

## Context

S01.5 needs a consistent, versioned, documented HTTP API with separate namespaces for files and (later) photos (FR-075). S01.4 needs chunked, resumable uploads (FR-004, FR-074). The API must be easy to call from an HTTP client in S01 (the exit criteria), from the web UI in S02, and from scripts with tokens in S03. It must be extendable without breaking clients (NFR-025).

## Options considered

### Option A: REST + JSON with an OpenAPI 3 contract
- **Pros:** universal client support; works with `curl`; OpenAPI enables generated clients and documentation; HTTP semantics (range, ETag, caching) fit file serving naturally.
- **Cons:** multiple round trips for complex views; conventions must be defined and enforced.

### Option B: GraphQL
- **Pros:** flexible queries for rich UIs.
- **Cons:** poor fit for binary streaming and range requests; caching is harder; heavier security surface (query cost).

### Option C: gRPC / Connect
- **Pros:** strong typing, streaming.
- **Cons:** not natively browser-friendly; awkward for plain-HTTP downloads and scripts.

### Option D: WebDAV as the primary API
- **Pros:** doubles as a network-drive protocol (S09).
- **Cons:** XML-based; weak pagination and search; awkward for a modern UI. It is better added as a separate surface in S09.

### Resumable uploads (sub-decision)

| Option | Pros | Cons |
|---|---|---|
| **tus 1.0 (open protocol)** | Documented standard with core, creation, termination, and checksum extensions. Mature clients (tus-js-client, Uppy). Server implementations available. | One more protocol surface to secure. |
| Custom chunked API | Full control. | Reinventing resumability; no off-the-shelf clients. |
| S3-style multipart | Well known. | Needs S3 semantics, which is overkill here. |

## Decision

**Recommendation (Proposed, not decided):**
- **REST + JSON**, versioned under **`/api/v1`**, with separate namespaces: `/api/v1/files`, `/api/v1/photos` (reserved until S04), `/api/v1/system`, and later `/api/v1/search`, `/users`, `/shares`, `/admin`.
- **OpenAPI 3** as the contract. CI fails if the spec and the implementation disagree.
- **Errors:** RFC 9457 problem details (`application/problem+json`) with stable machine-readable `code` values and a correlation ID. No internal details are exposed.
- **Item addressing:** a normalized, area-relative path passed as a query parameter or JSON field (e.g. `GET /api/v1/files/items?path=/Documents/a.pdf`). This avoids dot-segment and `%2F` ambiguities in URL paths. The server normalizes and validates every path (S01.6).
- **Operations:** move, copy, and rename are `POST /api/v1/files/operations/{move|copy|rename}` with a JSON body including `on_conflict` (`fail` | `rename` | `overwrite`, default `fail`).
- **Pagination:** cursor-based (`cursor`, `limit`), with sort parameters.
- **Downloads:** `GET /api/v1/files/content?path=…` with `Range`, `ETag`, `Last-Modified`, and a safe `Content-Disposition`.
- **Resumable uploads:** the **tus 1.0** protocol (core + creation + termination + checksum extensions) at `/api/v1/files/uploads/`. Sessions are stored in internal app data, and completed files are finalized by atomic rename (FR-074).

## Consequences

- **Easier:** HTTP-client demos in S01; generated TypeScript client in S02; standard upload clients.
- **Harder:** keeping the spec in sync (mitigated by the CI check); securing the tus surface (auth hooks from S03).
- **Required:** a route-conventions document; a shared error catalogue; API docs served locally without a CDN (I6).

## Approval record

> _Pending user review (S002)._

## Links

- **Related requirements:** FR-004, FR-005, FR-073, FR-074, FR-075, FR-077, NFR-001, NFR-025
- **Related ADRs:** ADR-0001, ADR-0003
- **Related stages:** S01.4, S01.5
- **Plan version:** 0.2.0
