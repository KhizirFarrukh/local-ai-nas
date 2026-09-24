# A002: Documentation audit (S01 final review)

| Field | Value |
|---|---|
| Audit ID | A002 |
| Date | 2026-09-24 |
| Session | S005 (`logs/sessions/2026-09-24_S005.md`, E122 onward) |
| Trigger | R12: the final review substage of stage S01 (task S01.7-T07) |
| Branch | `docs/S01.7-T07-audit` (from `develop` at `ecfde5c`) |
| Plan version at start | 1.1.2 |
| Plan version at end | 1.1.3 (PATCH for F-003 and F-004) |
| Status | **Complete** (report delivered in S005) |

## 1. Scope

**Sources audited** (order of authority, `templates/audit-checklist.md`):
1. The user's recorded instructions:
   - `bootstrap/initial-prompt.json` and `prompts/P002`–`P004`;
   - the verbatim USER entries in session logs S001–S005 (S005: E001–E116).
2. `README.md`.
3. `RULES.md` (1.5.0).
4. `plan.md` (1.1.2).
5. The rest:
   - ADR-0001 to ADR-0020;
   - `stages/S01-basic-nas.md`;
   - `dependencies.md`;
   - audit A001 and its README proposal.
6. Git history (`develop`, the task branches, `main`) as evidence.

**Files audited:**
- Every file under `code-agent-docs/`. The archive was checked for identity with git only.
- `README.md`, `AGENTS.md`, and `CLAUDE.md`.
- For the first time, the product documentation that S01 created: `docs/**`, `scripts/README.md`, and the README's Development section.
- To check the register: `go.mod` and the CI workflow.

**Method:**
- Scripted, read-only checks, kept in the agent's scratchpad as the checklist advises (`audit_links.py`, `audit_checks.py`, and one-off checks):
  - links and anchors;
  - backticked paths and placeholders;
  - plan versions against git;
  - `go.mod` and CI actions against the register;
  - IDs, ADR links, and pinned versions;
  - prompt identity;
  - log numbering and times against commit times.
- Manual review of meaning and consistency.
- The resumability test.

