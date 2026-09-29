# ADR-0035: Secure erase of retired drives

| Field | Value |
|---|---|
| Number | ADR-0035 |
| Status | Proposed |
| Date proposed | 2026-09-29 (session S007) |
| Date of last status change | 2026-09-29 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- Drive retirement offers "securely erase before disposal or sale" (P007, FR-336). It cannot be undone (I10) and must never touch the wrong drive.

**Verified 2026-09-29:** hdparm's `--security-erase` and `--security-erase-enhanced` are marked "DANGEROUS" in hdparm(8); ATA security commands fail on drives in the "frozen" state (commonly set by the BIOS) and are often blocked by USB bridges. For NVMe, `nvme sanitize` (block erase, overwrite, or crypto erase) works on the whole device, while `nvme format --ses=1` (user data erase) works per namespace (nvme-cli documentation and guides). **Unverified:** support detection details per drive, to check at S14.11.

## Options considered

### Option A: Firmware erase where supported, overwrite otherwise
- NVMe: sanitize (crypto erase if supported, else block erase); else format with `--ses=1`.
- SATA: ATA security erase when the drive is not frozen and not behind a blocking USB bridge.
- Otherwise: a full overwrite, with a verification read of samples.
- **Pros:** the fastest and most thorough method each drive supports.
- **Cons:** detection logic per drive type; frozen drives need a suspend-resume or re-plug, which is explained, never forced.

### Option B: Overwrite only
- **Pros:** one method.
- **Cons:** slow (many hours on large drives); on SSDs an overwrite does not reach over-provisioned space.

## Decision

**Recommended: Option A.** The console shows the method, the estimated time, the drive's model and serial, and requires the typed confirmation (FR-203); a completion record with the serial and method is kept. Could: blinking the drive's LED (ledctl) to confirm which one it is.

## Consequences

- **Required:** helper operations for sanitize, format, security erase, and overwrite, each refusing the OS drive and any drive holding NAS data (S14.2); tests on virtual disks only (NFR-038).

## Approval record

_Pending: put to the user with the P007 report (S007 E047). Decided when its stage is planned in detail (S14.11)._

## Links

[ADR-0029](ADR-0029-privileged-storage-helper.md) · plan FR-203, FR-336
