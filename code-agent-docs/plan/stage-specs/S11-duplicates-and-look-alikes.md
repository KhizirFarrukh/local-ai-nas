# S11: Duplicate and look-alike management

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](../PLAN_INDEX.md)). Moved unchanged from `plan.md` section 10.12 in version 1.9.0 (ADR-0044).

### 10.12 S11: Duplicate and look-alike management

- **Origin:** User-defined features (P005), placed by the planner; burst grouping added by the user in S007 (E004).
- **Goal:** Find exact and resolution-variant duplicates of photos, find duplicate files, group look-alike photos and bursts into stacks, and let the user resolve all of these safely.
- **User requirements (quoted from P005 and S007):**
  > "Detect literal duplicate photos using hashing." "Detect duplicates that are the same photo at different resolutions." "For different-resolution duplicates, the user chooses which one to keep."
  > "Group photos that are almost the same but have minor differences." "Show the user that it is a grouped photo." "The user chooses which photo in the group is displayed on the grid. By default, the first photo is displayed." "The user can choose to keep one photo and delete the rest." "AI may be used to detect near-identical photos."
  > "Handle duplicate files in the files area." "The user chooses whether to keep both, or keep one and which one." "The user chooses whether to put a shortcut to the kept file in the deleted file's place."
  > "you also need to add auto grouping of burst photos" (S007)
- **Placement (P005):** after S10. It needs metadata (S05), per-user ownership (S07), the trash (S08.1) for safe deletion and undo, and network access (S09), because shortcuts must also work over shares.
- **Status:** Not started

#### S11.1: Hashing foundation and backfill
- **Goal:** Every item has the hashes that detection needs, and lookups scale.
- **Scope:** confirm that content hashes (stored from the S01 follow-up and S04.2) and perceptual hashes (S04.4) exist for every item; a backfill job for missing ones (including files uploaded before FR-211 and files added outside the app); the similarity index (ADR-0022); threshold calibration on a labelled fixture set.
- **Deliverables:** backfill job with progress; similarity index built from stored hashes; labelled fixture set (exact copies, re-saved and re-encoded copies, crops, bursts, "similar but different" pairs); calibration report with the chosen thresholds.
- **Depends on:** S04.2, S04.4, S05.1, S04.3; ADR-0021, ADR-0022 Accepted.
- **Requirements:** FR-211, FR-212, FR-213, NFR-034.
- **Acceptance criteria:**
  1. After the backfill, every item in both areas has its content hash, and every image its perceptual hash.
  2. A lookup in a 100,000-photo library returns candidates without comparing all pairs (benchmark recorded).
  3. The thresholds and their measured false-positive rates on the fixture set are recorded, and exact duplicates have no false positives.
- **Risks/notes:** RK-32 (wrong matches). Hashes of HEIC and RAW depend on the formats supported (Q26).
- **Status:** Not started

#### S11.2: Photo duplicate detection
- **Goal:** Exact duplicates and resolution variants are found within each user's photos.
- **Scope:** exact duplicates by content hash; resolution variants by perceptual hash, aspect ratio, and date taken (FR-151); per-user scope only (I5); incremental scans on every upload, scheduled scans, and on-demand scans as background jobs; the per-user upload-time policy for exact duplicates (Skip and report, Keep both, Ask); videos: exact duplicates only.
- **Deliverables:** detection jobs; duplicate groups with a type ("Exact", "Resolution variant") and the evidence; upload-time policy setting; cross-area handling per Q43.
- **Depends on:** S11.1, S07.
- **Requirements:** FR-150, FR-151, FR-153, FR-154, FR-159, FR-022.
- **Acceptance criteria:**
  1. Exact duplicates and resolution variants in the fixture set are grouped with the right type, and "similar but different" pairs are not.
  2. No group ever contains items of two users (tested with two users and identical files).
  3. A new upload is checked incrementally; the upload policy behaves as set (skip and report, keep both, ask).
