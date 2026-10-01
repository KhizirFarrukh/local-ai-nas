# S05: Media metadata

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](../PLAN_INDEX.md)). Moved unchanged from `plan.md` section 10.6 in version 1.9.0 (ADR-0044).

### 10.6 S05: Media metadata

- **Origin:** User-defined
- **Goal:** Every photo has a sidecar JSON file that is the source of truth for its metadata.
- **User requirements (quoted):**
  > "Stage 5 is media metadata work."
  > "From the README: each photo's own metadata is stored in a JSON file alongside the photo."
- **Status:** Not started
- **Security (1.12.0, P010; rule R15):** S05.3: ExifTool only through the sandbox wrapper with core-generated names (its `-stay_open` mode reads arguments line by line), version 12.38 or newer (NFR-064, NFR-065); **Tier 1**. S05.8: the hostile media corpus and sidecar JSON fuzzing.

#### S05.1: Sidecar schema v1
- **Goal:** A stable, versioned, forward-compatible sidecar format.
- **Scope:**
  - Proposed ADR for the JSON schema; naming convention (`IMG_0001.jpg.lainas.json` since 1.9.0, D1); `schemaVersion`.
  - Reserved sections for ownership and access (S07) and AI results (S17), so later stages need no breaking changes.
  - A machine-readable JSON Schema file for validation; an identifying marker (Q11).
- **Deliverables:** schema ADR; `schema/sidecar/v1.json`; typed models; example sidecars.
- **Depends on:** S04.1.
- **Requirements:** FR-023, FR-024, FR-100, NFR-007.
- **Acceptance criteria:**
  1. The JSON Schema validates the examples and rejects malformed sidecars.
  2. The `access` and `ai` sections are defined, so S07 and S17 need no `schemaVersion` bump for their base data.
  3. Every sidecar carries `schemaVersion` and the identifying marker.
- **Risks/notes:** Schema churn (RK-14): review carefully before acceptance.
- **P005 change (1.4.0):** schema v1 also reserves the sections `hashes` (content and perceptual hash), `stack` (stack ID, cover flag, user-locked decisions), `duplicates` (the "not duplicate of" list), `mergedFrom` (merged-metadata provenance), and `optimizationHistory` (8.7). Face boxes in the `ai` section use normalized coordinates (0–1), so resizing in S12 keeps them valid.
- **P006 addition (1.6.0, [Planner addition], pending Q54):** A reserved `livePhoto` section in the sidecar schema: the motion component's file name, content hash, and source type (FR-218).
- **P008 (1.9.0):** the sidecar name is `<full original filename>.lainas.json` (FR-356, D1); v1 reserves `mediaId` (FR-348) and `provenance` (FR-222); `access` is a non-authoritative mirror (FR-359, D2); fixtures where a Takeout-style `.json` and our sidecar sit side by side.
- **Status:** Not started

#### S05.2: Sidecar manager
- **Goal:** The single, safe writer of sidecars.
- **Scope:** create, read, update, delete; atomic writes (temp file then rename); locking for concurrent writers; validation against the schema; recovery from corrupt sidecars; preservation of unknown fields; foreign JSON detection.
- **Deliverables:** sidecar manager module; quarantine area in internal data; recovery routine.
- **Depends on:** S05.1.
- **Requirements:** FR-024, FR-025, FR-030, FR-101, NFR-006.
- **Acceptance criteria:**
  1. Killing the process during writes never leaves a partial or empty sidecar (fault injection).
  2. Concurrent updates to one sidecar never lose an update.
  3. A corrupt sidecar is detected, quarantined, reported, and rebuilt from the media without losing recoverable user fields.
  4. Unknown fields survive read-modify-write, and foreign JSON is never overwritten.
- **Risks/notes:** Windows file locking needs retries (8.4).
- **Status:** Not started

#### S05.3: Metadata extraction
- **Goal:** Accurate metadata for every media item.
- **Scope:**
  - EXIF, XMP, and IPTC for images: date taken with timezone handling, camera, lens, dimensions, orientation, GPS.
  - Video metadata: duration, codec, creation date, GPS where present.
  - HEIC and RAW if supported per S04.1; date fallback with its source recorded.
- **Deliverables:** extraction job; ExifTool in `-stay_open` mode via a custom Go wrapper, and ffprobe for video (ADR-0012); fixture expectations.
- **Depends on:** S05.2, S04.3.
- **Requirements:** FR-010, FR-011, FR-018, FR-019, FR-020, FR-098.
- **Acceptance criteria:**
  1. Extraction runs as a job for every ingested item and writes the results to the sidecar.
  2. Fixture expectations are met (dates with and without offsets, GPS, orientation).
  3. Items without a capture date get a fallback date with its source recorded.
  4. Unreadable metadata is recorded as such without failing the item.
