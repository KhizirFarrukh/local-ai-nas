# CURRENT_STATE

**Last updated:** 2026-09-24 12:59 +0500 (session S005)
**Plan version:** 1.1.1 (`code-agent-docs/plan.md`), **Approved baseline** 1.0.0 (S005) + the setup-script requirement (1.1.0, S005 E015)
**Current phase:** Implementation: S01 (Basic NAS implementation) in progress; S01.1 and S01.2 done; S01.3, S01.5, and S01.6 in progress

## Active stage and task
- **Active stage:** S01 (Basic NAS implementation), **In Progress** (`stages/S01-basic-nas.md`); S01.1 and S01.2 **Done**; S01.6 **In Progress** (T01, T02 done; T03–T06 later); S01.5 **In Progress** (T01–T04 done; T05/T06 later); S01.3 **In Progress**.
- **Active task:** **S01.3-T06** Download (branch `feat/S01.3-T06-download`): done, waiting for CI and the merge; next **S01.3-T07** rename and move

## In progress (write-ahead)
- S01.3-T06: committed; push, CI, merge into `develop`.

## Last completed
- `develop` verified complete (S005): merge `c535cda` brought `cf60f72` (plan 0.4.0, which PR #3 had put on `main` only) and the user's `bda8321` (prompts 3 and 4).
- Approval-stage decisions D-01–D-14 applied (S005 E007–E010). **Plan 1.0.0 baseline and S01 approved** (S005 E012).
- **S01.3-T06 done** (S005 E077): `GET /api/v1/files/content` with ranges and conditional requests via `http.ServeContent`, 412/416 as problems (new codes `precondition_failed`, `range_not_satisfiable`), `attachment` Content-Disposition (RFC 6266/8187), nosniff + sandbox CSP; details from the open handle.
- **S01.3-T05 done and merged** (CI run 35971985758 green; S005 E074): `PUT /api/v1/files/content`: declared size checked first (411/413/507 before any byte), temp file in the target folder + fsync + atomic commit (Link for fail/rename, Rename for overwrite), hidden temp names, per-route body limits; the simple-upload part of S01.4-T05 done early.
- **S01.3-T04 done and merged** (CI run 35970636485 green; S005 E072): `POST /api/v1/files/folders` with `parents` and `on_conflict` (fail/rename/overwrite); every new name checked before anything is created; strict JSON bodies (`api/body.go`); the resolver reports Windows device names with the reserved_name rule.
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
- Plan 1.1.0: the user's requirement to record every dependency and build a setup script per platform (FR-149, NFR-032, S11.2, Q41; `dependencies.md` section 12; RULES 1.5.0) (S005 E015–E016).

## Next steps
1. **S01.3** (core file operations): T01 FilesService interface on os.Root with hooks (internal/files); then list (cursor pagination), details (MIME, ETag), create folder, simple streamed upload (PUT /content, free-space guard, name rules), download (Range/ETag), rename, move, copy, delete. **Spec-first:** each endpoint goes into `api/openapi.yaml` first, then `go generate ./internal/api`, then the strict operation (a missing one does not compile); add `github.com/oapi-codegen/runtime` v1.7.0 with the first operation that has parameters (register, R6). Each endpoint: an error case in `errorCases` (the contract test requires it), a review in `docs/api/conventions.md`, and fake-service tests that invalid input never reaches the service.
2. **S01.4-T06 must also clean stale `.local-ai-nas-tmp-*` files** (S01 change log, T05 row). Then S01.6-T03..T06 (remember the container bind note in the S01 change log), S01.4, S01.5-T05/T06, S01.7.
3. **Q18 (library size)** is needed before S01.7; ask the user when S01.7 comes close.
4. CI: every push runs `.github/workflows/ci.yml`. `gh` is not installed; the repository is public. Watch a commit's run through the public REST API (`/repos/KhizirFarrukh/local-ai-nas/actions/runs?head_sha=<sha>`, then `/jobs`; 60 anonymous requests per hour) or the run's web page (its "Status" field). Step logs need sign-in: reproduce Linux failures with `GOOS=linux go test -c` binaries in the WSL Ubuntu distro, run from the package directory under /mnt/c when a test reads repo files.
5. Every finished branch: merge it into `develop` myself and push (RULES User Preferences, S005).

## Blocked or waiting on user
- Q18 (library size) is needed before S01.7 only.

## Open questions (short list; full text in plan.md section 5)
- ★ **For S01:** Q18 library size (before S01.7)
- **Partly open:** Q5 native Windows/macOS installers (S11.2) · Q6 AI speed expectations · Q26 image formats (HEIC/RAW) · Q32 photos exposure over shares
- **Open for later stages:** Q10, Q11, Q13, Q14, Q15, Q19, Q27–Q31, Q33–Q36, Q39, Q40
- **Answered in S005:** Q1 (multi-platform), Q16 (closed), Q22 (AGPL-3.0), Q37 (none), Q38 (S01–S11)

## Pointers
- Latest session log: `code-agent-docs/logs/sessions/2026-09-24_S005.md`
- Active stage document: `code-agent-docs/stages/S01-basic-nas.md` (In Progress)
- Audit report: `code-agent-docs/audits/A001-2026-09-24-documentation-audit.md` (decisions received in section 9)
- ADRs: `code-agent-docs/decisions/ADR-0001` … `ADR-0020` (0019 Proposed; 0012 superseded in part by 0020; the rest Accepted, including 0003 since S005)
- Dependency register: `code-agent-docs/dependencies.md` (section 12: deployment prerequisites per platform, the input for the S11.2 setup scripts)
- Rules: `code-agent-docs/RULES.md` (v1.5.0: targeted plan reading; merge-into-develop; dependency record + per-platform setup scripts) · Prompts: `code-agent-docs/prompts/` (P002–P004)