- **Risks/notes:** Q43 (the same photo in both areas), Q45 (upload default).
- **P006 addition (1.6.0, [Planner addition], pending Q54):** A pair is one item for duplicate detection; the still decides (FR-218).
- **Status:** Not started

#### S11.3: Photo duplicate review and resolution
- **Goal:** The user resolves photo duplicates with full information and nothing is lost by mistake.
- **Scope:** side-by-side comparison (FR-152); the recommended pick; bulk rules with a preview of every group (I10); metadata merge onto the kept item with provenance; removal to the trash; "not duplicates" marks; reclaimable space per group and total.
- **Deliverables:** resolution API; metadata-merge routine (through the sidecar manager, I9); bulk-rule engine with a preview; "not duplicates" store in sidecars.
- **Depends on:** S11.2, S08.1, S05.5, S07.5.
- **Requirements:** FR-152, FR-155, FR-156, FR-157, FR-158, FR-159, NFR-033.
- **Acceptance criteria:**
  1. Keeping one item moves the others to the trash; restoring them brings back the items, their sidecars, and their album and share memberships.
  2. Albums, favorites, tags, descriptions, and shares of removed items appear on the kept item, and the sidecar records their origin.
  3. A bulk rule shows every group's result before it applies, and nothing changes without confirmation.
  4. A group marked "not duplicates" never comes back, even after a rescan.
- **Risks/notes:** Merging shares must keep I5 (a share granted on a removed item moves to the kept item only for the same owner).
- **Status:** Not started

