# API versioning

How the local-ai-nas API changes over time without breaking its clients (S01.5-T02).

## The version is in the path

- The current version is **v1**: every endpoint is under `/api/v1`. The offline documentation at `/api/docs` is not versioned.
- A test checks that the API registers no route outside `/api/v1` and `/api/docs`. The web interface (S02) answers every path outside `/api` and is not part of the versioned API.
- The spec's `info.version` is the version of the document itself. It changes with every edit, while the path version changes only for breaking changes.

## What may change within v1

These changes are **compatible**. They can happen at any time, so clients must tolerate them:

- New endpoints.
- New optional query parameters and new optional request-body fields.
- New fields in response bodies. Clients must ignore fields they do not know.
- New values for `code` and `rule` in problems. Clients should treat an unknown code like its HTTP status.
- New enumeration values in *responses*, such as a new item `kind`. Clients should handle unknown values, for example by showing the item as generic.
- Wording changes in `title` and `detail` of problems. Programs use `code` and `rule`, never the text.
- New response headers.

## What is a breaking change

These changes are **not allowed within v1**. They need a new version (`/api/v2`):

- Removing or renaming an endpoint, a parameter, a request field, or a response field.
- Changing the type or format of a field, or the meaning of an existing value.
- Making an optional parameter or field required, or adding a new required one.
- Removing an enumeration value that clients may send, or changing a default (for example `on_conflict=fail`).
- Changing the status code or the problem `code` that an existing situation produces.
- Tightening validation so that requests that used to succeed now fail. Security fixes are the exception: they may refuse dangerous input, and the change is noted in the changelog.

## Deprecation

1. The deprecated element is marked `deprecated: true` in the spec, and the release notes say why and what replaces it.
2. Responses from a deprecated endpoint carry a `Deprecation` header (RFC 9745) and, once a removal date is set, a `Sunset` header (RFC 8594).
3. It keeps working for at least **two minor releases** or **six months**, whichever is longer, after the release that deprecated it.
4. Removal happens only in a new API version. `/api/v1` and `/api/v2` then run side by side during the same period.

## Before 1.0 of the product

Until the first stable release (S11.7), the API is **pre-release**. Breaking changes may still happen within v1 when the design needs them, but each one is recorded in the changelog and in the session log. From the first stable release on, the rules above apply without exception.
