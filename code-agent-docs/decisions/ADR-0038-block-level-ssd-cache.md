# ADR-0038: Block-level SSD cache (lvmcache) and an optional LVM layer

| Field | Value |
|---|---|
| Number | ADR-0038 |
| Status | Proposed |
| Date proposed | 2026-09-29 (session S007) |
| Date of last status change | 2026-09-29 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- A block-level cache speeds up everything, including metadata and small random reads (P007 tier 3, FR-340, Could).

**Verified 2026-09-29 (lvmcache(7)):** a cache attaches to an **existing LVM logical volume** (`lvconvert --type cache --cachepool` or `--cachevol`, or `--type writecache`); writethrough is the default mode ("any data written will be stored both in the cache and on the origin LV"); writeback "delays writing data blocks from the cache back to the origin LV"; removal with `--uncache` or `--splitcache`. So the pool's filesystem must sit on an LVM logical volume **from creation**.

## Options considered

### Option A: Create pools with an optional LVM layer now; offer lvmcache later (writethrough only)
- **Pros:** a block cache can be added without recreating the pool.
- **Cons:** one more layer in every pool (complexity, a little overhead), decided at S14.5.

### Option B: No block-level cache
- **Pros:** simpler pools.
- **Cons:** only the file read cache (ADR-0036) is available.

## Decision

**Open: asked in Q70.** Recommendation: **Option B for the first version**, because the application-level cache covers the user's request (large, frequently used files) on every platform; revisit if metadata-heavy workloads are slow on the reference Pi.

## Consequences

- If Option A is chosen: ADR-0028's pool layout adds LVM, and S14.5 creates it; tests on loop devices.

## Approval record

_Pending: put to the user with the P007 report (S007 E047). Decided when its stage is planned in detail (S15.1 (affects S14.5 and ADR-0028))._

## Links

[ADR-0028](ADR-0028-pool-filesystem.md) · [ADR-0036](ADR-0036-ssd-read-cache.md) · plan FR-340, Q70
