# S04: Media management

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](../PLAN_INDEX.md)). Moved unchanged from `plan.md` section 10.5 in version 1.9.0 (ADR-0044).

### 10.5 S04: Media management

- **Origin:** User-defined
- **Goal:** A separate photos area with Google Photos style management.
- **User requirements (quoted):**
  > "Stage 4 is media management implementation."
  > "Media storage is separate from file storage."
  > "The NAS has two folders at the root, files and photos, and they never intersect."
  > "Content moves between them only when the user selects a file to copy or move to files or photos."
- **Status:** Not started

#### S04.1: Photos area rules and domain model
- **Goal:** Define what a media item is and how it is stored inside `photos/`, and enforce separation from `files/`.
- **Scope:**
  - Definition of a media item; supported formats (videos, HEIC, and RAW are open questions: Q26).
  - How media is organized on disk inside `photos/` (proposed ADR, Q40).
  - I1 enforced in the service layer; a photos API namespace separate from the files API.
- **Deliverables:** ADR for the photos on-disk layout; `PhotosService`; content-based media type detection; `/api/v1/photos` namespace.
- **Depends on:** S03 (Done), S01.2, S01.5.
- **Requirements:** FR-017, FR-018, FR-019, FR-020, FR-092, FR-093, NFR-026.
- **Acceptance criteria:**
  1. The photos service can address only `photos/<ns>/`, and the files service only `files/<ns>/` (tests).
  2. Media type is determined from content, and unsupported types are rejected.
  3. The photos API is separate from the files API and follows S01.5 conventions.
  4. The supported-format list matches the answer to Q26.
