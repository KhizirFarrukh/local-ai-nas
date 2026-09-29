# ADR-0034: Detecting newly installed drives

| Field | Value |
|---|---|
| Number | ADR-0034 |
| Status | Proposed |
| Date proposed | 2026-09-29 (session S007) |
| Date of last status change | 2026-09-29 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- The NAS should notice a new drive without a restart where the hardware allows it (P007, FR-328), identify drives by serial number, model, and WWN (never device names), and never mount anything automatically (threat model: a malicious USB drive).

**Verified 2026-09-29:** the kernel sends device events on netlink `NETLINK_KOBJECT_UEVENT` ("kernel messages to user space"); listening to a netlink multicast group needs root or `CAP_NET_ADMIN` (netlink(7)). On Windows, top-level windows receive `WM_DEVICECHANGE` with `DBT_DEVICEARRIVAL` when devices arrive (Microsoft Learn), and disks can be listed through WMI (`Win32_DiskDrive`).

## Options considered

### Option A: The helper listens to kernel events on Linux; re-scan elsewhere
- **Pros:** instant detection on Linux without extra daemons (the helper already runs as root); no udev dependency.
- **Cons:** events need filtering (only whole disks); a re-scan is still needed after missed events.

### Option B: Through udev (udevadm monitor or libudev)
- **Pros:** enriched device properties.
- **Cons:** containers and minimal systems may lack udev; a C library or a subprocess.

### Option C: Periodic re-scan only
- **Pros:** simplest, every platform.
- **Cons:** delay.

## Decision

**Recommended: Option A on Linux, with Option C as the fallback everywhere** (a scheduled re-scan and a "Scan for drives" button in the admin console). Windows: a periodic WMI re-scan (a hidden window for `WM_DEVICECHANGE` is optional, Could). macOS: re-scan (DiskArbitration is Could).

## Consequences

- **Required:** a known-drive registry keyed by serial and WWN (S14.3); the threat model entry for hostile hot-plug devices (S03.1); detection tests with loop devices (S14.13).

## Approval record

_Pending: put to the user with the P007 report (S007 E047). Decided when its stage is planned in detail (S14.3 (extends ADR-0029))._

## Links

[ADR-0029](ADR-0029-privileged-storage-helper.md) · plan FR-202, FR-328
