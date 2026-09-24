# API errors

Every error response of the local-ai-nas API has the same format: an RFC 9457 **problem details** object, sent with the media type `application/problem+json`. The code is in [`internal/apperr`](../../internal/apperr/apperr.go).

This page describes the format and the codes that exist so far. The full catalogue, with the codes for each endpoint and contract tests, comes with the API layer (S01.5-T03).

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
| `correlation_id` | The request ID. The same ID is in the `X-Request-ID` response header and in every server log line for this request. |

Unexpected server errors (`internal`) always have the same generic `detail`. The actual cause is only in the server log, under the `correlation_id`. A crash inside a request handler is handled the same way.

## Codes

| Code | HTTP status | Meaning |
|---|---|---|
| `invalid_request` | 400 | The request is malformed: bad parameters or body. |
| `invalid_name` | 400 | A file or folder name breaks the name rules. |
| `outside_root` | 400 | A path would leave your storage area. |
| `not_found` | 404 | The file, folder, or upload does not exist. |
| `conflict` | 409 | The request clashes with the current state, for example the target already exists. |
| `too_large` | 413 | The upload or request is over a size limit. |
| `internal` | 500 | An unexpected server error. See the server log under the `correlation_id`. |
| `not_available` | 501 | The feature is part of the API but not available yet (for example the photos API in stage 1). |
| `insufficient_storage` | 507 | The write would use the free space kept in reserve. |

Codes are stable: a code keeps its meaning, and new codes are only added.
