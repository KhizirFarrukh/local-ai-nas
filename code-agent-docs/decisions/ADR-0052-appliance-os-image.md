# ADR-0052: The appliance OS image ("local-ai-nas OS")

| Field | Value |
|---|---|
| Number | ADR-0052 |
| Status | Proposed |
| Date proposed | 2026-10-01 (session S007, E076) |
| Date of last status change | 2026-10-01 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

The user asked for "NAS OS" support (E076). One reading, decided by the agent under the user's delegation, is an appliance: an image people flash and switch on, which becomes a working NAS without a terminal (release R14, FR-391). The production machine is a Raspberry Pi (NFR-051); the deployers (S14.2) already install and harden the NAS on Raspberry Pi OS and Debian (NFR-073, NFR-077).

**Verified on 2026-10-01** (research R004): Raspberry Pi builds Raspberry Pi OS with pi-gen and published rpi-image-gen (2025) for custom images, with profiles, image layouts, and SBOM handling.

## Options considered

### Option A: Our own Linux distribution
- **Cons:** a large, permanent maintenance burden (kernel, security updates, hardware support); excluded (NG12).

### Option B: Raspberry Pi OS Lite (and Debian for x86-64) with the NAS preinstalled (recommended)
- **Build:** rpi-image-gen (or pi-gen) for the Pi image; a Debian-based installer image for x86-64 later in R14; both run the same steps as the deployers, so a flashed system equals a deployed one.
- **First boot:** grows the filesystem, generates secrets and the TLS certificate (NFR-074, NFR-067), then shows the address and a **one-time setup code** on an attached screen and writes it to the boot partition; the setup page on the LAN accepts only that code, which expires after setup (otherwise the S03 rule stands: setup only on the machine itself or with the command line). No default passwords; SSH off unless the Raspberry Pi Imager settings turn it on.
- **Updates:** the distribution's unattended security updates; the NAS through its signed updates (FR-381); no A/B system images in the first version.
- **Signing:** images are release artifacts, signed and verifiable like every release (FR-382).

### Option C: Only documentation ("install Raspberry Pi OS, then run the deployer")
- **Cons:** does not give the user a NAS OS; the deployer already covers this path.

## Decision

**Option B (recommendation while Proposed).** To decide when R14 is planned: rpi-image-gen or pi-gen (licenses and maintenance to verify), the x86-64 installer approach, and how the setup code is shown on headless boards (the boot partition and, optionally, an LED pattern or the router's device name).

## Consequences

- **Easier:** a NAS without a terminal; the same hardening as the deployers.
- **Harder:** images to build, test on real Pi boards, and keep patched for every release (RK-64).
- **Required:** threat model entries (T-74); a Raspberry Pi test device in R14; register entries for the build tool.

## Approval record

_Not yet approved._

## Links

- **Related requirements:** FR-391, FR-388, FR-149, FR-381, FR-382, NFR-073, NFR-077
- **Related ADRs:** ADR-0006 (development environment and packaging), ADR-0049 (release signing), ADR-0050 (service account), ADR-0053 (host management)
- **Related stages:** R14, S14.2
- **Plan version:** 1.13.0
