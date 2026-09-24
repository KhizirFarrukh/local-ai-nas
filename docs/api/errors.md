# API errors

Every error response of the local-ai-nas API has the same format: an RFC 9457 **problem details** object, sent with the media type `application/problem+json`. The code is in [`internal/apperr`](../../internal/apperr/apperr.go).

This page describes the format and the codes. The spec [`api/openapi.yaml`](../../api/openapi.yaml) has the same format as the `Problem` schema and lists, for every endpoint, the error responses it can give. A contract test checks every error response against that schema (S01.5-T03).

## Format

```json
{
  "type": "about:blank",
  "title": "Not Found",
  "status": 404,
  "detail": "no item at /docs/report.pdf",
  "code": "not_found",
  "correlation_id": "3f9c2a7e0b1d4c6f8a2e5b7d9c1f3a5e"
}
```

| Field | Meaning |
|---|---|
| `type` | Always `about:blank` for now, so `title` is the standard HTTP status text (RFC 9457 section 4.2.1). |
| `title` | The HTTP status text. |
| `status` | The HTTP status code, the same as the response status. |
| `detail` | A human-readable explanation for this occurrence. It never contains internal details such as server paths or stack traces. |
| `code` | A stable, machine-readable code. Programs should use this, not `title` or `detail`. |
| `rule` | Only for some codes: the exact rule that was broken, for example `reserved_name` for `invalid_name` (see [Name rules](#name-rules)). |
| `correlation_id` | The request ID. The same ID is in the `X-Request-ID` response header and in every server log line for this request. |

Unexpected server errors (`internal`) always have the same generic `detail`. The actual cause is only in the server log, under the `correlation_id`. A crash inside a request handler is handled the same way.

## Codes

| Code | HTTP status | Meaning |
|---|---|---|
| `invalid_request` | 400 | The request is malformed: bad parameters or body. |
| `invalid_name` | 400 | A file or folder name breaks the name rules. |
| `outside_root` | 400 | A path would leave your storage area. |
| `not_found` | 404 | The file, folder, or upload does not exist, or there is no endpoint at this path. |
| `conflict` | 409 | The request clashes with the current state, for example the target already exists. |
| `method_not_allowed` | 405 | The endpoint exists, but not with this method. The `Allow` header lists the methods it accepts. |
| `length_required` | 411 | An upload without a `Content-Length` header. The server needs the size first, to check the limits and the free space before storing anything. |
| `precondition_failed` | 412 | A conditional request (`If-Match`, `If-Unmodified-Since`) does not match the file's current version. |
| `too_large` | 413 | The upload or request is over a size limit. |
| `range_not_satisfiable` | 416 | The `Range` of a download lies outside the file. The `Content-Range` header gives the file's size (`bytes */<size>`). |
| `too_large_for_sync` | 422 | The operation is valid but too big to run within one request, for example a copy over the synchronous copy limits (`copy.sync_max_items`, `copy.sync_max_bytes`). Larger copies become background jobs in a later stage. |
| `internal` | 500 | An unexpected server error. See the server log under the `correlation_id`. |
| `not_available` | 501 | The feature is part of the API but not available yet (for example the photos API in stage 1). |
| `insufficient_storage` | 507 | The write would use the free space kept in reserve. |

Codes are stable: a code keeps its meaning, and new codes are only added.

## Name rules

A name the API is asked to create (an upload, a new folder, the target of a rename, move, or copy) must follow these rules, on every operating system, so the storage works the same on Linux and Windows. An invalid name is refused with `invalid_name` and the `rule` below. Names are never changed silently. Files that already exist on disk with other names can still be read.

| `rule` | The name is refused when |
|---|---|
| `empty_name` | it is empty |
| `dot_name` | it is `.` or `..`, or reads as dots in compatibility form (such as fullwidth `．．` or `‥`) |
| `reserved_name` | it is a Windows device name, with any extension and in any case: `CON`, `PRN`, `AUX`, `NUL`, `CONIN$`, `CONOUT$`, `COM0`–`COM9`, `LPT0`–`LPT9`, `COM¹`–`COM³`, `LPT¹`–`LPT³` (for example `aux.txt`), or it starts with `.local-ai-nas-tmp-`, the prefix of the server's temporary files |
| `forbidden_character` | it contains any of `<` `>` `:` `"` `/` `\` `\|` `?` `*` |
| `control_character` | it contains a control character (U+0000 to U+001F, or U+007F) |
| `trailing_dot_or_space` | it ends with a dot or a space |
| `name_too_long` | it is longer than 255 bytes in UTF-8 |
| `path_too_long` | the whole path inside your storage area is longer than 4096 bytes |
| `invalid_utf8` | it is not valid UTF-8 |
| `lookalike_separator` | it contains a character that looks like `/` or `\` and that some tools turn into one: `／` `＼` `∕` `⁄` `∖` `⧵` `⧸` `⧹` `﹨` (other fullwidth characters, such as `？`, are fine) |

Paths are also normalized to Unicode NFC: a name sent in decomposed form (common on macOS) is the same file as its composed form.
