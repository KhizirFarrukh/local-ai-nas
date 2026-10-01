# ADR-0047: TLS and certificate policy

| Field | Value |
|---|---|
| Number | ADR-0047 |
| Status | Proposed |
| Date proposed | 2026-10-01 (session S007) |
| Date of last status change | 2026-10-01 (session S007) |
| Supersedes | none (amends the TLS part of ADR-0010 once accepted) |
| Superseded by | none |

## Context

The approved S03 design (stage document 4.4) gives LAN access over HTTPS only, TLS 1.2 minimum, Go's default suites, a self-signed ECDSA P-256 certificate, and HSTS on every HTTPS answer. The user's requirement UR-2 (secure connections between every client and the server) and P010 refine it (NFR-067):

1. **HSTS with a self-signed certificate is harmful:** once a browser has seen HSTS for a host, it allows no click-through on a certificate warning. When the self-signed certificate is renewed (397-day validity) or regenerated, users are locked out (threat T-66).
2. **AEAD-only suites for TLS 1.2.** Go's default list for TLS 1.2 has been changing between versions; whether it still contains CBC suites in Go 1.27 is **Unverified** (2026-10-01). Setting the list explicitly removes the doubt.
3. **Pinning by clients:** the pairing QR code carries the certificate's SHA-256 fingerprint (FR-364, NFR-069), so apps trust on first use with verification and never "ignore certificate errors".
4. **Plain HTTP on loopback** stays in the approved design; P010 asks for HTTPS everywhere after the localhost-only phase. This is question Q87.

## Options considered

### Option A: Keep the approved design
- **Cons:** the HSTS lockout; suites depend on Go's defaults.

### Option B: The approved design with the P010 refinements (recommended)
- TLS 1.3 preferred, TLS 1.2 minimum; TLS 1.2 limited to ECDHE with AES-GCM or ChaCha20-Poly1305 (explicit list); checked with testssl.sh in the S03 gate.
- **HSTS only when a trusted certificate is in use** (an owner-provided certificate from a trusted CA, or ACME in R06).
- The fingerprint in the console, in `local-ai-nas tls generate`, and in the pairing QR code.
- Owner-provided certificates; ACME with R06; a local CA option may come later.
- Loopback HTTP per the answer to Q87.

## Decision

**Option B (recommendation while Proposed).** S03.4-T01 changes accordingly (proposed in the S03 stage document, P010).

## Consequences

- **Easier:** no lockouts; predictable suites.
- **Harder:** a second rule for when HSTS applies.
- **Required:** tests for both cases (self-signed: no HSTS; trusted: HSTS); testssl.sh in the gate.

## Approval record

_Not yet approved._

## Links

- **Related requirements:** NFR-067, FR-088, NFR-020, NFR-069, FR-364, FR-276
- **Related ADRs:** ADR-0010 (authentication and sessions; TLS part amended)
- **Related stages:** S03.4, R06, R07, R08
- **Plan version:** 1.12.0
