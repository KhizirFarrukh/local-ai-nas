# S12: Storage optimization

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](../PLAN_INDEX.md)). Moved unchanged from `plan.md` section 10.13 in version 1.9.0 (ADR-0044).

### 10.13 S12: Storage optimization

- **Origin:** User-defined feature (P005), placed by the planner.
- **Goal:** Let users reduce the resolution and quality of existing and future media, with a live preview, full metadata preservation, and an undo window.
- **User requirements (quoted from P005):**
  > "Let the user reduce the resolution of existing photos and videos to save storage." "Scope choices: all media, selected items, photos only, videos only, or media in a certain group or classification." "An option to reduce the resolution of future uploads automatically." "The user chooses how much quality reduction to apply." "A live preview of one sample photo shows what it will look like after reduction."
  > "Reduction to a fixed resolution: when the aspect ratio does not match the target, the user chooses whether the width or the height is matched." "Alternatively, the user can choose percentage scaling." "The user can limit compression to certain types or groups of media."
- **Placement:** after S11, which provides look-alike stacks as a scope filter; it uses the trash (S08.1) for undo.
- **Status:** Not started
- **Security (1.12.0, P010; rule R15):** the only stage where ExifTool writes, and only on files the core created (NFR-065); the tools run in the sandbox (NFR-064).

#### S12.1: Image optimization engine
- **Goal:** Resize and re-encode an image exactly as asked, keeping all its metadata.
- **Scope:** resize modes (match width, match height, fit within, percentage); orientation-aware targets; never upscale; the rounding rule; quality; keeping the format (conversion is a Could); RAW and animated images excluded by default; metadata preservation (8.23); atomic replacement with verification.
- **Deliverables:** resize arithmetic module (pure, heavily unit-tested); libvips and ExifTool pipeline; verification step.
- **Depends on:** S04.4, S05.3, ADR-0012.
- **Requirements:** FR-176, FR-180, FR-182, FR-183, FR-184, FR-185, FR-186, FR-187, FR-190, FR-193, NFR-039.
- **Acceptance criteria:**
  1. The user's examples hold: 1920×1080 to a 1024×768 target gives 1024×576 (match width), 1365×768 (match height), and 1024×576 (fit within); 50% turns 4000×3000 into 2000×1500.
  2. Items at or below the target are not resized; nothing is stretched or cropped.
  3. EXIF, XMP, and IPTC survive for JPEG, HEIC, and PNG fixtures, and orientation is applied exactly once.
- **Risks/notes:** HEIC output needs an HEVC encoder in libheif (x265 is GPL; kvazaar is BSD), recorded in the register. RK-33.
- **Status:** Not started

#### S12.2: Video optimization engine
- **Goal:** Re-encode videos smaller with predictable quality and their metadata intact.
- **Scope:** FFmpeg transcoding with the resize modes and presets (2160p, 1440p, 1080p, 720p, 480p), never upscaling; CRF quality; the codec policy (ADR-0025, Q48); audio kept (optional re-encode); creation time, GPS, and rotation preserved; hardware acceleration evaluated.
- **Deliverables:** video pipeline; ADR-0025 decided; metadata checks for MP4 and MOV.
- **Depends on:** S04.8 (FFmpeg setup), S12.1 (shared arithmetic).
- **Requirements:** FR-176, FR-180, FR-190, FR-192.
- **Acceptance criteria:**
  1. Output plays in the target browsers (per the codec policy) and in the S04.8 player.
  2. Creation time, GPS, and rotation survive for MP4 and MOV fixtures.
  3. Resolution presets never upscale.
- **Risks/notes:** CPU cost on weak hardware (RK-29); encoder licenses (RK-36).
- **Status:** Not started

#### S12.3: Scope selection and estimation
- **Goal:** The user selects exactly the media to change and knows the effect first.
- **Scope:** all scope filters (FR-177, FR-178), combinable; the face-group and classification filters exist as hooks, activated in S17.11; item counts; skip reasons; a sampling-based estimate of the space saved.
- **Deliverables:** scope query builder over the index; estimator.
- **Depends on:** S06, S11.4, S12.1, S12.2.
- **Requirements:** FR-177, FR-178, FR-188.
- **Acceptance criteria:**
  1. Each filter, alone and combined, selects exactly the expected fixture items.
  2. The estimate is within a documented tolerance of the actual savings on the fixture library.
- **Risks/notes:** None.
- **P006 addition (1.6.0, [Planner addition], pending Q54):** Optimization keeps a pair's motion part unchanged by default, with an option to include it (FR-218).
- **Status:** Not started

