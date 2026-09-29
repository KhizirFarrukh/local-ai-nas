# S17: AI features

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](../PLAN_INDEX.md)). Moved unchanged from `plan.md` section 10.16 in version 1.9.0 (ADR-0044).

### 10.16 S17: AI features

- **Origin:** User-defined
- **Stage ID:** S17 since 1.4.0 (was S12; see 10.18).
- **Position rule:** always the last stage (invariant I8).
- **Goal:** Optional, fully local AI that classifies photos and groups faces, with results stored in sidecars and used by search.
- **User requirements (quoted):**
  > "AI work is always the last stage."
  > "From the README: the AI is opt-in and runs locally alongside the NAS software."
  > "It identifies what kind of photo each one is and classifies it automatically, so searching 'receipts' shows receipt photos even though the user never labelled them."
  > "It detects human faces and groups photos by similar faces. One photo can contain multiple faces."
  > "Classifications are stored in the photo's JSON metadata file, so search reads metadata and never runs the AI model at query time."
  > "It includes an auto photo classification system."
- **Status:** Not started

#### S17.1: AI architecture and opt-in
- **Goal:** A safe, optional, offline AI foundation.
- **Scope:**
  - A separate, optional worker process or container, with the NAS fully functional when it is off (I7).
  - Opt-in toggle (per install or per user: Q35).
  - Hardware detection (CPU, GPU); resource limits and scheduling (throttling, running when idle).
  - Model direction decided in 0.3.0 (ADR-0018). S17.1 selects exact variants by benchmark (license, size, accuracy, CPU performance), GPU execution providers, and embedding storage.
  - Models obtained once with explicit user consent (bundled or downloaded), then run fully offline.
- **Deliverables:** ADR-0018 updated with the chosen variants; AI worker skeleton (ADR-0017); internal job API; model manager with checksums and license display; opt-in settings.
- **Depends on:** S14 and S15 (Done); AI is always last (I8).
- **Requirements:** FR-031, FR-032, FR-136, NFR-002, NFR-004, NFR-005.
- **Acceptance criteria:**
  1. With the worker stopped or not installed, every non-AI test passes.
  2. AI is off by default and enabled only by explicit opt-in.
  3. Models download only with consent, are checksum-verified, and then work with the network disabled.
  4. The exact model variants are recorded in ADR-0018 with license, size, accuracy, and CPU performance, and all satisfy NFR-029.
  5. Resource limits and idle scheduling are configurable and respected.
- **Risks/notes:** Training-data note on permissive face models (ADR-0018); ONNX Runtime wheels for Python 3.14 unverified (ADR-0017); CPU performance (RK-06).
- **P008 (1.9.0):** hardware-aware execution providers with CPU fallback, quantized variants, CLIP ViT-B/32 as a candidate (FR-361). FastEmbed only as a convenience layer running strictly from local files (verified: `specific_model_path` skips downloads; `local_files_only`); the license of the exact CLIP weights is checked here (the OpenAI CLIP repository is MIT; the Hugging Face card of `openai/clip-vit-base-patch32` carries no license tag).
- **Status:** Not started

#### S17.2: AI processing pipeline
- **Goal:** Every photo is processed once, reliably, and reprocessed only when needed.
- **Scope:** jobs on the S04.3 job system; processing new photos on ingest; backfilling the existing library with progress, pause, and resume; idempotent processing; model name and version recorded in the sidecar; reprocessing when the model changes.
- **Deliverables:** AI job types; backfill controller; staleness tracking.
- **Depends on:** S17.1, S04.3.
- **Requirements:** FR-035, FR-036, FR-037, NFR-018.
- **Acceptance criteria:**
  1. New photos are processed automatically after ingest (when opted in).
  2. Backfill shows progress and can pause and resume, including across restarts.
  3. Processing is idempotent, model@version is recorded, and changing the model marks items for reprocessing.
- **Risks/notes:** None.
- **Status:** Not started

#### S17.3: Auto photo classification
- **Goal:** Photos are classified automatically and found by meaning.
- **Scope:**
  - A category taxonomy (e.g. documents, receipts, screenshots, food, pets, landscapes, people, vehicles) that users can extend.
  - Multiple labels per photo with confidence scores; confidence thresholds.
  - Mapping labels into the S06.5 synonym dictionary, so `receipts` also matches `invoice` and `voucher`.
- **Deliverables:** classifier integration; taxonomy file; thresholds; dictionary mapping.
- **Depends on:** S17.2, S06.5.
- **Requirements:** FR-033, FR-034, FR-138.
- **Acceptance criteria:**
  1. The taxonomy includes at least the listed categories, and users can add categories.
  2. Each photo can carry several labels with confidence, and thresholds are configurable.
  3. Searching `receipts` finds unlabelled receipt photos through AI labels and synonyms.
  4. Classification meets the S17.12 evaluation targets.