- **Risks/notes:** External tool licenses and availability on native installs (ADR-0012, plan 8.20).
- **P005 change (1.4.0, the user's burst request in S007):** extraction also records burst identifiers where present (Apple `BurstUUID` in the maker notes; Google `XMP-GCamera:BurstID` and `BurstPrimary`) and sub-second capture times, for burst grouping (FR-165, S11.4).
- **P006 addition (1.6.0, [Planner addition], pending Q54):** Extract the content identifiers and the motion-photo markers with ExifTool (tags verified in research R001) (FR-217).
- **Status:** Not started

#### S05.4: Offline reverse geocoding
- **Goal:** Place names from GPS, fully offline.
- **Scope:** converting GPS coordinates to place names (city, region, country) using a bundled offline dataset with no network calls. Dataset and granularity decided in 0.3.0: GeoNames `cities500` + admin1/country tables, custom k-d tree (ADR-0013).
- **Deliverables:** bundled `cities500` dataset with attribution (docs + About page); geocoding job; dataset version stored in sidecars.
- **Depends on:** S05.3.
- **Requirements:** FR-062, NFR-001.
- **Acceptance criteria:**
  1. Items with GPS get city, region, and country in the sidecar, with the dataset version.
  2. Geocoding makes no network calls (tested with the network disabled).
  3. Fixture coordinates resolve as expected, including edge cases (borders, sea, 0,0).
- **Risks/notes:** GeoNames is CC BY 4.0, so attribution is required.
- **Status:** Not started

#### S05.5: User-editable metadata
- **Goal:** Users can describe and correct their photos.
- **Scope:** description, user tags, date and time correction, location correction, all written to the sidecar; an info panel and editing UI in the photos GUI.
- **Deliverables:** edit API; info panel; edit forms.
- **Depends on:** S05.2, S04.7.
- **Requirements:** FR-015, FR-099, FR-014.
- **Acceptance criteria:**
  1. Edits persist to the sidecar and survive re-extraction.
  2. A corrected date or location replaces the extracted one in the timeline and place fields, while the original values are kept.
  3. The info panel shows extracted and edited metadata, and editing works on phone layouts.
- **Risks/notes:** None.
- **Status:** Not started

#### S05.6: Sidecar lifecycle
- **Goal:** Sidecars stay correct through every operation.
- **Scope:** the sidecar follows its media on rename, move, copy, delete, and cross-area transfer (what happens when a photo moves to files is an open question: Q27); the timeline switches to date taken.
- **Deliverables:** lifecycle hooks on photos operations and transfer; the timeline order switch.
- **Depends on:** S05.2, S04.6.
- **Requirements:** FR-026, FR-103, FR-013.
- **Acceptance criteria:**
  1. After every app operation on a photo, its sidecar has the matching name and location (a test per operation).
  2. Cross-area transfer applies the rule chosen in Q27.
  3. The timeline orders by date taken (with the FR-011 fallback) with no client change.
- **Risks/notes:** None.
- **Status:** Not started

#### S05.7: Reconciliation, integrity, and migrations
- **Goal:** Find and repair inconsistencies, and evolve the schema safely.
- **Scope:**
  - A scan that finds media without sidecars, orphaned sidecars, and stale metadata, with repair actions.
  - A schema migration framework (vN to vN+1) with backups and dry-run mode.
  - The policy for external changes (Q10); a CLI migrate command.
- **Deliverables:** reconciliation job and report; repair actions; migration framework; CLI commands.
- **Depends on:** S05.2, S04.3.
- **Requirements:** FR-027, FR-028, FR-029, FR-068, FR-102, NFR-017.
- **Acceptance criteria:**
  1. A scan reports every planted inconsistency in a test library and repairs it when asked.
  2. A migration dry-run reports changes without writing, and a real run backs up affected sidecars first.
  3. Migrations are idempotent.
  4. Orphaned sidecars are never deleted silently.
- **Risks/notes:** The watcher in S09.4 reuses this reconciler.
- **P008 (1.9.0):** the reconciler keeps item IDs across external moves and assigns IDs to unknown files (FR-347); it never imports access data from a hand-edited sidecar (FR-359).
- **Status:** Not started

#### S05.8: Testing and stage review
- **Goal:** Prove sidecar correctness across varied real-world media, then close the stage.
- **Scope:** a fixture set of photos and videos with varied EXIF, GPS, timezones, and corrupt metadata; round-trip tests; documentation; completion record; user sign-off.
- **Deliverables:** scripted fixture generator; round-trip suite; sidecar format documentation; completion record.
- **Depends on:** S05.1–S05.7.
- **Requirements:** NFR-007, NFR-014.
- **Acceptance criteria:**
  1. The fixture set covers the cases in section 12 and is generated by a script.
  2. Round-trip tests (extract → write → read → edit → write) preserve every field.
  3. The completion record is written and the user's sign-off is recorded.
- **Risks/notes:** None.
- **P006 addition (1.6.0, [Planner addition], pending Q54):** Fixtures as in S04.9, through the metadata pipeline (FR-217).
- **Status:** Not started

**Design notes (S05):**
- The reserved `access` and `ai` sections avoid breaking changes in S07 and S17.
- The sidecar manager is the single writer.
- S04's timestamp fallback is replaced here.
- The geocoding dataset is versioned so places can be re-geocoded later.
- The reconciler is reused by S08.2 and S09.4.

**Exit criteria (quoted):** "Every photo has a valid sidecar with extracted and user-edited metadata, and sidecars stay correct through every file operation."

---
