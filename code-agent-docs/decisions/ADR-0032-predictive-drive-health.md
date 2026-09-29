# ADR-0032: Predictive drive health rules

| Field | Value |
|---|---|
| Number | ADR-0032 |
| Status | Proposed |
| Date proposed | 2026-09-29 (session S007) |
| Date of last status change | 2026-09-29 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- The user asked for replacing a drive that is "failing or seeming to fail" (P007). Most home NAS data loss is a single failing drive. The MVP turns raw SMART data into four statuses: Healthy, Watch, Replace soon, Replace now (FR-332).
- Platforms: Linux (smartmontools), Windows (the storage reliability counters of Windows, or smartctl for Windows), Raspberry Pi drives often behind **USB bridges**, which may not pass SMART.

**Verified 2026-09-29:** smartctl runs short and long self-tests (`-t short`, `-t long`), including on NVMe; reports NVMe health from the SMART/Health log (critical warning byte, percentage used, media and data-integrity errors); outputs JSON (`-j`); supports USB bridges through `-d sat` and bridge-specific types, and notes that **some USB bridges do not pass SMART** (Debian's smartctl(8) manual). Windows' `Get-StorageReliabilityCounter` reports temperature, read and write errors (corrected, uncorrected), wear, power-on hours, and start-stop and load cycles (Microsoft Learn).

## Options considered

### Option A: Rules over SMART attributes with trends, in the core
Read `smartctl -j` (through the helper on Linux; directly where allowed), store readings, and apply documented rules.
- **Pros:** explainable statuses; trend detection (growth, not only thresholds); the same rules on every platform.
- **Cons:** rule tuning; vendor differences in attribute meanings.

### Option B: Only the drive's own overall health verdict
- **Pros:** simple.
- **Cons:** the verdict usually flips to failing too late; misses growing reallocated or pending sectors.

### Option C: A statistical failure model
- **Pros:** can be more accurate on large fleets.
- **Cons:** needs training data; not explainable; overkill for a home NAS.

## Decision

**Recommended: Option A.** Proposed rules (tuned in S10.3 with recorded SMART fixtures, NFR-038):

- **Replace now:** overall health failed; a failed short self-test; pending or offline-uncorrectable sectors growing within a day; NVMe critical warning set (spare below threshold, media read-only, reliability degraded); read errors seen by the NAS itself.
- **Replace soon:** reallocated sectors growing over a week; any pending sectors; a failed extended self-test; NVMe percentage used ≥ 90% or media errors growing; reported uncorrectable errors (187) growing.
- **Watch:** a small, stable number of reallocated sectors; interface CRC errors (reported as a likely cable or USB-bridge problem); temperature above the drive's limit; NVMe percentage used ≥ 80%.
- **Healthy:** none of the above. **Health data not available:** SMART cannot be read (USB bridge, container, virtual disk), with the reason; never shown as Healthy (RK-24).
- **Self-tests:** short weekly, extended monthly, at quiet hours (configurable in the console).

## Consequences

- **Required:** the fixtures of healthy, degrading, and failed drives (12.3); alerts through the MVP channels (FR-221); the helper operation "run self-test" on Linux (S14.2 extends it).

## Approval record

_Pending: put to the user with the P007 report (S007 E047). Decided when its stage is planned in detail (S10.3 (MVP-A))._

## Links

[P007](../prompts/P007-drive-lifecycle-and-ssd-cache.json) · plan FR-129, FR-221, FR-332 · RK-24
