# CURRENT_STATE

**Last updated:** 2026-09-28 11:53 +0500 (session S007)
**Plan version:** 1.4.1 (`code-agent-docs/plan.md`), **Approved baseline** 1.0.0 (S005); 1.4.0 adds P005 (15 stages, 120 substages; packaging is now S13 and AI S15, table in plan 10.18)
**Current phase:** **S01 (Basic NAS) Done** (signed off, S005 E126; P005 follow-up tasks open). **S02 (NAS GUI) In Progress** (approved S006 E009); S02.1–S02.7 Done; S02.8 remains (after the S01 follow-ups)

## Active stage and task
- **Active stage:** **S02: NAS GUI**, **In Progress** (`stages/S02-nas-gui.md`; approved in S006 E009, with "Skip Safari": S02 checks Chrome, Edge, and Firefox). S01 is **Done** (`stages/S01-basic-nas.md`, completion record in section 13).
- **Active task:** none. Next: the last **S01 follow-up task** S01.4-T07 (content hash for resumable uploads; S01.2-T07 and S01.3-T10 are done), then **S02.8**.
- **The user's instruction (S007 E010):** after P005, resume S02 and work until the stage is complete, and also build the P005 items that belong to the current or earlier stages (the S01 follow-up tasks).

## In progress (write-ahead)
- none.

## Last completed
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
1. **Still open from the P005 report** (not blocking S02): Q42–Q49 and Q51; accepting or removing the planner additions (labelled in plan section 3); ADR-0022–ADR-0029 (Proposed); re-confirmation of S01 with its follow-ups; the RAID 0/1 reading of E008. ADR-0021 and the CI trigger were decided in E013.
2. **S02.8** (`stages/S02-nas-gui.md`, section 5), after the S01 follow-ups: (tests, including the regression tests for bugs S02-B01 to B10; the cross-browser report; the guide; audit A003; CI at stage completion; the user's sign-off with a hands-on walkthrough).
3. **S01 follow-ups** (P005, Q50: now): S01.2-T07 portable storage root (no ADR needed); S01.3-T10 and S01.4-T07 content hash (ADR-0021 Accepted: SHA-256). Each writes its own tests (S01's testing substage is closed).
4. Follow-up from S01 (the user's decision): run `scripts/perf-baseline.sh` on the Raspberry Pi and the mini-PC when available; also measure SHA-256 there (ADR-0021).
5. **Endpoint workflow (spec-first):** spec → `go generate ./internal/api` → strict operation; an error case in `errorCases`/`bodyErrorCases`; a review row in `docs/api/conventions.md`; a fake-service test that invalid input never reaches the service.
6. **CI (the user's preference, S007 E002, E013):** CI runs only when a stage-completion tag `S<NN>-done` is pushed (create it on `develop` after the stage's testing substage, before sign-off) or by hand. Run the checks locally before each commit and merge. For the stage-end run: `gh` is not installed; watch a run through the public REST API (`/repos/KhizirFarrukh/local-ai-nas/actions/runs?head_sha=<sha>`, then `/jobs`) or its web page. Reproduce Linux failures with `GOOS=linux go test -c` binaries in WSL.
7. Every finished branch: merge it into `develop` myself (`--no-ff`) and push (RULES User Preferences).

## Blocked or waiting on user
- **Stray folder `C:\c`** (holds only an empty `Users` tree, left by an S006 command): deleting it was blocked by a safety check, so the user deletes it.

## Open questions (short list; full text in plan.md section 5)
- **New in 1.4.0 (P005):** Q42 stack cover · Q43 cross-area duplicates · Q44 retention of replaced originals · Q45 upload default for exact duplicates · Q46 shortcuts over shares · Q47 mdadm and Linux-only pools · Q48 video codec · Q49 who applies optimization policies · Q51 confirm M3 = S01–S13. **Q50 answered** (S007 E010: now).
- **Partly open:** Q5 native Windows/macOS installers (S13.2) · Q6 AI speed expectations · Q26 image formats (HEIC/RAW) · Q32 photos exposure over shares
- **Open for later stages:** Q10, Q11, Q13, Q14, Q15, Q19, Q27–Q31, Q33–Q36, Q39, Q40, Q41
- **Answered in S005:** Q1 (multi-platform), Q16 (closed), Q18 (100k photos + 100k files), Q22 (AGPL-3.0), Q37 (none), Q38 (S01–S11 in the IDs of that time, through the pre-AI release; since 1.4.0 that release is S13, so M3 is S01–S13, to confirm in Q51)

## Pointers
- Latest session log: `code-agent-docs/logs/sessions/2026-09-28_S007.md` (current); S006 is closed
- Stage documents: `code-agent-docs/stages/S02-nas-gui.md` (**In Progress**); `code-agent-docs/stages/S01-basic-nas.md` (**Done**; P005 follow-up tasks at the end of section 5; completion record in section 13)
- Audit reports: `code-agent-docs/audits/A002-2026-09-24-documentation-audit.md` (the S01 final review); `A001-2026-09-24-documentation-audit.md`
- ADRs: `code-agent-docs/decisions/ADR-0001` … `ADR-0029` (0019 and 0022–0029 Proposed; 0021 Accepted in S007; 0012 superseded in part by 0020; the rest Accepted)
- Dependency register: `code-agent-docs/dependencies.md` (section 12: deployment prerequisites per platform, the input for the S13.2 setup scripts)
- Rules: `code-agent-docs/RULES.md` (v1.7.1: invariant I10; CI only at stage completion, on a tag `S<NN>-done` or by hand) · Prompts: `code-agent-docs/prompts/` (P002–P005)
