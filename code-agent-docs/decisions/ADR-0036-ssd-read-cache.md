# ADR-0036: SSD read cache of original files

| Field | Value |
|---|---|
| Number | ADR-0036 |
| Status | Proposed |
| Date proposed | 2026-09-29 (session S007) |
| Date of last status change | 2026-09-29 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- The user asked for "caching into ssd of most used (typically large) files/photos, if configured" (P007, FR-337). Off by default; a dead SSD must never lose data or cause a wrong read (NFR-046); users never see each other's cached items or statistics (NFR-047).
- On a **Raspberry Pi 5** an NVMe SSD on the PCIe connector is much faster than hard drives on USB, so the cache is most useful there (S007 E046).

## Options considered

### Option A: An application-level, content-addressed read cache
Entries keyed by content hash (FR-211); admission by a frequency filter; eviction by recency and frequency; write-through only.
- **Pros:** every platform; no reformatting; authorization happens before any cache read; identical content cached once; a moved file keeps its entry.
- **Cons:** covers file reads only (not database or metadata I/O, which fast internal data covers, ADR-0037).

### Option B: A block-level cache under the filesystem (lvmcache or dm-cache; Linux)
- **Pros:** speeds up everything.
- **Cons:** needs an LVM layer from pool creation; Linux only (ADR-0038 decides whether to offer it).

### Option C: bcache
- **Cons:** requires reformatting the cached drive. Not recommended.

## Decision

**Recommended: Option A** with:
- **Admission:** a TinyLFU-style frequency sketch; a file is admitted after N reads in a window (proposed 3 in 7 days, Q72) and at least a minimum size (proposed 8 MB), within area and type filters; copies run in the background, throttled.
- **Eviction:** W-TinyLFU (a small recency window in front of a frequency-protected main area), within a size budget and a free-space reserve; pinned items (per user, private) are not evicted within the pin budget.
- **Consistency:** served only when the content hash, size, and modification time match the file's current record; temp-write, sync, verify, then publish; a scheduled scrub.
- **Wear:** a daily write budget; admission pauses when it is used up.
- **Write policy:** write-through only; write-back is a not-scheduled candidate (11a).

## Consequences

- **Required:** fault-injection and leak tests (S16.8); privacy-safe statistics (aggregated only); the cache excluded from backups and quotas.

## Approval record

_Pending: put to the user with the P007 report (S007 E047). Decided when its stage is planned in detail (S16.1)._

## Links

[P007](../prompts/P007-drive-lifecycle-and-ssd-cache.json) · [ADR-0021](ADR-0021-content-hash-algorithm.md) · plan FR-337, FR-339, NFR-046, NFR-047, NFR-048
