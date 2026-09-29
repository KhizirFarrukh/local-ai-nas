# CURRENT_STATE

**Last updated:** 2026-09-29 20:43 +0500 (session S007)
**Plan version:** 1.6.0 (`code-agent-docs/plan.md`; P006: MVP additions FR-217–FR-221 pending Q54, release roadmap R01–R12 in section 11b), **Approved baseline** 1.0.0 (S005); 1.4.0 adds P005 (15 stages, 120 substages; packaging is now S13 and AI S15, table in plan 10.18)
**Current phase:** **S01 (Basic NAS) Done** (signed off, S005 E126; P005 follow-up tasks done in S007). **S02 (NAS GUI) In Progress** (approved S006 E009); S02.1–S02.7 Done; S02.8 In Progress (T01–T04 done; T05 waits for the sign-off)

## Active stage and task
- **Active stage:** **S02: NAS GUI**, **In Progress** (`stages/S02-nas-gui.md`; approved in S006 E009, with "Skip Safari": S02 checks Chrome, Edge, and Firefox). S01 is **Done** (`stages/S01-basic-nas.md`, completion record in section 13).
- **Active task:** **S02.8-T05**: the stage-end CI run again (the walkthrough fixes are on `develop`), then the user's sign-off. The P005 follow-up tasks of S01 are all done.
- **The user's instruction (S007 E010):** after P005, resume S02 and work until the stage is complete, and also build the P005 items that belong to the current or earlier stages (the S01 follow-up tasks).

## In progress (write-ahead)
- **S02.8-T05: the stage-end CI must run again** before the S02 sign-off: `develop` has the walkthrough fixes (S02.3-T05, S02.4-T05). The tag `S02-done` points at 494e616; moving it again needs the user's permission (asked in S007 E042), or the user starts "Run workflow". Then: the completion record's test results and the cross-browser report's CI sentence, and the sign-off (walkthrough, Narrator check, README proposals R-13–R-15, S01 re-confirmation, and the P006 questions Q54 and Q34 first).

