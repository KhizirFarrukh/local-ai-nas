# CURRENT_STATE

**Last updated:** 2026-09-25 07:27 +0500 (session S006)
**Plan version:** 1.3.0 (`code-agent-docs/plan.md`), **Approved baseline** 1.0.0 (S005) + the setup-script requirement (1.1.0, S005 E015)
**Current phase:** **S01 (Basic NAS) Done**, signed off by the user (S005 E126). **S02 (NAS GUI) In Progress** (approved S006 E009); S02.4-T01 done

## Active stage and task
- **Active stage:** **S02: NAS GUI**, **In Progress** (`stages/S02-nas-gui.md`; approved in S006 E009, with "Skip Safari": S02 checks Chrome, Edge, and Firefox). S01 is **Done** (`stages/S01-basic-nas.md`, completion record in section 13).
- **Active task:** **S02.4-T01** (branch `feat/S02.4-T01-uploads`): done, waiting for CI and the merge; next **S02.4-T02** files, folders, drag and drop

## In progress (write-ahead)
- S02.4-T01: committed on `feat/S02.4-T01-uploads`; push, CI, merge into `develop` (branches may be stacked while CI runs).

## Last completed
- **S02.1-T01 done** (S006 E011): `web/` SvelteKit project (exact versions, pnpm 12.6.0); format, lint, check, and build pass on Windows and Linux (WSL); packages registered; Node.js 24.19.0 on Windows (newest in winget), 24.21.0 in WSL.
- **S02 plan merged** (CI run 36056943504 green; `develop` a3b4b87).
- **S02 approved** (S006 E009): "Approve S02 (Recommended)", and "Skip Safari" for the cross-browser checks. Plan 1.3.0.
- **Testing approach change merged** (CI run 36054572413 green; `develop` bdc9832).
- **Testing approach changed** (S006 E001–E004), at the user's instruction:
  - Code first, written to be testable. Each stage's final testing substage writes its unit, integration, and system/application tests.
  - Coverage is reported on every push but not enforced until the stage end (80%).
  - A bug is fixed at once, and its test is written at the stage end.
  - RULES 1.6.0 (R6), plan 1.2.0, the stage template, `docs/testing.md`, and CI updated.