- **Risks/notes:** Quality on hard categories (RK-07).
- **Status:** Not started

#### S17.4: Face detection
- **Goal:** Find every usable face.
- **Scope:** detection of every face in a photo, with bounding boxes and quality scores; thresholds to ignore tiny or blurry faces.
- **Deliverables:** detector integration; quality scoring; thresholds.
- **Depends on:** S17.2.
- **Requirements:** FR-039, FR-137.
- **Acceptance criteria:**
  1. Faces above the thresholds are detected with a box and a quality score.
  2. Faces below the thresholds are ignored.
  3. Detection meets the S17.12 evaluation targets.
- **Risks/notes:** None.
- **Status:** Not started

#### S17.5: Face recognition and grouping
- **Goal:** Photos of the same person are grouped.
- **Scope:** face embeddings; clustering of similar faces into groups; incremental assignment of new faces to existing groups; one photo belonging to several face groups. Where embeddings are stored is decided by the agent at S17 per ADR-0018 (likely SQLite blobs with brute-force cosine); the sidecar holds group references and boxes, per the README.
- **Deliverables:** embedding storage decision recorded in ADR-0018 (internal data, 8.14); clustering; incremental assignment; face-group registry (Q13).
- **Depends on:** S17.4.
- **Requirements:** FR-040.
- **Acceptance criteria:**
  1. Embeddings are stored as the ADR defines, and sidecars hold boxes and group references.
  2. Similar faces cluster into groups, and one photo can be in several groups.
  3. New faces are assigned incrementally without re-clustering the library.
- **Risks/notes:** None.
- **Status:** Not started

#### S17.6: User corrections
- **Goal:** Users fix mistakes, and the fixes stick.
- **Scope:** name a group; merge and split groups; remove a wrongly assigned face; "not this person"; hide a group; reject AI tags. Corrections persist and are never overwritten by reprocessing.
- **Deliverables:** correction API; constraint store used by clustering.
- **Depends on:** S17.5.
- **Requirements:** FR-038, FR-042, FR-043, FR-044, FR-045.
- **Acceptance criteria:**
  1. Users can name, merge, split, and hide groups, remove faces, and mark "not this person".
  2. Corrections and tag rejections survive reprocessing and model changes (tested).
- **Risks/notes:** None.
- **Status:** Not started

#### S17.7: Sidecar writing and search integration
- **Goal:** AI results are searchable like any other metadata.
- **Scope:** writing tags and faces to the sidecar's `ai` section (schema from S05.1); updating the index; activating the `face:` operator and AI tags in `tag:` and free-text search.
- **Deliverables:** AI result writer (via the sidecar manager); index mapping; `face:` activation.
- **Depends on:** S17.3, S17.5, S06.
- **Requirements:** FR-033, FR-047, FR-056.
- **Acceptance criteria:**
  1. Tags and faces are written to the sidecar `ai` section through the sidecar manager.
  2. `tag:`, free text, and `face:` find AI results.
  3. With the AI worker stopped, search still returns AI results (I4, tested).
- **Risks/notes:** None.
- **Status:** Not started

#### S17.8: AI GUI
- **Goal:** AI features are usable and controllable from the GUI.
- **Scope:** an Explore view for browsing auto-classifications (People, Things, Places); face group management; an AI settings page with the opt-in toggle, status, progress, model information, and pause.
- **Deliverables:** Explore view; people management UI; AI settings page.
- **Depends on:** S17.6, S17.7.
- **Requirements:** FR-036, FR-041, FR-139.
- **Acceptance criteria:**
  1. Explore shows People, Things, and Places.
  2. Face groups can be managed in the GUI.
  3. The AI settings page shows opt-in, status, progress, and model information, and has a pause control.
- **Risks/notes:** None.
- **Status:** Not started

#### S17.9: AI privacy and multi-user rules
- **Goal:** AI respects privacy and access rules.
- **Scope:** classifications and face groups are per user and follow S07 access rules; the handling of faces in shared photos is defined; opting out deletes all AI-derived data (sidecar `ai` sections and embeddings) on request.
- **Deliverables:** per-user AI data model; shared-photo face policy; opt-out deletion job.
- **Depends on:** S17.5, S07.
- **Requirements:** FR-046, FR-140, NFR-011, NFR-024.
- **Acceptance criteria:**
  1. AI data is per user and access-checked like the photos it describes.
  2. Faces in shared photos follow the documented policy.
  3. Opting out deletes all AI-derived data on request (tested).
