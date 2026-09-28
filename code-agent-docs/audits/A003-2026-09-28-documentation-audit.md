# A003: Documentation audit (S02 final review)

| Field | Value |
|---|---|
| Audit ID | A003 |
| Date | 2026-09-28 |
| Session | S007 (`logs/sessions/2026-09-28_S007.md`, E024) |
| Trigger | R12: the final review substage of stage S02 (task S02.8-T04) |
| Branch | `docs/S02.8-T04-audit-a003` (from `develop` at `cea37c2`) |
| Plan version at start | 1.4.1 |
| Plan version at end | 1.4.2 (PATCH for F-002) |
| Status | **Complete**; one item waits for the user (F-004, README proposal) |

## 1. Scope

**Sources audited** (order of authority, `templates/audit-checklist.md`):
1. The user's recorded instructions:
   - `bootstrap/initial-prompt.json` and `prompts/P002`–`P005`;
   - the verbatim USER entries in session logs S006 and S007.
2. `README.md`.
3. `RULES.md` (1.7.1).
4. `plan.md` (1.4.1).
5. The rest:
   - ADR-0001 to ADR-0029;
   - `stages/S01-basic-nas.md` (with the P005 follow-up tasks) and `stages/S02-nas-gui.md`;
   - `dependencies.md`;
   - audits A001 and A002.
6. Git history (`develop`, the task branches, `main`) as evidence.

**Files audited:**
- Every file under `code-agent-docs/`. The archive was checked for identity with git only.
- `README.md`, `AGENTS.md`, and `CLAUDE.md`.
- The product documentation, with S02's new files:
  - `docs/**`, including `docs/guide/web-interface.md` and `docs/reports/S02-cross-browser.md`;
  - `scripts/README.md`;
  - the README's Development section.
- To check the register: `go.mod`, `web/package.json`, the CI workflow, and the tools the tests use.

**Method:**
- Scripted, read-only checks, kept in the agent's scratchpad as the checklist advises (`audit_links.py` from A002, and `audit_a003.py`):
  - links, anchors, backticked paths, and placeholders;
  - plan versions against git, and the archive against every committed plan;
  - archived prompts against the user's originals;
  - ADR sections and statuses;
  - `go.mod`, `web/package.json`, and CI actions against the register and the stage document;
  - requirement IDs;
  - log numbering and times against commit times.
- Manual review of meaning and consistency, including:
  - the guide's shortcut table against `web/src/lib/files/keys.ts`;
  - the stage IDs in current documents against plan 10.18.
- Commands in the README's new rows, run as written (group K).
- The resumability test.

**Interpretation notes** (A001's and A002's notes still apply):
- These backticked paths are not broken paths:
  - branch names such as `docs/S02.8-T03-guide-report`;
  - planned files and packages of later stages (`internal/photos`, `internal/jobs`, `docs/third-party-notices.md`, the S13.2 setup scripts, the S03 threat model);
  - notation (`testdata/fuzz/FuzzXxx`, "`ADR-0001` … `ADR-0029`");
  - paths relative to `docs/` in `docs/README.md`.
- "Placeholder" as a feature word (the Photos and Settings placeholders of S02.2) is not an unfinished placeholder.

### 1.1 Inventory at start (S007, 2026-09-28 13:20)

| Area | Files |
|---|---|
| `code-agent-docs/` root | `RULES.md` (326 lines), `CURRENT_STATE.md` (58), `plan.md` (3,277), `dependencies.md` (253) |
| `bootstrap/`, `prompts/` | `initial-prompt.json`; `P002`–`P005` (P005 is new since A002) |
| `decisions/` | ADR-0001 … ADR-0029 (ADR-0021 to ADR-0029 new since A002) |
| `stages/` | `S01-basic-nas.md` (493, Done), `S02-nas-gui.md` (500, In Progress) |
| `templates/` | stage, session log, ADR, audit checklist |
| `logs/sessions/` | S001–S006 (closed), S007 (current, E001–E023 at start) |
| `archive/plan-history/` | `plan_v0.1.0` … `plan_v1.4.0` (15 files) |
| `audits/` | A001 and A002, each with a README proposal |
| Repository root | `README.md`, `AGENTS.md`, `CLAUDE.md`, `LICENSE`; the user's `prompts/1-…` to `5-…` |
| Product documentation | `docs/README.md`, `docs/licensing.md`, `docs/testing.md`, `docs/storage-root.md`, `docs/api/{conventions,errors,versioning,usage}.md`, `docs/perf/S01-baseline.md`, `docs/guide/web-interface.md`, `docs/reports/S02-cross-browser.md`, `scripts/README.md`, `deploy/README.md` |

