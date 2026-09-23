# ADR-0002: API style: REST with an OpenAPI contract; router and code generation

| Field | Value |
|---|---|
| Number | ADR-0002 |
| Status | **Accepted** |
| Date proposed | 2026-09-24 (session S002) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none |
| Superseded by | none |

> **History:** drafted in S002 as *Proposed* ("API style and resumable upload protocol"). Updated in S003 to the user's decision (P003). The resumable-upload sub-decision now lives in its own ADR, **ADR-0008**, because P003 lists it separately.

## Context

S01.5 needs a consistent, versioned, documented HTTP API (FR-075) with separate namespaces for files and photos. The API must work from plain HTTP clients (S01 exit criteria), the web UI (S02, with a generated TypeScript client), and scripts with tokens (S03). It must extend without breaking clients (NFR-025).

## Options considered

### Option A: REST over HTTP with an OpenAPI 3 contract (chosen)
- **Pros:** universal clients; HTTP semantics (Range, ETag, caching) fit file serving; generated clients and documentation.
- **Cons:** conventions must be defined and enforced (done below).

### Option B: GraphQL
- **Cons:** poor fit for binary streaming and ranges; harder caching; a larger security surface.

### Option C: gRPC / Connect
- **Cons:** not natively browser- or `curl`-friendly for downloads.

### Option D: WebDAV as the primary API
- **Cons:** XML; weak pagination and search. It is added separately for network drives (ADR-0015).

### Router (agent decides)
- **Go standard library `net/http.ServeMux`** (method and wildcard patterns since Go 1.22): no dependency.
- **chi v5.3.2** (MIT, verified): route groups and middleware composition.

### Spec workflow (agent decides)
- **Spec-first**, with the contract at `api/openapi.yaml` and code generated from it.
- **Code-first**, with the spec generated from Go annotations.

## Decision

**REST over HTTP, versioned under `/api/v1`, with an OpenAPI 3 specification at `api/openapi.yaml` as the contract** (user decision, P003).

**Conventions (carried over from the S002 draft):**
- **Namespaces:** `/api/v1/files`, `/api/v1/photos` (reserved until S04; returns problem `not_available`), `/api/v1/system`, and later `/search`, `/users`, `/shares`, `/admin`.
- **Errors:** RFC 9457 problem details (`application/problem+json`) with stable `code` values and a correlation ID. No internals are exposed.
- **Addressing:** the item path is a normalized, area-relative **query parameter or JSON field** (e.g. `GET /api/v1/files/items?path=/Docs/a.pdf`). This avoids `%2F` and dot-segment ambiguity in URL paths.
- **Operations:** `POST /api/v1/files/operations/{rename|move|copy}` with `on_conflict` = `fail` (default) | `rename` | `overwrite`.
- **Pagination:** cursor-based (`cursor`, `limit`), with `sort` and `order`.
- **Downloads:** `GET /api/v1/files/content?path=…` with `Range`, `ETag`, `If-None-Match`, `Last-Modified`, and a safe `Content-Disposition`.
- **Resumable uploads:** see **ADR-0008** (tus).

### Implementation details chosen by agent
| Detail | Choice | Reason |
|---|---|---|
| Router | **Standard library `net/http.ServeMux`** (Go 1.22+ patterns). Middleware as plain `func(http.Handler) http.Handler` chains. **chi v5.3.2** is the documented fallback if grouping or middleware needs outgrow it | "Prefer the Go standard library" (P003); the route set is small and regular |
| Spec workflow | **Spec-first**: hand-written `api/openapi.yaml`. Go server interfaces and models are generated with **oapi-codegen v2.8.0** (Apache-2.0, verified; `std-http` strict server target), with runtime `github.com/oapi-codegen/runtime` **v1.7.0** (Apache-2.0). Generated code is committed | The contract is reviewed as a document; generated interfaces make the handlers fail to compile if the spec changes |
| Spec/code agreement in CI | CI runs `go generate ./...` and fails on `git diff --exit-code`. Integration tests also validate responses against the spec | Satisfies "CI must fail if the spec and the code disagree" |
| TypeScript client (S02) | **openapi-typescript 7.13.0** (types) + **openapi-fetch 0.17.0** (typed fetch), both MIT, verified | Small runtime, types only from the spec; regenerated in CI with a drift check |
| API docs page (S01.5) | **Redoc 2.5.4** (MIT, verified) standalone bundle, **vendored** and served by the core (`go:embed`), with no CDN. Also evaluated: swagger-ui-dist 5.33.0 (Apache-2.0) and @scalar/api-reference 1.71.0 (MIT) | Offline docs (I6); read-only, a single file |
| Tool pinning | oapi-codegen pinned via a `tool` directive in `go.mod` | Reproducible generation |

## Consequences

- **Easier:** reviewable contract; generated server interfaces and TypeScript types; HTTP-client demos; offline docs.
- **Harder:** hand-writing the spec (mitigated by generated interfaces that fail to compile on mismatch); tus endpoints live outside the generated set (documented in the spec as an external protocol).
- **Required:** route-conventions document (S01.5-T01); error catalogue (S01.5-T03); CI drift check (S01.5-T05).

## Approval record

> Decision made by the user in plan change request #3 (`code-agent-docs/prompts/P003-technology-stack.json`, `technology_stack.api`). The router, spec workflow, and TypeScript generator are "agent decides" items (listed above for the user to object to)
> (2026-09-24, session S003). Verified in S003 log E005/E007.

## Links

- **Related requirements:** FR-005, FR-073, FR-075, FR-077, NFR-001, NFR-025
- **Related ADRs:** ADR-0001, ADR-0004, ADR-0005, ADR-0008 (tus, split out of this ADR), ADR-0009 (web UI client)
- **Related stages:** S01.5 onward
- **Plan version:** 0.3.0
