# S01: Basic NAS implementation

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](../PLAN_INDEX.md)). Moved unchanged from `plan.md` section 10.2 in version 1.9.0 (ADR-0044).

### 10.2 S01: Basic NAS implementation

- **Origin:** User-defined
- **Goal:** A reliable storage service that manages the files area through an API, with the two-area layout in place.
- **User requirements (quoted):**
  > "Stage 1 is basic NAS implementation."
  > "The NAS has two folders at the root: files and photos."
- **Scope note (from P002):** S01 creates both the `files/` and `photos/` roots but implements only the files area. The photos area is managed from S04. There is no GUI and no authentication yet, so the server must bind to localhost only.
- **Status:** **Done** (signed off by the user on 2026-09-24, S005; completion record in `stages/S01-basic-nas.md` section 13)

#### S01.1: Project foundation
- **Goal:** Establish the approved stack, repository, tooling, and conventions, so every later change is built, checked, and tested the same way.
- **Scope:**
  - Set up the Accepted stack (P003, 0.3.0): Go (ADR-0001), REST + OpenAPI spec-first (ADR-0002), repository layout (ADR-0004), testing, linting, and CI (ADR-0005), dev environment (ADR-0006), SQLite + goose migrations (ADR-0007).
  - Repository structure; dependency management; linting, formatting, type checking.
  - Test framework; CI pipeline (Linux + Windows); development environment (local + Docker dev setup).
  - Configuration system (config file + environment variables); structured logging; error-handling conventions.
  - LICENSE and dependency-license policy (Q22); `.gitattributes` line-ending policy (S001 finding); `.editorconfig`.
  - Carries over the old v0.1.0 "Stage 0" content.
- **Deliverables:**
  - Go module with pinned toolchain and tool directives; repository skeleton per ADR-0004.
  - SQLite database with migrations wired at startup (ADR-0007).
  - CI workflow; config, logging, and error modules.
  - Developer setup documentation.
- **Depends on:** plan baseline approval (1.0.0); S01 stage document approved; Q22 answered in S005: AGPL-3.0. The stack ADRs (0001, 0002, 0004–0007) were Accepted via P003, and ADR-0003 in S005.
- **Requirements:** NFR-008, NFR-009, NFR-013, NFR-014, NFR-016, NFR-025, NFR-029, NFR-030.
- **Acceptance criteria:**
  1. CI runs lint, format check, type check, tests, and a dependency-license check on every PR, on Linux and Windows, and a deliberately failing test turns it red.
  2. A fresh clone can be set up and the server started with the documented commands on Windows and Linux.
  3. Invalid configuration stops startup with a clear message, and environment variables override file values.
  4. Logs are structured, carry a request ID, and never contain secrets (tested).
  5. All errors use one documented format. Unexpected exceptions return a generic body with a correlation ID.
- **Risks/notes:** The stack ADRs were accepted in 0.3.0 (P003), and ADR-0003 and Q22 (AGPL-3.0) in S005. The remaining gate is the approval of the S01 stage document.
- **P008 (1.9.0):** follow-up F5, built in S03 before sessions: `synchronous=FULL` for the whole database with throttled high-frequency writes (NFR-056, ADR-0007 amendment).
- **Status:** Done (S005, 2026-09-24; details in `stages/S01-basic-nas.md`)

#### S01.2: Storage layout and configuration
- **Goal:** Create and validate the storage root with `files/` and `photos/`, place internal data outside both, and make the layout ready for per-user namespaces.
- **Scope:**
  - Configurable storage root; creation and validation of `files/` and `photos/` on startup.
  - Internal app data outside both areas (I2).
  - Per-user namespace layout and an owner on every item (ADR-0003).
  - Disk space checks; startup health checks and a health endpoint.
- **Deliverables:** accepted ADR-0003; layout initializer; namespace resolver; owner model; free-space guard; `GET /api/v1/system/health`.
- **Depends on:** S01.1.
- **Requirements:** FR-069, FR-070, FR-071, FR-072, NFR-026.
- **Acceptance criteria:**
  1. Starting against an empty root creates `files/`, `photos/`, the default namespace, and internal data. Starting again changes nothing.
  2. Configurations that place internal data inside an area (or an area inside internal data) are rejected at startup.
  3. Every resolved item carries an owner and a namespace, and no request can address outside its namespace.
  4. Writes that would push free space below the configured reserve are refused with a clear error.
  5. The health endpoint reports each startup check: root writable, temp and areas on one filesystem, free space, config valid.
