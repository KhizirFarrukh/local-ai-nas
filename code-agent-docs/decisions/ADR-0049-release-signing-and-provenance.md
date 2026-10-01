# ADR-0049: Release signing and provenance

| Field | Value |
|---|---|
| Number | ADR-0049 |
| Status | Proposed (the direction was decided by the user, decision D2; the details are confirmed when this ADR is accepted) |
| Date proposed | 2026-10-01 (session S007) |
| Date of last status change | 2026-10-01 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

A tampered release or update is a backdoor into every installation (threat T-57; the user's requirement UR-5). People who deploy the NAS must be able to check that what they downloaded is genuine, and the NAS's updater must refuse anything that is not.

**The user's decision D2 (2026-10-01, verbatim):** "Sigstore / Cosign + Minisign (Modern, Low-Overhead Code Signing). Use Cosign (Sigstore) for container images and GitHub release assets. It supports keyless signing via OIDC (linked to your GitHub workflow/identity) or standard key-based signing, avoiding the bloat of traditional GPG key management. Provide an automated checksums.txt file signed with Minisign or Cosign alongside every release binary/container so users deploying the NAS can verify file integrity and build provenance (slsa-provenance) with a single command."

**Verified on 2026-10-01:** GitHub artifact attestations are available in all public repositories and use the Sigstore public-good instance, with a public transparency log (GitHub documentation); `github.com/aead/minisign` is an MIT-licensed, pure-Go implementation of Minisign that verifies signatures, including pre-hashed ones and trusted comments (latest release v0.3.0).

## Options considered

### Option A: GPG signatures
- **Cons:** key management overhead; the user's decision chose otherwise.

### Option B: Cosign keyless plus Minisign (the user's decision, D2)
- **Cosign keyless** signs container images and release assets from the GitHub Actions release workflow: the identity is bound to the repository and workflow and recorded in the public transparency log; no long-lived key.
- **`checksums.txt`** (SHA-256 of every artifact) signed with Cosign **and** Minisign.
- **SLSA build provenance:** GitHub artifact attestations (recommended: built into GitHub, free for public repositories) or the SLSA GitHub generator.
- **SBOM:** SPDX (ISO/IEC 5962) recommended for the binaries; CycloneDX acceptable for images if the tooling favors it.
- **Reproducible builds:** `-trimpath`, the toolchain version recorded, build information embedded.
- **One-command verification:** `scripts/verify-release.sh` and `scripts/verify-release.ps1` check the Cosign signature, the checksums, and the provenance, and print pass or fail.
- **Updater (FR-381):** embeds the Minisign public key and verifies the signed update manifest offline with the small pure-Go library, avoiding the Sigstore verification stack in the core binary; then checks each file's SHA-256.
- **Minisign key protection:** used only in a protected `release` environment that needs the owner's approval; stored encrypted; backed up offline by the owner; rotation and revocation documented; the next public key announced in advance in a signed release.

## Decision

**Option B**, as the user decided (D2). To confirm on acceptance: the Cosign and Minisign tool versions and their pinned checksums; attestations through GitHub or the SLSA generator; the SBOM tool and format; the exact verification commands for Linux, macOS, and Windows.

## Consequences

- **Easier:** anyone can verify a release with one command; the updater cannot be fed a tampered update.
- **Harder:** the owner keeps the Minisign key safe and approves each release in the protected environment.
- **Required:** S14.3 (updater), S14.7 (release workflow, scripts, documentation); register entries for Cosign, Minisign, and the Go library when adopted.

## Approval record

_Direction decided by the user (D2, above). The ADR as a whole is not yet approved._

## Links

- **Related requirements:** FR-381, FR-382, NFR-078, NFR-079, NFR-061
- **Related ADRs:** ADR-0005 (CI)
- **Related stages:** S13, S14.3, S14.7
- **Plan version:** 1.12.0
