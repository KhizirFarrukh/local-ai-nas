# S16: SSD caching (new in 1.7.0, P007)

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](../PLAN_INDEX.md)). Moved unchanged from `plan.md` section 10.15a in version 1.9.0 (ADR-0044).

### 10.15a S16: SSD caching (new in 1.7.0, P007)

- **Origin:** User-defined (P007).
- **User requirement (quoted, P007):** "one more thing to add, caching into ssd of most used (typically large) files/photos, if configured."
- **Goal:** If configured, the most-used and large files and photos are served from an SSD, and internal data can live on fast storage, with no risk to data if the SSD fails.
- **Placement:** directly after the drive stage (S15), before AI (I8), in milestone M4 "Drives and storage". It reuses disk health (S10.3), the helper for preparing an SSD on Linux (S15.2), and the new-drive wizard (S15.9). The application-level cache works on every platform; on a Raspberry Pi 5 an NVMe SSD makes the biggest difference (8.33).
- **Depends on:** S15, S10.3, S04.4, S04.8, S06, S08.3.
- **Write policy:** write-through only; write-back is a not-scheduled candidate (11a).
- **Status:** Not started

#### S16.1: Architecture and decisions
- **Goal:** The cache is designed and its decisions are made.
- **Scope:** the ADRs for the read cache (ADR-0036: admission and eviction), fast internal data (ADR-0037), and the block-level cache (ADR-0038, Q70); the threat model entries (cache files never exposed; privacy of statistics); Raspberry Pi measurements of an NVMe SSD against USB hard drives.
- **Deliverables:** ADR decisions; threat model entries; a benchmark note.
- **Depends on:** S15 (Done), S10.3, S08.3.
- **Requirements:** FR-337, FR-340, NFR-046, NFR-047.
- **Acceptance criteria:**
  1. The ADRs are decided with the user.
  2. The threat model lists the cache and its failure modes.
- **Risks/notes:** RK-44, RK-45.
- **Status:** Not started

#### S16.2: Fast internal-data placement
- **Goal:** Busy internal data can live on an SSD safely.
- **Scope:** moving the index, thumbnails, and transcode cache (and optionally the database, Q71) to an SSD with the migration engine's copy, verify, and switch steps in a short maintenance window; trash, temp uploads, and replaced originals stay on the storage root (A18); SSD failure rebuilds derived data and restores the database from the metadata backup.
- **Deliverables:** move wizard; failure handling.
- **Depends on:** S16.1, S08.3.
- **Requirements:** FR-338.
- **Acceptance criteria:**
  1. Moving and moving back work with every item verified.
  2. Removing the SSD rebuilds derived data and never loses a file.
- **Risks/notes:** RK-44, RK-45.
- **Status:** Not started

#### S16.3: Cache store
- **Goal:** A safe, content-addressed store on the SSD.
- **Scope:** entries named by content hash in NAS-owned directories; temp-write, sync, verify, publish; a scheduled scrub; a size budget and free-space reserve; bypass on failure; cleanup of deleted content; excluded from backups and quotas.
- **Deliverables:** cache store.
- **Depends on:** S16.1.
- **Requirements:** FR-337, NFR-046.
- **Acceptance criteria:**
  1. A corrupted entry is detected and never served.
  2. A full or missing SSD switches the cache to bypass without failed requests.
- **Risks/notes:** RK-44, RK-45.
- **Status:** Not started

#### S16.4: Admission, eviction, pinning, and prewarming
- **Goal:** The right files are on the SSD.
- **Scope:** admission after N reads in a window and a minimum size (Q72), with area and type filters and a frequency filter; eviction by recency and frequency (ADR-0036); per-user private pins within a pin budget; prewarm rules at quiet hours; the daily write budget.
- **Deliverables:** policies with tests.
- **Depends on:** S16.3.
- **Requirements:** FR-337, FR-339, NFR-048.
- **Acceptance criteria:**
  1. Popular large files stay cached while one-off large reads do not push them out.
  2. Pins are private and never evicted within the budget; admission pauses when the write budget is used up.
- **Risks/notes:** RK-44, RK-45.
- **Status:** Not started

#### S16.5: Read-path integration and consistency
- **Goal:** Every read benefits, and none is ever wrong.
- **Scope:** downloads (with ranges), photo and video viewing, streaming of originals and transcoding input, previews, WebDAV reads, and exports go through the cache after authorization (I5); an entry is served only when its hash, size, and modification time match; writes always go to the storage root.
- **Deliverables:** cache hooks in the read paths.
- **Depends on:** S16.3, S16.4.
- **Requirements:** FR-337, NFR-046, NFR-047.
- **Acceptance criteria:**
  1. A file changed outside the app is served fresh, never from a stale entry.
  2. Authorization tests prove no user reads another user's cached item.
- **Risks/notes:** RK-44, RK-45.
- **Status:** Not started

#### S16.6: Endurance and power
- **Goal:** The SSD lasts, and drives rest.
- **Scope:** SSD wear monitoring (reused health evaluator), TRIM, and the optional hard-drive spin-down (Could, Q73) with a warning about wear from frequent spin-ups.
- **Deliverables:** wear alerts; TRIM schedule; spin-down option.
- **Depends on:** S16.3, S10.3.
- **Requirements:** FR-341, NFR-048.
- **Acceptance criteria:**
  1. Wear near end of life raises an alert.
  2. Spin-down, if approved, respects the configured idle time.
- **Risks/notes:** RK-44, RK-45.
- **Status:** Not started

#### S16.7: Cache GUI and metrics
- **Goal:** The admin configures and understands the cache in the console.
- **Scope:** the console's cache settings (enable, choose the SSD, budgets, thresholds, filters, pins, prewarm, write budget, clear and bypass) with a setup benchmark that warns if the SSD is not faster; privacy-safe metrics (hit rate, bytes served, space, write budget, estimated time saved); per-user pin status; "Use as SSD cache" in the new-drive wizard.
- **Deliverables:** console pages; metrics.
- **Depends on:** S16.2–S16.6, S15.9.
- **Requirements:** FR-337, FR-338, FR-339, NFR-047.
- **Acceptance criteria:**
  1. Changes take effect without a restart.
  2. Statistics never list another user's file names.
- **Risks/notes:** RK-44, RK-45.
- **Status:** Not started

#### S16.8: Testing, performance, and stage review
- **Goal:** Prove the cache is fast and harmless, then close the stage.
- **Scope:** hit-rate and speed benchmarks with the cache off and on (on the Raspberry Pi profile and, when available, the real Pi); consistency tests with external changes; fault injection (SSD removed mid-read, corrupted cache file, full SSD); leak tests for items and statistics; the R12 documentation audit; completion record; user sign-off.
- **Deliverables:** test suites; benchmark report; audit; completion record.
- **Depends on:** S16.1–S16.7.
- **Requirements:** NFR-046, NFR-047, NFR-048, NFR-049.
- **Acceptance criteria:**
  1. Repeated reads of large files come from the SSD measurably faster on reference hardware.
  2. Removing or corrupting the SSD never causes a wrong read, a lost file, or downtime.
  3. The completion record is written and the user's sign-off is recorded.
- **Risks/notes:** RK-44, RK-45.
- **Status:** Not started

**Design notes (S16):** The cache is never a source of truth: everything in it can be deleted without loss (I2). Reads are authorized before the cache is consulted (I5). Only write-through.

**Exit criteria (P007):** With the cache configured, repeated reads of large files come from the SSD measurably faster on reference hardware. Removing or corrupting the SSD never causes a wrong read, a lost file, or downtime.

---