- **Risks/notes:** HEIC/RAW decoders have licensing implications (ADR-0012, `dependencies.md`).
- **P006 addition (1.6.0, [Planner addition], pending Q54):** Domain model: paired items, a Live Photo or motion photo as one library item (FR-217, FR-218).
- **1.10.0 (the user's answer to Q40):** the photos layout is **hybrid**: uploaded folders keep their structure; loose photos go to `YYYY/MM/` by date taken; the layout ADR records the rules (date source and fallback, name conflicts, refiling when a date is corrected).
- **Status:** Not started

#### S04.2: Media ingestion
- **Goal:** Get media into the photos area safely, without duplicates or junk.
- **Scope:**
  - Upload into the photos area; only media types accepted.
  - Duplicate detection by content hash; import batches; handling of corrupt or unreadable media.
  - _Planner addition, pending Q39:_ server-side import from a host folder.
- **Deliverables:** photo upload (reusing S01.4); hashing during streaming; duplicate check; batch model and report.
- **Depends on:** S04.1, S01.4.
- **Requirements:** FR-003, FR-010, FR-022, FR-094, FR-143.
- **Acceptance criteria:**
  1. Uploading non-media to the photos area is refused with a clear reason.
  2. An exact duplicate is detected and reported, and no second copy is stored unless the user chooses to keep it.
  3. A batch with some corrupt files stores the valid ones and reports the corrupt ones.
  4. Ingested items appear in the photos API with hash, size, type, and file timestamps.
- **Risks/notes:** Post-processing (thumbnails) is queued through S04.3, which is built before S04.4 in the stage's execution order.
- **P005 change (1.4.0):** exact-duplicate detection at ingest uses the stored content hash (FR-211), with the default "Skip and report" (FR-022, Q45). The full per-user policy arrives in S11.2 (FR-154). Resolution variants are never blocked at upload.
- **P006 addition (1.6.0, [Planner addition], pending Q54):** Pairing at ingest, including a half that arrives late or in the other order (FR-217).
- **P008 (1.9.0):** ingest is fully asynchronous: dropping thousands of files enqueues cheap jobs in batches and the API stays responsive (NFR-054); backpressure limits queued work per user and overall.
- **Status:** Not started

#### S04.3: Background job system
- **Goal:** One reliable background-work system for the whole project.
- **Scope:**
  - A persistent job queue with retries, progress reporting, concurrency limits, and survival across restarts.
  - Built here and reused by S05, S06, S08, S09, and S17.
  - It also replaces the S01.4 and S03.6 simple schedulers.
- **Deliverables:** custom SQLite-backed queue per ADR-0011 (leases, backoff, priorities); worker pool; priorities; per-user job context; internal progress API.
- **Depends on:** S03.2 (internal DB).
- **Requirements:** FR-095, NFR-012.
- **Acceptance criteria:**
  1. Jobs survive a server restart and resume or retry.
  2. Failures retry with backoff up to a limit, then are marked failed with the error.
  3. Per-type concurrency limits hold under load.
  4. Progress of long jobs is available through the API.
  5. Every job carries the user it acts for, which is the hook for I5 checks in S07.4.
- **Risks/notes:** None.
- **P008 (1.9.0):** grows the minimal job foundation built in S03 (FR-351) instead of creating it: priorities, per-type limits, the full monitor; lower OS priority for external tools (NFR-055, 8.41).
- **Status:** Not started

#### S04.4: Thumbnails and previews
- **Goal:** Fast visual browsing without ever loading originals in grids.
- **Scope:** several thumbnail sizes; video poster frames; generation as background jobs; storage in internal app data, never inside `photos/` (I2); regeneration on demand.
- **Deliverables:** thumbnail job; rendition store keyed by content hash; regeneration command.
- **Depends on:** S04.2, S04.3.
- **Requirements:** FR-012, FR-019, NFR-003.
- **Acceptance criteria:**
  1. Every supported item gets the configured sizes, stored only in internal data.
  2. Generation never blocks uploads or browsing.
  3. Deleting the thumbnail cache and regenerating restores all thumbnails.
  4. Videos get a poster frame (if videos are supported per Q26).
  5. Thumbnails are served only after the same authorization check as the original.
- **Risks/notes:** None.
- **P005 change (1.4.0):** a 64-bit perceptual hash is computed for every image from a small rendition produced with the thumbnails, and stored in the sidecar and the index (FR-212, ADR-0022). (libvips runs as a separate process, so the core has no decoded image to reuse; the small rendition costs no extra read of the original.)
- **P008 (1.9.0):** thumbnail tools run at lower OS priority where the platform honors it (NFR-055).
- **Status:** Not started

#### S04.5: Library organization
- **Goal:** Organize the library without ever duplicating files.
- **Scope:** timeline (sorted by file timestamps until S05 provides date taken); albums as virtual collections that reference items; favorites; hide or archive.
- **Deliverables:** timeline query (keyset pagination, day grouping); album model and API (storage per Q13); favorites; hidden and archived flags.
- **Depends on:** S04.2.
- **Requirements:** FR-013, FR-016, FR-096.
- **Acceptance criteria:**
  1. The timeline lists all items newest first, grouped by day, with keyset pagination.
  2. Adding to an album never copies a file, and deleting an album never deletes items.
  3. Favorites and hidden or archived items can be set, unset, and filtered.
  4. The timeline API is designed to switch to date taken in S05.6 without a client change.
- **Risks/notes:** Album storage must satisfy I2 (Q13).
- **Status:** Not started

#### S04.6: Cross-area transfer
- **Goal:** Content moves between `files/` and `photos/` only by explicit user action.
- **Scope:** explicit copy and move between files and photos, in both directions; validation (only media may enter photos); behavior of metadata during transfer (coordinated with S05.6, Q27); conflict handling; audit logging.
- **Deliverables:** `TransferService` (the only code path touching both areas); transfer API; audit events.
- **Depends on:** S04.1, S04.2, S03.6.
- **Requirements:** FR-097, FR-090, NFR-026.
- **Acceptance criteria:**
  1. Copy and move work in both directions, only on explicit request.
  2. Non-media sent to photos is refused.
  3. Every transfer creates an audit event.
  4. Name conflicts follow the S01.6 policy (fail, rename, overwrite).
  5. A move is atomic from the user's view: the item is in exactly one area at any time.
- **Risks/notes:** Until S05 exists there are no sidecars to carry. S05.6 completes the metadata behavior.
- **P006 addition (1.6.0, [Planner addition], pending Q54):** A pair is transferred as a unit (FR-218).
- **Status:** Not started

#### S04.7: Photos GUI
- **Goal:** A Google Photos–style interface for the photos area.
- **Scope:**
  - Timeline grid with date grouping and virtualized infinite scroll.
  - Lightbox viewer with zoom, swipe, and video playback.
  - Album UI; photo upload UI; multi-select.
  - "Copy/Move to Photos" in the files GUI and "Copy/Move to Files" in the photos GUI.
- **Deliverables:** photos views, lightbox, album screens, and transfer actions in both GUIs.
- **Depends on:** S04.4, S04.5, S04.6, S02.
- **Requirements:** FR-013, FR-014, FR-016, FR-021, FR-097.
- **Acceptance criteria:**
  1. The timeline scrolls smoothly through 50,000 items.
  2. The lightbox supports zoom, swipe or arrow navigation, and video playback with seeking.
  3. Albums can be created, renamed, deleted, and filled from a multi-selection.
  4. Cross-area copy and move exist only as explicit, confirmed actions in both GUIs.
- **Risks/notes:** None.
- **P006 addition (1.6.0, [Planner addition], pending Q54):** The "Live" badge and motion playback, with "play" and "still only" (FR-218).
- **Status:** Not started

#### S04.8: Video streaming and quality levels
- **Goal:** Play videos from either area with a live quality menu (Auto, Original, 1080p, 720p, 480p, 360p). Lower qualities are produced on demand and cached.
- **Scope:**
  - _Added in 0.4.0 at the user's request (S004 E005, answers E008; ADR-0020)._
  - HLS playback: master and level playlists, segments. hls.js player with a manual quality menu and Auto (adaptive bitrate); native HLS fallback.
  - Hybrid on-demand transcoding: the original streams without transcoding. Lower levels are transcoded on first request into cached HLS segments (FFmpeg subprocess sessions, keyframe-aligned 4 s segments, restart on seek).
  - A transcode cache in internal data with a size cap and LRU eviction (a job on S04.3).
  - Hardware-encoder detection with CPU fallback; concurrency and maximum-quality limits in settings.
  - The same player in the photos lightbox (S04.7) and the files-area video preview (S02.6, upgraded).
  - Authorization of every playlist and segment request.
- **Deliverables:** streaming endpoints (in `api/openapi.yaml`); transcoding session manager; cache and eviction job; encoder detection in health; a shared Svelte player component with the quality menu; Docker device-passthrough documentation.
- **Depends on:** S04.3, S04.4, S04.7, S02.6, S03.5.
- **Requirements:** FR-019, FR-144, FR-145, FR-146, FR-147, FR-148, NFR-031, NFR-024.
- **Acceptance criteria:**
  1. A video plays in the photos lightbox and in the files preview, and the user can switch between Auto and each available level during playback without restarting.
  2. The original quality plays without any transcoding. A lower level starts within the NFR-031 target and is served from the cache on later plays.
  3. The transcode cache never exceeds its size cap. Least-recently-used levels are evicted, and eviction never touches originals.
  4. With several viewers, concurrent transcodes stay within the configured limit, and the NAS API keeps its NFR-003 latency.
  5. Playlist and segment requests without authorization are refused (tested), and cache contents are never listed.
- **Risks/notes:**
  - Weak CPUs (RK-29). H.264 encoder licensing (libx264 is GPL; RK-30, audit A001 D-04). Manual quality selection on iOS depends on hls.js support there (RK-31).
  - Side effect **confirmed by the user in S005 (D-14)**: originals in codecs a browser cannot play become playable through the transcoded levels (ADR-0020).
- **P008 (1.9.0):** interactive streaming transcodes keep normal priority; when one GPU serves transcoding and AI, playback wins and AI jobs pause (NFR-055).
- **Status:** Not started

#### S04.9: Testing and stage review
- **Goal:** Prove separation, correctness, and performance, then close the stage.
- **Scope:** tests proving that no operation places an item in the other area implicitly; performance testing with a large library (e.g. 50,000 items); _video streaming tests (added in 0.4.0)_; documentation; completion record; user sign-off.
- **Deliverables:** separation test suite; 50k-item performance report; streaming test suite and cross-browser results; photos user guide; completion record.
- **Depends on:** S04.1–S04.8.
- **Requirements:** NFR-026, NFR-003, NFR-027, NFR-031.
- **Acceptance criteria:**
  1. Every API endpoint and job that writes files is exercised, and none places an item in the other area implicitly.
  2. With 50,000 items, timeline page loads meet NFR-003 on reference hardware.
  3. Streaming tests pass on current Chrome, Edge, Firefox, and Safari (macOS and iOS): switching levels during playback, Auto, seeking, cache eviction under the size cap, and refusal of unauthorized segment requests.
  4. Documentation and the completion record are written, and the user's sign-off is recorded.
- **Risks/notes:** Synthetic library generator needed (section 12).
- **P006 addition (1.6.0, [Planner addition], pending Q54):** Fixtures: Apple pairs uploaded in both orders, Google and Samsung motion photos, unpaired MOV files (FR-217, FR-218).
- **P008 (1.9.0):** the responsiveness test: browsing and search within NFR-003 while a 10,000-file import runs (NFR-054).
- **Status:** Not started

**Design notes (S04):**
- Where an S04 feature depends on data that only arrives in S05 (such as date taken), it uses a simple fallback now, which S05 replaces.
- The job system is the shared infrastructure for all later background work.
- `TransferService` is the single crossing point between areas.
- Photos namespaces are per user from ADR-0003.
- Thumbnail authorization uses the S03.5 policy, ready for S07.4.

**Exit criteria (quoted):** "Photos can be uploaded, browsed on a timeline, organized into albums, and transferred between areas only by explicit user action."

---