**Interpretation notes** (A001's notes still apply):
- These backticked paths are not broken paths:
  - planned files (ADR layouts; files of later stages such as `docs/third-party-notices.md` in S11);
  - branch names such as `docs/P002-staged-roadmap`;
  - historical text in session logs.
- "Placeholder" as a feature word (the S02.2 settings placeholder, the S01.2-T06 photos placeholder) is not an unfinished placeholder.

### 1.1 Inventory at start (S005, 2026-09-24 23:26)

| Area | Files |
|---|---|
| `code-agent-docs/` root | `RULES.md` (310 lines), `CURRENT_STATE.md` (91), `plan.md` (2,617), `dependencies.md` (205) |
| `bootstrap/`, `prompts/` | `initial-prompt.json`; `P002`, `P003`, `P004` (no new prompt file since A001) |
| `decisions/` | ADR-0001 … ADR-0020 |
| `stages/` | `S01-basic-nas.md` (429) |
| `templates/` | stage, session log, ADR, audit checklist |
| `logs/sessions/` | S001–S004 (closed), S005 (current, E001–E121 at start) |
| `archive/plan-history/` | `plan_v0.1.0` … `plan_v1.1.1` (9 files) |
| `audits/` | A001 report and README proposal |
| Repository root | `README.md`, `AGENTS.md`, `CLAUDE.md`, `LICENSE`; the user's `prompts/1-…` to `4-…` |
| Product documentation | `docs/README.md`, `docs/licensing.md`, `docs/testing.md`, `docs/api/{conventions,errors,versioning,usage}.md`, `docs/perf/S01-baseline.md`, `scripts/README.md` |

**Git:**
- `develop` is at `ecfde5c`, with 61 task commits and 53 merge commits since S005 started (06:19).
- `main` is at `121444e` (2026-09-24 01:34, before S005), so nothing has been merged into `main` in S005, as the rules require.

## 2. Summary

**11 findings** (F-001 to F-011).

| Severity | Fixed | Needs user decision | Deferred | Accepted as-is | Total |
|---|---|---|---|---|---|
| Critical | 1 (F-010) | 0 | 0 | 0 | **1** |
| Major | 2 (F-005*, F-011) | 1 (F-005*) | 0 | 0 | **2** |
| Minor | 7 (F-001, F-003, F-004, F-006, F-007, F-008, F-009) | 0 | 0 | 1 (F-002) | **8** |
| **Total** | **10** | **1** | **0** | **1** | **11** |

\* F-005 is fixed (the README text was restored). The new wording it replaced is now a proposal for the user (R-11).

- **The one Critical finding:** CURRENT_STATE's first next step still listed all of S01.7 "in order: T01 … T08", although T01–T06 were done. A fresh agent could have started S01.7 again from T01. Fixed.
- **The two Major findings:**
  - In S01.7-T05 the agent changed the README status line, the user's original text, outside the sections its stage document allows. It is restored, and the new wording is proposed in `A002-readme-proposal.md`.
  - The dependency register did not list the system tools that the scripts, the guide, and the README need: Git, optional Docker, Bash and the POSIX utilities, curl, PowerShell, and a headless browser. They are added (F-011).
- The rest is small: a stale path, two stale "pending Q1" notes in the plan (plan 1.1.3), a pointer-file sentence, a files-table row, CURRENT_STATE ordering, and new checklist lines for product documentation.

## 3. Check results

| Group | Result | Evidence |
|---|---|---|
| **A. Structure** | Pass | See below. |
| **B. Prompt fulfillment** | Pass; F-002 (informational) | See below. |
| **C. Rules and pointers** | F-006 | RULES 1.5.0 has R1–R12 in full. The 1.4.0 and 1.5.0 changelog rows cite their approvals (S005 E007/D-09, E008, D-03, E015). User Preferences record commit behavior, branching, merge-into-develop, the dependency record, and stacked branches. Behavior matches: every S005 task branch was merged `--no-ff` into `develop` by the agent, and `main` was not touched. |
| **D. Plan** | F-003, F-004 | See below. |
| **E. README alignment** | F-005 | A001's approved items R-01–R-08 and R-10 are in place. S01 changed the License section (S01.1-T02) and the Development section (S01.1-T06, S01.6-T06, S01.7-T05/T06). The status line changed in S01.7-T05 is outside those sections (F-005). The feature list is unchanged since A001. |
| **F. ADRs** | Pass | ADR-0001–0020 are numbered without gaps, and each has the six template sections. ADR-0019 is Proposed; ADR-0012 and ADR-0020 link to each other both ways (partial supersession); the rest are Accepted with approval records. The Unverified items are deferred to their stages (A001: S06.5, S09.1, S12.x). |
| **G. Dependency register** | F-011 | Every direct module and `tool` in `go.mod` is listed at its `go.mod` version, as is every GitHub Action in `ci.yml` (6, at their pinned versions). Vendored Redoc is listed with its checksum. The versions in section 7 of the stage document match. Section 12 (per-platform prerequisites) is present. The system tools used by the scripts and guides were missing (F-011), found while preparing S01.7-T08 and before this audit's merge. |
| **H. Stage documents** | F-007 | Only S01 has a document (just-in-time). It has sections 1–13, every task has acceptance criteria, and the approval record is quoted (S005 E012). There are no placeholders, and its status (In Progress) matches CURRENT_STATE. The files-table row for README lacked S01.1-T02 (F-007). |
| **I. CURRENT_STATE** | F-001, F-009, F-010 | 91 lines, under 150. It had a stale path (F-001), a stale phase line, ordering, and pointers (F-009), and wrong next steps (F-010). |
| **J. Session logs and git** | Pass | See below. |
| **K. Cross-document consistency** | F-008 | See below. |
| **L. Resumability** | F-010 (first run); **Pass** (re-run) | Section 6. |

**A. Structure:**
- Every folder in the documentation map exists, as do all required and root files.
- There are no orphan files. `docs/` and `scripts/` are product files, outside the map's scope.
- The plan archive holds 0.1.0 to 1.1.2 (10 files after this audit). Each is byte-identical to `plan.md` in git at that version.
- Scripts shown as `scripts/*.sh` are executable in git. Two were not, and were fixed in S01.7-T06 before this audit (E121).

**B. Prompt fulfillment:**
- There is no new prompt file since A001.
- Each user message in S005 has a USER entry: E001, E003, E007, E008, E012, E015, E020, E049, E065, E069, E079, E109, and E116.
- The archived prompts are identical to the user's originals in git. On this PC, the working copies differ only in line endings (F-002).

**D. Plan:**
- The header version, 1.1.2, equals the latest revision entry, and every earlier version is archived.
- S01.1–S01.6 are Done and S01.7 is In Progress, as in the stage document.
- Every FR/NFR ID referenced in the stage document, CURRENT_STATE, the S005 log, and the product docs exists (181 defined).
- All 20 ADRs are linked from the plan.
- Answered questions cite where they were answered (Q1, Q16, Q18, Q22, Q37, Q38).
- Stale "pending Q1" text remained in NFR-031 and S06.8 (F-003, F-004).

**J. Session logs and git:**
- S001–S004 are closed with summaries. S005 is the current session, with E001–E122 in sequence, no gaps, and times in order.
- Log times equal the commit times, within a minute, for the sampled tasks: S01.5-T06, S01.6-T06, and S01.7-T03 to T06.
- Every commit is traceable to the log. The only commit without a task ID is the user's `bda8321`, recorded in E004.
- The commits follow Conventional Commits with task IDs. `git status` was clean between tasks.
- R9 archiving is not due (5 logs).

**K. Cross-document consistency:**
- Task IDs in the log and in CURRENT_STATE all exist in the stage document.
- **Links:** 0 broken links or anchors in all tracked Markdown outside `archive/`.
- **Placeholders:** 13 hits, all deferred with a stage or explained: ADR-0015 "to be confirmed in S09.1", the register's synonym dictionary "TBD" (S06.5), and "placeholder" used as a feature word.
- **Terminology:** "Files area" and "Photos area" are capitalized as product names in the README and lowercase in technical docs; the plan uses both. This is informational, not a finding.
- **Product docs:** every code block in `docs/api/usage.md` was run as written on Linux and in Windows PowerShell 5.1 (E119), and the demo scripts run in CI.
- The checklist had no checks for product documentation (F-008).

## 4. Findings

| ID | Severity | Group | Location | Description | Outcome |
|---|---|---|---|---|---|
| F-001 | Minor | I, K | `CURRENT_STATE.md` (S01.3-T04 bullet) | The path `api/body.go` should be `internal/api/body.go`. | **Fixed** |
| F-002 | Minor | B | Working copies on the development PC | The archived prompts (`bootstrap/initial-prompt.json`, `prompts/P002`–`P004`) and the user's `prompts/2-…`–`4-…` have CRLF line endings on this PC. They were checked out before `.gitattributes` set `eol=lf`. In git, every blob is LF and the archive equals the user's originals byte for byte, and a fresh clone gets LF. | **Accepted as-is**: the repository content is correct, and the user's own `prompts/` files are left alone. If wanted, removing a file and running `git checkout -- <file>` refreshes it. |
| F-003 | Minor | D | `plan.md` NFR-031 | It still said "(targets to confirm with Q1)", although Q1 was answered in S005. | **Fixed** (plan 1.1.3): "(measured on the Q1 platforms, see Q1)", worded like NFR-003. |
| F-004 | Minor | D | `plan.md` S06.8 | "Reference hardware pending Q1", and the library size as "e.g. 100,000 photos plus 100,000 files", although Q1 and Q18 are answered. | **Fixed** (plan 1.1.3): it names the Q1 hardware and the Q18 size. |
| F-005 | **Major** | E | `README.md` line 7 (status line) | In S01.7-T05 the agent replaced the user's original status line ("This project is in its initial stage. Features described below are the planned scope and are not yet implemented.") with a new one. The stage document allows README changes only in the License, Development, and API usage sections. | **Fixed** (the user's text is restored, identical to `7c37389`), and **needs the user's decision**: proposal R-11, together with R-12 (the roadmap tick), in `A002-readme-proposal.md`, for the S01 sign-off. |
| F-006 | Minor | C | `AGENTS.md` | `CLAUDE.md` says "Keep it in sync with `AGENTS.md`", but `AGENTS.md` had no matching sentence. | **Fixed** |
| F-007 | Minor | H | `stages/S01-basic-nas.md` section 6 | The `README.md` row named only S01.1-T06 and S01.7-T05, and only the Development and API usage sections. S01.1-T02 also changes the README (its License section, by the task's own text). | **Fixed**: the row lists the License section and S01.1-T02. |
| F-008 | Minor | A, E, K | `templates/audit-checklist.md` | The checklist had no checks for product documentation. It needs three: scripts executable in git, README edits within their stage's sections, and doc commands actually run. | **Fixed**: three checks added, with a note in the template header. |
| F-009 | Minor | I | `CURRENT_STATE.md` | The phase line still said "S01.7 … next". A note left over from S01.4-T03 sat under "In progress". The two oldest "Last completed" items were listed first. The pointers did not list A002. | **Fixed** |
| F-010 | **Critical** | I, L | `CURRENT_STATE.md` "Next steps" | Step 1 listed all of S01.7 "in order: T01 integration suite, … T08". T01–T06 were already done, so a fresh agent could have started the stage over. | **Fixed**: step 1 is finishing A002, and step 2 is T08, with the exact questions for the user's sign-off. |
| F-011 | **Major** | G | `dependencies.md` section 3 | The register listed no system tools, although the README requires Git (and optional Docker), `scripts/*.sh` need Bash and the POSIX utilities, the guide and the demos need curl (7.87+ for `--url-query`) and PowerShell, and `check-api-docs-offline.sh` needs a headless Edge or Chrome. R6 and the user's rule ("keep record of all dependencies needed", S005 E015) cover such tools. The first pass of group G checked only `go.mod` and CI. | **Fixed**: a sub-table in section 3 with versions, licenses, purpose, and verification. None of them is needed on a release install (section 12). |

## 5. Prompt fulfillment matrix

No new prompt file came after A001. The matrix covers P004's rule and the user's instructions given in S005.

| Instruction (verbatim; log) | Status | Evidence |
|---|---|---|
| P004 / R12: "Run a documentation audit … as part of every stage's final review substage" | Done for S01 | This report (S01.7-T07) |
| "Accept all (Recommended)" for D-01–D-07, D-09, D-11, D-14 (E007) | Done | A001 section 9 |
| "when you are done with a branch, you yourself should do a merge of that branch into develop (another rule you should remember)" (E008) | Done | RULES 1.4.0 User Preferences; every S005 task branch merged by the agent (for example `b1e8c3e`, `73aba59`, `ecfde5c`) |
| "Approve S01 (Recommended)" and "Approve as 1.0.0 (Recommended)" (E012) | Done | Stage document section 11; plan 1.0.0 |
| Keep a record of every dependency; a separate setup script per platform (E015) | Partial, as planned | `dependencies.md` covers every module, tool, CI action, and vendored file, and section 12 lists runtime prerequisites per platform. The setup scripts are S11.2 (FR-149, NFR-032). |
| "AGPL-3.0-or-later (Recommended)" (E020) | Done | `LICENSE`; plan 1.1.1 |
| "100k photos + 100k files (Recommended)" (Q18, E116) | Done | Plan 1.1.2; NFR-003; `docs/perf/S01-baseline.md` |
| "continue" (E049, E065, E069, E079, E109, and later) | Done | S01 continued task by task in the stage's execution order |

## 6. Resumability test

Acting as a new agent: read `AGENTS.md`, then follow R1 (RULES, CURRENT_STATE, the plan's targeted sections, the latest log).

| Question | First run (before the fixes) | Re-run (after the fixes) |
|---|---|---|
| What is this project? | Clear (README, plan section 1) | Clear |
| Which stage and task are active? | S01; S01.7-T07 (the "Active task" line) | Same |
| What was the last completed action? | Unclear: "Last completed" began with two S005-morning items, and S01.7-T06 came third | S01.7-T06, merged (`ecfde5c`) |
| What exactly is the next action? | **Wrong:** "S01.7 in order: T01 …" (F-010) | Finish A002, then T08: completion record, then ask the user for the sign-off with the two review items |
| What is waiting on the user? | The throughput deviation only | The throughput deviation and README proposals R-11/R-12, both at the S01 sign-off |
| What rules must I follow? | RULES.md 1.5.0 (pointer correct) | Same |

## 7. Decisions needed from the user

1. **README proposals R-11 (status line) and R-12 (roadmap: stage 1 done)** (F-005), in `A002-readme-proposal.md`.
   - Recommendation: approve both, applied right after the S01 sign-off.
   - It is asked together with the sign-off in S01.7-T08.

The stage's other open review item, the throughput deviation in `docs/perf/S01-baseline.md`, is not a documentation finding. It is asked at the same time (S01 change log, S01.7-T04 row).

## 8. Closing: what the next audit should watch

1. **Examples after login (S03):** stage 3 adds login and HTTPS, which changes every command in `docs/api/usage.md` and in the demo scripts. Re-run the guide's code blocks and the demos whenever the API changes. The new checklist line in K covers this.
2. **Register scope:** group G should compare the register with every tool the scripts, docs, and CI call, not only with `go.mod` (F-011).
3. **README:** apply R-11/R-12 only if approved. Keep task edits inside the sections the stage document names (new checklist line in E).
4. **Plan status at the stage end:** S01 → Done in the plan (a PATCH) and in the stage document at the sign-off.
5. **Session length:** S005 is one very long session (over 1,000 log lines). R9 counts log files, not lines. Closing sessions at natural breaks keeps each log readable.
6. **From A001, still open:** the Unverified register items at their stages; whether to commit the audit scripts (the checklist keeps them out of the repository unless the user asks).