#### S11.4: Look-alike stacks and bursts
- **Goal:** Near-identical photos and bursts are grouped into stacks the user controls.
- **Scope:** detection per FR-160 (looser perceptual threshold, time window, same camera) and bursts per FR-165 (camera burst identifier first, then the shot sequence); automatic stacking (never deletes); the cover (default: the camera's burst primary, else the first photo per Q42); user actions (choose the cover, remove a photo, unstack, merge stacks, "keep one, delete the rest", manual stacks); user decisions locked against regrouping; settings (on/off, sensitivity, time window); stacks visible only to the owner; persistence in each member's sidecar (I3); behavior in timeline, albums, and search.
- **Deliverables:** stack detector and burst grouper; stack API; sidecar `stack` section writer; search index fields filled (`stack_id`, `stack_cover`); ADR-0023 decided.
- **Depends on:** S11.1, S05.3 (burst identifiers, date taken), S06.2.
- **Requirements:** FR-160, FR-161, FR-162, FR-163, FR-165, FR-166, FR-167, FR-168.
- **Acceptance criteria:**
  1. Burst fixtures with Apple `BurstUUID` and Google `GCamera:BurstID` form one stack each; bursts without identifiers are grouped by the shot sequence; the camera's primary shot is the cover.
  2. Look-alike fixtures form stacks at the default sensitivity, and "similar but different" pairs do not.
  3. A chosen cover, a removal, an unstack, and a manual stack survive a rescan and an index rebuild from sidecars.
  4. People a stacked photo is shared with see it as an individual photo.
  5. "Keep one, delete the rest" asks first and moves the rest to the trash.
- **Risks/notes:** RK-32. Q42 (cover definition). The AI enhancement is S17.11 (FR-164).
- **Status:** Not started

#### S11.5: File duplicate detection
- **Goal:** Identical files in the files area are found cheaply.
- **Scope:** grouping by size, then a partial hash (first and last blocks), then the full hash (stored hashes are reused, so unchanged files are never re-read); files added outside the app hashed by the watcher and reconciliation; per-user scope, shared files never deletion candidates; empty files excluded; incremental, scheduled, and on-demand scans; folder ignore list; upload-time check with keep both, skip, or shortcut.
- **Deliverables:** file duplicate job; ignore-list setting; upload-time check in the upload flow (tus hook and simple upload).
- **Depends on:** S11.1, S09.4, S07.
- **Requirements:** FR-169, FR-172, FR-173, FR-153.
- **Acceptance criteria:**
  1. Identical files under different names and folders are grouped; files of equal size but different content are not; empty files never are.
  2. Files in an ignored folder are never suggested.
  3. An upload matching an existing file offers keep both, skip, or a shortcut, and each choice gives the matching result.
- **Risks/notes:** None beyond RK-32 (exact matches only, so the risk is low).
- **Status:** Not started

#### S11.6: File duplicate resolution and shortcuts
- **Goal:** The user keeps one copy and, if wanted, leaves shortcuts that behave correctly in every case.
- **Scope:** resolution options (keep all as "not duplicates"; keep one; bulk rules with a preview); shortcuts at the removed copies' locations; the representation chosen in ADR-0024 (recommended: app-level shortcut records by stable file ID); every behavior rule in FR-175; how shortcuts appear over network shares (Q46).
- **Deliverables:** shortcut records and API; listing, download, search, and quota integration; WebDAV representation; ADR-0024 decided.
- **Depends on:** S11.5, S08.1, S09.2, S10.2.
- **Requirements:** FR-170, FR-171, FR-174, FR-175, NFR-033.
- **Acceptance criteria:**
  1. Opening or downloading a shortcut serves the target; moving or renaming the target keeps the shortcut working.
  2. Trashing the target shows "target missing" with a restore option; restoring it brings the shortcut back to life; deleting a shortcut never deletes the target; deleting a target warns how many shortcuts point to it.
  3. Shortcuts use no quota, cannot be shared, and never grant access to a target the user cannot read (I5).
  4. Over WebDAV, shortcuts appear as ADR-0024 decides.
- **Risks/notes:** Q46. Symlinks and hard links are rejected in ADR-0024 (the S01.6 symlink policy; hard links free no space and edit both copies).
- **Status:** Not started

#### S11.7: Duplicates and stacks GUI
- **Goal:** Everything above is usable from the GUI.
- **Scope:** a Duplicates view for photos and for files (groups, comparison, recommended pick, bulk rules, reclaimable space); stack badges in the grid, timeline, albums, and search results; the stack viewer and its actions; shortcut items in file listings; related settings (upload policy, stacking, ignore list).
- **Deliverables:** Duplicates view; stack viewer; badges; settings pages.
- **Depends on:** S11.3, S11.4, S11.6, S04.7.
- **Requirements:** FR-152, FR-158, FR-161, FR-166, FR-175.
- **Acceptance criteria:**
  1. Every S11 action can be done from the GUI, by mouse, touch, and keyboard.
  2. Each destructive action shows a preview and asks first (I10).
- **Risks/notes:** None.
- **Status:** Not started

#### S11.8: Testing, audit, and stage review
- **Goal:** Prove accuracy, safety, and privacy, then close the stage.
- **Scope:** false-positive tests on the fixture set; large-library performance; metadata-merge tests; shortcut edge cases (target moved, trashed, restored, deleted); privacy tests proving no cross-user matches; burst tests; the stage's unit, integration, and system tests (S006); the R12 documentation audit; completion record; user sign-off.
- **Deliverables:** test suites; accuracy and performance report; audit report; completion record.
- **Depends on:** S11.1–S11.7.
- **Requirements:** NFR-034, NFR-033, NFR-024; verification of every S11 requirement.
- **Acceptance criteria:**
  1. The accuracy targets of S11.1 hold on the fixture set, and exact duplicates have no false positives.
  2. No test finds a cross-user match or leak (results, counts, timing).
  3. The shortcut edge cases pass.
  4. Coverage and the audit meet the stage-end rules; the completion record is written and the user's sign-off is recorded.
- **Risks/notes:** The fixture set must be license-clean and generated or consented (12.3).
- **Status:** Not started

**Design notes (S11):** Duplicate groups are recomputed from stored hashes; the durable data is the user's decisions (sidecars for photos, the internal DB for files, both backed up). Nothing is deleted except through the trash after confirmation (I10). Matching never crosses users (I5).

**Exit criteria (P005):** "Duplicates and look-alikes are found accurately within each user's own library. Every removal goes through preview, confirmation, and trash. Shortcuts behave correctly in every edge case." _(1.4.0: and bursts are grouped automatically, S007.)_

---
