# Documentation audit checklist (reusable)

<!--
Template for documentation audits under RULES.md R12.
Created in audit A001 (session S004) from P004. Written generally so that every later audit (A002, A003, ...) can reuse it.
UPDATE THIS CHECKLIST whenever a new document type, folder, rule, or invariant is added (R12 / P004).

How to use:
1. Copy the "Report structure" section into code-agent-docs/audits/A<NNN>-<YYYY-MM-DD>-<slug>.md and fill it AS YOU GO. An interrupted audit must keep its progress.
2. Write the in-progress entry to CURRENT_STATE.md first (R2 write-ahead).
3. Work through groups A to L in order. Record "Pass" or finding IDs (F-001, ...) per check.
4. Classify every finding (severity + outcome), apply the direct fixes (Critical first), prepare the decisions, re-check, then report.
-->

## Sources of truth (order of authority)
1. The user's recorded instructions: `bootstrap/initial-prompt.json`, every `prompts/P<NNN>-*.json`, and the verbatim USER entries in session logs.
2. `README.md` (the project vision).
3. `RULES.md`.
4. `plan.md`.
5. ADRs, stage documents, `dependencies.md`, audits.
6. Git history (`git log --all`, `git show`) as evidence of what actually happened.

## Fix policy
- **Fix directly** (log each fix with its finding ID):
  - Required files or sections whose content is fully specified or recorded elsewhere.
  - Broken links and paths.
  - Unambiguous ID, version, numbering, or cross-reference mismatches.
  - A stale CURRENT_STATE.
  - Register/ADR drift.
  - Missing changelog rows for approved changes.
  - Missing archived plan versions recoverable exactly from git.
  - Formatting, typos, or unclear wording that does not change meaning.
- **Needs user decision** (record the options and a recommendation; do not change):
  - Anything that changes scope, requirements, stage content, invariants, or an Accepted ADR.
  - Conflicts where it is unclear which document is correct.
  - Rule changes that are not pre-approved.
  - README changes (write a proposal file instead: `audits/A<NNN>-readme-proposal.md`).
  - Any missing approval record.
- **Never:**
  - Invent history (use "Unknown").
  - Create or backfill an approval the user did not give ("Approval not recorded").
  - Edit past session-log entries (add a new entry, or a clearly labelled retroactive note at the end of the old log).
  - Mark something verified that was not actually verified.
- **Plan versioning:** if plan.md changes, archive it first and make one PATCH bump for the whole audit, with a revision entry citing the audit and its trigger.

## Severity levels
- **Critical:** would make a fresh agent resume incorrectly, break an invariant, or misrepresent an approval. Examples: a wrong next step, a missing rule, a missing pointer file, a claimed approval that was never given.
- **Major:** a required deliverable or section is missing, traceability is broken, or IDs, versions, or statuses are inconsistent.
- **Minor:** formatting, typos, or wording that does not change meaning.

**Outcomes:** Fixed · Needs user decision · Deferred (names the stage that resolves it) · Accepted as-is (informational, with a reason).

## Check groups

### A. Structure
- [ ] Every folder in the RULES.md documentation map exists.
- [ ] Every required file exists: RULES, CURRENT_STATE, plan, dependencies, all templates, bootstrap prompt, every archived prompt, audit reports.
- [ ] Root files exist: README.md, AGENTS.md, and every editor-native pointer file listed in RULES.md.
- [ ] No orphaned files: every file is covered by the documentation map (or the map needs an entry).
- [ ] `archive/plan-history/` holds every superseded plan version listed in the revision history, and each equals `git show <commit>:code-agent-docs/plan.md`.

### B. Prompt fulfillment
- [ ] For every user prompt: each definition_of_done item, execution step, and pre-approved change is marked Done / Partial / Missing with evidence.
- [ ] Every archived prompt is valid, complete (not truncated), and byte-identical to the user's original.
- [ ] Every prompt and user message has a USER entry in a session log.

### C. Rules and pointers
- [ ] RULES.md contains every rule (R1 onward) fully written out, with nothing summarized away compared with the prompts that defined them (normalized text comparison).
- [ ] Every approved rule change is present, and every changelog row cites its approval (prompt or user message).
- [ ] Required sections exist: quick start, project invariants, documentation map, pointer files, User Preferences, changelog.
- [ ] User Preferences records the answers the user gave (commit behavior, …). An unanswered question is an open finding.
- [ ] No two rules contradict each other, and the actual behavior (git log) follows the recorded preferences.
- [ ] Pointer files are short, point to correct paths, and give the correct reading order.

