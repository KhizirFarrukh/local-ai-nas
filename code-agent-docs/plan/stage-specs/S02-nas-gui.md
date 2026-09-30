# S02: NAS GUI

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](../PLAN_INDEX.md)). Moved unchanged from `plan.md` section 10.3 in version 1.9.0 (ADR-0044).

### 10.3 S02: NAS GUI

- **Origin:** User-defined
- **Goal:** A graphical application that lets people use the NAS without touching the API.
- **User requirements (quoted):**
  > "Stage 2 is the NAS software (GUI) that helps in interacting with the actual NAS."
- **Status:** **Done** (S007, 2026-09-29; signed off by the user: "you know what, assume its tested, continue to next stage"; completion record in `stages/S02-nas-gui.md` section 13)

#### S02.1: GUI technology and design foundation
- **Goal:** Choose the GUI approach and build the design system and API client that every screen uses.
- **Scope:**
  - GUI approach and framework decided (0.3.0): SvelteKit static SPA + TypeScript + Tailwind CSS, embedded in the Go binary (ADR-0009; Q25 answered). S02.1 confirms toolchain versions (Node.js LTS; TypeScript/svelte-check compatibility) and records them in `dependencies.md`.
  - Design system: typography, color, spacing, components, icons.
  - Light and dark themes; API client layer generated from OpenAPI.
- **Deliverables:** toolchain confirmation recorded; web app skeleton served by the NAS (go:embed); component library; theme tokens; generated API client; frontend lint, type-check, and tests in CI.
- **Depends on:** S01 (Done), S01.5.
- **Requirements:** FR-078, FR-083, NFR-001, NFR-014.
- **Acceptance criteria:**
  1. The GUI is served by the NAS at the same origin as the API and makes no external requests (fonts, icons, and scripts are bundled).
  2. Light and dark themes switch at runtime and follow the OS setting by default.
  3. The API client is generated from the committed OpenAPI spec, and CI fails on drift.
  4. Frontend lint, type checks, and unit tests run in CI.
- **Risks/notes:** TypeScript 7 compatibility with svelte-check was not verified in S003. Pin a compatible TypeScript version if needed.
- **Status:** Done (S006, 2026-09-25; details in `stages/S02-nas-gui.md`)

#### S02.2: App shell and navigation
- **Goal:** A stable layout and navigation frame that every feature plugs into.
- **Scope:** main layout; navigation with Files, Photos (placeholder until S04), and Settings (placeholder); routing; global loading and error states; notifications.
- **Deliverables:** shell, router, error boundary, notification system.
- **Depends on:** S02.1.
- **Requirements:** FR-079.
- **Acceptance criteria:**
  1. Files, Photos, and Settings are reachable with deep links and browser back and forward.
  2. API errors show a consistent, human-readable message, never a blank screen.
  3. Long operations show progress, and completions and failures raise notifications.
- **Risks/notes:** The placeholders are replaced in S04.7 (Photos) and S10.5 (Settings).
- **Status:** Done (S006, 2026-09-25; details in `stages/S02-nas-gui.md`)

#### S02.3: File browser
- **Goal:** Browse the files area comfortably, even with very large folders.
- **Scope:** list and grid views; breadcrumbs; folder navigation; sorting; virtualized lists for very large folders; empty states.
- **Deliverables:** file browser views.
- **Depends on:** S02.2, S01.3.
- **Requirements:** FR-002; FR-214, FR-215 (1.5.0).
- **Acceptance criteria:**
  1. A folder of 50,000 items scrolls smoothly and loads pages on demand.
  2. Sorting by name, size, date, and type matches the API order.
  3. The URL reflects the current folder, and reloading restores it.
  4. Empty and error states show a clear next action.
