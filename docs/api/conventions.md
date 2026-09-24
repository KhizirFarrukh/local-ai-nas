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

## Download

`GET /api/v1/files/content?path=` sends a file (`HEAD` sends only the headers). Folders and symbolic links are refused with `400`.

- **Ranges:** `Range: bytes=…` gives `206` with `Content-Range`: one range, an open end (`bytes=500-`), a suffix (`bytes=-500`), or several ranges (`multipart/byteranges`). A range outside the file gives `416 range_not_satisfiable` with `Content-Range: bytes */<size>`.
- **Versions:** every download has a strong `ETag` (the same as the item's `etag`) and `Last-Modified`. `If-None-Match` and `If-Modified-Since` give `304`; `If-Range` resumes only while the file is unchanged; `If-Match` and `If-Unmodified-Since` give `412 precondition_failed` when the file has changed.
- **Headers:** `Content-Type` is the file's media type. `Content-Disposition` is always `attachment`, with an ASCII `filename` and the exact UTF-8 name in `filename*` (RFC 6266, RFC 8187). `Cache-Control: private, no-cache` lets a browser keep a copy but revalidate it with the ETag. `X-Content-Type-Options: nosniff` and `Content-Security-Policy: default-src 'none'; sandbox` keep an uploaded HTML or SVG file from running as part of the site.

## Rename and move

- `POST /api/v1/files/operations/rename` with `{"path", "new_name", "on_conflict"}` renames an item in its folder. `new_name` is one name and must follow the name rules. A change of case only (`a.txt` to `A.txt`) works on every disk.
- `POST /api/v1/files/operations/move` with `{"from", "to", "on_conflict"}` moves an item to the full path `to`, so a move can also rename. The parent folder of `to` must exist. A folder cannot be moved into itself or below itself (`400`).
- Both answer `200` with the item at its new path. With `overwrite`, a file replaces a file atomically; a folder is never replaced or merged, and never replaces a file (`409`).

## Copy

- `POST /api/v1/files/operations/copy` with `{"from", "to", "on_conflict"}` copies a file, or a folder with everything in it, to the full path `to` (its parent must exist). Modification times are kept. The answer is `201` with the copy, or `200` when `overwrite` replaced a file.
- The whole source is checked before anything is written: the synchronous copy limits (`copy.sync_max_items`, default 1000 files and folders; `copy.sync_max_bytes`, default 1 GiB) give `422 too_large_for_sync` (background copies come in a later stage); every name the copy creates must follow the name rules; a symbolic link or special file in the tree is refused (`400`), because links are never followed; and the free-space reserve applies (`507`). A folder cannot be copied into itself (`400`).
- The copy is invisible until it is complete: it is written under a hidden temporary name, synced to disk, and then renamed in one step. A failed or cancelled copy leaves nothing behind.

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
| `GET /api/v1/files/content` | 1–7 met. Path through the resolver; only regular files, opened without following links and checked to be the file that was looked up (3); the path is the only parameter (5); the file is streamed by `http.ServeContent` from the open handle, never read whole (6); errors 400/404/405/409/412/416/500 in the spec; 412 and 416 are turned from ServeContent's plain text into problems and validated against the schema in the download tests, the rest in the contract test (2, 4); single, open-ended, suffix, past-the-end, unsatisfiable, malformed, and multiple ranges, every conditional header, HEAD, Content-Disposition encoding with a standard parser, and FuzzAPI (7). |
| `POST /api/v1/files/operations/rename` | 1–7 met. Source and target through the resolver, the new name through the name rules; temporary names are not found (3); strict JSON body and `on_conflict` checked before the service, which a fake-service test proves (5); small body under the default limit (6); errors 400/404/405/409/413/500 in the spec and in the contract test (2, 4); file, folder, case-only, and same-name renames, each conflict policy, invalid names, and FuzzAPI (7). |
| `POST /api/v1/files/operations/move` | 1–7 met as for rename; in addition the target's parent must be a folder, and a folder moving into itself is refused by comparing the target's ancestors with the source as file-system objects, which also catches case variants on Windows (3); moves of files, folders with contents, and names that only share a prefix, concurrent moves to one name (exactly one wins with `fail`, all succeed with `rename`, no file lost), and the no-hard-link fallback (7). |
| `POST /api/v1/files/operations/copy` | 1–7 met. Source and target through the resolver; every created name through the name rules during a scan that also refuses links, so nothing is followed (3); strict JSON body and `on_conflict` before the service, which a fake-service test proves (5); files are streamed with a 1 MiB buffer and synced, the limits and the free space are checked before the first byte (6); errors 400/404/405/409/413/422/500/507 in the spec, 422 validated in the API tests and 507 in the files tests, the rest in the contract test (2, 4); byte-identical trees with kept times, each conflict policy, every refusal leaving the disk unchanged, cancellation, concurrent copies to one name, and FuzzAPI (7). |
| Other file endpoints (S01.3, S01.4) | Reviewed when they are added. |