### D. Plan
- [ ] The header version equals the latest revision entry, and every previous version is archived.
- [ ] Every required section is present and non-empty.
- [ ] All stages are present and in order. The AI stage is last (I8).
- [ ] Every substage has all required fields: goal, scope, deliverables, depends on, requirement IDs, acceptance criteria (2–5), risks/notes, status.
- [ ] Every stage ends with a testing and review substage that includes the R12 audit.
- [ ] Traceability: every FR/NFR maps to a substage; every referenced ID exists; no duplicates; deprecated IDs are marked, not deleted.
- [ ] Dependencies: every "depends on" exists; no cycles; no dependency on a later stage.
- [ ] Open questions: answered ones cite where; open ones name what they block; none answered without a user decision; baseline-approval decisions are grouped.
- [ ] Invariants are worded identically in plan.md and RULES.md.
- [ ] The chosen-stack table matches the ADRs' current statuses and links.
- [ ] No stale text contradicts a newer decision (grep for superseded technologies and answered-question markers such as "pending Q…").

### E. README alignment
- [ ] Every README feature maps to at least one requirement.
- [ ] Every place where README.md is out of date with **approved** decisions is written up as a proposal (current text → proposed text → reason with source), never as a direct edit.

### F. ADRs
- [ ] Numbering is sequential with no gaps or duplicates. File names match titles (or the exception is documented).
- [ ] Every ADR has every template section.
- [ ] Every status is valid and matches what the approving prompt or user specified. Accepted ADRs cite their approval.
- [ ] Superseded ADRs link both ways (including partial supersession notes).
- [ ] Every ADR is linked from plan.md, and dependency-bearing ADRs are linked from the register.
- [ ] Every item marked "Unverified" gets a new verification attempt, and the result is recorded.

### G. Dependency register
- [ ] Every dependency, external tool, dataset, and model named in plan.md, an ADR, or a stage document is in the register (alternatives named only in "Options considered" go in the alternatives list).
- [ ] Versions and licenses match the ADRs and stage documents.
- [ ] Every row has a verification status. License items needing the user's attention are flagged (⚠).

### H. Stage documents
- [ ] Only the active or next stage has a stage document (just-in-time rule). Any others are flagged.
- [ ] Each stage document covers all of its substages, with tasks `S<NN>.<n>-T<NN>`, each with acceptance criteria.
- [ ] It contains every R3 section, including the approval record (empty if not approved) and the R12 audit task in the final substage.
- [ ] No placeholders remain.
- [ ] Every listed dependency has an ADR link and a register entry.
- [ ] Its status is consistent with CURRENT_STATE.md.

### I. CURRENT_STATE.md
- [ ] Every field matches reality: plan version, phase, active stage and task, last completed, next steps, blockers, open questions, pointers.
- [ ] No stale "In progress" entry (a leftover means an interrupted action: investigate and record).
- [ ] Under about 150 lines.
- [ ] The first next step is something a **fresh agent** can do immediately, and every git or branching instruction works on the current branch state (e.g. do not branch off a branch that lacks `code-agent-docs/`).

### J. Session logs and git
- [ ] Session numbers are sequential. Every closed log has startup, entries, and a closing summary.
- [ ] Every commit (all branches) that touches documentation corresponds to a logged action. Unlogged commits get a reconciliation entry in the current log.
- [ ] `git status` is clean or explained. Commit messages follow R7. The branch and PR workflow follows User Preferences.
- [ ] R9 archiving is applied if more than 20 logs exist.
- [ ] Timestamps come from the system clock and agree with git commit times.

### K. Cross-document consistency
- [ ] Stage, substage, task, requirement, ADR, and invariant IDs are identical everywhere.
- [ ] Terminology is consistent (areas, sidecar naming, operators, component names).
- [ ] All relative links and backticked `code-agent-docs/…` paths resolve.
- [ ] Placeholder search (TODO, TBD, FIXME, XXX, "to be confirmed", "placeholder", `<date>`, empty sections). Each hit is resolved, deferred with a named stage, or explained as notation.
- [ ] Dates are consistent and plausible across logs, revision history, ADRs, and changelogs.

### L. Resumability test
- [ ] Acting as a brand-new agent with no chat history, read only AGENTS.md, then follow the R1 reading order.
- [ ] Answer from the documents alone: What is this project? What stage and task are active? What was the last completed action? What exactly is the next action? What is waiting on the user? What rules must I follow?
- [ ] Any question that cannot be answered clearly and correctly is a **Critical** finding.

## Report structure (`audits/A<NNN>-<YYYY-MM-DD>-<slug>.md`)
1. Header: audit ID, date, session, trigger, plan version at start and at end, status.
2. Scope: files and sources audited, method, interpretation notes, inventory.
3. Summary: finding counts by severity and by outcome.
4. Check results: every group A–L with "Pass" or finding IDs.
5. Findings table: ID, severity, group, location, description, outcome, fix reference or options + recommendation.
6. Prompt fulfillment matrix.
7. Resumability test: questions and answers (initial run and re-run after fixes).
8. Decisions needed from the user: a numbered list with options and recommendations, Critical first.
9. Closing: what the next audit should watch.

**Automation hint:** the checks for A, B, C, D, F, H, J, and K can be scripted (structure, JSON validity, text comparison, ID/link/placeholder scans, git-vs-log hashes). Keep scripts outside the repository unless the user asks to commit them.
