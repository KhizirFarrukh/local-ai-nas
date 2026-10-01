# ADR-0050: Service account, file permissions, and mount options per platform

| Field | Value |
|---|---|
| Number | ADR-0050 |
| Status | Proposed |
| Date proposed | 2026-10-01 (session S007) |
| Date of last status change | 2026-10-01 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

The NAS must never let stored content run on the host (invariant I12), and other users or programs on the host must not read its data, keys, or database (threats T-42, T-43; the user's requirements UR-4 and UR-5; NFR-063, NFR-073, NFR-077). Today log files are 0600; the S03.5-T05 task already plans `.local-ai-nas` 0700 and the database files 0600 (bug S03-B02). The deployers (S14.2) will create the service account.

## Options considered

### Option A: Run as the installing user with default permissions
- **Cons:** every program of that user can read the data; files can be executable.

### Option B: A dedicated account and strict modes (recommended)
- **Account:** Linux: a system user `local-ai-nas` with no login shell, owning the storage root; Windows: a virtual service account (`NT SERVICE\local-ai-nas`), not LocalSystem; Docker: a fixed non-root UID.
- **Modes:** umask 027; user files 0640 and folders 0750 (never an execute bit; I12); internal data 0700; the configuration and secret files 0600; Windows: equivalent ACLs (the service account and Administrators only).
- **Startup check:** warn when the storage root, internal data, or configuration are readable or writable by others; **refuse to start** when a secret file (TLS key, secrets key) is readable by others.
- **Mounts:** where the NAS controls the mount (its own drives and pools, S15.5): `noexec,nosuid,nodev`.
- **systemd hardening** (NFR-077) and the Windows firewall rule for Private networks are set by the deployers.

## Decision

**Option B (recommendation while Proposed).** The modes and the startup check come first (S03.5-T05, proposed for approval); the account and the service units with the deployers (S14.2); the mounts with the pools (S15.5).

## Consequences

- **Easier:** a compromised other program on the host cannot read the data; nothing stored can be executed from the NAS's own drives.
- **Harder:** files copied in by other users of the host need the right group; the guide explains it.
- **Required:** tests of the modes on Linux; an ACL check on Windows; the deployers' scripts.

## Approval record

_Not yet approved._

## Links

- **Related requirements:** NFR-073, NFR-063, NFR-077, NFR-037, FR-149
- **Related ADRs:** ADR-0003 (storage layout), ADR-0029 (storage helper)
- **Related stages:** S03.5, S14.2, S15.5
- **Plan version:** 1.12.0
