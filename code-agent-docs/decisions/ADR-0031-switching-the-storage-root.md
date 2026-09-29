# ADR-0031: Switching the storage root after a migration

| Field | Value |
|---|---|
| Number | ADR-0031 |
| Status | Proposed |
| Date proposed | 2026-09-29 (session S007) |
| Date of last status change | 2026-09-29 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- After a migration (ADR-0030) the NAS must start using the new drive, and the old one must stay untouched for the rollback window.
- With Docker (ADR-0006), the storage root is a bind mount into the container; a new path would need a changed Compose file.
- On Linux the storage helper (ADR-0029) can mount drives; on Windows and macOS the user prepares and mounts the target.

## Options considered

### Option A: A stable mount point, remounted by the helper (Linux)
The storage root always lives at the same path (e.g. `/srv/local-ai-nas`). The switch unmounts the old drive's filesystem there and mounts the new one, and mounts the old drive at a rollback path.
- **Pros:** configuration and Docker bind mounts stay unchanged; rollback is the reverse remount.
- **Cons:** needs the helper; the service (or container) must be stopped during the swap, because open files keep the old mount busy; mount propagation into a running container is not relied on.

### Option B: Change the configured path and restart
- **Pros:** works everywhere, no helper.
- **Cons:** Docker Compose files must change; users of the old path (scripts, other programs) break.

## Decision

**Recommended: Option A on Linux with the helper; Option B elsewhere** (and on Linux without the helper). The switch happens inside the maintenance window: stop the service or container, remount (A) or rewrite the configuration (B), start, and run the post-switch check. The container is **restarted**, not relied on to see a mount change live, so no special mount propagation is needed. To confirm in S15.6 with Docker's documentation on bind mounts.

## Consequences

- **Easier:** Docker users never edit their Compose file after a drive change.
- **Required:** the setup scripts (S14.2) install the NAS with the stable mount point on Linux; the switch and rollback are tested with Docker and natively (S15.13).

## Approval record

_Pending: put to the user with the P007 report (S007 E047). Decided when its stage is planned in detail (S15.6)._

## Links

[ADR-0030](ADR-0030-storage-migration-engine.md) · [ADR-0029](ADR-0029-privileged-storage-helper.md) · [ADR-0006](ADR-0006-dev-environment-and-packaging.md)
