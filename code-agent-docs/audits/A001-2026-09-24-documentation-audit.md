# A001: Documentation audit

| Field | Value |
|---|---|
| Audit ID | A001 |
| Date | 2026-09-24 |
| Session | S004 (`logs/sessions/2026-09-24_S004.md`) |
| Trigger | P004 (`code-agent-docs/prompts/P004-documentation-audit.json`) |
| Branch | `docs/P004-documentation-audit` (stacked on `docs/P003-technology-stack` at `828b95c`) |
| Plan version at start | 0.3.0 |
| Plan version at end | 0.3.1 (the separate video-streaming change afterwards takes it to 0.4.0) |
| Status | **Complete** (report delivered in S004) |

## 1. Scope

**Sources audited** (P004 order of authority):
1. The user's recorded instructions: `bootstrap/initial-prompt.json`, `prompts/P002-staged-development-roadmap.json`, `prompts/P003-technology-stack.json`, and the verbatim USER entries in session logs S001–S003.
2. `README.md`.
3. `RULES.md`.
4. `plan.md`.
5. ADRs (`decisions/ADR-0001` … `ADR-0019`), the stage document `stages/S01-basic-nas.md`, and `dependencies.md`.
6. Git history (all branches) as evidence.

**Files audited:** every file under `code-agent-docs/` (inventory below), plus `README.md`, `AGENTS.md`, `CLAUDE.md`, and the user's root `prompts/` folder.

**Method:**
- Scripted checks (scratchpad scripts, re-runnable) for structure, IDs, links, placeholders, tables, versions, and text comparison.
- Manual review for meaning, consistency, and resumability.
- Registry lookups for re-verification of "Unverified" items.

**Interpretation notes:**
- "Dependency" (group G) means anything the project uses, may use (candidate or deferred), or evaluated with a verified version. Other alternatives named only in ADR "Options considered" are listed in the register under "Alternatives named in ADRs" without verification.
- Angle-bracket notation such as `<root>`, `<internal>`, and `<ns>` in stage and plan text is **notation** for configured paths, not an unfinished placeholder. Templates (`templates/*.md`) are placeholders by design and are excluded from the placeholder search.

### 1.1 Inventory at start (S004, 2026-09-24 00:58)

| Area | Files |
|---|---|
| `code-agent-docs/` root | `RULES.md` (290 lines), `CURRENT_STATE.md` (46), `plan.md` (2,536), `dependencies.md` (134) |
| `bootstrap/` | `initial-prompt.json` (424) |
| `prompts/` | `P002-staged-development-roadmap.json` (851), `P003-technology-stack.json` (366), `P004-documentation-audit.json` (244, archived this session) |
| `decisions/` | ADR-0001 … ADR-0019 (19 files), `.gitkeep` |
| `stages/` | `S01-basic-nas.md` (388), `.gitkeep` |
| `templates/` | `stage-template.md` (145), `session-log-template.md` (58), `adr-template.md` (62) |
| `logs/sessions/` | `2026-09-23_S001.md` (129), `2026-09-23_S002.md` (184), `2026-09-24_S003.md` (293), `2026-09-24_S004.md` (current) |
| `archive/plan-history/` | `plan_v0.1.0.md` (723), `plan_v0.2.0.md` (2,535), `.gitkeep` |
| `archive/sessions/` | `.gitkeep` only |
| `audits/` | this report (created this session) |
| Repository root | `README.md`, `AGENTS.md`, `CLAUDE.md`; user folder `prompts/` with `1-…`, `2-…` (committed by the user) and `3-…`, `4-…` (**untracked**) |

**Git (all branches):**