**Git:**
- `develop` is at `cea37c2`.
- `main` is at `121444e` (2026-09-24 01:34), so nothing has been merged into `main` since before S005, as the rules require.

## 2. Summary

**8 findings** (F-001 to F-008).

| Severity | Fixed | Needs user decision | Deferred | Accepted as-is | Total |
|---|---|---|---|---|---|
| Critical | 0 | 0 | 0 | 0 | **0** |
| Major | 2 (F-003, F-005) | 1 (F-004) | 0 | 0 | **3** |
| Minor | 4 (F-001, F-002, F-006, F-007) | 0 | 0 | 1 (F-008) | **5** |
| **Total** | **6** | **1** | **0** | **1** | **8** |

- **No Critical finding.** The resumability test passes (section 6).
- **The three Major findings:**
  - CURRENT_STATE's "Next steps" began with the open P005 questions and listed S02.8 work that was already done. The "Active task" line was right, so a fresh agent would still have resumed correctly. Fixed (F-003).
  - The README's features and roadmap do not show the four feature areas the user added in P005, and stage 2 is not ticked. This is written up as proposals R-13 to R-15 for the user (F-004).
  - The register lacked the test browsers and the tool that made the media fixtures. It was found and fixed in S02.8-T03's register pass, before this audit (F-005).
- The rest is small: the stage document's files table, stale plan text about ADR-0021 (plan 1.4.2), register and stage-document drift fixed in T03, and missing task-start entries.

## 3. Check results

| Group | Result | Evidence |
|---|---|---|
| **A. Structure** | Pass | See below. |
| **B. Prompt fulfillment** | Pass | See below and section 5. |
| **C. Rules and pointers** | Pass | RULES 1.7.1 has R1–R12 in full and invariants I1–I10. The 1.6.0, 1.7.0, and 1.7.1 changelog rows cite their approvals (S006 testing approach; P005 for I10 and the checklist change; S007 E002 and E013 for CI). User Preferences record every answer, including the CI trigger. Behavior matches: every S006 and S007 task branch was merged `--no-ff` into `develop` by the agent, `main` was not touched, and CI has not run since the tag-only trigger (E013). `AGENTS.md` and `CLAUDE.md` are identical apart from their own names. |
| **D. Plan** | F-002 | See below. |
| **E. README alignment** | F-004 | S02 changed only the Development section (S02.1-T05, S02.8-T03), as the stage document allows. A002's R-11 and R-12 are applied. The P005 features and stages, and stage 2, are not in the Features and Roadmap sections (F-004). |
| **F. ADRs** | Pass | ADR-0001–0029 are numbered without gaps, and each has the six template sections. ADR-0021 is Accepted with its approval (S007 E013). ADR-0019 and ADR-0022–0029 are Proposed, as P005 requires ("Deciding any 'Proposed' ADR on the user's behalf" is out of scope). All 29 are linked from the plan. The Unverified register items belong to later stages (S04.8, S06.5, S10.3, S12.1, S12.2, S14, S15) and are deferred to them. |
| **G. Dependency register** | F-005, F-006, F-007 | Every direct module and `tool` in `go.mod`, every package in `web/package.json` (at its exact version, in both the register and stage section 7 after F-007), and every GitHub Action in `ci.yml` is listed. Every row has a verification status. Section 12: S02 adds nothing the running NAS needs on the target (the interface is embedded in the binary); Node.js and pnpm are listed as from-source build tools. The test browsers and the fixture tool were missing (F-005). |
| **H. Stage documents** | F-001, F-007 | S01 (Done) and S02 (active) have documents; no other stage has one (just-in-time). S02 has sections 1–13, every task has acceptance criteria, the approval is quoted (S006 E009), and T01–T03 are Done, T04 and T05 Not started, matching CURRENT_STATE. The S01 follow-up tasks S01.2-T07, S01.3-T10, and S01.4-T07 are Done. No placeholders. The files table did not match what was built (F-001). |
| **I. CURRENT_STATE** | F-003 | 58 lines, under 150. Plan version, phase, active stage and task, and blockers were right; the next steps were not in order (F-003). The write-ahead entry for this audit was written first. |
| **J. Session logs and git** | F-008 | See below. |
| **K. Cross-document consistency** | Pass | See below. |
| **L. Resumability** | **Pass** (first run and re-run) | Section 6. |

