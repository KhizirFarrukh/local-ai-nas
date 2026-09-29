# ADR-0042: Protection against other websites before login (Host and Origin checks)

| Field | Value |
|---|---|
| Number | ADR-0042 |
| Status | Proposed |
| Date proposed | 2026-09-30 (session S007) |
| Date of last status change | 2026-09-30 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

Until sessions and CSRF tokens exist (S03.3, S03.5), the API answers any request that reaches it. A web page the user visits can make the browser send requests to the NAS: cross-site form posts, and **DNS rebinding**, where a hostile domain resolves to `127.0.0.1` or the NAS's LAN address, so its requests look same-origin (threat T-19, bug S03-B01, P008 F2, FR-350, NFR-057).

## Options considered

| Option | Pros | Cons |
|---|---|---|
| **(a) Host allow-list + Origin check on state-changing requests + no CORS** | Stops rebinding and cross-site writes without accounts; stays useful after login | Non-browser clients must send a correct Host (they do) |
| (b) (a) plus a per-install token (P008) | Also stops non-browser local processes | Sessions and CSRF tokens arrive in the same stage (S03), so the token would be built and removed within weeks |
| (c) Wait for sessions | No extra work | The GUI can delete files from another website until then |

## Decision

**(a).** The per-install token of P008 is not built (an adaptation to the real state: S03 is already in progress).
- **Host allow-list per listener:** the loopback listener accepts `127.0.0.1`, `localhost`, and `[::1]` with its port; with the container exception (`server.allow_container_bind`) also the names in a new `server.allowed_hosts` setting; the LAN listener (S03.4) its configured names, the host name, `<hostname>.local`, and its own addresses. Anything else gets `421 Misdirected Request` as a problem.
- **Origin check:** every request that is not GET, HEAD, or OPTIONS must have an `Origin` that is one of the listener's own origins; a foreign `Origin` (including `null`) gets `403`. A **missing** `Origin` is allowed: browsers always send it on such requests (MDN: "same-origin requests except for GET or HEAD requests"), so its absence means a non-browser client such as a script, the S01 demo, or the tests.
- **Referrer policy:** the app's `Referrer-Policy` changes from `no-referrer` to `same-origin`. Referrers still never leave the NAS, and browsers then send the real origin on the NAS's own non-GET requests (MDN: certain referrer policies make `Origin` `null` for non-GET requests outside cors mode). The F2 system test confirms it in Chromium, Firefox, and Edge.
- **CORS:** no `Access-Control-Allow-Origin` anywhere (tusd's CORS stays off).
- These checks stay after login; sessions add CSRF tokens (S03.5).

## Consequences

- **Easier:** closes bug S03-B01 before accounts exist.
- **Harder:** reverse proxies must pass the original Host (documented in S14).
- **Required:** the first part of S03.5-T02 (built first in S03); tests in S03.10-T02.

## Approval record

_Proposed with plan 1.9.0 (P008). The user asked for the review's suggestions to be "added into the plan so these can also be implemented along the way" and approved building the follow-up tasks ("Approve all, F2 first (Recommended)", S007 E059); acceptance of this ADR's decision is asked before its first task starts._

## Links

plan FR-350, NFR-057, 8.39; threat model T-19; stage document S03 section 4.5 · [ADR-0010](ADR-0010-security-building-blocks.md) · P008 F2
