# API conventions

The rules every local-ai-nas API endpoint follows (S01.5-T01, ADR-0002). New endpoints are checked against the [review checklist](#review-checklist) at the end.

The contract is [`api/openapi.yaml`](../../api/openapi.yaml) (OpenAPI 3.0.3). The Go server code is generated from it (spec-first), so the spec and the server cannot drift apart.

## Base path and areas

- Every endpoint is under **`/api/v1`**. The offline API documentation is under `/api/docs`. Nothing else is served. Versioning is described in [versioning.md](versioning.md).
- The storage has two areas that never mix:
  - `/api/v1/files/...` is the **files area**.
  - `/api/v1/photos/...` is **reserved** for the photos area. Until the media stages (S04), every request there gets `501` with the problem code `not_available`.
- `/api/v1/system/...` holds server-level endpoints, such as `GET /api/v1/system/health`.

## Namespaces

Each user has their own namespace in each area (ADR-0003). The API always shows the caller's namespace as `/`: a client never names a namespace and cannot reach another one. In S01 there is one owner, `u0001`, and no login yet (S03).

## Naming

- Paths use lowercase words and plural nouns for collections: `/files/items`, `/files/folders`, `/files/uploads`.
- Actions that are not plain create, read, update, or delete are `POST` requests to `/…/operations/<verb>`, for example `/files/operations/move`.
- JSON field names and query parameters are `snake_case`: `on_conflict`, `next_cursor`, `mod_time`.

## Addressing files and folders

- A file or folder is addressed by its **path** in the namespace. The path goes in the `path` query parameter (or a body field such as `from` and `to` for operations), never in the URL path. So names with any allowed characters work without special URL routing.
- A path starts with `/` (the namespace root), uses `/` as the only separator, and is UTF-8. It is percent-encoded once, as usual for query parameters.
- The server **refuses** paths that could leave the namespace, such as `..` segments, backslashes, drive letters, UNC-like `//` paths, and double-encoded separators. It normalizes paths to Unicode NFC. The full list is in [errors.md](errors.md#name-rules) and in the resolver (S01.6).
- **Names the API creates** must follow the [name rules](errors.md#name-rules). A name that breaks a rule is refused with `invalid_name` and the `rule`. It is never changed silently.

## Requests and responses

- Request and response bodies are JSON (`application/json`, UTF-8). The exceptions are file content (`application/octet-stream`) and tus uploads.
- Times are RFC 3339 in UTC with a `Z` suffix, for example `2026-09-24T07:15:00Z`.
- Sizes are whole numbers of bytes.
- Booleans in query parameters are `true` or `false`.
- Unknown JSON fields in a request body are refused with `invalid_request`, so typos surface instead of being ignored.

## Listing: pagination and sorting

- Lists are paged with an **opaque cursor**. A response has `items` and, when more remain, `next_cursor`. The client passes `cursor=<next_cursor>` to get the next page. A cursor is only valid for the same listing, sort, and order.
- `limit` sets the page size: default 100, maximum 1000.
- `sort` is one of `name` (default), `size`, `mod_time`, `kind`. `order` is `asc` (default) or `desc`. Ties are always broken by name, so pages never overlap or skip items.

## Conflicts

Operations that create an item (simple upload, tus upload, new folder, copy, move, rename) take `on_conflict`:

| Value | When the target already exists |
|---|---|
| `fail` (default) | The request fails with `409 conflict`. |
| `rename` | The new item gets a free name in the pattern `name (1).ext`, `name (2).ext`, and so on. The response shows the name used. |
| `overwrite` | The existing file is replaced atomically. Folders are never merged or overwritten. |

## Simple upload

`PUT /api/v1/files/content?path=&on_conflict=` stores the request body, the raw file content, as the file at `path`. The Content-Type of the body is not used.

- `Content-Length` is required (`411 length_required` without it). The server checks the size against `uploads.max_file_size` (`413 too_large`) and the free-space reserve (`507 insufficient_storage`) before it reads the body. The parent folder must exist, and with `on_conflict=fail` an existing target is refused (`409`) before the body is read as well. A client that sends `Expect: 100-continue` does not send a refused body at all.
- The content goes to a hidden temporary file in the target folder, which is synced to disk and then renamed to its name in one step. A partly written file is never visible: listings skip temporary files (names starting with `.local-ai-nas-tmp-`, which the name rules reserve), and the target name appears only when the file is complete. A body shorter or longer than `Content-Length`, or a dropped connection, leaves nothing behind.
- The answer is `201` with the new item (with `on_conflict=rename`, under the name used), or `200` when `on_conflict=overwrite` replaced a file.
- For large files, or on unreliable networks, use the resumable upload below.

## Errors

- Every error response is an RFC 9457 problem (`application/problem+json`) with a stable `code`, sometimes a `rule`, and the `correlation_id`. See [errors.md](errors.md).
- Unknown endpoints get `404 not_found`, and a known endpoint with the wrong method gets `405 method_not_allowed` with an `Allow` header. Both are problems too.

## Request IDs

Every response has an `X-Request-ID` header. A client may send its own (1 to 128 characters: letters, digits, `-`, `_`, `.`, `:`) and the server keeps it. The same ID is the `correlation_id` of a problem and appears in every server log line for the request.

## Resumable uploads (tus)

Large uploads use the **tus 1.0.0** protocol at `/api/v1/files/uploads/` (S01.4). tus is an external, published protocol, so the spec describes it but does not redefine it. The client sends the target in the upload metadata:

| tus metadata key | Meaning |
|---|---|
| `target_path` | The path of the finished file, as above. Required. |
| `on_conflict` | `fail`, `rename`, or `overwrite`. Default `fail`. |
| `sha256` | Optional lowercase hex SHA-256 of the whole file. The server checks it before the file appears. |

## Review checklist

Every endpoint is reviewed against this list before it is merged. The table after the list records the review of each endpoint.

1. The path is under `/api/v1` and follows the naming rules. File operations are in `/files`; nothing new goes into the reserved `/photos`.
2. The spec (`api/openapi.yaml`) describes it, including every parameter, body, response, and error. The handler implements the generated interface.
3. Paths go through the resolver, and created names through the name rules. No handler touches the disk directly.
4. Every error is a problem with a documented code, and the spec lists the error responses.
5. Input is validated before the service layer: types, ranges, and enumerations. Invalid input gives a 4xx, never a 5xx.
6. Large bodies are streamed, never read whole into memory. Size limits and the free-space reserve are checked before writing.
7. Tests cover success, each documented error, and invalid input.

| Endpoint | Checklist result |
|---|---|
| `GET /api/v1/system/health` | 1–4 and 7 met; 5 and 6 nothing to check (no input, no request body). In the spec since S01.5-T03 and served by the generated strict handler since S01.5-T04. A failing check gives 503 with the report, not a problem, by design. |
| `* /api/v1/photos`, `* /api/v1/photos/…` | 1–4 and 7 met; 5 and 6 nothing to check. Always `501 not_available`; in the spec as `/photos` since S01.5-T03. |
| `GET /api/v1/files/items` | 1–7 met. Path through the resolver (3); paging values checked before the service, which a fake-service test proves (5); no request body (6 n/a); errors 400/404/405/500 in the spec and in the contract test (2, 4); folder, file, paging, and error tests plus FuzzAPI (7). |
| `POST /api/v1/files/folders` | 1–7 met. Path through the resolver, and the new folder plus every missing parent through the name rules before anything is created (3); the body must be JSON with no unknown fields and no trailing data, and `on_conflict` is checked against its enumeration before the service, which a fake-service test proves (5); the body is small and capped by the server's body limit, over which the answer is 413 (6); errors 400/404/405/409/413/500 in the spec and in the contract test (2, 4); created, renamed, existing, and parents tests, the invalid-input test, and FuzzAPI (7). |
| `PUT /api/v1/files/content` | 1–7 met. Path through the resolver and the name through the name rules (3); `on_conflict`, the declared size (411, 413), the free space (507), the parent, and a `fail` conflict are checked before the body is read, which a fake-service test and a real-client test with `Expect: 100-continue` prove (5); the body is streamed with a 1 MiB buffer into a temporary file, fsynced, and committed atomically; its route has its own body limit, `uploads.max_file_size` (6); errors 400/404/405/409/411/413/500/507 in the spec, all but 507 in the contract test and 507 in the files tests (2, 4); byte-identity (SHA-256), each conflict policy, concurrent renames, invisibility during the upload, failures that leave nothing, and FuzzAPI (7). |
| Other file endpoints (S01.3, S01.4) | Reviewed when they are added. |