- **S01.7-T08 done and merged** (CI run 36044157203 green; `develop` e5311a8; S005 E127).
- **S01 Done: the user signed off** (S005 E126): "Sign off S01 (Recommended)". The throughput deviation was accepted ("Accept, measure later (Recommended)"), and README R-11/R-12 were approved and applied ("Apply both (Recommended)"). Plan 1.1.4.
- **S01.7-T07 done and merged** (CI run 36042107504 green; S005 E123, E124): audit A002, 11 findings (1 Critical and 2 Major fixed; 7 Minor fixed; 1 accepted; the register now lists the system tools); plan 1.1.3; README proposal R-11/R-12 for the sign-off.
- **S01.7-T06 done and merged** (CI run 36040515533 green, both `demo` jobs included; S005 E121): `scripts/demo.sh` and `scripts/demo.ps1` are green on Windows (PowerShell 5.1, recorded) and on Linux, and CI job `demo` runs both. tusd's `BodyReadError` is now logged as WARN, and scripts that were not executable are fixed.
- **S01.7-T05 done and merged** (CI run 36039328692 green; S005 E119): `docs/api/usage.md`, every operation with curl for Linux/macOS and Windows PowerShell, each block run as written on Linux (WSL) and in Windows PowerShell 5.1; README "Use the API"; docs and scripts indexes.
- **S01.7-T04 done and merged** (CI run 36037945955 green; S005 E117): `docs/perf/S01-baseline.md`: listing and memory targets met; throughput over loopback on NVMe below 80% (a deviation for the user's review at sign-off); downloads now keep sendfile.
- **Q18 answered** (S005 E116): 100,000 photos + 100,000 files per installation (confirms A12); plan 1.1.2.
- **S01.7-T03 done and merged** (CI run 36030776899 green; S005 E115): Unicode, empty, sparse (Linux), deep, and 10,000-entry cases end to end; a Range on an empty file no longer yields Go's invalid `Content-Range`.
- **S01.7-T02 done and merged** (CI run 36030178058 green; S005 E113): the attack corpus end to end at every operation and tus; malicious names, header injection, oversized inputs, tus metadata abuse; fixes for `ENAMETOOLONG` (500 → 404) and the Windows colon rule.
- **S01.7-T01 done and merged** (CI run 36029417608 green; S005 E111): `TestIntegration` runs the real program and reaches 81 operation statuses over HTTP (every declared one except 7 listed with reasons); the spec's wrong tus HEAD 412 was removed.
- **S01.5 closed and merged** (CI run 36028388074 green; S005 E109): T06 offline docs (vendored Redoc 2.5.4, checksum registered; `/api/docs/`; offline render checked with headless Edge); the tus server refuses foreign upload IDs with 404 (fuzz finding); all 5 criteria checked (S01 change log).
- **S01.5-T05 done and merged** (CI run 36004028905 green; S005 E107): `TestErrorStatusesAreInTheSpec` and `TestTusStatusesAreInTheSpec` (every status a case gets is declared); the tus operations complete in the spec; the CI drift check verified by a deliberate drift (red) and its fix (green); a missing operation does not compile (checked).
- **S01.4 closed and merged** (S005 E105): T06 cleanup (`uploads.expiry`, `Server.Cleanup`, `files.CleanTemp`, `internal/schedule`); closure test `TestResumeAtAnyPoint`; all 5 criteria checked (S01 change log).
- **S01.4-T05 done** (S005 E104): `chunkLimit` refuses a tus request declaring more than `uploads.max_chunk_size` with 413 before tusd reads it.
- **S01.4-T04 done and merged** (CI run 36001374478 green, including the 10 GiB memory job; S005 E103): `TestMemoryBound` (simple upload, tus, copy, download; heap growth < 256 MiB) in a CI `memory` job, 10 GiB on Linux and 1 GiB on Windows.
- **S01.4-T03 done and merged** (CI run 36000680420 green; S005 E100): finished uploads become files through `files.Local.CommitUpload` (MoveInto + shared commit), the copy fallback across file systems, SHA-256 check, `Item-Path` header, fault-injection tests; tus rows in `TestConflictMatrix`.
- **S01.4-T02 done and merged** (CI run 35999796296 green; S005 E098): the creation hook runs `files.Local.CheckUpload` (the simple upload's checks, shared); refusals are problems and store nothing.
- **S01.4-T01 done and merged** (CI run 35999368786 green; S005 E096): `internal/uploads` embeds tusd v2.10.1 at `/api/v1/files/uploads/` (sessions in SQLite, problems for every error, codes `locked` and `unavailable`); FuzzAPI covers it; a Windows path-aliasing fix in the resolver.
- **S01.6 closed and merged** (CI run 35996941498 green; S005 E093): T06 bind guard (loopback only; host names resolved; container-only exception `server.allow_container_bind`); Unicode look-alike gap fixed (resolver + name rules + corpus); all 5 criteria checked (S01 change log).
- **S01.6-T05 done and merged** (CI run 35995930046 green; S005 E091): `storage.Locks`; the files service locks the target folder during commits only; downloads retry a replaced file; `TestStressConcurrentWriters` (50 writers × 3 policies with readers: one consistent result, no partial or temporary files).
- **S01.6-T04 done and merged** (S005 E090): `TestConflictMatrix` (72 cases) and `TestConcurrentRenameCollisions` (every operation); tus finalize joins with S01.4-T03.
- **S01.6-T03 done and merged** (CI run 35994898119 green; S005 E088): links are never followed by any operation (a path through a link → 400, inside or outside the area), listed without a target, never created by the API (architecture test); link tests run wherever the OS allows links.
- **S01.3 closed and merged** (CI run 35994100122 green; S005 E085): T09 delete (`DELETE /api/v1/files/items`, `recursive`); closure tests over a real server (every operation) and for confinement (10 operations × escape paths and links; sentinels unchanged); `storage.IsEscape` maps os.Root escape refusals to `outside_root` (they were 500). All 5 substage criteria checked (S01 change log).
- **S01.3-T08 done and merged** (CI run 35993223778 green; S005 E083): `POST /api/v1/files/operations/copy`: a scan first (limits → 422 `too_large_for_sync`, name rules, links refused, free space), a hidden temp file or temp tree committed in one step, times kept; settings `copy.sync_max_items`/`sync_max_bytes`; `renameIfFree` fixes a Windows race in folder renames (also for T07 moves).
- **S01.3-T07 done and merged** (CI run 35992364545 green; S005 E080): rename and move endpoints; folder-into-itself refused through `os.SameFile` ancestors (catches case variants on Windows); case-only renames; atomic no-replace for files (hard link + remove, undone if the remove fails); folders never replaced or merged.
- **S01.3-T06 done and merged** (CI run 35972700242 green; S005 E077): `GET /api/v1/files/content` with ranges and conditional requests via `http.ServeContent`, 412/416 as problems (new codes `precondition_failed`, `range_not_satisfiable`), `attachment` Content-Disposition (RFC 6266/8187), nosniff + sandbox CSP; details from the open handle.
- **S01.3-T05 done and merged** (CI run 35971985758 green; S005 E074): `PUT /api/v1/files/content`: declared size checked first (411/413/507 before any byte), temp file in the target folder + fsync + atomic commit (Link for fail/rename, Rename for overwrite), hidden temp names, per-route body limits; the simple-upload part of S01.4-T05 done early.
- **S01.3-T04 done and merged** (CI run 35970636485 green; S005 E072): `POST /api/v1/files/folders` with `parents` and `on_conflict` (fail/rename/overwrite); every new name checked before anything is created; strict JSON bodies (`internal/api/body.go`); the resolver reports Windows device names with the reserved_name rule.
- **S01.3-T03 done and merged** (CI run 35969766390 green; S005 E070): item details with MIME (extension, else sniff) and a strong ETag (size + mtime + file ID) that changes on rewrite and on replacement.
- **S01.3-T02 done and merged** (CI run 35969161697 green; S005 E068): `GET /api/v1/files/items` (folder page or file details), cursor pagination and 4 sort keys; 10,000-entry acceptance test; runtime v1.7.0 linked.
- **S01.3-T01 done and merged** (CI run 35968332439 green; S005 E066): `files.Service` + hooks + `Local.Stat`; the API reaches files only through the interface (depguard + architecture test, both verified against a deliberate violation).
- **S01.5-T04 done and merged** (CI run 35967714415 green; S005 E064): spec-first pipeline (oapi-codegen v2.8.0 → `internal/api/gen`), strict GetHealth, error hooks → problems, spec↔route test, FuzzAPI (2.24 M requests, no 5xx).
- **S01.5-T03 done and merged** (CI run 35967063660 green; S005 E062): Problem schema and error responses in `api/openapi.yaml`; 404/405 are problems; the contract test validates every error response and requires an error case per route.
- **S01.5-T02 done and merged** (CI run 35966560636 green; S005 E060): `docs/api/versioning.md`; `internal/api` route table + `New`; every route is under /api/v1 or /api/docs (test).
- **S01.5-T01 done and merged** (CI run 35966117590 green; S005 E058): `docs/api/conventions.md` with the review checklist.
- **S01.6-T02 done and merged** (CI run 35965767428 green; S005 E056): `storage.ValidateName` / `ValidateNewPath` with 9 rules reported in the problem's `rule` field; FuzzValidateName proves every accepted name is creatable on NTFS (136 k execs) and Linux.
- **S01.6-T01 done and merged** (CI run 35965241227 green; S005 E054): path rules (UNC, drive letters, encoded separators, dot/space runs, NFC) with a 56-case attack corpus, an os.Root bypass test, and FuzzResolve (3.49 M execs clean).
- **S01.2 closed and merged** (CI run 35964600803 green; S005 E052): T06 done (`/api/v1/photos` → 501 not_available); all 6 S01.2 tasks Done; substage acceptance checked (S01 change log).
- **S01.2-T05 done and merged** (CI run 35964271426 green; S005 E050): health checks config, storage_writable, same_filesystem (warn + upload_finalize_mode=copy on a split), free_space, database; run at startup and on the health endpoint.
- **S01.2-T04 done and merged** (CI run 35950125517 green; S005 E048): `storage.DiskFree` (Statfs / GetDiskFreeSpaceEx) and `SpaceGuard` (507 insufficient_storage before any byte is stored).
- **S01.2-T03 done and merged** (CI run 35949731329 green; S005 E046): `storage.Resolver` (Resolve with safe path rules; OpenRoot as the second layer), `files.Item` with OwnerID, and the architecture test of `internal/files`.
- **S01.2-T02 done and merged** (CI run 35949441925 green; S005 E044): `storage.db_dir` / `storage.logs_dir` relocation; `Layout.Init` rejects every overlap of internal data with the areas (and with uploads) before creating anything.
- **S01.2-T01 done and merged** (`develop` 843c6fd; CI run 35948981464 green; S005 E042): `internal/storage` Layout + Init (creates `files/u0001`, `photos/u0001`, `.local-ai-nas/{tmp/uploads,db,logs}`; real, writable directories; unknown root entries reported and left alone); `serve`/`migrate` use it.
- **S01.1 closed** (S005 E040): T06 done (Docker dev environment, example config, README Development, CI image job; CI run 35948272549 green). All 11 S01.1 tasks Done; substage acceptance checked (S01 change log).
- **S01.1-T11 done and merged** (branch CI green; S005 E037): `cmd/local-ai-nas` (serve, migrate up|status, version; graceful shutdown) and `internal/health`; the smoke test runs the program as a separate process; a manual run on Windows is healthy.
- **S01.1-T10 done and merged** (branch CI green via the badge; S005 E035): `internal/db` (WAL, writer + query-only readers, goose migrations, schema settings/uploads); licenses all allowed; no vulnerabilities.
- **S01.1-T09 done and merged** (branch CI green; S005 E033): `internal/apperr` (kinds with stable codes, RFC 9457 problems, Write, Recover); 100% package coverage.
- **S01.1-T08 done and merged** (branch CI green; S005 E031): `internal/logging` (RotatingFile, New with stderr + file, RequestID, AccessLog with redaction); coverage of internal/... 94.8%.
- **S01.1-T07 done and merged** (`develop` 69a0f7b; branch CI green; S005 E029): `internal/config` (settings table; file/env/flag precedence; strict keys; typed values; all errors name key + source); coverage of internal/... 95.4%.
- **S01.1-T05 done and merged** (`develop` f662664; S005 E027): CI green on GitHub (run 35945085750, 7 jobs), red on a deliberate failing test (run 35945229679), green again after the revert (run 35945387694); arm64/amd64/windows artifacts produced.
- **S01.1-T04 done and merged** (`develop` dc9c0a0; S005 E025): `internal/testutil` + tests (unit, integration, fuzz seeds), `scripts/coverage.sh` (83.0%, gate works), `docs/testing.md`; 45 s of fuzzing found nothing.
- **S01.1-T03 done and merged** (`develop` f0cc48c; S005 E023): `.golangci.yml` (v2) + `scripts/install-golangci-lint.sh`; clean on the skeleton; depguard, errcheck, and gofmt violations fail (then removed).
- **S01.1-T02 done and merged** (`develop` b0cc2ec; S005 E021): `LICENSE` (AGPL v3), README License section, `docs/licensing.md`, `scripts/allowed-licenses.txt`, `scripts/check-licenses.sh`; the check passes, and a fixture with an Unlicense license fails it (then reverted).
- **S01.1-T01 done and merged** (`develop` 0100920; S005 E019): `go.mod` (go 1.27, toolchain go1.27.1; oapi-codegen, govulncheck, go-licenses as `tool` directives), skeleton packages, `api/openapi.yaml` stub, `.gitattributes`/`.editorconfig`/`.gitignore`. Build, vet, and test pass on Windows; Linux amd64/arm64 cross-build + vet pass; renormalize makes no changes.
- Approval-stage decisions D-01–D-14 applied (S005 E007–E010). **Plan 1.0.0 baseline and S01 approved** (S005 E012).
- `develop` verified complete (S005): merge `c535cda` brought `cf60f72` (plan 0.4.0, which PR #3 had put on `main` only) and the user's `bda8321` (prompts 3 and 4).
- Plan 1.1.0: the user's requirement to record every dependency and build a setup script per platform (FR-149, NFR-032, S11.2, Q41; `dependencies.md` section 12; RULES 1.5.0) (S005 E015–E016).

## Next steps
1. **S02 in order** (`stages/S02-nas-gui.md`, section 5):
   - S02.1-T01 toolchain and project (install pnpm 12.6.0; Node.js 24 LTS; create `web/`; register every package), on `feat/S02.1-T01-toolchain`;
   - then T02 embedding and serving, T03 API client, T04 design system, T05 CI;
   - then S02.2 to S02.7, and S02.8 (tests, report, guide, audit A003, sign-off).
   - **Testing approach (R6, S006):** tasks in S02.1–S02.7 deliver testable code only, checked by running them. S02.8 writes the unit (Vitest), integration, and system (Playwright: Chromium, Firefox, Edge) tests, and the regression tests for bugs recorded during the stage (section 12).
2. Follow-up from S01 (the user's decision): run `scripts/perf-baseline.sh` on the Raspberry Pi and the mini-PC, over gigabit Ethernet as well, when they are available (`docs/perf/S01-baseline.md`).
3. **Endpoint workflow (spec-first), for later API work:** spec → `go generate ./internal/api` → strict operation; an error case in `errorCases`/`bodyErrorCases`; a review row in `docs/api/conventions.md`; a fake-service test that invalid input never reaches the service.
4. CI: every push runs `.github/workflows/ci.yml`. `gh` is not installed; the repository is public. Watch a commit's run through the public REST API (`/repos/KhizirFarrukh/local-ai-nas/actions/runs?head_sha=<sha>`, then `/jobs`; 60 anonymous requests per hour) or the run's web page (its "Status" field). Step logs need sign-in: reproduce Linux failures with `GOOS=linux go test -c` binaries in the WSL Ubuntu distro, run from the package directory under /mnt/c when a test reads repo files.
5. Every finished branch: merge it into `develop` myself and push (RULES User Preferences, S005).

## Blocked or waiting on user
- Nothing.

## Open questions (short list; full text in plan.md section 5)
- ★ **For S01:** none (Q18 answered in S005: 100k photos + 100k files)
- **Partly open:** Q5 native Windows/macOS installers (S11.2) · Q6 AI speed expectations · Q26 image formats (HEIC/RAW) · Q32 photos exposure over shares
- **Open for later stages:** Q10, Q11, Q13, Q14, Q15, Q19, Q27–Q31, Q33–Q36, Q39, Q40
- **Answered in S005:** Q1 (multi-platform), Q16 (closed), Q18 (100k photos + 100k files), Q22 (AGPL-3.0), Q37 (none), Q38 (S01–S11)

## Pointers
- Latest session log: `code-agent-docs/logs/sessions/2026-09-25_S006.md` (current); S005 is closed
- Stage documents: `code-agent-docs/stages/S02-nas-gui.md` (**In Progress**); `code-agent-docs/stages/S01-basic-nas.md` (**Done**; completion record in section 13)
- Audit reports: `code-agent-docs/audits/A002-2026-09-24-documentation-audit.md` (the S01 final review) and `A002-readme-proposal.md`; `A001-2026-09-24-documentation-audit.md` (decisions received in section 9)
- ADRs: `code-agent-docs/decisions/ADR-0001` … `ADR-0020` (0019 Proposed; 0012 superseded in part by 0020; the rest Accepted, including 0003 since S005)
- Dependency register: `code-agent-docs/dependencies.md` (section 12: deployment prerequisites per platform, the input for the S11.2 setup scripts)
- Rules: `code-agent-docs/RULES.md` (v1.6.0: targeted plan reading; merge-into-develop; dependency record + per-platform setup scripts; code first, tests at the end of each stage) · Prompts: `code-agent-docs/prompts/` (P002–P004)