- **Risks/notes:** Includes a **prototype task** confirming that @tanstack/svelte-virtual works with Svelte 5 (ADR-0009), with a custom windowing fallback. Performance on low-end phones is tested in S02.7.
- **Follow-up (1.5.0, the user's walkthrough in S007):** folder sizes (FR-214) and the added and modified dates (FR-215) in the list and the grid, with a new API endpoint for a folder's size and an `added_time` field and sort key. Task S02.3-T05 in `stages/S02-nas-gui.md`; the substage stays Done.
- **Status:** Done (S006, 2026-09-25; details in `stages/S02-nas-gui.md`)

#### S02.4: Uploads and downloads
- **Goal:** Easy, robust uploads and downloads from the GUI.
- **Scope:** button and drag-and-drop upload of files and folders; upload queue with progress, pause, resume, and cancel (using S01.4); single and multi-item download (streamed zip).
- **Deliverables:** upload manager (tus client); streamed ZIP endpoint (server addition); download actions.
- **Depends on:** S02.3, S01.4.
- **Requirements:** FR-003, FR-004, FR-006, FR-080; FR-216 (1.5.0).
- **Acceptance criteria:**
  1. Dropping a folder tree uploads it with its structure.
  2. Pause, resume (including after a page reload or network drop), and cancel work for large files.
  3. Multi-item download streams a ZIP without the server buffering it in memory.
  4. Per-file errors (limit, conflict, disk full) are shown clearly.
- **Risks/notes:** Browser support for folder drag-and-drop varies (checked in S02.8).
- **Follow-up (1.5.0, the user's walkthrough in S007):** finished uploads refresh the folder on screen, folder uploads too (bug S02-B12), and a single uploaded item is scrolled into view and blinks twice (FR-216). Task S02.4-T05 in `stages/S02-nas-gui.md`; the substage stays Done.
- **Status:** Done (S006, 2026-09-25; details in `stages/S02-nas-gui.md`)

#### S02.5: File operations UI
- **Goal:** Every file operation available from the GUI, individually and in bulk.
- **Scope:** create folder, rename, move, copy (with a folder picker), delete with confirmation; multi-select; context menus; keyboard shortcuts; name-conflict dialogs.
- **Deliverables:** operation dialogs, folder picker, selection model, shortcut map.
- **Depends on:** S02.3, S01.3, S01.6.
- **Requirements:** FR-007, FR-021, FR-077, FR-081.
- **Acceptance criteria:**
  1. All S01 operations work from the GUI on single items and multi-selections.
  2. Name conflicts open a dialog (skip, rename, overwrite), and the choice is passed to the API.
  3. Delete always asks for confirmation and states that deletion is permanent (no trash until S08).
  4. The main actions have keyboard shortcuts and context-menu entries.
- **Risks/notes:** Permanent delete until S08 is a data-loss risk (RK-19).
- **P008 (1.9.0):** follow-up F4, built in S03: a Trash page (list, restore, delete permanently, empty trash) and delete dialogs that say items go to the trash (FR-354).
- **Status:** Done (S006, 2026-09-25; details in `stages/S02-nas-gui.md`)

#### S02.6: File previews
- **Goal:** Preview common file types safely inside the GUI.
- **Scope:** previews for images, text and code, PDF, audio, and video (streamed via range requests); a clear fallback for unsupported types.
- **Deliverables:** preview components (with a bundled PDF viewer); safe serving headers for previews.
- **Depends on:** S02.3, S01.3.
- **Requirements:** FR-082, NFR-022 (safe defaults ahead of S03).
- **Acceptance criteria:**
  1. Supported types preview inline, and audio and video seek via range requests without a full download.
  2. SVG, HTML, and other active content never execute scripts in the app's origin (tested).
  3. Large text files preview only a bounded first portion.
  4. Unsupported types show a fallback with a download action.
- **Risks/notes:** The full CSP arrives in S03.5. S02 already serves active content as attachment or sandboxed. The video player is built so that S04.8 can add HLS playback and the quality menu without replacing it (ADR-0020).
- **Status:** Done (S006, 2026-09-25; details in `stages/S02-nas-gui.md`)

#### S02.7: Responsiveness and accessibility
- **Goal:** The GUI works well on phones and tablets and for keyboard and screen-reader users.
- **Scope:** phone and tablet layouts; full keyboard navigation; screen-reader labels; sufficient color contrast.
- **Deliverables:** responsive layouts; accessibility fixes; automated accessibility checks in CI.
- **Depends on:** S02.2–S02.6.
- **Requirements:** NFR-015.
- **Acceptance criteria:**
  1. All S02 flows work at 360 px phone width and at tablet widths.
  2. Every action is reachable by keyboard with a visible focus indicator.
  3. Automated accessibility checks report no serious violations on the main screens.
  4. Contrast meets WCAG 2.1 AA in both themes.
- **Risks/notes:** None.
- **Status:** Done (S007, 2026-09-28; details in `stages/S02-nas-gui.md`)

#### S02.8: Testing and stage review
- **Goal:** Prove that a non-technical user can do everything from S01 through the GUI, then close the stage.
- **Scope:** component tests; end-to-end tests of the main user flows; cross-browser check; documentation; completion record; user sign-off.
- **Deliverables:** component and end-to-end suites in CI; cross-browser report; GUI user guide section; completion record.
- **Depends on:** S02.1–S02.7.
- **Requirements:** NFR-027, NFR-014.
- **Acceptance criteria:**
  1. End-to-end tests cover browse, upload (with resume), download, rename, move, copy, delete, and preview, and pass in CI.
  2. Cross-browser check done on current Chrome, Edge, and Firefox. Safari is not checked in S02 (the user's decision, S006).
  3. The GUI user guide section is written.
  4. The completion record is written and the user's sign-off is recorded.
- **Risks/notes:** Safari testing may need a Mac or a cloud device service. That is a local tooling question, not a runtime dependency.
- **Status:** Done (S007, 2026-09-29; details in `stages/S02-nas-gui.md`)

**Design notes (S02):**
- The GUI uses only the public API, so S03 adds authentication without GUI rewrites.
- The component library and API client are reused by every later GUI substage (S03.8, S04.7, S05.5, S06.7, S07.6, S08.7, S09.5, S10, S17.8).
- The app is still localhost-only.

**Exit criteria (quoted):** "A non-technical user can do everything from S01 through the GUI."

---