- **Risks/notes:** The layout affects S07.2 and S09. ADR-0003 avoids a later data move by creating namespaces now.
- **P005 follow-up (1.4.0):** the storage root must not assume a single physical disk and must be relocatable, so it can later move onto a pool (S15.6, NFR-036). The upload temp folder and the trash always follow the root's filesystem (A18). Recorded as a follow-up task in `stages/S01-basic-nas.md` (S01.2-T07); S01 stays Done. **Done in S007** (Q50: `docs/guide/storage-root.md`).
- **Status:** Done (S005, 2026-09-24; details in `stages/S01-basic-nas.md`)

#### S01.3: Core file operations
- **Goal:** A service layer and API that perform every basic file operation on the files area, confined to it.
- **Scope:** list directory with pagination and sorting; item details (size, modified time, type); create folder; upload (single request); download with HTTP range support; rename; move; copy; delete. All confined to the files area.
- **Deliverables:** `FilesService` interface and local-filesystem implementation with before/after operation hooks; `/api/v1/files` endpoints.
- **Depends on:** S01.2. Also uses the path resolver and name validation from S01.6 and the API conventions from S01.5, which are therefore built first in the stage's execution order (see the stage document and 10.17).
- **Requirements:** FR-003, FR-005, FR-007, FR-073.
- **Acceptance criteria:**
  1. Every operation works over HTTP against a real temporary filesystem and has integration tests.
  2. A 10,000-entry folder lists in correct, stable pages for every sort order.
  3. Range requests return 206 with the correct bytes, and unsatisfiable ranges return 416.
  4. No operation can read or write outside the caller's files namespace (tested).
  5. Endpoints reach the filesystem only through the service interface (architecture test).
- **Risks/notes:** Copying large folders is synchronous in S01, with limits. It moves onto the job system in S04.3.
- **P005 follow-up (1.4.0):** simple uploads compute a content hash while the file streams, with no extra read, and store it with the item (FR-211). The algorithm is standard-library SHA-256 (ADR-0021, Accepted in S007). **Done in S007** (S01.3-T10, Q50).
- **P008 (1.9.0):** follow-ups F1 and F4, built in S03: stable item IDs with the items table, the backfill, and `id` in the API (FR-346, FR-347 on demand, FR-349; ADR-0040); minimal trash with restore and purge (FR-354).
- **Status:** Done (S005, 2026-09-24; details in `stages/S01-basic-nas.md`)

#### S01.4: Large file handling
- **Goal:** Upload and download files of any size reliably, with bounded memory, and never expose partial files.
- **Scope:**
  - Chunked, resumable uploads (tus via embedded tusd, ADR-0008); streaming I/O everywhere.
  - Configurable size limits; temp file then atomic rename.
  - Cleanup of abandoned uploads; optional checksum verification.
- **Deliverables:** resumable upload endpoints; upload session store in `.local-ai-nas/tmp/uploads/`; cleanup task behind a scheduler interface; limit configuration.
- **Depends on:** S01.2, S01.3, S01.5.
- **Requirements:** FR-004, FR-074, NFR-006, NFR-021.
- **Acceptance criteria:**
  1. An upload interrupted at any point resumes from the last confirmed offset, and the final file is byte-identical (hash verified).
  2. Uploading and downloading a 10 GB file raises server memory by less than a fixed bound (proposed: 256 MB).
  3. An in-progress upload never appears in listings or downloads, and a crash mid-upload leaves no partial file in `files/`.
  4. Uploads over the configured limit are refused before data is stored.
  5. Abandoned uploads are deleted after the configured expiry.
- **Risks/notes:** Atomic rename requires one filesystem (A18), which S01.2 checks.
- **P005 follow-up (1.4.0):** resumable uploads also store a content hash (FR-211). The hash state is serialized between chunks so the file is still read only once; if a resume cannot restore the state, the assembled file is hashed once before the atomic rename (8.26). **Done in S007** (S01.4-T07, Q50).
- **P008 (1.9.0):** follow-up F3, built in S03: the minimal durable job foundation replaces the temporary cleanup scheduler; long operations become jobs (`202 Accepted`); the operation journal with recovery at startup; idempotency keys (FR-351–FR-353, NFR-053; ADR-0011 amendment, ADR-0041).
- **Status:** Done (S005, 2026-09-24; details in `stages/S01-basic-nas.md`)

#### S01.5: API layer
- **Goal:** A consistent, versioned, documented HTTP API that later stages extend without breaking clients.
- **Scope:**
  - Resource design with separate route namespaces for files and (reserved) photos; `/api/v1` versioning.
  - Consistent error format (RFC 9457); request validation.
  - OpenAPI specification; API documentation served locally.