**A. Structure:**
- Every folder in the documentation map exists, as do all required and root files. The new `docs/guide/` and `docs/reports/` are product documentation, indexed in `docs/README.md`.
- The plan archive holds 0.1.0 to 1.4.0, and 1.4.1 after this audit (16 files). Each is byte-identical to a committed `plan.md` with that version.
- Every `scripts/*.sh` file is executable in git (mode 100755).

**B. Prompt fulfillment:**
- P005 is the one new prompt since A002. Its archive is byte-identical to the user's `prompts/5-feature-additions-prompt.json` in git, and it is valid JSON. So are P002–P004.
- Every user message in S006 and S007 has a USER entry.
- The matrix is in section 5.

**D. Plan:**
- The header version, 1.4.1, equals the latest revision entry, and every earlier version is archived.
- S02.1–S02.7 are Done and S02.8 is In Progress, as in the stage document.
- Every FR and NFR ID referenced in the S02 stage document, CURRENT_STATE, the S007 log, the guide, and the report exists (252 defined).
- Every stage ends with a testing and review substage (S11.8, S12.8, S13.7, S14.10, S15.12), and plan 9 and 12.1 add the R12 audit to each.
- **Stale text** (F-002): Q50 said "the content hash still waits for ADR-0021 to be Accepted", and the P005 notes in S01.2, S01.3, and S01.4 said ADR-0021 was Proposed and the follow-ups were still to be scheduled. ADR-0021 was Accepted in E013, and the follow-ups were done in E017–E019.

**J. Session logs and git:**
- S001–S006 are closed with summaries (S006: E042). S007 is the current session, with E001–E023 in sequence, no gaps, and times in order.
- Log times equal the commit times, within a minute, for S01.4-T07 (11:59), S02.8-T01 (12:33), T02 (13:09), and T03 (13:15 and 13:16).
- All 46 non-merge commits on `develop` since S006 began carry a task or session ID found in the S006 or S007 log, except the user's own `c7089fa` ("added prompt #5"), which the S007 log records (startup).
- R9 archiving is not due (7 logs).
- Six tasks have a TASK DONE entry but no TASK START entry or write-ahead (F-008).

