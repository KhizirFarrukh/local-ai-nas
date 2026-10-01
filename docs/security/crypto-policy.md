# Cryptography policy

| Field | Value |
|---|---|
| Version | 1.0 (2026-10-01, plan change request #10) |
| Plan requirement | NFR-080 (this policy), NFR-067 (TLS), NFR-074 (secrets) |
| ISO/IEC 27001:2022 | Control 8.24, use of cryptography (based on public summaries) |
| Decisions | ADR-0010 (passwords and sessions), ADR-0021 (SHA-256), ADR-0047 (TLS and certificates, Proposed), ADR-0048 (secrets, Proposed), ADR-0049 (release signing, Proposed; direction decided by the owner, D2) |

**No custom cryptography, ever.** Only vetted libraries: the Go standard library and `golang.org/x/crypto`, the platform's libraries on phones and desktops, and the tools named below.

## 1. Randomness

| Where | Source |
|---|---|
| Go | `crypto/rand` only for keys, tokens, IDs that must be unguessable, salts, nonces; `math/rand` is forbidden in security code (lint) |
| Python (AI worker) | `secrets` |
| Browser | `crypto.getRandomValues` |

Tokens and secrets have at least 128 bits of entropy (public links: at least 128 bits, R09).

## 2. Algorithms

| Purpose | Algorithm | Notes |
|---|---|---|
| Passwords | **Argon2id** with the ADR-0010 parameters, PHC format | Built in S03.2-T02. Rules from NIST SP 800-63B: a minimum length (12 characters), long passwords allowed, no forced composition rules; breached-password check in R05 (FR-268) |
| Integrity of content | **SHA-256** (ADR-0021) | Content hashes since S01.3-T10 |
| Message authentication (webhooks, signed tokens) | **HMAC-SHA-256** | Constant-time comparison |
| Secrets at rest (SMTP and webhook credentials, two-factor secrets) | **AES-256-GCM** or **XChaCha20-Poly1305** | The key is kept outside the database (NFR-074, ADR-0048), so a database backup alone does not reveal them |
| Session and API tokens | Random values; only their **SHA-256** is stored | Shown once (S03.3) |
| TLS | See section 3 | |
| Release signatures | **Cosign keyless** (Sigstore) and **Minisign** (Ed25519) | The owner's decision D2; the updater verifies Minisign (FR-381) |
| Two-factor codes | TOTP (RFC 6238) | S03.7 |

## 3. TLS (NFR-067)

- **HTTPS for every connection that leaves the computer.** The LAN listener is HTTPS only (S03.4). Plain HTTP exists only on loopback (the approved S03 design) and stays there as the one documented exception: that traffic never leaves the computer (Q87, decided 2026-10-01 under the owner's delegation); it can be turned off.
- **TLS 1.3 preferred, TLS 1.2 minimum.** For TLS 1.2 only ECDHE with AEAD cipher suites (AES-GCM, ChaCha20-Poly1305), set explicitly; Go's default list for TLS 1.2 is checked for the Go version in use (Unverified for Go 1.27 on 2026-10-01; ADR-0047). TLS 1.3 suites are Go's (all AEAD).
- **Checked with testssl.sh** in the S03 security gate.
- **Certificates:** a self-signed ECDSA P-256 certificate generated at first setup is the LAN default; its SHA-256 fingerprint is shown in the web interface and carried in the pairing QR code, so clients pin it (trust on first use with verification, never "ignore certificate errors"). Owner-provided certificates are supported; trusted ACME certificates come with remote access (R06).
- **HSTS only with a trusted certificate.** With a self-signed certificate, HSTS would stop browsers from letting the user continue past a certificate warning after the certificate is renewed, and lock users out (P010 change to the S03.4 design).
- **Outbound TLS always verifies certificates.** There is no "skip verification" option anywhere in the code (NFR-071).

## 4. Keys and secrets

| Secret | Created | Stored | Rotation |
|---|---|---|---|
| TLS private key | First setup or `local-ai-nas tls generate` | `.local-ai-nas/state/tls/`, 0600, restricted ACL on Windows | Regenerate before expiry (warned 30 days before); after a suspected leak at once |
| Secrets encryption key (ADR-0048) | First start, from `crypto/rand` | Outside the database, 0600 | Re-encrypt on rotation; after a suspected leak |
| Session tokens | Sign-in | SHA-256 only | Expire; "sign out everywhere" |
| API tokens and device tokens | Created by the user or by pairing | SHA-256 only | Revoke and recreate |
| Internal API token (AI worker, storage helper) | First start | 0600 | On every upgrade, or after a suspected leak |
| Minisign release key | The owner, once | Encrypted, in a protected GitHub environment; an offline backup with the owner | Documented procedure; the next public key announced in a signed release |

Procedures for each rotation are part of [incident response](incident-response.md).

## 5. What is not allowed

- Home-made algorithms, modes, or protocols.
- MD5 or SHA-1 for any security purpose (they may appear only where an external format requires them, never as protection).
- ECB mode, unauthenticated encryption, fixed or reused nonces.
- Keys, secrets, or passwords in the repository, container images, logs, or command lines.
