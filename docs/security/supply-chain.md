# Supply chain security

| Field | Value |
|---|---|
| Version | 1.0 (2026-10-01, plan change request #10) |
| Plan requirements | NFR-078 (repository and CI), NFR-079 (dependencies), FR-381 (signed updates), FR-382 (release integrity), FR-363 (dependency security review, S13) |
| Aligned with | NIST SP 800-218 (SSDF), SLSA, OpenChain (ISO/IEC 5230 and ISO/IEC 18974:2023), SPDX (ISO/IEC 5962); ISO/IEC 27001:2022 controls 5.19–5.22, 8.4, 8.19, 8.30 (based on public summaries) |

**Goal:** no backdoor gets into the code, its dependencies, the build, or a release, and anyone can check that a release is genuine.

## 1. Why it matters (recent examples, checked 2026-10-01)

- **March 2025, `tj-actions/changed-files` (CVE-2025-30066):** attackers moved the action's version tags to a malicious commit that printed CI secrets into public workflow logs, in thousands of repositories.
- **March 2026, `aquasecurity/trivy-action` and `setup-trivy`:** attackers force-pushed 76 of 77 version tags of the Trivy action to credential-stealing code for about 12 hours (March 19–20) and published a malicious Trivy release. This project's CI uses `trivy-action` by tag; its runs (September 2026) were outside that window, but the same attack would have reached it. This is why actions are pinned by commit SHA (section 3).
- **AI-suggested packages:** coding assistants sometimes suggest package names that do not exist; attackers register such names ("slopsquatting"). The coding agent of this project verifies every package before adding it (rule R15).

## 2. Dependencies (NFR-079)

Every new dependency:
1. Is **justified** (the standard library or an existing dependency cannot do it) and recorded in the dependency register in the same commit (RULES R6).
2. Is **verified as the genuine package**: the exact name, the publisher or organization, the repository it comes from, recent maintenance, and its license; for packages suggested by an AI or found by search, the name is checked against the project's own documentation.
3. Has a **compatible license** (`docs/dev/licensing.md`).
4. Is **pinned with a checksum**: `go.sum`, `pnpm-lock.yaml`, and `uv.lock` (AI worker); the tools in CI and the deployers are pinned to exact versions with checksums.
5. Is **scanned** (govulncheck, `pnpm audit`, pip-audit, Trivy) and covered by Dependabot alerts.
6. Gets a look at the **size of its dependency tree** before it is added; fewer is better.

**npm install scripts are disabled** except for an allow-list (pnpm's setting), so a package cannot run code when it is installed.

The full review of every component, with upgrades and downgrades, is stage S13 (before packaging) and repeats before every release (FR-363).

## 3. Repository and CI (NFR-078)

| Control | Status |
|---|---|
| GitHub Actions pinned to full commit SHAs (with the version in a comment), not tags | Planned (foundation task S03.5-T06) |
| Workflow token permissions set to the minimum per job (`permissions:` in every workflow) | Planned (S03.5-T06) |
| No secrets in workflows that pull requests from forks can start | Planned (S03.5-T06); CI already starts only on stage tags or by hand |
| gitleaks pre-commit and pre-push hooks, committed `.gitleaks.toml`, a one-time scan of the whole history (decision D3) | Planned (S03.5-T07) |
| The coding agent never bypasses hooks (`--no-verify`) and never weakens a check to make a build pass | In force (rule R15) |
| **Owner tasks** (account settings only the owner can change): secret scanning with push protection; Dependabot alerts and security updates; private vulnerability reporting; branch protection; two-factor authentication on the owner's GitHub account | Waiting on the owner (see the checklist below) |

**Owner checklist** (repository Settings on GitHub; all free for public repositories, checked 2026-10-01 in GitHub's documentation):
1. Settings → Code security: turn on **Secret scanning** and **Push protection**.
2. Settings → Code security: turn on **Dependabot alerts** and **Dependabot security updates**.
3. Settings → Code security: turn on **Private vulnerability reporting**.
4. Settings → Branches (rulesets): for `main`: no direct pushes (pull requests only), no force pushes, no deletion. For `develop`: **no force pushes and no deletion only**, so the agent's own merges into `develop` (the owner's standing preference) keep working.
5. Account → Password and authentication: **two-factor authentication** on (an authenticator app or a security key).
6. Later, for releases (S14.7): a protected `release` environment that needs the owner's approval.

## 4. Releases (FR-382, the owner's decision D2)

- **Reproducible Go builds:** `-trimpath`, a recorded toolchain version, build information embedded; anyone can rebuild and compare.
- **SBOM** for the binaries and images, in SPDX (ISO/IEC 5962) or CycloneDX (format chosen in ADR-0049).
- **SLSA build provenance** attestations (GitHub artifact attestations, free for public repositories and written to the public Sigstore transparency log; or the SLSA GitHub generator; ADR-0049).
- **Cosign keyless signatures** on container images and release assets, made by the release workflow (identity bound to the repository and workflow).
- **`checksums.txt`** (SHA-256 of every artifact), signed with Cosign **and** Minisign.
- **One-command verification:** `scripts/verify-release.sh` and `scripts/verify-release.ps1` print a clear pass or fail; the install guide shows the command.

## 5. Updates (FR-381)

The NAS's updater embeds the Minisign public key. It downloads the update manifest, verifies its Minisign signature offline, then checks each file's SHA-256 against the manifest, and refuses anything unsigned or mismatched. Update checks are opt-in (I6).

## 6. Signing keys

- The Minisign secret key is used only in a protected GitHub Actions environment that needs the owner's approval; it is stored encrypted, backed up offline by the owner, and has a documented rotation and revocation procedure ([incident response](incident-response.md), Part B).
- Cosign keyless needs no long-lived key; the identity is the release workflow.
- Android and Windows app signing (R07, R08; question Q86) follow the same key-protection rules.

## 7. NIST SSDF in this project (summary)

| SSDF practice group | Here |
|---|---|
| Prepare the organization (PO) | RULES (R1–R15), this policy set, roles in the security policy |
| Protect the software (PS) | Repository and CI controls (section 3); signed releases (section 4); release archives with provenance |
| Produce well-secured software (PW) | Threat model per stage, the secure coding standard, linters, reviews, the stage security gate |
| Respond to vulnerabilities (RV) | [Vulnerability management](vulnerability-management.md) |