| Commit | Time | Author | Subject |
|---|---|---|---|
| `828b95c` | 2026-09-24 00:53 | agent (user's identity) | docs(logs): close session S003 and update CURRENT_STATE [P003] |
| `18d9331` | 00:52 | agent | docs(adr): record technology stack decisions [P003] |
| `5ab149b` | 00:27 | agent | docs(plan): staged roadmap v0.2.0 with S01 stage plan [P002] |
| `fd23abb` | 2026-09-23 23:55 | user | added staged development prompt |
| `c8bcee7` | 23:47 | agent | docs(logs): close session S001 and update CURRENT_STATE [S001] |
| `2b5c1a7` | 23:47 | agent | docs: bootstrap agent documentation system and plan v0.1.0 [S001] |
| `c34cc21` | 23:34 | user | rename file |
| `9f56922` | 23:30 | user | added initial prompt |
| `1c7faba` | 23:19 | user | add readme.md |

- **Branches:** `main` and `develop` are at `c34cc21`. Docs branches are stacked: `docs/S001-bootstrap-agent-docs` → `docs/P002-staged-roadmap` → `docs/P003-technology-stack` → `docs/P004-documentation-audit` (this audit). All are pushed except this one; no PRs are open.

## 2. Summary

**32 findings** (IDs F-001–F-034; F-004 and F-029 were not used).

| Severity | Fixed | Needs user decision | Deferred | Accepted as-is | Total |
|---|---|---|---|---|---|
| Critical | 1 (F-027) | 0 | 0 | 0 | **1** |
| Major | 7 (F-011, F-013, F-016, F-021, F-024, F-025, F-028) | 6 (F-005, F-009, F-012, F-014, F-018, F-022*) | 0 | 0 | **13** |
| Minor | 12 (F-001, F-002, F-006, F-008, F-010, F-017, F-019, F-023, F-026, F-030, F-033, F-034) | 3 (F-003*, F-007, F-015) | 1 (F-032) | 2 (F-020, F-031) | **18** |
| **Total** | **20** | **9** | **1** | **2** | **32** |

\* Partly fixed (register or description updated); the remaining part needs a user decision.

- **What was fixed:**
  - The Critical resumability hazard (CURRENT_STATE branching instruction).
  - Stale and contradictory plan text (plan 0.3.0 → **0.3.1**, PATCH).
  - Register gaps and re-verifications, among them TypeScript pinned to 6.0.3 and the Debian FFmpeg build found to be GPL-3.0+.
  - The S01 dependency table and the R12 audit task.
  - Documentation-map wording.
  - Git/log reconciliation.
- **Pre-approved P004 changes applied:** `audits/` in the documentation map, `templates/audit-checklist.md`, rule **R12**, and the audit step in the stage template (RULES 1.3.0, changelog rows a–e).
- **README.md unchanged.** Proposal: `audits/A001-readme-proposal.md`.
- **Out of audit scope:** the user's mid-session request for video streaming with live quality switching (S004 E005–E009) is handled **after** this audit as a separate plan change (0.3.1 → 0.4.0).

## 3. Check results (A–L)

Scripted checks: `audit_a001.py` (102 pass / 9 raw findings / 17 info on the first run; the raw findings were triaged below, and 2 were false positives), `reverify_a001.py` (group F/G re-verification), plus manual review.

### A. Structure: **findings F-001, F-002, F-003**
- Pass: all expected folders exist (`stages/`, `decisions/`, `logs/sessions/`, `archive/plan-history/`, `archive/sessions/`, `templates/`, `bootstrap/`, `prompts/`; `audits/` created this session).
- Pass: all expected files exist (RULES, CURRENT_STATE, plan, dependencies, three templates, bootstrap prompt, P002, P003).
- Pass: root files `README.md`, `AGENTS.md`, `CLAUDE.md` exist.
- Pass: plan history is complete. The revision history lists 0.1.0, 0.2.0, 0.3.0 (current). `plan_v0.1.0.md` equals `git show c8bcee7:code-agent-docs/plan.md` and `plan_v0.2.0.md` equals `git show 5ab149b:…` (byte-for-byte after line-ending normalization).
- Findings: the documentation map does not cover `audits/` (F-001); `.gitkeep` wording (F-002); the root `prompts/` folder is undocumented and two prompt files are untracked (F-003).

### B. Prompt fulfillment: **Pass** (matrix in section 5)
- All four archived prompts are valid JSON, complete (contain `final_instruction`), and byte-identical to the user's originals in the root `prompts/` folder.
- Every prompt has its prescribed USER entry: S001 E001, S002 E001, S003 E001, S004 E001.
- Every definition-of-done item of the bootstrap prompt, P002, and P003 is Done. Items with deviations point to findings (F-009, F-025).

### C. Rules and pointers: **findings F-005, F-006, F-007**
- Pass: every sentence of the bootstrap `operating_rules` appears in RULES.md after normalization (quotes, backticks, list numbering). The only differences are the P002-approved R3 hierarchy wording and a trailing period in the R10 precedence list. No rule was summarized away.
- Pass: P002 changes are present (R3 hierarchy, R4 pre-1.0 versioning, Project invariants I1–I8), as are P003 changes (I9, register in the documentation map, R6 addition).
- Pass: required sections present (Quick start, Documentation map, Bootstrap pointer files, User Preferences, changelog). All 10 changelog rows cite an approval (bootstrap prompt, the user's S001 answers, P002, P003).
- Pass: User Preferences records the commit preference (S001 answers).
- Pass: no two rules contradict each other (manual review of R1–R11 and the invariants).
- Pass: `AGENTS.md` and `CLAUDE.md` are 18 lines each, with correct paths and reading order (RULES → CURRENT_STATE → plan → latest log).
- Findings: the git workflow deviated from the recorded preference (F-005); session/stage ID confusion (F-006); the startup cost of reading all of plan.md (F-007).

### D. Plan: **findings F-008 to F-017**
- Pass: header 0.3.0 equals the latest revision entry, and every previous version is archived.
- Pass: all 17 required sections are present and non-empty (including 2a, 2b, 7 "Chosen technology stack", 11a).
- Pass: 12 stages in order, S12 last. All 92 P002 substages are present, each with all 8 fields and 2–5 acceptance criteria.
- Pass: requirement traceability. 173 IDs (FR-001–FR-143, NFR-001–NFR-030), no duplicates, FR-001 marked Deprecated (not deleted), every non-deprecated ID linked from a substage, no undefined ID referenced.
- Pass: dependencies. Every "Depends on" names an existing substage or stage; no substage depends on a later stage; no cycles. The only cycle reported by the script (`S01.1 → S01.1`) is a wording artefact (F-008).
- Pass: invariants I1–I9 are worded identically in plan.md and RULES.md (plan's I9 carries only an "Added in 0.3.0" annotation).
- Pass: the chosen-stack table (section 7) matches ADR statuses and links (17 Accepted; 0003 and 0019 Proposed).
- Pass (with F-015): open questions. Answered ones cite where; every open one names what it needs; none were answered without a user instruction, except the judgment call in F-015.
- Findings: F-008–F-017.

### E. README alignment: **finding F-018**
- Pass: every README feature maps to at least one requirement:

  | README feature | Requirements |
  |---|---|
  | NAS storage | FR-002–FR-007, FR-069 |
  | Timeline and albums | FR-013, FR-016 |
  | EXIF | FR-010 |
  | Sidecar JSON | FR-023–FR-030 |
  | Opt-in local AI | FR-031, NFR-002 |
  | Auto-classification | FR-033, FR-034, FR-138 |
  | Face grouping, several faces per photo | FR-039, FR-040 |
  | Results stored in sidecars, AI runs once | FR-033, FR-035 |
  | Search fields | FR-047 |
  | Word forms, synonyms, typos | FR-048–FR-050 |
  | Operators before/after/on/place/tag/face/type | FR-056–FR-058 |
  | Privacy principles | NFR-001, NFR-002, I6, I7 |
  | Deployment plans | FR-131, FR-132, NFR-008 |
  | Offline geocoding | FR-062 |
  | Face naming and merging | FR-042, FR-043 |
  | User accounts and permissions | FR-065, FR-112 |
- Finding: the README is out of date compared with approved decisions (F-018). Proposal in `audits/A001-readme-proposal.md`.

### F. ADRs: **findings F-019 to F-023**
- Pass: ADR-0001 to ADR-0019 are sequential with no gaps or duplicates, and every ADR has all template sections (Context, Options considered, Decision, Consequences, Approval record, Links).
- Pass: every status is valid and matches P003 (ADR-0003 is outside P003 and stays Proposed; ADR-0019 is Proposed as P003 specifies). Every Accepted ADR's approval record cites P003. No ADR is Superseded.
- Pass: every ADR is linked from plan.md section 7, and the dependency-bearing ADRs are linked from the register.
- **Re-verification of "Unverified" items (2026-09-24, S004):**

  | Item | Result |
  |---|---|
  | Node.js LTS | **v24.21.0** (Krypton), MIT. Resolved |
  | Perl in Debian trixie | **5.40.1-6+deb13u1**. Resolved |
  | libde265 in trixie | **1.0.15-1+deb13u2**. Debian copyright lists LGPL-3+, GPL-3+, BSD-4-clause (F-023) |
  | Debian trixie FFmpeg build | `debian/rules` has **`--enable-gpl` and `--enable-version3`**, not `--enable-nonfree`. So it is a GPL-3.0-or-later build (F-022) |
  | ONNX Runtime 1.30.0 | wheels for **cp311–cp314**, so Python 3.14 is supported. Resolved |
  | TypeScript 7 + svelte-check | svelte-check 4.7.6 declares `typescript: ^5.0.0 \|\| ^6.0.0`, so **TypeScript 7 is not supported**; latest 6.x is **6.0.3** (F-021) |
  | RapidOCR models | the README states that converted models are under the Apache License 2.0; `MODEL_LICENSE` was not found at the repository root (404). **Partially verified**; re-check at S12.10 |
  | Python license | the PSF agreement wording was not found on docs.python.org/3/license.html. **Still unverified** (re-check at S12.1) |
  | Synonym dictionary seed | not chosen yet. **Deferred** to S06.5 |
- Findings: F-019 (ADR-0001 file name), F-020 (post-writing amendments, informational), F-021, F-022, F-023.

### G. Dependency register: **findings F-021, F-022, F-023, F-024**
- Pass: every row has a verification status. The license-attention items are flagged with ⚠.
- Pass: versions and licenses match the ADRs (58 register versions cross-checked in S003; re-run in the re-check step).
- Finding F-024: items named in ADRs, the plan, or the S01 document are missing from the register:
  - Samba (ADR-0019); Docker Engine, Compose, buildx (ADR-0006).
  - The GitHub Actions used by CI (ADR-0005, S01.1-T05).
  - axe-core (plan S02.7, 12.1); sqlite-vec (ADR-0018); LibRaw (ADR-0012, deferred); ONNX Runtime GPU execution providers (ADR-0017, deferred).
  - `golang.org/x/text` (S01 document).
- Updated rows: TypeScript (F-021), FFmpeg (F-022), libheif/libde265 (F-023), plus the resolved re-verifications (Node.js, Perl, ONNX Runtime on Python 3.14).

### H. Stage documents: **findings F-025, F-026**
- Pass: only `stages/S01-basic-nas.md` exists (just-in-time rule). It covers S01.1–S01.7 with tasks `S01.x-Tyy`, each with acceptance criteria.
- Pass: every R3 section is present, including an empty Approval record ("Not yet approved"). Status Planned is consistent with CURRENT_STATE.
- Finding F-025: the dependency table has no ADR column, and `golang.org/x/text` has no version (written as "latest at the time").
- Finding F-026: with the new R12 (pre-approved in P004), S01.7 needs a documentation-audit task before sign-off.

### I. CURRENT_STATE.md: **findings F-027 (Critical), F-028, F-016**
- Pass: plan version, phase, active stage and task, blockers, and pointer paths match reality at the start of S004. The file is 46 lines (under 150). The "In progress" entry was clean ("none", S003 closed) before this audit wrote its own write-ahead.
- Finding F-027 (**Critical**): next step 4 tells a fresh agent to branch `feat/S01.1-T01-go-module` off `develop`. But `develop` (`c34cc21`) contains **no** `code-agent-docs/`, so the agent would start S01 without rules or plan. The S003 closing summary had the condition "(after the docs PRs are merged)", but CURRENT_STATE did not.
- Finding F-028: the first "Next steps" item is a **user** action ("The user reviews…"), not something a fresh agent can act on.
- F-016 applies here too: Q37 and Q38 are listed under "later stages" although they block the baseline approval.

### J. Session logs and git: **findings F-003, F-005, F-030, F-031**
- Pass: session numbers S001–S004 are sequential. S001, S002, and S003 each have Startup, Entries, and a Closing summary (S004 is exempt until it ends).
- Pass: every commit on every branch is referenced by hash in some session log (script). Log-closing commits are recorded in the next session's startup (`c8bcee7` in S002, `828b95c` in S004).
- Pass: agent commit messages follow R7: Conventional Commits with `[S001]`/`[P002]`/`[P003]` identifiers, since no task IDs exist yet. The user's own commits (`add readme.md`, and so on) are outside R7.
- Pass: R9 archiving is not triggered (4 logs, fewer than 20).
- Git status: the documentation changes of S001–S003 are committed and pushed. The workflow deviations are F-005; the untracked root prompt files are F-003.
- Findings:
  - F-030: S002 E003 estimated S001's close as "about 23:50", but git shows 23:47.
  - F-031: S002 E009 used a non-clock timestamp.

### K. Cross-document consistency: **findings F-032 (plus F-008, F-010, F-011, F-013 found here)**
- Pass: requirement, ADR, stage, substage, task, and invariant IDs are identical everywhere they appear (script: no undefined FR/NFR/ADR reference in any live document).
- Pass: all relative Markdown links resolve, and every backticked `code-agent-docs/…` path exists (live documents).
- Pass: terminology. "files area" and "photos area", the sidecar naming `IMG_0001.jpg.json`, operator names, and component names are used consistently. "internal data" and "internal app data" are used as synonyms (acceptable).
- Pass: dates are consistent and plausible. S002 crosses midnight and says so. Plan 0.2.0 is dated 2026-09-24 because it was written after midnight. RULES 1.1.0 rows are dated 2026-09-23 because they were applied before midnight. Exception: F-030.
- Placeholder search (templates excluded):
  - False positives: "placeholder" as a feature term (Photos/Settings placeholder pages; S01.2-T06 "Photos area placeholder"); the file-name patterns `<YYYY-MM-DD>_S<NNN>` in RULES; the S01 change-log row describing removed placeholders.
  - Real deferred markers are F-032.

### L. Resumability test: **initial run failed on 2 questions (F-027, F-028); re-run in section 6**

## 4. Findings

| ID | Severity | Group | Location | Description | Outcome | Fix reference / options and recommendation |
|---|---|---|---|---|---|---|
| F-001 | Minor | A | RULES.md, documentation map | `audits/` (P004) is not in the documentation map | Fixed | Pre-approved P004 change (a): map entry added |
| F-002 | Minor | A | RULES.md, documentation map | The map says "Empty folders contain a .gitkeep", but `stages/` and `decisions/` are no longer empty and keep theirs (R9 forbids deleting) | Fixed | Wording clarified: `.gitkeep` files are kept (R9) |
| F-003 | Minor | A, J | Repository root `prompts/`; git status | The user's original prompt files live in root `prompts/`, which no document describes. `prompts/3-…` and `prompts/4-…` are **untracked** (1 and 2 were committed by the user) | Fixed (description) + **Needs user decision** (commit) | Map now describes root `prompts/` as user-managed originals. Decision D-10: commit them yourself as before, or let the agent include them in the audit PR. **Recommendation:** let the agent include them |
| F-005 | Major | C | RULES.md User Preferences; git branches | The preference is "one branch per stage/task off `develop`, pushed, PR into `develop`". Since S002, branches are **stacked** on unmerged docs branches (reason logged), and **no PR was opened** (`gh` not installed). This is a deviation without an explicit approval | **Needs user decision** | D-03: (a) open/merge the docs PRs in order now; (b) install and authorize `gh` so the agent can open PRs; (c) add a preference clarifying that stacking is allowed while earlier docs PRs are unmerged. **Recommendation:** (a) + (c), (b) optional |
| F-006 | Minor | C, L | RULES.md, documentation map | Startup evidence: session IDs `S001` (three digits) look like stage IDs `S01` (two digits) | Fixed | The map now states the distinction explicitly. The naming itself is unchanged |
| F-007 | Minor | C, L | RULES.md R1 step 3 | Startup evidence: R1 requires reading plan.md **completely** every session. It is 2,536 lines, which makes startup slow and context-heavy | **Needs user decision** (rule change) | D-09: allow "read sections 2a, 5, 7, 10.1 and the active stage's section of 10 completely; other sections on demand". **Recommendation:** approve |
| F-008 | Minor | D | plan.md S01.1 "Depends on" | The text "The S01.1 ADRs were Accepted via P003" makes S01.1 appear to depend on itself (the script reported a cycle) | Fixed | Reworded to "the stack ADRs (0001, 0002, 0004–0007)" |
| F-009 | Major | D | plan.md 10.14 | Six planner changes to the user's substages have **no approval record**: S01 and S04 execution order; S03.2 scope; S04.2 host import; S11.7 extended to include the stage review (so S11 ends with review under the P002 name "Release"); S12.6 tag rejection | **Needs user decision** | D-02: approve or reject each with the plan baseline. **Recommendation:** approve all six |
| F-010 | Minor | D | plan.md 6.1, GUI row | "Web app served by the NAS (pending Q25)" is stale; Q25 was answered by P003 (ADR-0009) | Fixed | Updated to SvelteKit static SPA, ADR-0009 |
| F-011 | Major | D | plan.md S12 design notes | "The model licensing ADR covers optional non-permissive packs (Q16)" **contradicts** Accepted ADR-0018 and NFR-029 (non-permissive weights such as InsightFace are excluded) | Fixed | Aligned with ADR-0018 (the ADR is authoritative for technology decisions) |
| F-012 | Major | D | plan.md S12.10 AC 2; FR-054 | S12.10 says semantic search results are "never computed at query time". P003 notes that semantic search needs a text model **at query time**, an exception to I4 that needs explicit approval. The plan text and the P003 note conflict | **Needs user decision** | D-06: (a) approve a narrow I4 exception for semantic search (local text encoder in the AI worker at query time; search works without it); (b) drop semantic search; (c) allow only precomputed forms (e.g. tag-vocabulary embeddings computed offline, no query-time model). **Recommendation:** decide with Q36; (c) by default, (a) only if you want free-text semantic search |
| F-013 | Major | D | plan.md 6.3 layout | `logs/` is described as "application + audit logs", but Accepted ADR-0007 stores the **audit log in SQLite** | Fixed | 6.3 aligned with ADR-0007 |
| F-014 | Major | D, H | plan.md FR-130/S10.5 vs S01 doc S01.1-T08 | S01 sends application logs to **stderr only**, but FR-130/S10.5 need an in-app **log viewer**, which requires logs the app can read. This is a forward-compatibility gap in an agent-chosen detail | **Needs user decision** | D-07: (a) stderr + a size-rotated JSON log file in `.local-ai-nas/logs/` from S01 (small in-house rotator, no dependency); (b) stderr only, with S10.5 reading journald/Docker logs (platform-specific); (c) recent log lines kept in SQLite. **Recommendation:** (a) |
| F-015 | Minor | D | plan.md Q16 | Q16 is marked "answered by implication of P003". The decision it gated (face model licensing) is decided by P003/NFR-029, but its sub-question (commercial use or distribution) was never answered by the user | **Needs user decision** | D-11: confirm Q16 is closed. **Recommendation:** confirm |
| F-016 | Major | D, I | plan.md section 5; CURRENT_STATE | The decisions needed for **plan baseline approval** (Q38 "needed by: roadmap baseline", Q37, the 10.14 flags, ADR-0003) are not grouped anywhere. CURRENT_STATE lists Q37/Q38 under "later stages" | Fixed | A "Needed for baseline approval" list was added to plan section 5 and CURRENT_STATE, derived only from the existing "Needed by" values |
| F-017 | Minor | D | plan.md section 5 legend | "★ = needed before S01 implementation can start", but Q1 and Q18 are only needed for the S01.7 performance baseline (their own "Needed by", and the S01 doc) | Fixed | Legend clarified. The per-question "Needed by" values are unchanged |
| F-018 | Major | E | README.md | The README predates approved decisions: the two-area layout (I1/I2), search across both areas and the operators `in:`/`ext:`/`size:` (P002), multi-user and sharing (P002), the single-writer rule (I9), the technology stack (Accepted ADRs), the deployment plan (ADR-0006), and the roadmap order (P002 user-defined stages) | **Needs user decision** | D-05: approve, edit, or reject `audits/A001-readme-proposal.md` (README itself unchanged) |
| F-019 | Minor | F | `decisions/ADR-0001-backend-language-framework.md` | The file name does not match the title "Core server language: Go". The name was kept in S003 because session logs (append-only) link to it | Fixed | Documented exception added to ADR-0001's history note |
| F-020 | Minor | F | ADR-0001, 0002, 0004, 0005, 0009, 0017, 0018 | These Accepted ADRs were amended within S003 after first being written (golangci-lint pinning correction; version mentions). This is logged in S003 E011/E014. No decision changed | Accepted as-is | Informational: recorded for transparency |
| F-021 | Major | F, G | `dependencies.md` TypeScript row; ADR-0009 | svelte-check 4.7.6 supports TypeScript `^5 \|\| ^6` only, but the register lists TypeScript 7.0.2 | Fixed | The register pins **TypeScript 6.0.3**. A dated verification note was appended to ADR-0009, whose decision already says "pin the version svelte-check supports" (decision unchanged) |
| F-022 | Major | F, G | `dependencies.md` FFmpeg row; ADR-0006/0012 | Verified: **Debian trixie FFmpeg is a GPL-3.0-or-later build** (`--enable-gpl --enable-version3`). The core image would ship GPL binaries as separate programs | Fixed (register) + **Needs user decision** | Register row verified and flagged ⚠. D-04 (with Q22): (a) accept Debian's GPL build as a separate program and provide source offers in third-party notices; (b) build an LGPL-only FFmpeg for the image; (c) make video features optional and ship FFmpeg as a separate add-on. **Recommendation:** (a) for now; revisit at S11.1 once the project license is chosen |
| F-023 | Minor | F, G | `dependencies.md` libheif/libde265 row | libde265 (trixie 1.0.15) is LGPL-3+ for the library; Debian's copyright file also lists GPL-3+ and BSD-4-clause parts | Fixed | Register row updated and flagged ⚠ |
| F-024 | Major | G | `dependencies.md` | Missing: Samba; Docker Engine/Compose/buildx; CI GitHub Actions (checkout v7.0.1, setup-go v7.0.0, golangci-lint-action v9.3.0, setup-buildx-action v4.4.1, trivy-action v0.36.0); axe-core 4.13.0 (MPL-2.0); sqlite-vec (MIT OR Apache-2.0); LibRaw (LGPL-2.1 OR CDDL-1.0); ONNX Runtime GPU EPs; golang.org/x/text v0.42.0. There is also no list of alternatives named in ADRs | Fixed | Rows added with verified versions and licenses (statuses Candidate/Deferred); new section "Alternatives named in ADRs" |
| F-025 | Major | H | `stages/S01-basic-nas.md` section 7 | The dependency table has no ADR column; `golang.org/x/text` has no version ("latest at the time") and no register entry | Fixed | ADR column added; x/text pinned as v0.42.0 (conditional use); register row added |
| F-026 | Minor | H | `stages/S01-basic-nas.md` S01.7 | R12 (pre-approved in P004) requires a documentation audit in every final review substage; S01.7 had none | Fixed | New task S01.7-T07 "Documentation audit (R12)"; completion and sign-off renumbered to S01.7-T08 (document not yet approved, so this is allowed; change-log row added) |
| F-027 | **Critical** | I, L | `CURRENT_STATE.md` next step 4 | It tells a fresh agent to branch off `develop`, which contains no `code-agent-docs/`. The agent would resume without rules or plan | Fixed | Prerequisite added: the docs PRs must be merged into `develop` in order first; until then, branch off the latest docs branch and say so |
| F-028 | Major | I, L | `CURRENT_STATE.md` next step 1 | The first next step is a user action, so a fresh agent cannot act on it immediately | Fixed | Rewritten as the agent's concrete first action (present the pending decisions from A001 section 7 and wait; no S01 code before approval) |
| F-030 | Minor | J | S002 log E003 | Estimated S001's close as "about 23:50"; git shows commits `2b5c1a7` and `c8bcee7` at 23:47 | Fixed | Reconciliation entry in the S004 log (the past entry is untouched, per R2) |
| F-031 | Minor | J | S002 log E009 | Timestamp "00:0x (sequence)", although the clock was available (R2) | Accepted as-is | Past entries cannot be edited. Noted in the S004 reconciliation entry |
| F-032 | Minor | K | `dependencies.md`; ADR-0015 | Deferred markers remain: synonym dictionary "TBD" (S06.5); RapidOCR model weights (partially verified, S12.10); Python license (unverified, S12.1); ADR-0015 details "to be confirmed in S09.1 with the threat model" | Deferred | Each names the stage that resolves it. No action needed now |
| F-033 | Minor | K | RULES.md documentation map (stages row) | The example stage document `S00-foundation.md` is obsolete (no S00 exists since plan 0.2.0) | Fixed | Example changed to `S01-basic-nas.md` (RULES 1.3.0 changelog row e) |
| F-034 | Minor | K | `archive/plan-history/plan_v0.3.0.md` (and earlier copies) | Relative links in archived plan copies (e.g. `decisions/ADR-…`) do not resolve from `archive/plan-history/`. They are verbatim copies and must stay unchanged | Fixed (documented) | Documentation-map note added (RULES 1.3.0 changelog row e); archives left verbatim |

## 5. Prompt fulfillment matrix

Evidence refers to files, sections, and commits. "Done*" means done with a deviation that is recorded as a finding.

### 5.1 Bootstrap prompt (definition_of_done)
| # | Item | Status | Evidence |
|---|---|---|---|
| 1 | code-agent-docs/ exists with the full folder structure | Done | Inventory (1.1); commit `2b5c1a7` |
| 2 | RULES.md contains R1–R11 in full, plus the extra sections | Done | Group C (normalized text comparison: every sentence present) |
| 3 | AGENTS.md at root (+ editor-native pointer) | Done | `AGENTS.md`, `CLAUDE.md` (18 lines each) |
| 4 | Three templates exist | Done | `templates/` |
| 5 | plan.md 0.1.0 with all 14 sections | Done | `archive/plan-history/plan_v0.1.0.md` (= `c8bcee7`) |
| 6 | CURRENT_STATE reflects plan awaiting review | Done (at the time) | `c8bcee7` |
| 7 | Session log S001 complete with closing summary | Done | `logs/sessions/2026-09-23_S001.md` |
| 8 | Prompt archived verbatim | Done | `bootstrap/initial-prompt.json` (byte-identical to root `prompts/1-…`) |
| 9 | No application code, no dependencies installed | Done | Git history: documentation only |
| 10 | User received summary + open questions | Done | S001 E018 |
- Execution steps 1–13: Done. Step 12 (git preference) is answered in S001 E014 and recorded in User Preferences.

### 5.2 P002 (definition_of_done)
| # | Item | Status | Evidence |
|---|---|---|---|
| 1 | Prompt archived as P002 | Done | `prompts/P002-…json` (byte-identical) |
| 2 | Previous plan archived | Done | `plan_v0.1.0.md` |
| 3 | plan.md bumped, 12 stages, every substage and field | Done | plan 0.2.0 (`5ab149b`); group D |
| 4 | S12 last; every stage ends with testing/review | Done* | S11.7 "Release" with scope extended to include the review (plan 10.14 item 5). Approval pending: F-009 |
| 5 | RULES has the pre-approved changes, invariants, changelog | Done | RULES 1.1.0 rows (a)–(e) |
| 6 | Stage template includes substages | Done | `templates/stage-template.md` section 5 |
| 7 | S01-basic-nas.md, Planned, full task breakdown | Done | `stages/S01-basic-nas.md` |
| 8 | S01.1 ADRs exist as Proposed | Done (at the time) | ADR-0001/0002 (+0003) in `5ab149b`. Later accepted by P003 |
| 9 | CURRENT_STATE and session log current, with closing summary | Done | S002 log closing summary |
| 10 | No code, no dependencies | Done | Git history |
| 11 | User received the step 11 report | Done* | Delivered combined with the P003 report at the user's request (S002 E010–E014, E019) |
- Pre-approved changes (a)–(e): Done (RULES changelog 1.1.0). Execution steps 1–11: Done.

### 5.3 P003 (definition_of_done)
| # | Item | Status | Evidence |
|---|---|---|---|
| 1 | Prompt archived as P003 | Done | `prompts/P003-…json` (byte-identical) |
| 2 | All ADRs exist with correct statuses, versions, licenses (or Unverified) | Done | ADR-0001–0019; S003 E005/E007/E011 |
| 3 | dependencies.md lists every dependency, tool, dataset, model | Done* | Created; gaps found by this audit (F-024) |
| 4 | RULES has I9, register in the map, R6 addition, changelog | Done | RULES 1.2.0 rows (a)–(c) |
| 5 | plan.md bumped, chosen-stack table, architecture, resolved questions, risks, revision entry | Done* | plan 0.3.0. "AI speed on CPU-only" was done by updating RK-06, not as a new row. Stale leftovers: F-010, F-011 |
| 6 | S01 doc concrete, no placeholders | Done* | `18d9331`. Residual: x/text version (F-025) |
| 7 | CURRENT_STATE and session log current, with closing summary | Done* | S003 closing summary. CURRENT_STATE hazards: F-027, F-028 |
| 8 | No code, no dependencies | Done | Git history |
| 9 | User received the step 14 report | Done | S003 E016 |
- Pre-approved changes (I9, register, R6 rule): Done. Execution steps 1–14: Done.

## 6. Resumability test

**Method:** act as a new agent with no chat history. Read `AGENTS.md` (or the auto-loaded `CLAUDE.md`), then follow R1: `RULES.md` → `CURRENT_STATE.md` → `plan.md` → the newest file in `logs/sessions/` → the stage document and its ADRs. Answer only from the documents.

### 6.1 Initial run (documents as of `828b95c`, start of S004)
| Question | Answer found | Clear and correct? |
|---|---|---|
| What is this project? | A self-hosted NAS with separate files and photos areas, sidecar JSON metadata, forgiving search, multi-user sharing, and optional local AI last (plan section 1, README) | Yes |
| What stage and task are active? | None. S01 is Planned (CURRENT_STATE, stage document) | Yes |
| What was the last completed action? | P003 applied (CURRENT_STATE "Last completed"; S003 closing summary) | Yes |
| What exactly is the next action? | CURRENT_STATE item 1 is "The user reviews …", a user action. Item 4 says to branch off `develop` | **No.** The agent's own next action was unclear (F-028), and the branching instruction would lose the docs (F-027, **Critical**) |
| What is waiting on the user? | Plan and S01 approval, ADR-0003, Q22, Q1, Q18 | Partly: the baseline-approval questions Q37/Q38 were filed under "later stages" (F-016) |
| What rules must I follow? | RULES.md R1–R11, invariants I1–I9, User Preferences | Yes |

**Result: failed** on the next-action question (Critical F-027, Major F-028) and partly on the waiting-on-user question (F-016).

### 6.2 Re-run after fixes (S004)
| Question | Answer found | Clear and correct? |
|---|---|---|
| What is this project? | As above | Yes |
| What stage and task are active? | None. S01 Planned. Audit A001 in progress, with a queued video-streaming plan change (CURRENT_STATE "In progress") | Yes |
| What was the last completed action? | P003 during the audit. After step 11: "Documentation audit A001" | Yes |
| What exactly is the next action? | CURRENT_STATE next step 1: present the A001 section 7 decisions and the review request to the user and wait; no S01 code before approval. Step 4 says to merge the docs PRs in order before any S01 branch, or branch off the latest docs branch and tell the user | **Yes** |
| What is waiting on the user? | The baseline decisions (Q38, Q37, the 10.14 flags, ADR-0003), the other A001 decisions, Q22 for S01.1-T02, and Q1/Q18 for S01.7 | **Yes** |
| What rules must I follow? | RULES.md R1–R12 (R12 new), invariants I1–I9, User Preferences | Yes |

**Result: pass.**

## 7. Decisions needed from the user

No Critical finding needs a decision; the only Critical finding (F-027) is fixed. The decisions are listed in the order they matter for the **plan baseline approval (1.0.0)** and for starting S01.

| # | Decision | Source | Options | Recommendation |
|---|---|---|---|---|
| D-01 | **Storage layout** (ADR-0003, Proposed) | Baseline; gates S01.2 | Accept · change (e.g. usernames instead of IDs as folder names) · reject | **Accept**: per-user folders (`files/u0001/`, `photos/u0001/`) from day one, internal data in `.local-ai-nas/` |
| D-02 | **Planner changes to your substages** (plan 10.14) | F-009; baseline | Approve all · approve some · reject | **Approve all six**: S01/S04 execution order, S03.2 scope, S04.2 host import, S11.7 includes the review, S12.6 tag rejection |
| D-03 | **Git workflow** (stacked branches, no PRs) | F-005 | (a) open/merge the docs PRs in order now · (b) install and authorize `gh` so the agent opens PRs · (c) add a preference allowing stacked docs branches while earlier PRs are unmerged | **(a) + (c)**; (b) optional |
| D-04 | **FFmpeg licensing** in the Docker image (the Debian build is GPL-3.0+) | F-022; with Q22 | (a) accept it as a separate program with source-offer notes · (b) build an LGPL-only FFmpeg · (c) ship video tools as an optional add-on | **(a)** for now; revisit at S11.1 |
| D-05 | **README proposal** R-01–R-09 | F-018 | Approve · edit · reject per item | **Approve R-01–R-08**; R-09 (technology section) optional |
| D-06 | **Semantic search vs invariant I4** | F-012; with Q36 | (a) approve a narrow I4 exception (query-time text model, optional) · (b) drop semantic search · (c) precomputed-only variant | **Decide with Q36**; default (c) |
| D-07 | **Application log persistence** (for the S10.5 log viewer) | F-014 | (a) stderr + a rotated JSON file in `.local-ai-nas/logs/` from S01 · (b) stderr only; the viewer reads journald/Docker · (c) recent logs in SQLite | **(a)** |
| D-08 | **First usable release (Q38)** and **not-scheduled candidates (Q37)** | F-016; baseline | Q38: M1 (S01–S03) / **M2 (S01–S06)** / M3 (S01–S11). Q37: add mobile backup, public links, remote access, or none now | **Q38: M2** (with the no-trash caveat, plan 11). **Q37: none now** |
| D-09 | **Startup reading rule** (R1 reads all of plan.md, 2,500+ lines) | F-007; rule change | Keep · allow a targeted read (2a, 5, 7, 10.1, the active stage section; the rest on demand) | **Allow the targeted read** (RULES change with changelog) |
| D-10 | **Untracked `prompts/3-…` and `prompts/4-…`** in the repository root | F-003 | You commit them as before · let the agent include them in the audit commit | **Let the agent include them** |
| D-11 | **Close Q16** (face model licensing) | F-015 | Confirm closed · keep open | **Confirm closed** (NFR-029 decides it) |
| D-12 | **Project license (Q22)**; needed before S01.1-T02 | ★ S01 | MIT · Apache-2.0 · GPL-3.0 · AGPL-3.0 | **AGPL-3.0** keeps network-hosted forks open, which is common for self-hosted apps. **Apache-2.0** if maximum reuse matters more. Every linked dependency is compatible with either |
| D-13 | **Q1 (hardware) and Q18 (library size)**; needed before S01.7 | ★ S01.7 | Answer when convenient | Answer before S01.7 so the performance targets can be set |

After these decisions: **approve plan.md as the 1.0.0 baseline** (R4) and **approve `stages/S01-basic-nas.md`** (R3). Implementation can then begin with S01.1-T01.

## 8. Closing: what the next audit should watch

1. **Plan size and startup cost** (F-007): plan.md is 2,500+ lines and still growing. Check whether R1 reading remains practical.
2. **Register drift:** every dependency added during S01 must land in `dependencies.md` in the same commit (R6). Compare `go.mod`, `pnpm-lock.yaml`, and `uv.lock` against the register.
3. **CURRENT_STATE branching instructions** after the docs PRs merge (F-027): make sure the next steps match the real branch state.
4. **Supersession links:** after the video-streaming change, ADR-0020 must link to ADR-0012 and ADR-0012 back to ADR-0020 (partial supersession).
5. **Open decisions D-01–D-13:** confirm that each was recorded (User Preferences, ADR approval records, plan revision entries) and that nothing was applied without an approval record.
6. **Unverified and deferred items:** synonym dictionary (S06.5), RapidOCR model weights (S12.10), Python license (S12.1), ADR-0015 details (S09.1). Re-check them at their stages.
7. **Timestamps:** the agent twice typed a timestamp by hand (S001, S003) and caught it. Keep checking log and CURRENT_STATE times against git commit times.
8. **Scripts:** the A001 check scripts lived in the agent's scratchpad, not the repository. If you want audits to be reproducible by others, decide whether to commit them (e.g. under `scripts/audit/`).