- **Risks/notes:** Biometric privacy (RK-09).
- **Status:** Not started

#### S17.10: Optional AI extensions
- **Goal:** Extra AI capabilities, if the user wants them.
- **Scope:** priority "Could", pending the user's decision (Q36): OCR so the text of receipts and documents is searchable; local semantic search using embeddings. _(1.4.0, P005: duplicate and similar-photo detection moved out; the baseline is S11 and the AI enhancement S17.11.)_ _Semantic search (S005, D-06): precomputed forms only by default (e.g. tag-vocabulary embeddings computed offline). A query-time text model is an I4 exception that needs separate approval._
- **Deliverables:** only the approved extensions.
- **Depends on:** S17.2, S17.7.
- **Requirements:** FR-054, FR-141.
- **Acceptance criteria (per approved extension):**
  1. It runs locally with no network calls.
  2. Results are stored and searched through the index, never computed at query time.
  3. It has its own evaluation.
- **Risks/notes:** Items not approved are marked "not required".
- **P006 extensions (1.6.0, Could, pending the user's decision like Q36):** G-150 pet recognition and grouping; G-151 smart memories from people, pets, and events; G-152 classification of documents in the files area (e.g. receipt PDFs); G-153 sensitive-content auto-hide suggestion; G-154 blurry photo and screenshot cleanup suggestions (extends S17.11); G-155 local speech-to-text subtitles and search for videos; G-156 sensitive-text redaction suggestions (needs OCR, Q36). They follow I4 (never at query time), I7, I9, and NG9.
- **P008 (1.9.0):** hybrid lexical and semantic search (FR-362, ADR-0043); the query-text encoder runs in the AI worker behind the internal API with a strict time limit (set here, e.g. 300 ms) and a small cache; on timeout or with the worker off, search silently uses lexical results (I4 reworded, D3; I7).
- **Status:** Not started

#### S17.11: AI-assisted library cleanup (new in 1.4.0, P005)
- **Goal:** AI improves look-alike grouping and unlocks AI-based scopes for optimization.
- **Scope:** embedding-based look-alike grouping (finds look-alikes not taken in a burst); smart cover suggestions (sharpest, eyes open, best exposure) that never override a user's choice; activation of the face-group and classification filters for storage optimization (S12.3) and upload policies (S12.6, post-AI follow-up hook).
- **Deliverables:** embedding-similarity grouping feeding the S11.4 stacks; cover suggester; filters activated.
- **Depends on:** S17.2, S17.3, S17.5, S11.4, S12.3, S12.6.
- **Requirements:** FR-164, FR-177, FR-179.
- **Acceptance criteria:**
  1. On the look-alike fixture set, AI grouping finds more true look-alikes than the S11 baseline at an equal or lower false-positive rate.
  2. A cover the user chose is never replaced by a suggestion.
  3. Face-group and classification scopes select the expected items; with AI off, S11 and S12 work unchanged (I7).
- **Risks/notes:** Optional like all AI (I7). The previous evaluation substage becomes S17.12 and stays last.
- **P006 (1.6.0):** G-154 (blurry photo and screenshot cleanup suggestions) extends this substage (Could).
- **Status:** Not started

#### S17.12: Evaluation, packaging, and stage review
- **Goal:** Prove AI quality and performance, package it as optional, and close the stage.
- **Scope:** a labelled evaluation set with accuracy targets; throughput benchmarks on CPU-only hardware; packaging the AI worker as an optional component (e.g. a Compose profile); documentation; completion record; user sign-off.
- **Deliverables:** evaluation set and report; benchmark report; `ai` Compose profile; AI documentation; completion record.
- **Depends on:** S17.1–S17.11.
- **Requirements:** NFR-028, NFR-004, NFR-002.
- **Acceptance criteria:**
  1. Classification and face grouping meet the accuracy targets on the evaluation set.
  2. CPU-only throughput is recorded on reference hardware.
  3. The AI worker ships as an optional component.
  4. The completion record is written and the user's sign-off is recorded.
- **Risks/notes:** Evaluation data must be license-clean or consented and kept out of the repository.
- **Status:** Not started

**Design notes (S17):**
- A separate worker process, with results written by the core's sidecar manager (single writer).
- Embeddings live in internal data, not in sidecars.
- Only permissively licensed models are used (NFR-029, ADR-0018). Non-commercial or research-only weights such as InsightFace are excluded.
- Per-user AI data follows S07.

**Exit criteria (quoted):** "With AI enabled, unlabelled receipt photos are found by searching 'receipts', faces are grouped correctly per the evaluation targets, and disabling AI leaves the NAS fully functional."

---
