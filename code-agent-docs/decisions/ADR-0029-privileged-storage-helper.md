# ADR-0029: Privileged storage helper

| Field | Value |
|---|---|
| Number | ADR-0029 |
| Status | Proposed |
| Date proposed | 2026-09-28 (session S007) |
| Date of last status change | 2026-09-28 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- Drive pools (S14) need root: listing drives with SMART, partitioning, creating md arrays, formatting, mounting, replacing members.
- The **core must never run as root** (NFR-037, RK-35). It faces the network and parses untrusted uploads.
- Deployments: native service on Linux (S13.2 setup scripts) or Docker Compose (S13.1). Pools are Linux-only (A22).
- Wrong disk operations destroy data (RK-34); every one must be deliberate, checked, and logged (I10).

## Options considered

### Option A: A separate helper service on the host
A small Go program run as root by systemd, reached over a local Unix socket, doing only allow-listed operations.
- **Pros:** the smallest possible privileged surface; operations are typed requests (never shell strings); the helper checks each request against its own drive inventory; works with native and Docker deployments (the socket is mounted into the container).
- **Cons:** a second program to install, update, and secure.
- **License / cost:** AGPL-3.0-or-later (project code).

### Option B: Run the core as root
- **Pros:** simplest.
- **Cons:** any bug in the core becomes a root compromise. Rejected (NFR-037).

### Option C: sudo rules for specific commands
- **Pros:** no new service.
- **Cons:** command arguments are hard to constrain safely in sudoers; no validation against the drive inventory; audit is only sudo's log.

### Option D: udisks2 over D-Bus with polkit rules
- **Pros:** an existing, maintained privileged daemon with drive, filesystem, and md RAID interfaces.
- **Cons:** polkit rules grant broad disk rights to the core's user; D-Bus inside containers is awkward; headless servers may not run udisks2; its md RAID interface's exact capabilities are **not verified** (to check in S14.1 if Option A is questioned).
- **License / cost:** GPL-2.0-or-later / LGPL (to verify).

## Decision

**Recommended: Option A, a separate helper service.**

- **Transport:** a Unix socket owned by root with group access for the NAS service user only (e.g. `/run/local-ai-nas/storage-helper.sock`, mode 0660); the helper also checks the peer's user ID (`SO_PEERCRED`).
- **Allow list:** discover drives, read SMART, create RAID 0/1, format (ADR-0028), mount and unmount at the configured pool path, set up monitoring, start a consistency check, replace a member, grow (where supported), stop a pool, import existing pools. Nothing else.
- **Validation:** every target drive must be in the helper's own inventory; the OS drive and drives holding mounted data (including the current storage root, outside the migration flow) are refused; destructive requests need the confirmation token issued after the user typed the phrase (FR-203).
- **Execution:** fixed command templates with arguments passed as separate strings (never through a shell); timeouts; progress events.
- **Audit:** every request, its result, and the calling user are logged by the helper (root-owned log) and mirrored to the core's security log (S03.6).
- **Installation:** an optional component of the Linux setup scripts (S13.2), added in S14.2; with Docker, the helper runs on the host.

## Consequences

- **Easier:** clear security boundary; testable allow list; the core stays unprivileged.
- **Harder:** two programs to version together (the socket protocol is versioned).
- **Required (follow-up work, constraints this imposes):** S03.1 lists the helper in the threat model; S14.2 builds it; S14.10 tests refusal of everything outside the allow list; `dependencies.md` records the host tools it runs.

## Approval record

_Pending: put to the user with the P005 report (S007)._

## Links

- **Related requirements:** FR-210, FR-203, NFR-037
- **Related ADRs:** ADR-0027 (pool approach), ADR-0028 (filesystem), ADR-0010 (security building blocks)
- **Related stages:** S03.1, S13.2, S14.2
- **Plan version:** 1.4.0 (concern 8.24)
