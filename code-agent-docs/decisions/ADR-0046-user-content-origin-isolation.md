# ADR-0046: User content origin isolation

| Field | Value |
|---|---|
| Number | ADR-0046 |
| Status | Proposed |
| Date proposed | 2026-10-01 (session S007) |
| Date of last status change | 2026-10-01 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

User files are served by the same origin as the app. Since S02.6 every download and preview sends `nosniff`, `Content-Security-Policy: default-src 'none'; sandbox`, and `attachment` for active types, so a file cannot run script with the app's session (threats T-21, T-22). A browser bug in rendering a file, or a missed header on a new route, would still act inside the app's origin. Serving content from a separate origin removes that class of risk (NFR-066; the user's requirement UR-3). The public release must have it (R09).

## Options considered

### Option A: Keep one origin with strict headers
- **Pros:** simple; already built.
- **Cons:** relies on every content route sending the right headers and on browsers honoring them.

### Option B: A separate origin for content (recommended)
- **LAN:** a second port on the same host (a different port is a different origin; the same certificate covers it). **Public:** a separate host name.
- App cookies are never sent to it, so access uses **short-lived, single-item URLs bound to the user** (an unguessable capability created by the app after authorization, valid for minutes, tied to the user and the item, never reusable for another item).
- Every response keeps the S02.6 headers.
- **Pros:** a rendering bug cannot reach the app's session or API.
- **Cons:** a second listener; URL capabilities must be designed with care (lifetime, binding, leakage through logs and the `Referer` header; `Referrer-Policy: no-referrer`).

### Option C: A sandboxed iframe on the same origin
- **Pros:** no second listener.
- **Cons:** weaker; the sandbox attribute can be wrong in one place.

## Decision

**Option B (recommendation while Proposed):** Should on the LAN, Must before R09. When it is built on the LAN is decided when this ADR is accepted (S03.5 or a later stage; at the latest R09).

## Consequences

- **Easier:** content rendering can never touch the app session.
- **Harder:** a second listener and port in the firewall rule; capability URLs for content.
- **Required:** the listener manager (S03.4) supports it; tests that app cookies never reach the content origin.

## Approval record

_Not yet approved._

## Links

- **Related requirements:** NFR-066, NFR-022, FR-005, FR-145; I12
- **Related ADRs:** ADR-0042 (local-origin protection), ADR-0047 (TLS)
- **Related stages:** S03, R09
- **Plan version:** 1.12.0