- **Deliverables:** route conventions document; error schema and codes; `api/openapi.yaml` (spec-first) with oapi-codegen-generated server interfaces committed; offline API docs page (vendored Redoc); CI spec-drift check (ADR-0002).
- **Depends on:** S01.1 (ADR-0002).
- **Requirements:** FR-075, NFR-001.
- **Acceptance criteria:**
  1. All endpoints are under `/api/v1`. Files endpoints are under `/api/v1/files`, and `/api/v1/photos` returns a documented "not available yet" error.
  2. Every error response validates against the documented error schema.
  3. Invalid input is rejected with 4xx before reaching the service layer (tested per endpoint).
  4. CI fails if the committed OpenAPI spec differs from the generated one.
  5. The API docs page works with the network disconnected.
- **Risks/notes:** The docs renderer is vendored and served by the core, with no CDN (I6, ADR-0002).
- **Status:** Done (S005, 2026-09-24; details in `stages/S01-basic-nas.md`)

#### S01.6: Safety baseline
- **Goal:** Path traversal is impossible, filenames are safe on every OS, concurrent operations are safe, and the server is reachable only from localhost.
- **Scope:**
  - Path normalization and traversal prevention.
  - Filename sanitization including names reserved on Windows, macOS, and Linux.
  - Symlink policy; name-conflict handling (fail / auto-rename / overwrite).
  - Concurrency safety on the same item; localhost-only binding by default.
- **Deliverables:** per-area path resolver; filename validator; enforced symlink policy; `on_conflict` parameter; per-path locking; bind-address guard.
- **Depends on:** S01.2.
- **Requirements:** FR-076, FR-077, NFR-010, NFR-019, NFR-020.
- **Acceptance criteria:**
  1. A traversal attack corpus is rejected on Linux and Windows: `..`, encoded variants, absolute paths, drive letters, UNC paths, NUL bytes, mixed separators, Unicode look-alikes.
  2. Names invalid on any supported OS are rejected with a clear reason: `CON`, `aux.txt`, `a:b`, a trailing dot or space, control characters, over-long names.
  3. Symlinks cannot be used to read or write outside the area.
  4. Concurrent writes to the same path produce one consistent result and no partial file (stress test).
  5. By default the server listens only on loopback. Configuring any other address is refused until S03 is Done.
- **Risks/notes:** The resolver and name validation are built early in S01's execution order because S01.3 depends on them.
- **P008 (1.9.0):** the protection against other websites (FR-350) is built first in S03 (S03.5-T02); the bind-address guard stays.
- **Status:** Done (S005, 2026-09-24; details in `stages/S01-basic-nas.md`)

#### S01.7: Integration, testing, and stage review
- **Goal:** Prove that S01 works end to end and is safe, then close the stage.
- **Scope:**
  - Integration tests against a real temporary filesystem.
  - Attack tests (traversal, malicious names).
  - Edge cases: Unicode names, empty files, very large files, deep nesting.
  - Basic performance check; documentation updates; completion record; user sign-off.
- **Deliverables:** integration, attack, and edge-case suites in CI; performance baseline report; README developer and API usage sections; scripted API demo; completion record.
- **Depends on:** S01.1–S01.6.
- **Requirements:** NFR-003 (S01 targets), NFR-014; verification of every S01 requirement.
- **Acceptance criteria:**
  1. All S01 tests pass in CI on Linux and Windows.
  2. The attack and edge-case suites cover every item listed in the scope.
  3. The performance baseline (listing, throughput, memory during a 10 GB transfer) is recorded against NFR-003.
  4. A scripted API demo manages files and folders end to end, including a resumed upload.
  5. The completion record is written and the user's sign-off is recorded.
- **Risks/notes:** Reference hardware (Q1, S005): the Windows 11 development PC, plus a Raspberry Pi and an x86-64 mini-PC when available. Library size: 100,000 photos + 100,000 files (Q18, S005).
- **Status:** Done (S005, 2026-09-24; details in `stages/S01-basic-nas.md`)

**Design notes (S01):**
- Storage access sits behind a service interface with hook points (trash, sharing checks, quotas, sidecar sync, indexing), so later features never touch every endpoint.
- Every item is modelled with an owner from the start, derived from its namespace, so S07 needs no core-model migration.
- SQLite (WAL) exists from S01 (ADR-0007): migrations, settings, and the upload-session index. Users and sessions tables are added by migrations in S03.2. tusd keeps its own upload data files in `<internal>/tmp/uploads/`.
- The cleanup scheduler sits behind an interface that the S04.3 job system implements later.
- `photos/` exists and is validated but has no API until S04.

**Exit criteria (quoted):** "Using only the API (e.g. an HTTP client), files and folders in the files area can be managed reliably, including large resumable uploads. All tests pass in CI."

---