**K. Cross-document consistency:**
- **Stage ID changes** (plan 10.18): the current documents use the 1.4.0 IDs in their current meaning (S11.1 backfill in `internal/files/hashes.go`; S13.1, S13.2, S13.4, S13.6, and S13.7 in `docs/`, `deploy/`, `scripts/README.md`, and RULES; S14 and S15 in CURRENT_STATE). The historical documents keep their old IDs and are correct for their dates (for example A002's "`docs/third-party-notices.md` in S11" is the old packaging stage).
- **Links:** 0 broken links or anchors in all tracked Markdown outside `archive/`.
- **Placeholders:** 20 hits, all deferred with a stage or explained: ADR-0015 "to be confirmed in S09.1", the register's synonym dictionary "TBD" (S06.5), and "placeholder" used as a feature word.
- **Product docs:**
  - The guide's shortcut table matches `shortcuts()` in `web/src/lib/files/keys.ts` (the source of the in-app "?" list), row by row. The preview keys (← →, Page Up and Page Down in PDFs) match `PreviewFrame.svelte` and `PdfView.svelte`.
  - The flows the guide describes are the ones the system tests drive (`web/tests/e2e`).
  - The README's new commands were run as written:
    - `pnpm exec playwright install chromium firefox`: exit 0, during this audit;
    - `pnpm test`: 198 passed, during this audit;
    - `pnpm coverage` and `pnpm test:e2e`: in S02.8-T01 and T02 (E021, E022).
  - The CI sentence matches `ci.yml` (the `web` job on Linux; `e2e` on Linux and Windows with `E2E_CHROME=1`).
- Terminology and dates are consistent.

## 4. Findings

| ID | Severity | Group | Location | Description | Outcome |
|---|---|---|---|---|---|
| F-001 | Minor | H, K | `stages/S02-nas-gui.md` section 6 | The files table still offered the planning-time alternatives ("`internal/webapp/` (or `internal/api/app.go` …)", "(or `internal/archives/`)", "(or a new `scripts/check-web-licenses.sh`)"). It named `web/e2e/**` for the system tests, which are in `web/tests/e2e/**`. It had no rows for the cross-browser report, the Playwright and Vite configs, `docs/testing.md`, `testdata/SOURCES.md`, or the test-output ignore files. | **Fixed**: the table names what was built (`internal/webapp/`, `internal/files/archive.go`, `scripts/check-web-licenses.mjs`, `web/tests/e2e/**`) and has the missing rows. |
| F-002 | Minor | D | `plan.md` Q50 (section 5); the P005 notes of S01.2, S01.3, and S01.4 | They still said ADR-0021 had to be Accepted before the follow-ups were built, and that their timing was open. ADR-0021 was Accepted in S007 E013, and the follow-ups were done in E017–E019. | **Fixed** (plan 1.4.2, PATCH): they say the follow-ups were done in S007. 1.4.1 is archived. |
| F-003 | **Major** | I, L | `CURRENT_STATE.md` "Next steps" | Step 1 was the open P005 questions ("not blocking S02"). Step 2 listed S02.8 work that was done (the tests, the report, the guide) together with what was left. The next action was stated correctly only in the "Active task" line. | **Fixed**: step 1 is finishing A003, step 2 is S02.8-T05 with what the sign-off asks, and step 3 is the P005 questions. The pointers list A003. |
| F-004 | **Major** | E | `README.md` Features and Roadmap sections; status line | The Features section has none of the four feature areas the user added in P005 (duplicate and look-alike photos, duplicate files with shortcuts, storage optimization, drive pools). The roadmap lists stages 1–7 and AI only, without the P005 stages, and stage 2 is not ticked. README edits outside the Development section need the user's approval. | **Needs the user's decision**: proposals R-13 (status line), R-14 (roadmap), and R-15 (features) in `A003-readme-proposal.md`, for the S02 sign-off. |
| F-005 | **Major** | G | `dependencies.md` section 3 | The register did not list the browsers the tests drive (Playwright's Chromium 153.0.8010.12 and Firefox 155.0; Edge and Chrome through Playwright's channels) or the tool that generated the committed media fixtures (PyAV 18.1.0 and NumPy 2.5.3, whose wheels bundle an FFmpeg build with GPL encoders). R6 and the user's rule ("keep record of all dependencies needed", S005 E015) cover such tools. | **Fixed** in S02.8-T03 (`1abbd95`), found by the register pass before this audit: three rows in section 3, all tools only, not shipped. |
| F-006 | Minor | G | `dependencies.md` section 4, the axe row | A stray `**Candidate** |` cell gave the row eight cells in a seven-column table, so its status read "Candidate" and the "Installed" note fell outside the table. | **Fixed** in S02.8-T03 (`1abbd95`). |
| F-007 | Minor | G, H | `stages/S02-nas-gui.md` section 7 | The vitest row said 5.0.1 (the planned version), but 5.0.2 is installed (S007 E020; the register already said 5.0.2). | **Fixed** in S02.8-T03 (`1abbd95`): "5.0.2 … (planned 5.0.1; a patch, S007 E020)". |
| F-008 | Minor | J | `logs/sessions/2026-09-28_S007.md`; `CURRENT_STATE.md` "In progress" | Six tasks have a TASK DONE entry but no TASK START entry, and no write-ahead in CURRENT_STATE while they ran: S02.7-T03 (E016), S01.2-T07 (E017), S01.3-T10 (E018), S01.4-T07 (E019), S02.8-T02 (E022), and S02.8-T03 (E023). R2 asks for a write-ahead before every multi-step action. Each task was short, committed and merged in one step, and fully logged at the end, so no trace was lost. | **Accepted as-is**: past entries are not rewritten. From this audit on, every task writes its write-ahead and a TASK START entry first (this audit did). |

## 5. Prompt fulfillment matrix

**P005** (the one new prompt since A002; applied in S007 E003–E012). The definition of done:

| # | Item | Status | Evidence |
|---|---|---|---|
| 1 | Prompt archived | Done | `prompts/P005-feature-additions.json`, identical to the user's file |
| 2 | Plan archived; MINOR bump; revision entry | Done | `plan_v1.3.0.md`; plan 1.4.0 revision row |
| 3 | The new stages with every substage and field; labels kept | Done, as the user amended it | S11 (8 substages), S12 (8), S14 (10). Pools are S14, not S13, by the user's decision in E008 ("… and that too in the end …") |
| 4 | Packaging and AI renumbered, AI last; no old IDs in their old meaning; the "Stage ID changes" table | Done, as amended | Packaging is S13 (not S14) and AI S15 (E008). Table in plan 10.18. Re-checked in this audit (group K) |
| 5 | `changes_to_existing_stages` applied, labelled P005 | Done | Plan notes in S01.2–S01.4, S03.1, S04.2, S04.4, S05.1, S05.3, S06.1, S08.1, S08.5, S09.2, S10.1–S10.3, S13.2 (E009) |
| 6 | Invariant I10 in RULES and plan, with a changelog row | Done | RULES 1.7.0; plan section 2a |
| 7 | Nine ADRs Proposed; register additions | Done | ADR-0021–0029 (ADR-0021 later Accepted by the user, E013); register rows |
| 8 | S01 stage document has the hashing and storage-root tasks | Done | S01.2-T07, S01.3-T10, S01.4-T07 (and built in E017–E019, as the user asked in E010) |
| 9 | Consistency check run and recorded | Done | S007 E011 |
| 10 | CURRENT_STATE and log current; closing summary | Partial, as expected | Current. S007 is still open, so its closing summary comes when the session closes |
| 11 | No application code; no real disks | Done | P005 was applied before any code in S007 (code resumed after E013); no disk was touched |
| 12 | Report to the user | Done | E012 |

**The user's instructions in S007:**

| Instruction (verbatim; log) | Status | Evidence |
|---|---|---|
| "dont run CI on every commit/push/checkpoint but CI should only run at the completion of a stage" (E002) | Done | RULES 1.7.0/1.7.1; `ci.yml` runs on `S[0-9][0-9]-done` tags and by hand only |
| "also, this maybe missing, but you also need to add auto grouping of burst photos" | Done | Plan S11 (bursts), S05.3 note (burst identifiers) |
| "and fix every issue you find in the prompt as you deem best" | Done | P005 report (E012), planner additions labelled in plan section 3 |
| "if working with raids is a complex problem, just do raid 0 and 1 implementation and that too in the end, and leave complex raid for later as planned non implemented work" (E008) | Done; the agent's reading awaits confirmation | Plan S14 (RAID 0 and 1), 11a (deferred) |
| "Keep working and once that is done, resume from the checkpoint … and keep working until the stage is complete, also make sure to work on the new stuff too if they were meant to be part of current or previous stages." (E010) | In progress | S02.7 done; S01 follow-ups done; S02.8-T01–T03 done; T04 (this audit) and T05 remain |
| "SHA-256 (Recommended)"; "Stage tag + manual (Recommended)" (E013) | Done | ADR-0021 Accepted; plan 1.4.1; RULES 1.7.1 |

## 6. Resumability test

Acting as a new agent: read `AGENTS.md`, then follow R1 (RULES, CURRENT_STATE, the plan's targeted sections, the latest log).

| Question | First run (before the fixes) | Re-run (after the fixes) |
|---|---|---|
| What is this project? | Clear (README, plan section 1) | Clear |
| Which stage and task are active? | S02; the "Active task" line: none, next S02.8-T04 then T05 | Same, plus the write-ahead entry for T04 |
| What was the last completed action? | S02.8-T03 (E023, `cea37c2`) | Same |
| What exactly is the next action? | Right in "Active task", but "Next steps" began with the P005 questions (F-003) | Finish A003; then T05: completion record, the `S02-done` tag and its CI run, the sign-off |
| What is waiting on the user? | The P005 questions and ADRs; the stray `C:\c` folder | The same, plus README proposals R-13–R-15 and the re-confirmation of S01, both at the S02 sign-off |
| What rules must I follow? | RULES.md 1.7.1 (pointer correct) | Same |

## 7. Decisions needed from the user

1. **README proposals R-13 (status line), R-14 (roadmap), and R-15 (features)** (F-004), in `A003-readme-proposal.md`.
   - Recommendation: approve R-13, R-14 option A, and R-15, applied right after the S02 sign-off.
   - Asked together with the sign-off in S02.8-T05.

## 8. Closing: what the next audit should watch

1. **Write-ahead discipline (F-008):** each task writes its write-ahead and a TASK START entry before it starts. The next audit checks that every TASK DONE entry has one.
2. **Files tables:** a stage document's files table is written before the code. When a task chooses between alternatives it names, the table should be updated in the same task (F-001).
3. **Test tools in the register:** new test browsers, fixture generators, and CI images belong in the register when they are first used (F-005).
4. **Plan notes that wait on decisions:** when a decision is made or a follow-up is done, grep the plan for the notes that waited on it (F-002).
5. **README after P005:** apply R-13–R-15 only if approved.
6. **Examples after login (S03):** stage 3 adds login and HTTPS, which changes every command in `docs/api/usage.md`, the demo scripts, and the system tests' server. Re-run them.
7. **Checklist:** A003 added three checks: test tools in the register (G), files tables against what was built (H), and task-start entries (J).
8. **From A001 and A002, still open:** the Unverified register items at their stages; whether to commit the audit scripts (the checklist keeps them out of the repository unless the user asks).