## Last completed
- **Stage-end CI green** (S007 E027–E028): run 36433983279, 15 of 15 jobs, on the tag `S02-done` (`develop` 494e616). The first run had failed in the Linux system tests because of two test races, fixed in `fix/S02.8-T05-e2e-resume-race`; the tag was moved with the user's permission.
- **S02.3-T05 and S02.4-T05 done** (S007 E042; the user's walkthrough): folder sizes (`GET /files/usage`), added and modified dates (`added_time`, sortable), uploads shown at once (bug S02-B12) and a single upload revealed with two blinks (`locate`); bug S02-B13 (a request storm from an effect) found by the new system tests and fixed. 210 web tests; system tests 142 passed, 2 skipped; Go coverage 89.7%.
- **Plan change request #6 applied** (S007 E034–E040; branch `docs/P006-release-roadmap`, merged into `feat/S02-walkthrough-fixes`): research R001 (18 services, 118 gaps); plan 1.6.0 (G13; NG5 changed; NG1–NG3 pending Q55–Q57; MVP additions FR-217–FR-221; releases R01–R12 with FR-222–FR-327 and NFR-040–NFR-043; sections 3.3, 11b, 11c; I11 proposed only); RULES 1.8.0 (R13 Releases, `research/` folder, keep-working rule); register section 13 (release candidates). No application code; no stage for R01–R12.
- **S02.8-T04 done** (S007 E024): documentation audit A003: 8 findings, no Critical; 6 fixed (plan 1.4.2), 1 accepted, 1 for the user (README proposals R-13–R-15, asked at the S02 sign-off).
- **S02.8-T03 done** (S007 E023): user guide `docs/guide/web-interface.md`; cross-browser report `docs/reports/S02-cross-browser.md` (Edge, Chromium, Firefox locally; Google Chrome in CI through `E2E_CHROME=1`); README Development section; register (test browsers, fixture tool).
- **S02.8-T02 done** (S007 E022): 42 Playwright system tests per browser (125 passed, 1 skipped in Chromium, Firefox, Edge), listing and ZIP64 integration tests, axe 0 violations in both themes; bug S02-B11 (a renamed item was no longer selected) fixed.
- **S02.8-T01 done** (S007 E021): 197 web unit and component tests (coverage of `web/src/lib` 95.4% statements, 82.9% branches), Go tests for archives and the web app (`internal/...` 89.9%), regression tests for S02-B01, B04, B07, B08, B09.
- **S01.4-T07 done; S01 follow-ups complete** (S007 E019): resumable uploads are hashed chunk by chunk (state saved per chunk, one full read only as a fallback); the client checksum no longer reads the file again.
- **S01.3-T10 done** (P005 follow-up; S007 E018): simple uploads and copies store a SHA-256 content hash (`content_hash` in item details), kept through rename, move, and delete, never reported for a file changed outside the NAS.
- **S01.2-T07 done** (P005 follow-up; S007 E017): the storage root can be copied or moved (`docs/storage-root.md`); unfinished uploads are repaired at start-up; bug S01-B01 (uploads locked for good on Windows) fixed with tusd's memory locker.
- **S02.7-T03 done; S02.7 closed** (S007 E016): token contrast computed for both themes (dark `danger` now #fa8585; preview arrows black/60); axe `color-contrast` clean on 11 screens × 2 themes × 2 widths in Edge and Firefox.
- **S02.7-T02 done** (S007 E015; merged into `develop`, CI at stage end): screen-reader names for items, live regions (notifications, selection, uploads), focus to the main area for empty folders, focus rings kept in Firefox (`focusFor`), level-one headings; axe clean in Edge and Firefox, both themes.
- **The user's decisions merged** (`develop` 2a371b3; S007 E013): ADR-0021 Accepted (SHA-256); CI only on a stage tag `S<NN>-done` or by hand (RULES 1.7.1; plan 1.4.1).
- **Plan change request #5 applied and merged** (`develop` 366d6da; S007 E003–E012): plan 1.4.0, RULES 1.7.0 (invariant I10; CI only at stage completion), ADR-0021–ADR-0029 (all **Proposed**), dependency register, audit checklist group K, S01 follow-up tasks (`stages/S01-basic-nas.md` section 5, end). New stages: **S11** duplicates, look-alike stacks, and bursts; **S12** storage optimization; **S14** drive pools, RAID 0 and 1 only (complex RAID deferred to plan 11a, the user's decision in S007 E008).
- **S02.7-T01 done and merged** (CI run 36112486230 green; `develop` 04ab2b4; S006 E040): phone and tablet layouts, 44 px touch targets, tap-to-open on touch.
- **S02.6 closed** (`develop` 617b1d9; S006 E034–E040): preview frame; image, audio, video; text and code; PDF with pdf.js; active-content safety.
- **S02.5 closed** (`develop` 90da220; S006 E030–E034): selection model (bug S02-B09), operation dialogs, conflict dialog, context menus and shortcuts.
- **S02.4 closed** (`develop` 0addb89; S006 E024–E030): upload manager (bugs S02-B06–B08), folder uploads, streamed ZIP archives, download actions.
- **S02.1–S02.3 closed** (`develop` 85d7154; S006 E011–E023): toolchain, embedding, API client, design system, CI (S02-B01, B02); shell, routes, errors, notifications (S02-B03–B05); virtualized list and grid views.
- **S02 approved** (S006 E009), plan 1.3.0. **Testing approach** (S006 E001–E004): code first, tests in each stage's final testing substage; RULES 1.6.0.
- **S01 Done** (S005 E126), plan 1.1.4: every task, CI run, and decision is in `stages/S01-basic-nas.md` (sections 12 and 13) and the S005 log.

## Next steps
1. **S02.8-T05** (`stages/S02-nas-gui.md`, section 5): run the stage-end CI again on `develop` (move the tag `S02-done` with the user's permission, or the user runs the workflow by hand); record its results in section 13 and the cross-browser report. Remaining: the user's sign-off with a hands-on walkthrough on this PC (including the Narrator listening check of S02.7-T02), README proposals R-13–R-15 (`audits/A003-readme-proposal.md`), and the re-confirmation of S01 with its follow-ups.
2. **Still open from the P005 report** (not blocking S02): Q42–Q49 and Q51; accepting or removing the planner additions (labelled in plan section 3); ADR-0022–ADR-0029 (Proposed); the RAID 0/1 reading of E008. ADR-0021 and the CI trigger were decided in E013. The S01 follow-ups are done (S007 E017–E019).
3. Follow-up from S01 (the user's decision): run `scripts/perf-baseline.sh` on the Raspberry Pi and the mini-PC when available; also measure SHA-256 there (ADR-0021).
4. **Endpoint workflow (spec-first):** spec → `go generate ./internal/api` → strict operation; an error case in `errorCases`/`bodyErrorCases`; a review row in `docs/api/conventions.md`; a fake-service test that invalid input never reaches the service.
5. **CI (the user's preference, S007 E002, E013):** CI runs only when a stage-completion tag `S<NN>-done` is pushed (create it on `develop` after the stage's testing substage, before sign-off) or by hand. Run the checks locally before each commit and merge. For the stage-end run: `gh` is not installed; watch a run through the public REST API (`/repos/KhizirFarrukh/local-ai-nas/actions/runs?head_sha=<sha>`, then `/jobs`) or its web page. Reproduce Linux failures with `GOOS=linux go test -c` binaries in WSL.
6. Every finished branch: merge it into `develop` myself (`--no-ff`) and push (RULES User Preferences).

## Blocked or waiting on user
- **Stray folder `C:\c`** (holds only an empty `Users` tree, left by an S006 command): deleting it was blocked by a safety check, so the user deletes it.

## Open questions (short list; full text in plan.md section 5)
- **New in 1.6.0 (P006), to ask at the next stop:** Q54 confirm the MVP additions (Live Photos and motion photos, phone auto-backup bridge, alert delivery) and **Q34 file versioning (recommended yes)** first; then Q52 AI position (after all releases, I8) · Q53 drive pools right after the MVP or at the end · Q55 NG1 reword · Q56 mobile apps (NG2) · Q57 non-destructive editing (NG3) · Q58 external libraries · Q59 map tiles · Q60 public accounts · Q61 exposure methods · Q62 locked folder · Q63 version labels · Q64 office engine · Q65 excluded features · Q66 invariant I11.
- **New in 1.4.0 (P005):** Q42 stack cover · Q43 cross-area duplicates · Q44 retention of replaced originals · Q45 upload default for exact duplicates · Q46 shortcuts over shares · Q47 mdadm and Linux-only pools · Q48 video codec · Q49 who applies optimization policies · Q51 confirm M3 = S01–S13. **Q50 answered** (S007 E010: now).
- **Partly open:** Q5 native Windows/macOS installers (S13.2) · Q6 AI speed expectations · Q26 image formats (HEIC/RAW) · Q32 photos exposure over shares
- **Open for later stages:** Q10, Q11, Q13, Q14, Q15, Q19, Q27–Q31, Q33–Q36, Q39, Q40, Q41
- **Answered in S005:** Q1 (multi-platform), Q16 (closed), Q18 (100k photos + 100k files), Q22 (AGPL-3.0), Q37 (none), Q38 (S01–S11 in the IDs of that time, through the pre-AI release; since 1.4.0 that release is S13, so M3 is S01–S13, to confirm in Q51)

## Pointers
- Latest session log: `code-agent-docs/logs/sessions/2026-09-28_S007.md` (current); S006 is closed
- Stage documents: `code-agent-docs/stages/S02-nas-gui.md` (**In Progress**); `code-agent-docs/stages/S01-basic-nas.md` (**Done**; P005 follow-up tasks at the end of section 5; completion record in section 13)
- Audit reports: `code-agent-docs/audits/A003-2026-09-28-documentation-audit.md` (the S02 final review, with `A003-readme-proposal.md`); `A002-2026-09-24-documentation-audit.md` (the S01 final review); `A001-2026-09-24-documentation-audit.md`
- ADRs: `code-agent-docs/decisions/ADR-0001` … `ADR-0029` (0019 and 0022–0029 Proposed; 0021 Accepted in S007; 0012 superseded in part by 0020; the rest Accepted)
- Dependency register: `code-agent-docs/dependencies.md` (section 12: deployment prerequisites per platform, the input for the S13.2 setup scripts)
- Research: `code-agent-docs/research/R001-2026-09-28-cloud-storage-feature-research.md` (P006)
- Rules: `code-agent-docs/RULES.md` (v1.8.0: R13 Releases; keep working, stop only when told or at a stage's end, stopping = save, commit, push; the agent's co-author line stays in commits, for transparency; invariant I10; CI only at stage completion, on a tag `S<NN>-done` or by hand) · Prompts: `code-agent-docs/prompts/` (P002–P005)
