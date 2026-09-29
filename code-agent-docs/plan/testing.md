# Testing strategy

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](PLAN_INDEX.md)). Moved unchanged from `plan.md` section 12 in version 1.9.0 (ADR-0044).

## 12. Testing strategy

### 12.1 Per-stage testing substages

| Stage | Testing substage | Focus |
|---|---|---|
| S01 | S01.7 | Integration on a real temp filesystem; traversal and malicious-name attacks; Unicode, empty, huge (sparse), and deeply nested files; performance baseline (NFR-003 S01 targets) |
| S02 | S02.8 | Component tests; Playwright end-to-end tests of the main flows; cross-browser; accessibility (axe) |
| S03 | S03.10 | Auth bypass, CSRF, and traversal regressions; route inventory (default deny); dependency and vulnerability scanning; static analysis; threat model review ; admin console route inventory and re-authentication (NFR-050); ARM64 and Raspberry Pi profile jobs (NFR-051) |
| S04 | S04.9 | Area-separation tests over every write path; 50k-item performance; video streaming (live level switching, Auto, seek, cache eviction, segment authorization; Chrome, Edge, Firefox, Safari on macOS and iOS) |
| S05 | S05.8 | Fixture library; round-trip tests; crash-injection during sidecar writes; migration dry-run and idempotency |
| S06 | S06.8 | Golden query set; parser property tests; rebuild-equivalence; 100k + 100k benchmarks |
| S07 | S07.7 | Cross-user refusal on every endpoint and job; search leak tests (results, counts, facets, suggestions, timing) |
| S08 | S08.7 | Trash and restore; integrity detection; disaster recovery from disk plus backup |
| S09 | S09.6 | Real clients (Windows/macOS/Linux) over WebDAV/SMB; permissions over the network; watcher and rename tests |
| S10 | S10.7 | Quota enforcement on every write path ; health rules against SMART fixtures; console completeness |
| S11 | S11.8 | Labelled duplicate and look-alike fixture set (false positives, exact duplicates never wrong); burst fixtures; large-library lookups; metadata merge; shortcut edge cases; no cross-user matches |
| S12 | S12.8 | Resize arithmetic (the user's examples); visual checks of previews; metadata preservation across JPEG, HEIC, PNG, and video; orientation; byte-identical revert; fault injection mid-job |
| S14 | S14.5–S14.7 | Full-system load tests; upgrade and rollback tests; release artifact tests |
| S15 | S15.13 | Simulated disks only (loop devices or VM disks): creation, failure, degraded mode, rebuild, growth, import, reboot re-assembly; a documented manual test on real hardware; the recovery drill ; | S16 | S16.8 | Cache off and on benchmarks; consistency with external changes; fault injection (SSD removed, corrupted entry, full SSD); leak tests for items and statistics |
drive lifecycle: hot-plug, upgrade, mirror conversion (with an interrupted resync), hot replacement, growth, rollback, retirement, secure erase on virtual disks, crash injection at every migration phase |
| S17 | S17.12 | Evaluation set accuracy; CPU throughput; AI-off regression (the whole non-AI suite passes with the worker stopped) |

Each final substage above **writes the stage's tests** (S006): unit, integration, and system/application tests for everything the stage built, plus the regression tests for bugs recorded during the stage. Tests named in the deliverables or acceptance criteria of earlier substages are written and checked there. The stage's code is written first, to be testable. Coverage of at least 80% (Go: `internal/...`) is an exit criterion.

**Raspberry Pi (1.7.0, the user's requirement):** from S03 on, the stage-end CI also runs the tests on an ARM64 runner (GitHub `ubuntu-24.04-arm`) and in a Raspberry Pi resource profile (Docker CPU and memory limits matching the chosen Pi, Q75), and each stage records its memory and speed against the NFR-051 budgets. Real-Pi measurements come when the user has one.

Each final review substage above also runs a **documentation audit** (R12, `templates/audit-checklist.md`).

**Releases (1.6.0, P006).** Each release R01–R12 becomes one or more stages just in time (11b.1); its testing substage adds the release's own focus (the exit criteria in 11b, and for R09 the gates in 11c) and an **upgrade test from the previous release with real migrated data** (NFR-017).

### 12.2 Test levels (all stages)

| Level | Covers |
|---|---|
| Unit | Pure logic: resolvers, validators, parser, schema, migrations, ranking |
| Integration | Real filesystem and SQLite; HTTP-level API tests |
| System / application | The whole program used as a user uses it: the binary over HTTP (S01: `TestIntegration`, the demo scripts); from S02 the GUI in a browser (end-to-end, Playwright) |
| Security | From S01.6; formalized in S03.10 |
| Performance | At the scales named in NFR-003 |

CI runs on Linux and Windows from S01.1. _Since S007 (the user's preference):_ CI runs **at the completion of each stage**, not on every commit or push: it starts when a stage-completion tag such as `S02-done` is pushed, or by hand ("Run workflow"; the user's choice, S007 E013). Plan wording such as "on every PR" (NFR-023) or "passes in CI" means the stage-completion run.

_1.9.0 (P008):_ a shared **crash-injection harness** for journaled operations (NFR-053); **local-origin security tests** (a foreign Host refused, cross-site posts refused, the GUI and scripts still working; NFR-057); **multilingual golden queries** (Urdu, Roman Urdu, mixed script; FR-357); a **background-load responsiveness test** (NFR-054); a **hybrid versus lexical evaluation** (FR-362); the **dependency review procedure** (S13.5).

### 12.3 Fixture sets
- **Files (S01):**
  - Names: Unicode (NFC/NFD, emoji, right-to-left), reserved names, long names.
  - Nesting: deep trees.
  - Sizes: empty files, sparse multi-GB files.
- **Media (S04–S05), generated by script so git stays small:**
  - EXIF variants: complete, missing, no offset, with offset, all orientations.
  - GPS: Karachi, Lahore, southern/western hemispheres, borders, antimeridian, 0,0, missing.
  - Formats: JPEG/PNG/WebP/GIF/HEIC, MP4/MOV, corrupt and zero-byte files.
  - Naming: same basename with different extensions; Google Takeout–style foreign `.json`.
- **Scale:** a synthetic generator for 50k photos (S04.9) and 100k photos + 100k files (S06.8, S14.5). It is not committed.
- **Duplicates and look-alikes (S11), 1.4.0:** generated by script from license-clean sources: exact copies under other names; re-saved, re-encoded, resized, and format-converted copies; crops and small edits; bursts with and without camera burst identifiers (Apple `BurstUUID`, Google `GCamera:BurstID`); "similar but different" pairs as negatives; identical files in two users' libraries (privacy tests).
- **Optimization (S12), 1.4.0:** images and videos with full EXIF, XMP, IPTC, ICC, GPS, and every orientation, in each supported format, with expected dimensions for each resize mode.
- **Pools (S15), 1.4.0:** loop-device or VM-disk sets of equal and unequal sizes. No test touches a real disk (NFR-038).
- **Drives (S10.3, S15), 1.7.0:** recorded SMART outputs (smartctl JSON) of healthy, degrading, and failed drives, NVMe and SATA, and drives behind USB bridges without SMART; virtual disks of different sizes for upgrades, mirror conversion, and growth.
- **SSD cache (S16), 1.7.0:** SSD failure injection (a removed or read-only cache directory, corrupted entries, a full disk).
- **AI (S17):** a labelled, license-clean or consented evaluation set kept outside the repository, with versioned reports.

- _1.9.0:_ **Google Takeout fixtures** (R01): synthetic archives built from the analysis of the user's sample data (A28, 8.42), never the sample itself; a Takeout-style `.json` next to a photo together with its `.lainas.json` sidecar (S05).

### 12.4 AI evaluation
- **Classification:** precision and recall per label at the chosen thresholds, CPU time per image.
- **Faces:** pairwise/BCubed precision and recall, clusters versus identities, manual merges needed per 100 faces.
- A model change must match or beat the previous report before adoption.

---
