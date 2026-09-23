# ADR-0006: Development environment and packaging: Docker Compose, multi-architecture

| Field | Value |
|---|---|
| Number | ADR-0006 |
| Status | **Accepted** (native Windows and macOS installs **deferred**, pending user decision) |
| Date proposed | 2026-09-24 (session S003) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none |
| Superseded by | none |

## Context

- NFR-008: one-command deployment. NFR-009 and NFR-030: linux/amd64 and linux/arm64.
- S01.1 needs a development environment. The developer works on Windows 11 (A16).
- The core relies on external tools (ExifTool, libvips with libheif, FFmpeg; ADR-0012) from S04/S05 onward. S01 needs none of them.
- The AI worker is optional (I7).

## Options considered

| Option | Pros | Cons |
|---|---|---|
| **Docker Compose primary** (chosen) | Reproducible. Bundles the external tools. Runs on common NAS OSes. The AI worker is a Compose profile | Windows and macOS hosts need Docker Desktop |
| Native install primary | No Docker | Every host must install ExifTool, libvips, libheif, FFmpeg, and Perl |
| Single self-contained binary including the tools | Simplest install | Impossible without cgo or static bundling of GPL/LGPL tools; license complexity |

## Decision

- **Primary deployment: Docker Compose** with **multi-architecture images for linux/amd64 and linux/arm64** (built with Docker buildx).
- **Core image** bundles ExifTool (with Perl), libvips with libheif, and FFmpeg. The Go binary is built with `CGO_ENABLED=0`.
- **AI worker:** a separate image, enabled with a Compose profile: `docker compose --profile ai up` (ADR-0017).
- **Secondary: native Linux install**, meaning the static binary plus a **systemd unit**. The external tools are installed from distribution packages, and the install guide covers them (S11.2, S11.4).
- **Deferred:** native **Windows and macOS** installs, pending the user's decision (Q5).
- **Development environment (S01.1):**
  - (a) Native: Go toolchain on Windows or Linux. This is enough for S01, which needs no external tools.
  - (b) `deploy/compose.dev.yaml` running the core in a container with the source bind-mounted, and a separate storage-root volume. The dev image already includes the external tools, so later stages need no host installs.

### Implementation details chosen by agent
| Detail | Choice | Reason |
|---|---|---|
| Base image | **`debian:trixie-slim`** (Debian 13, stable) for the core runtime image, and the official `golang` image as the build stage | glibc; all tools packaged for amd64 and arm64; security updates. Verified packaged versions in trixie: ExifTool 13.25, vips 8.16.1, FFmpeg 7.1.5, libheif 1.19.8. These are older than upstream (13.59 / 8.18.6 / 9.0.2 / 1.23.5) but security-maintained |
| HEIC plugins | Install only the libheif **decoder** plugins (libde265, LGPL). Do **not** install the x265 **encoder** plugin (GPL), because the NAS never encodes HEIC | Avoids a GPL component that is not needed |
| Image pinning | Base images pinned by digest. Package versions recorded in `docs/third-party-notices.md` at build time | "Pin every version" (P003) |
| Image scanning | Trivy on every image build (ADR-0005) | NFR-023 |

## Consequences

- **Easier:** consistent environments; tools bundled; arm64 support; the optional AI profile.
- **Harder:** multi-arch builds in CI (QEMU emulation for arm64 is slower); keeping Debian tool versions current.
- **Required:**
  - Third-party notices for bundled tools, including FFmpeg's build configuration (GPL or LGPL, checked in S11.1).
  - The install guide for native Linux (S11.4) must cover the tools and `fs.inotify.max_user_watches` (ADR-0016).

## Approval record

> Decision made by the user in plan change request #3 (`code-agent-docs/prompts/P003-technology-stack.json`, `technology_stack.packaging`: "Docker Compose as the primary deployment method … Native Windows and macOS installs, pending user decision")
> (2026-09-24, session S003). Debian package versions verified via sources.debian.org (S003 log E007).

## Links

- **Related requirements:** NFR-008, NFR-009, NFR-030, FR-131, FR-132
- **Related ADRs:** ADR-0001, ADR-0004, ADR-0005, ADR-0012, ADR-0016, ADR-0017
- **Related stages:** S01.1 (dev environment), S11.1, S11.2, S12.11
- **Plan version:** 0.3.0