#### S12.4: Live preview service
- **Goal:** The result of any setting can be seen in seconds, without harming the server.
- **Scope:** on-the-fly preview rendering for images and short video clips; a cache; limits (concurrency, size, rate) so previews stay responsive and never overload the server; never saved.
- **Deliverables:** preview endpoint and renderer; limits; cache in internal data (I2).
- **Depends on:** S12.1, S12.2.
- **Requirements:** FR-181, NFR-035.
- **Acceptance criteria:**
  1. A typical photo preview renders within the bound set here on reference hardware.
  2. Rapid setting changes are debounced, and a flood of preview requests is limited without affecting other users.
- **Risks/notes:** RK-29 on weak hardware.
- **Status:** Not started

#### S12.5: Bulk optimization jobs and safety
- **Goal:** Optimizing many items is safe, visible, and reversible.
- **Scope:** dry run; confirmation with counts and the estimate (I10); jobs with progress, pause, resume, cancel; failed items skipped and reported; originals kept for the retention period and revertible per item or per job (ADR-0026, Q44); the delete-at-once option with its warning; quota updates; `optimizationHistory` in sidecars; skipping items already optimized.
- **Deliverables:** optimization job; originals store; revert API; ADR-0026 decided.
- **Depends on:** S12.1–S12.4, S08.1 (S08.5 only if approved), S10.2.
- **Requirements:** FR-189, FR-191, FR-190, NFR-033, NFR-006.
- **Acceptance criteria:**
  1. A dry run changes nothing and reports what would change.
  2. Killing the server mid-job never leaves a broken or missing item (fault injection).
  3. Reverting an item or a job restores the originals byte-identical, with their sidecar data.
  4. Space is freed only when originals are removed, and quotas follow.
- **Risks/notes:** RK-33.
- **Status:** Not started

#### S12.6: Upload policies
- **Goal:** Future uploads are optimized automatically, as the user set.
- **Scope:** per-user policies (conditions and actions, on/off); applied at ingest after the upload completes and before the item appears; the original kept for the retention period or discarded at once, as chosen; items marked, with the original resolution in the info panel; preview while creating a policy; the post-AI follow-up hook for classification conditions (activated in S17.11); who may apply policies (Q49).
- **Deliverables:** policy model and ingest hook; policy editor API.
- **Depends on:** S12.5, S04.2.
- **Requirements:** FR-179, FR-190.
- **Acceptance criteria:**
  1. An upload matching a policy is stored optimized, with metadata intact and the original kept or discarded as set.
  2. Uploads not matching any policy are untouched.
- **Risks/notes:** Q49.
- **Status:** Not started

#### S12.7: Storage optimization GUI
- **Goal:** A clear, step-by-step way to shrink the library.
- **Scope:** a wizard (scope, settings, preview, estimate, confirm); job progress; revert; the upload policy editor; the storage report.
- **Deliverables:** wizard; job view; policy editor; storage report.
- **Depends on:** S12.3–S12.6.
- **Requirements:** FR-181, FR-194, FR-183.
- **Acceptance criteria:**
  1. The whole flow works by mouse, touch, and keyboard, and the percentage mode says it applies per dimension.
  2. Nothing changes before the confirmation step, which shows the counts and the estimate.
- **Risks/notes:** None.
- **Status:** Not started

#### S12.8: Testing, audit, and stage review
- **Goal:** Prove correct arithmetic, preserved metadata, and safe reverts, then close the stage.
- **Scope:** resize arithmetic tests including the user's examples; metadata preservation across JPEG, HEIC, PNG, and video formats; orientation tests; revert tests; performance; the stage's unit, integration, and system tests; the R12 documentation audit; completion record; user sign-off.
- **Deliverables:** test suites; report; audit; completion record.
- **Depends on:** S12.1–S12.7.
- **Requirements:** NFR-039, NFR-033, NFR-035; verification of every S12 requirement.
- **Acceptance criteria:**
  1. All arithmetic and metadata tests pass on every supported format.
  2. Revert restores byte-identical originals.
  3. Coverage and the audit meet the stage-end rules; the completion record is written and the user's sign-off is recorded.
- **Risks/notes:** None.
- **Status:** Not started

**Design notes (S12):** The app never replaces an original on its own (NG4, NFR-006): only an optimization the user starts or a policy the user enables does, after verification, keeping the original for the retention period. Face boxes are normalized, so AI data stays valid.

**Exit criteria (P005):** "Users can safely shrink their library with exactly the scope and settings they choose, see the result before committing, keep all metadata, and undo within the retention period."

---
