# ADR-0053: Host management through the privileged helper

| Field | Value |
|---|---|
| Number | ADR-0053 |
| Status | Proposed |
| Date proposed | 2026-10-01 (session S007, E076) |
| Date of last status change | 2026-10-01 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

A NAS OS manages its host from the web interface: network, time, OS updates, SSH, firewall, power, services (FR-388, release R14; the user's request E076). The core never runs as root (NFR-037); disk operations already go through a small privileged storage helper with an allow-list, caller authentication, and an audit log (ADR-0029, NFR-070). Host management is Tier 1 under the security program (R15): a mistake here can open the NAS to the network or lock the admin out.

## Options considered

### Option A: Run the core as root
- **Cons:** violates NFR-037; any core bug becomes a full host compromise.

### Option B: Extend the privileged helper with typed host operations (recommended)
- **Typed operations only, no shell:** set the network configuration (through the system's own tool: NetworkManager on Raspberry Pi OS since Bookworm, or systemd-networkd/ifupdown on Debian; to verify per platform), set the time zone and NTP, show and apply OS updates (unattended-upgrades status; apply security updates now), turn SSH on or off and manage authorized keys, apply the firewall rule set, restart, shut down, schedule a restart, show service status and host logs.
- **Safety:** every operation re-authenticated (`AdminRecent`) and audited; network and firewall changes use keep-or-revert (as S03.9-T05) so a wrong setting reverts unless confirmed from the new address; the helper validates every value itself (it never trusts the core).
- **Platforms:** Debian-family Linux (the appliance, Raspberry Pi OS, Debian). Not on Windows, NAS platforms (R13, the host manages itself), or Docker (no host access).

### Option C: A third-party management tool (e.g. Cockpit) next to the NAS
- **Pros:** much exists already.
- **Cons:** a second web interface with its own sign-in and attack surface; not integrated with the console.

## Decision

**Option B (recommendation while Proposed).** The allow-list and the tools per platform are verified when R14 is planned.

## Consequences

- **Easier:** one console for the whole NAS; the core stays unprivileged.
- **Harder:** a larger helper; careful tests for lock-out cases.
- **Required:** threat model entries (T-73); abuse tests for every operation (R15); NFR-070 applies (peer credentials and a token on the socket).

## Approval record

_Not yet approved._

## Links

- **Related requirements:** FR-388, FR-389, NFR-037, NFR-070, NFR-076, NFR-077
- **Related ADRs:** ADR-0029 (storage helper), ADR-0039 (admin console), ADR-0052 (appliance image)
- **Related stages:** R14, S03.9 (keep-or-revert pattern)
- **Plan version:** 1.13.0
