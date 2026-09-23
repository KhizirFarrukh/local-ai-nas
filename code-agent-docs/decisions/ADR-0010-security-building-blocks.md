# ADR-0010: Security building blocks: Argon2id, server-side sessions, CSRF tokens, optional TOTP, TLS

| Field | Value |
|---|---|
| Number | ADR-0010 |
| Status | **Accepted** |
| Date proposed | 2026-09-24 (session S003) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none |
| Superseded by | none |

## Context

S03 makes the NAS safe to reach from the LAN: FR-064, FR-085–FR-091, NFR-010, NFR-022. This ADR fixes the building blocks. The threat model (S03.1) may add controls but does not change these choices without a superseding ADR.

## Options considered

| Area | Options | Chosen |
|---|---|---|
| Password hashing | Argon2id; bcrypt; scrypt | **Argon2id** (`golang.org/x/crypto/argon2`) |
| Sessions | Server-side sessions with an opaque cookie; stateless JWT | **Server-side sessions in SQLite** (revocable; nothing sensitive in the cookie) |
| CSRF | Synchronizer token; double-submit cookie; SameSite only | **Synchronizer token** + SameSite + Origin check |
| 2FA | TOTP library; hand-written RFC 6238 | **pquerna/otp** (if S03.7 is approved) |
| TLS | Go `crypto/tls`; reverse proxy only | **`crypto/tls`** built in; a reverse proxy remains possible |

## Decision

- **Argon2id** via **`golang.org/x/crypto/argon2`** (x/crypto **v0.57.0**, BSD-3-Clause).
- **Server-side sessions in SQLite** (ADR-0007). The cookie holds **only a random opaque ID** and is **HttpOnly, Secure, SameSite**.
- **CSRF tokens** for all state-changing requests.
- **TOTP** via **`github.com/pquerna/otp` v1.5.0** (Apache-2.0), **only if S03.7 is approved** (Q33).
- **HTTPS** via Go's **`crypto/tls`**, with user-provided certificates or a generated self-signed certificate.

### Implementation details chosen by agent
| Detail | Choice | Reason |
|---|---|---|
| Argon2id parameters | **t=3, m=64 MiB, p=4, 16-byte random salt, 32-byte tag** (RFC 9106 "second recommended option"), stored in PHC string format. Rehash on login when the parameters change. At most 2 concurrent hash operations (memory-DoS guard). S03.2 benchmarks on reference hardware and may lower `m` only to the OWASP minimum (m=19 MiB, t=2, p=1) with a recorded reason | A strong, standard parameter set that is still feasible on a 4 GB arm64 board |
| Session ID | 256-bit random value from `crypto/rand`, base64url-encoded. The database stores only its SHA-256, so a leaked database does not reveal live cookies | Standard practice |
| Cookie | Name `__Host-lan_session` under HTTPS (Path=/, no Domain). `HttpOnly; Secure; SameSite=Lax`. Idle timeout 7 days (configurable) and absolute lifetime 30 days | `__Host-` prefix hardening. Lax keeps links from other apps working, and CSRF tokens protect state changes |
| CSRF | Per-session synchronizer token, sent by the SPA in the `X-CSRF-Token` header on every non-GET request. `Origin`/`Referer` must match the NAS origin | Defense in depth |
| TLS | TLS 1.2 minimum, 1.3 preferred, Go default cipher suites. Self-signed certificates are ECDSA P-256 via `crypto/x509`, with SANs for configured hostnames and IPs and a 397-day validity | Modern defaults; SANs are required by browsers |
| TOTP maintenance note | pquerna/otp has low activity (last push 2025-08). If it becomes unmaintained before S03.7, implement RFC 6238 with `crypto/hmac` in about 50 lines (new ADR) | Keeps the fallback cheap |

## Consequences

- **Easier:** revocable sessions and "log out everywhere"; no tokens with long-lived claims; standard password security.
- **Harder:** a database lookup per request (cheap with SQLite). Memory use for Argon2id is bounded by the concurrency guard.
- **Required:** CSRF middleware exemptions only for token-authenticated script access (API tokens, S03.3) and the tus endpoints, which are protected by custom headers, SameSite cookies, and an Origin check (ADR-0008). All of this is reviewed in the S03.1 threat model.

## Approval record

> Decision made by the user in plan change request #3 (`code-agent-docs/prompts/P003-technology-stack.json`, `technology_stack.security`). Argon2id parameters are chosen by the agent per "Record the chosen parameters in the ADR"
> (2026-09-24, session S003). Versions and licenses verified in S003 log E005.

## Links

- **Related requirements:** FR-064, FR-085, FR-086, FR-087, FR-088, FR-091, NFR-010, NFR-020, NFR-022
- **Related ADRs:** ADR-0007, ADR-0008, ADR-0009
- **Related stages:** S03 onward
- **Plan version:** 0.3.0
