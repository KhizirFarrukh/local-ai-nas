# RULES: Permanent Operating Rules for AI Agents on local-ai-nas

**RULES.md version:** 1.9.0
**Created:** 2026-09-23 (session S001)
**Source:** `operating_rules` in `code-agent-docs/bootstrap/initial-prompt.json`, transcribed in full with the original rule IDs.

## Purpose

These are the permanent operating rules for any AI agent working on **local-ai-nas**. Every session follows them without the user repeating them. They exist so the project can survive a total loss of chat memory: everything important is written to disk continuously, and a fresh agent instance can resume exactly where the previous one stopped.

**Mindset.** You are a disciplined engineer who plans before building and never jumps straight to writing code. Every unit of work follows this cycle: **plan, record, get approval, execute, test, review, record again.** Documentation is part of the product, not an afterthought. Assume your chat memory can be wiped at any moment. When something is uncertain, ask. When you assume, write the assumption down.

---

## Quick start (R1 startup checklist)

Do this at the start of **every** session and every new chat, even if the user's first message asks for something else:

1. Read `code-agent-docs/RULES.md` completely (this file).
2. Read `code-agent-docs/CURRENT_STATE.md`.
3. Read `code-agent-docs/plan.md`: sections 2a (invariants), 5 (open questions), 7 (chosen stack), and 10.1 (roadmap overview), plus the **complete** section of the active (or next) stage in section 10. Read the other sections when the work needs them. _(Changed in RULES 1.4.0, approved by the user in S005, decision D-09.)_
4. Read the most recent session log in `code-agent-docs/logs/sessions/`. If it has no closing summary, also read the one before it and reconstruct what was in progress.
5. Read the stage document for the active stage (`code-agent-docs/stages/`) and any ADRs it references (`code-agent-docs/decisions/`).
6. Run `git log` and `git status`. Report any changes that the logs do not describe to the user before continuing.
7. Create a new session log with the next sequential session number.
8. Give the user a short resume summary (current stage and task, last completed action, next planned action, open questions or blockers). Then proceed, or wait for confirmation if the next action is ambiguous.

**No code may be written until this checklist is complete.**

---

## Project invariants

No code, plan change, stage document, or ADR may violate these invariants without **explicit user approval**. A change to an invariant is itself a RULES.md change: it needs approval and goes in the changelog. They are repeated in `plan.md` ("Project invariants").

- **I1:** The storage root contains exactly two user-data areas: `files/` and `photos/`. They never intersect. An item moves between them only when the user explicitly copies or moves it.
- **I2:** Internal application data (database, search index, caches, thumbnails, trash, configuration) lives outside `files/` and `photos/`.
- **I3:** Photo sidecar JSON files are the source of truth for descriptive photo metadata. Ownership, access, and sharing are authoritative in the database; sidecars keep a read-only mirror of them. The search index is a cache that can always be rebuilt from disk.
- **I4:** Search reads only the index. No model reads user content in a request path. The one exception: when the optional AI worker is running and semantic search is enabled, a small local text model may embed the typed search text, within a strict time limit, falling back to normal search.
- **I5:** A user cannot access another user's files or photos unless they have been explicitly shared. This is enforced server-side on every access path: API, downloads, previews, thumbnails, search, network shares, and background jobs.
- **I6:** Everything runs locally. No telemetry. No network calls at runtime except for features the user has explicitly enabled.
- **I7:** AI is optional and opt-in. The NAS must be fully functional with AI disabled.
- **I8:** AI work is always the last stage of the roadmap. Any stage added in the future is inserted before it, and the AI stage is renumbered.
- **I9:** Only the core server writes sidecar files and the search index. Other processes, including the AI worker, submit results to the core server, which validates and writes them.
- **I10:** Destructive bulk operations (deleting duplicates, reducing media resolution or quality, creating or changing drive pools) always show a preview of what will change, require explicit confirmation, and keep an undo window wherever technically possible. Where undo is impossible (e.g. erasing drives to create a pool), the confirmation says so plainly.

---

## R1: Session start protocol

**Applies:** At the start of every session and every new chat, without exception, even if the user's first message asks for something else.

**Steps:**

1. Read `code-agent-docs/RULES.md` completely.
2. Read `code-agent-docs/CURRENT_STATE.md`.
3. Read `code-agent-docs/plan.md`: sections 2a (invariants), 5 (open questions), 7 (chosen stack), and 10.1 (roadmap overview), plus the **complete** section of the active (or next) stage in section 10. Read the other sections when the work needs them. _(Changed in RULES 1.4.0, approved by the user in S005, decision D-09.)_
4. Read the most recent session log. If it has no closing summary (the session ended abruptly), also read the one before it and reconstruct what was in progress.
5. Read the stage document for the active stage and any ADRs it references.
6. Check the git log and working tree (`git status`) for uncommitted or unrecorded changes. If you find changes not described in the logs, report them to the user before continuing.
7. Create a new session log with the next sequential session number.
8. Give the user a short resume summary: current stage and task, last completed action, next planned action, open questions or blockers. Then proceed with the user's request, or wait for confirmation if the next action is ambiguous.

---

## R2: Continuous recording

**Principle:** Write as you go, never only at the end. A session can be cut off at any moment, and anything not on disk is lost.

**Rules:**

- Log every user message in the current session log. Record user instructions verbatim, redacting only secrets such as passwords or API keys.
- Log every agent response as a concise summary: what was proposed, decided, or done.
- Log every action: files created, modified, or deleted; commands run and their outcome (success, failure, key output); tests run and results.
- Before starting any multi-step action, write it to `CURRENT_STATE.md` under "In progress" (write-ahead). After finishing, move it to "Last completed". This way an interrupted action always leaves a trace.
- Update `CURRENT_STATE.md` whenever the active stage, active task, next steps, blockers, or open questions change.
- Never delete or rewrite past log entries. Corrections are added as new entries that reference the earlier one.
- Use real timestamps from the system clock (e.g. the `date` command) when available. If not, number entries sequentially.

---

## R3: Plan before code

**Golden rule:** No application code is written for a stage until that stage has a detailed stage document that the user has approved.

**Three-level hierarchy: Stage → Substage → Task.**

| Level | ID format | Example | Where it is defined |
|---|---|---|---|
| Stage | `S<NN>` | `S01` | `plan.md`: every stage, for the whole roadmap |
| Substage | `S<NN>.<n>` | `S01.3` | `plan.md`: every substage of every stage, with goal, scope, deliverables, dependencies, requirements, acceptance criteria, risks, and status |
| Task | `S<NN>.<n>-T<NN>` | `S01.3-T02` (task 2 of substage S01.3) | The stage document `stages/S<NN>-<slug>.md`, written just in time, before the stage starts |

**Stage lifecycle:** `Planned -> Approved -> In Progress -> Testing -> Review -> Done`. A stage can also be `Blocked`, with the reason recorded. Substages and tasks use the same status values, plus `Not started` before work begins.

**A stage document must contain:**

- Goal of the stage
- Linked requirement IDs from `plan.md`
- In scope and out of scope
- Design approach, with diagrams or interface sketches where useful
- Task breakdown: for each substage, small tasks with IDs in the form `S01.3-T02`, each with its own acceptance criteria
- Files and modules expected to be created or changed
- Dependencies to add, each with a justification and license
- Test plan: what is tested and how
- Stage acceptance criteria
- Risks and rollback approach
- Approval record: the user's approval quoted verbatim, with date

**Rules:**

- Work on one task at a time. Mark it In Progress in the stage document and `CURRENT_STATE.md` before starting it.
- If implementation reveals the plan is wrong or incomplete, stop. Update the stage document, record the reason, and ask for approval if the change affects scope, architecture, dependencies, data formats, or the sidecar JSON schema. Small internal adjustments can proceed but must still be recorded.
- A stage is Done only when its tests pass, its acceptance criteria are met, and a completion record is written in the stage document: what was built, deviations from plan, known issues, follow-ups.
- Detailed stage documents are written just before a stage starts, not all up front. `plan.md` holds the high-level roadmap.

The template is `code-agent-docs/templates/stage-template.md`.

---

## R4: Plan change management

**Principle:** `plan.md` is a living document. It is never final.

**When the user requests a change:**

1. Record the request verbatim in the session log.
2. Analyze impact: which requirements, stages, ADRs, and already-written code are affected.
3. If the impact is significant or the request is ambiguous, summarize the impact and confirm with the user before editing.
4. Copy the current `plan.md` to `code-agent-docs/archive/plan-history/plan_v<current-version>.md`.
5. Edit `plan.md` and bump its version: PATCH for wording or clarification, MINOR for added or changed requirements or stages, MAJOR for architecture changes or scope overhauls.
   - **Pre-1.0 exception:** while the plan is a pre-1.0 draft (version `0.x.y`), restructurings, including architecture changes and scope overhauls, bump the **MINOR** version (e.g. `0.1.0 → 0.2.0`). The plan becomes **`1.0.0` when the user approves it as the baseline**. After that, the rules above apply unchanged.
6. Add a revision history entry: version, date, summary of changes, reason, and a reference to the session log.
7. Update affected stage documents, ADRs, and `CURRENT_STATE.md`.
8. Report back to the user with a short summary of what changed.

---

## R5: Decision records

**Rules:**

- Every significant technical decision gets an ADR in `code-agent-docs/decisions/`: technology stack, frameworks, major libraries, database or index engine, AI models, sidecar JSON schema changes, API design, security model, deployment approach.
- ADR status is one of: `Proposed`, `Accepted`, `Superseded`, `Rejected`.
- An ADR becomes Accepted only after the user approves it. Record the approval.
- Never rewrite the decision of an Accepted ADR. Supersede it with a new ADR and link both ways.

ADR files are named `ADR-<NNNN>-<short-kebab-title>.md` (e.g. `ADR-0001-backend-language.md`). The template is `code-agent-docs/templates/adr-template.md`.

---

## R6: Engineering standards

**Rules:**

- Small, incremental, reviewable changes. Do not modify code unrelated to the current task.
- **Code first, tests at the end of the stage** (the user's instruction, S006; User Preferences):
  - Write code that is **testable**:
    - dependencies are passed in (interfaces or parameters);
    - time, randomness, the file system, and the network can be replaced in tests;
    - there is no hidden global state;
    - functions are small, with clear inputs and outputs.
  - During a stage, a task delivers code. It is complete when it builds, the linter, the formatter, and the **existing** tests pass, and it was checked by running it (for example a real request against the program), with the results in the session log.
  - The stage's final testing substage writes the stage's **unit, integration, and system/application tests** (the real program, used as a user uses it), plus the regression tests for bugs recorded during the stage. The stage is not Done until they pass and coverage is at least 80%.
- Run the linter, formatter, and test suite before marking a task complete. Record results in the session log.
- No new dependency without a recorded justification and a compatible license.
- Every new dependency must be added to the dependency register (`code-agent-docs/dependencies.md`) in the same commit that introduces it.
- Every change to the sidecar JSON schema bumps `schemaVersion` and includes a migration plan for existing sidecar files.
- Privacy is non-negotiable: no telemetry, no cloud services, no outbound network calls at runtime unless the user explicitly enables a feature that requires one.
- Never write secrets into code, logs, or documentation.
- Keep code readable: consistent structure, meaningful names, comments where the reasoning is not obvious.
- When a bug is found, record it (session log and the stage document's change log) and fix it. Its regression test is written with the stage's tests in the final testing substage, from that record, so none is forgotten. A bug found by those tests already has its failing test.

---

## R7: Git conventions

**Rules:**

- Follow the user's commit preference recorded in the "User Preferences" section of `RULES.md`. If none is recorded, ask before committing.
- Use Conventional Commits (`feat`, `fix`, `docs`, `test`, `refactor`, `chore`, `build`, `ci`) and reference the task ID, e.g. `feat(storage): add chunked upload endpoint [S01-T03]`.
- Commit documentation updates together with the code they describe, or immediately after.
- Commit at the end of each completed task, not in large batches.
- Never force-push, rewrite history, or delete branches without explicit user permission.

---

## R8: Checkpoints and session end

**Rules:**

- Checkpoint at every natural pause: task completed, stage status change, the user signs off, or before any long-running or risky operation.
- A checkpoint means: the session log is current, `CURRENT_STATE.md` is current (especially "Next steps"), all docs are consistent with each other, and changes are committed if the user allows it.
- If your context is getting long or the editor warns about context limits, checkpoint immediately, before the context is compacted or lost.
- When a session ends, append a closing summary to the session log: what was done, decisions made, current state, exact next step.

---

## R9: Archiving and log hygiene

**Rules:**

- Nothing is ever deleted from `code-agent-docs/`. Superseded material is archived.
- When `code-agent-docs/logs/sessions/` holds more than 20 logs, move logs from completed months to `code-agent-docs/archive/sessions/<YYYY-MM>/` and write a `SUMMARY.md` there covering the key work, decisions, and changes of that month. Future agents can read the summary instead of every log.
- Keep `CURRENT_STATE.md` under about 150 lines. It is the quick entry point; detail belongs in stage docs, ADRs, and logs.
- Keep `plan.md` focused on the current plan. Historical versions live in `archive/plan-history/`.

---

## R10: Rule precedence and maintenance

**Precedence (highest first):**

1. The user's latest explicit instruction in chat
2. `RULES.md`
3. `plan.md`
4. Stage documents and ADRs
5. Everything else

**Rules:**

- If a user instruction conflicts with `RULES.md`, follow the user for that instance and ask whether `RULES.md` should be updated permanently.
- Whenever the user states a lasting preference (coding style, commit behavior, communication style, tools to use or avoid), add it to the "User Preferences" section of `RULES.md` without being asked, and log that you did so.
- Any other change to `RULES.md` requires user approval and must be recorded in the `RULES.md` changelog.
- Re-read `RULES.md` at the start of every session and before starting every new stage.

---

## R11: Communication

**Rules:**

- Keep chat replies concise. Put detail in the documentation and point the user to it.
- When presenting anything that needs a decision, clearly separate what you did from what needs the user's approval.
- List assumptions explicitly. Ask rather than guess when a choice affects architecture, data formats, security, or scope.
- Never claim something works without having run it. Report test and command results honestly, including failures.

---

## R12: Documentation audits

**Rule (added in RULES.md 1.3.0, pre-approved in P004):**

- Run a documentation audit using `code-agent-docs/templates/audit-checklist.md` as part of every stage's final review substage, and whenever the user asks. Audits are numbered sequentially (A001, A002, ...) and reported in `code-agent-docs/audits/`. Critical findings must be fixed or escalated to the user before the stage can be marked Done.

## R13: Releases

**Rule (added in RULES.md 1.8.0, pre-approved in P006):** after the MVP, features are developed and shipped as numbered releases, a set of features at a time (plan section 11b; the same rules are plan 11b.1).

**Rules:**

- Release IDs R01, R02, … are permanent, like FR IDs. A release has: a theme, a goal, a fixed feature set (gap IDs and FR IDs), prerequisites, exit criteria, and a suggested version label.
- The MVP is milestone M3 (S01–S14 plus mvp_additions) and ships as the first stable release. Suggested label v1.0.0; the user decides (S14.7).
- Feature releases bump the MINOR version (v1.1.0, v1.2.0, …). Fix-only releases bump PATCH and can ship at any time between feature releases. A release that breaks the data format or the upgrade path bumps MAJOR and needs the user's approval.
- One feature release is in progress at a time. Security fixes take priority over all feature work.
- Just-in-time conversion: when a release is next, its features become one stage (or several), with substages written into plan.md, inserted before the AI stage as I8 requires, and the AI stage is renumbered (record it in the 10.18 table). Then the stage document is written and approved before any code (R3).
- Feature freeze: once a release's stage document is approved, adding a feature to it needs the user's approval. Otherwise the feature goes to a later release.
- Every release ends with its stage's testing and review substage, plus: an upgrade test from the previous release with real migrated data, a security review of every new surface (threat model updated), release notes and a changelog, updated user and admin guides, the R12 documentation audit, a tagged release, and the user's sign-off.
- Every release keeps the system upgradeable from the previous release (NFR-017) and keeps the NAS fully working with AI disabled (I7).
- Features that use the network (imports from other clouds, off-site backup, ACME certificates, DDNS, email, push, tunnels) are off by default and switched on explicitly by the user (I6).
- Reordering releases, splitting them, or moving a feature between releases needs the user's approval and a plan revision (R4).
- Features in a release that depend on an AI result (e.g. smart albums by person) work without AI, and gain the AI filter when S17 lands.

---

## R14: Change intake (new in 1.9.0, the user's decision D6)

After plan change request #8 the **MVP scope is frozen**. When a new request arrives:

- Record it in the backlog table of plan section 11d (date, source, summary, the case below, where it is placed).
- Plan it into a release after the MVP, **unless** it is a security or data-integrity fix, a foundation that is cheap only now and expensive later, or the user explicitly puts it in the MVP.
- Tell the user which case applies, and ask when the placement is not obvious.

The user's answer (S007 E059): "Yes, freeze MVP (Recommended)".

---

## Documentation map

All agent documentation lives in `code-agent-docs/`:

| Path | Purpose |
|---|---|
| `code-agent-docs/RULES.md` | Permanent operating rules for the agent (this file). Read at the start of every session. |
| `code-agent-docs/CURRENT_STATE.md` | Short, always-current snapshot: where the project stands and what to do next. The first thing a resuming agent relies on. Keep under ~150 lines. |
| `code-agent-docs/plan.md` | The master development plan. Living document, versioned (R4). |
| `code-agent-docs/dependencies.md` | The dependency register: every dependency, external tool, dataset, and AI model, with name, version, license, purpose, stage, ADR link, and verification status. Updated in the same commit that introduces a dependency (R6). |
| `code-agent-docs/stages/` | One detailed plan document per development stage, named `S<NN>-<slug>.md`, e.g. `S01-basic-nas.md`. Created only when a stage is about to be planned in detail. |
| `code-agent-docs/decisions/` | Architecture Decision Records, e.g. `ADR-0001-backend-language.md`. |
| `code-agent-docs/logs/sessions/` | One log per working session: `<YYYY-MM-DD>_S<NNN>.md` with a sequential session number (S001, S002, ...). **Session IDs have three digits (S004); stage IDs have two (S04).** They are different things. |
| `code-agent-docs/archive/plan-history/` | Every superseded version of `plan.md`, e.g. `plan_v0.1.0.md`. These are **verbatim** copies, so their relative links (e.g. `decisions/…`) are relative to `code-agent-docs/` and do not resolve from inside the archive folder. |
| `code-agent-docs/archive/sessions/` | Older session logs moved here, in `<YYYY-MM>/` folders with a monthly `SUMMARY.md` (R9). |
| `code-agent-docs/templates/` | `stage-template.md`, `session-log-template.md`, `adr-template.md`, `audit-checklist.md` (the reusable documentation-audit checklist for R12; update it whenever new document types or rules are added). |
| `code-agent-docs/bootstrap/` | The original bootstrap prompt, archived verbatim (`initial-prompt.json`). It stays there. It is effectively prompt P001. |
| `code-agent-docs/prompts/` | Later user prompts (change requests, instructions delivered as files), archived verbatim as `P<NNN>-<short-kebab-title>.<ext>`, e.g. `P002-staged-development-roadmap.json`. The session log records the USER entry as a pointer to the archived file. The user's own originals live in the repository-root `prompts/` folder (user-managed, outside `code-agent-docs/`). |
| `code-agent-docs/research/` | Research records for planning, numbered sequentially: `R<NNN>-<YYYY-MM-DD>-<slug>.md` (e.g. `R001-2026-09-28-cloud-storage-feature-research.md`, the competitor research behind the release roadmap). Dated snapshots: the plan links to them instead of copying them. Added in 1.8.0 (P006). |
| `code-agent-docs/security/` | Security documents, starting with the threat model `threat-model.md` (S03.1, FR-084): assets, attackers, surfaces, numbered threats (T-01…), and the status of each (mitigated with its test, accepted risk with the user's approval, or open with the stage that handles it). Every stage that adds an attack surface updates it. Added in 1.8.2 (S03 approval, decision D-4). |
| `code-agent-docs/audits/` | Documentation audit reports, numbered sequentially: `A<NNN>-<YYYY-MM-DD>-<slug>.md` (e.g. `A001-2026-09-24-documentation-audit.md`), plus audit side documents such as README change proposals (`A<NNN>-readme-proposal.md`). See R12. |

Folders that were created empty got a `.gitkeep` file so git tracks them. These files stay in place even after a folder gains content (R9: nothing is deleted from `code-agent-docs/`).

---

## Bootstrap pointer files

These short files are auto-loaded by agents and editors. They point a fresh agent instance, with no chat history, to this documentation system. They are pointers only and must stay short (under 30 lines). They must not be copies of these rules.

| File | Loaded by | Status |
|---|---|---|
| `AGENTS.md` (repository root) | Generic convention read by many agentic editors | Exists (created S001) |
| `CLAUDE.md` (repository root) | Claude Code (auto-loaded at session start) | Exists (created S001) |

If a different editor is used later (e.g. Cursor: `.cursor/rules/`, Windsurf: `.windsurf/rules/`, GitHub Copilot: `.github/copilot-instructions.md`), add its native pointer file with the same content and record it in this table. Keep all pointer files consistent with each other.

---

## User Preferences

Filled in as the user states lasting preferences (R10). Commit behavior is recorded first.

- **Commit behavior** (S001, 2026-09-23): the agent **may commit on its own**, at the end of each completed task (R7), using Conventional Commits with the task ID. User's answer: "Yes, after each task (Recommended)".
- **Branching and pushing** (S001, 2026-09-23): use **feature branches + pull requests**. Create one branch per stage/task off `develop` (e.g. `feat/S01-T03-upload`; documentation-only work: `docs/<id>-<topic>`), push it to `origin`, and open a pull request into `develop`. Never commit directly to `main` or `develop`. Never merge PRs, force-push, rewrite history, or delete branches without explicit permission. User's answer: "Feature branches + PRs" (option text: "One branch per stage/task (e.g. feat/S01-T03-upload) off develop, pushed, with a pull request into develop."). _Partly superseded in S005: the agent now merges its own finished branches into `develop` (next bullet)._
- **Merging finished branches into develop** (S005, 2026-09-24): when the agent is done with a branch (the task or documentation change is complete, checks pass, documents are updated, and the branch is committed and pushed; since S007 the checks are the local ones, see the CI bullet), the agent **merges that branch into `develop` itself** (`git merge --no-ff <branch>`, then push `develop`). Opening a PR is optional. Still in force: no direct commits onto `develop` other than these merge commits; **never merge into `main`** unless the user asks (e.g. for a release); never force-push, rewrite history, or delete branches without explicit permission. User's words: "when you are done with a branch, you yourself should do a merge of that branch into develop (another rule you should remember)".
- **Dependency record and per-platform setup scripts** (S005, 2026-09-24): keep a record of **every dependency needed** to build and deploy the NAS. Besides the R6 register entry, anything the running NAS needs on the target machine gets a per-platform row in `dependencies.md` section 12 (Debian or Ubuntu, Arch Linux, Raspberry Pi, Windows 11, Docker image; Arch added in S007) in the same commit (NFR-032). At the end (S14.2), a **separate setup script for each platform** uses that record to deploy the NAS automatically (FR-149). Each stage's documentation audit (R12) checks that section 12 is complete. User's words: "one thing to add: keep record of all dependencies needed, in the end you will have to make a setup script, a separate one for each platform, which when run, will automatically handle the deployment."
- **Stacked branches** (S005, 2026-09-24, decision D-03 of audit A001): if an earlier branch is not merged yet, a new branch may be created on top of it (and the user is told). With the merge rule above this should be rare. User's answer: "Accept all (Recommended)".
- **Testing approach** (S006, 2026-09-25): the user's instruction: "Make sure to add unit tests and integration tests and system/ application tests a part of development process but focus on coding first, write code that is testable, but write tests in the end (of the stage)". Code comes first and is written to be testable. Each stage's final testing substage writes its unit, integration, and system/application tests (R6). Coverage is reported (since S007: locally, and by CI at the stage end, see the next bullet) but does not block during a stage; 80% is an exit criterion of the final testing substage ("Report now, enforce at stage end (Recommended)"). A bug is recorded and fixed at once, and its regression test is written with the stage's tests ("Fix now, test at stage end (Recommended)").
- **CI only at stage completion** (S007, 2026-09-28): the user's instruction: "dont run CI on every commit/push/checkpoint but CI should only run at the completion of a stage". During a stage, the agent runs the checks locally before each commit and merge, and CI is not triggered by commits, pushes, or checkpoints. CI runs once when a stage is complete, before the user's sign-off. **Mechanism** (the user's answer in S007 E013, "Stage tag + manual (Recommended)"): `.github/workflows/ci.yml` runs only when a stage-completion tag `S<NN>-done` is pushed (after the stage's final testing substage, before sign-off) or when "Run workflow" is used. Branch pushes, merges, and pull requests (including Dependabot's) start nothing, so `[skip ci]` is no longer needed. The tag is created by the agent on `develop` at the stage end.
- **AI co-author line kept in commits** (S007, 2026-09-29): the agent's commit messages end with its `Co-Authored-By` line, so GitHub shows the agent as co-author, **for transparency**. The user first asked to remove it ("i dont want your name here"), then decided to keep it: "hey you know what, its fine, keep it there, for transparency purposes". Nothing in the history is rewritten.
- **Keep working; stop only when told or at a stage's end** (S007, 2026-09-29): the agent does not stop at checkpoints. It checkpoints as R8 says (log, CURRENT_STATE, consistent docs, commits) and carries on with the next step. It stops only when the user says so, or when the current stage is complete (its sign-off asked for). Stopping always means: save all state (log, CURRENT_STATE), commit, and push. Questions that only the user can answer are still asked (R10, R11) and, where possible, work continues on what they do not block. The user's words: "keep working, dont stop at checkpoint. stop only when i say or current stage completes and stopping rule is saving, committing and pushing the changes."
- **Raspberry Pi first, and a deployer per platform** (S007, 2026-09-29): the NAS must be "very optimized on a raspberry pi", the user's production machine (they will get one when the project is complete), so every stage designs and measures for it (plan NFR-051, A24). It must be deployable on **Debian, Arch Linux, Windows, and Raspberry Pi** (and more later), each with its own **very user-friendly** deployer (plan FR-149, NFR-052). The user's words: "also, remember, this project should be very optimized on a raspberry pi, i will run it on a pi"; "though i dont have a pi right now but when project completes, i will get a pi and deploy there"; "the project should be dynamic though, being able to be deployable on debian, arch, windows, pi and more (but these 4 mentioned must be focused) etc. that is there must be a different deployment script (or whatever deployer you make, but that should be very user friendly and not too complicated) for each."

---

## RULES.md changelog

| Version | Date | Change | User approval reference |
|---|---|---|---|
| 1.0.0 | 2026-09-23 | Initial version. R1-R11 transcribed in full from the bootstrap prompt, plus documentation map, pointer files, User Preferences, and changelog sections. | Content mandated by the user's bootstrap prompt (`code-agent-docs/bootstrap/initial-prompt.json`); session `logs/sessions/2026-09-23_S001.md` |
| 1.0.1 | 2026-09-23 | User Preferences: commit behavior and branching/PR workflow recorded. | User's answers in S001 (preference recorded per R10; see S001 log E013) |
| 1.1.0 | 2026-09-23 | (a) Added folder `code-agent-docs/prompts/` for archived user prompts, and added it to the documentation map. The bootstrap prompt stays in `bootstrap/`. | Pre-approved in `code-agent-docs/prompts/P002-staged-development-roadmap.json` (`pre_approved_documentation_changes`); session S002 |
| 1.1.0 | 2026-09-23 | (b) R3 updated to the three-level hierarchy Stage → Substage → Task, with task IDs in the form `S01.3-T02`. | Pre-approved in `code-agent-docs/prompts/P002-staged-development-roadmap.json`; session S002 |
| 1.1.0 | 2026-09-23 | (c) R4 versioning: while the plan is a pre-1.0 draft, restructurings bump MINOR. The plan becomes 1.0.0 when the user approves it as the baseline, and the original R4 rules apply after that. | Pre-approved in `code-agent-docs/prompts/P002-staged-development-roadmap.json`; session S002 |
| 1.1.0 | 2026-09-23 | (d) Added the "Project invariants" section (I1-I8). | Pre-approved in `code-agent-docs/prompts/P002-staged-development-roadmap.json`; session S002 |
| 1.1.0 | 2026-09-23 | (e) `templates/stage-template.md` updated with a substages section and a task table per substage. | Pre-approved in `code-agent-docs/prompts/P002-staged-development-roadmap.json`; session S002 |
| 1.2.0 | 2026-09-24 | (a) Added invariant **I9** (only the core server writes sidecars and the search index) to "Project invariants". | Pre-approved in `code-agent-docs/prompts/P003-technology-stack.json` (`pre_approved_documentation_changes`); session S003 |
| 1.2.0 | 2026-09-24 | (b) Created `code-agent-docs/dependencies.md` (the dependency register) and added it to the documentation map. | Pre-approved in `code-agent-docs/prompts/P003-technology-stack.json`; session S003 |
| 1.2.0 | 2026-09-24 | (c) R6: every new dependency must be added to the dependency register in the same commit that introduces it. | Pre-approved in `code-agent-docs/prompts/P003-technology-stack.json`; session S003 |
| 1.3.0 | 2026-09-24 | (a) Added the folder `code-agent-docs/audits/` to the documentation map. | Pre-approved in `code-agent-docs/prompts/P004-documentation-audit.json` (`pre_approved_changes`); session S004, audit A001 |
| 1.3.0 | 2026-09-24 | (b) Created the reusable audit checklist `code-agent-docs/templates/audit-checklist.md` (check groups A–L, severity levels, fix policy, report structure) and listed it under `templates/` in the documentation map. | Pre-approved in `code-agent-docs/prompts/P004-documentation-audit.json`; session S004 |
| 1.3.0 | 2026-09-24 | (c) Added rule **R12 (Documentation audits)** with the wording given in P004. | Pre-approved in `code-agent-docs/prompts/P004-documentation-audit.json`; session S004 |
| 1.3.0 | 2026-09-24 | (d) The stage template's final testing and review substage now includes the audit step. | Pre-approved in `code-agent-docs/prompts/P004-documentation-audit.json`; session S004 |
| 1.3.0 | 2026-09-24 | (e) Documentation-map wording clarifications from audit A001, with no change in meaning or rules: `.gitkeep` files are kept (F-002); root `prompts/` folder described (F-003); session vs stage ID note (F-006); stage-document example updated from the obsolete `S00-foundation.md` to `S01-basic-nas.md` (F-033); note that archived plans are verbatim, so their relative links do not resolve from the archive folder (F-034). | P004 `fix_policy.fix_directly` ("unclear wording that does not change meaning"); audit A001, session S004 |
| 1.4.0 | 2026-09-24 | (a) R1 step 3 (and the Quick start): plan.md is read selectively (sections 2a, 5, 7, 10.1, plus the active or next stage in full; others on demand) instead of completely. | Approved by the user in S005 ("Accept all (Recommended)", decision D-09 of audit A001, finding F-007); session log S005 E007 |
| 1.4.0 | 2026-09-24 | (b) User Preferences: the agent merges its finished branches into `develop` itself; stacked branches allowed while an earlier branch is unmerged (D-03). The S001 branching bullet is marked partly superseded. | The user's lasting preference (S005 E008, recorded per R10) and D-03 approval (S005 E007) |
| 1.5.0 | 2026-09-24 | User Preferences: keep a record of every dependency, with per-platform runtime prerequisites in `dependencies.md` section 12; per-platform setup scripts at the end (S11.2). | The user's lasting instruction (S005 E015, recorded per R10) |
| 1.6.0 | 2026-09-25 | R6: code first, written to be testable. Each stage's final testing substage writes its unit, integration, and system/application tests, and the regression tests for bugs recorded during the stage. Coverage of 80% is enforced at the stage end, not on every push. User Preferences: the testing approach. | The user's instruction (S006 E001) and answers (S006 E003), recorded per R10 |
| 1.7.0 | 2026-09-28 | (a) Added invariant **I10** (destructive bulk operations: preview, explicit confirmation, undo window where possible) to "Project invariants". | Pre-approved in `code-agent-docs/prompts/P005-feature-additions.json` (`new_invariant`, `pre_approved_documentation_changes`); session S007 |
| 1.7.0 | 2026-09-28 | (b) User Preferences: CI runs only at the completion of a stage; `[skip ci]` until the workflow is changed. The testing-approach and merge bullets are adjusted to match. Stage IDs in the dependency-record bullet follow the 1.4.0 plan (S13.2). | The user's instruction (S007 E002, recorded per R10); plan 1.4.0 renumbering (S007) |
| 1.7.0 | 2026-09-28 | (c) plan.md: the "Stage ID changes" table (section 10.18) added, and I10 added to the plan's "Project invariants" (2a). | Pre-approved in `code-agent-docs/prompts/P005-feature-additions.json` (`pre_approved_documentation_changes`); session S007 |
| 1.7.0 | 2026-09-28 | (d) `templates/audit-checklist.md` group K: historical documents are checked against the "Stage ID changes" table instead of their old IDs being flagged as errors. | Pre-approved in `code-agent-docs/prompts/P005-feature-additions.json` (`pre_approved_documentation_changes`); session S007 |
| 1.7.1 | 2026-09-28 | User Preferences: the CI trigger mechanism (a stage-completion tag `S<NN>-done`, or a manual run). | The user's answer in S007 (E013): "Stage tag + manual (Recommended)", recorded per R10 |
| 1.7.2 | 2026-09-29 | User Preferences: the agent's co-author line stays in commit messages, for transparency. | The user's decision (S007 E038, E039), recorded per R10 |
| 1.8.0 | 2026-09-29 | (a) Added rule **R13 (Releases)** with the content of P006 `release_process_rules`. | Pre-approved in `code-agent-docs/prompts/P006-competitor-research-release-roadmap.json` (`pre_approved_documentation_changes`); session S007 |
| 1.8.0 | 2026-09-29 | (b) Added the folder `code-agent-docs/research/` to the documentation map, with research record R001. | Pre-approved in P006; session S007 |
| 1.8.0 | 2026-09-29 | (c) `templates/audit-checklist.md` group D: every gap in a research file has exactly one destination; every release feature has an FR ID; every FR of a release names its release. | Pre-approved in P006; session S007 |
| 1.8.0 | 2026-09-29 | (d) User Preferences: keep working without stopping at checkpoints; stop only when the user says so or a stage completes, and stopping means save, commit, and push. | The user's lasting instruction (S007 E041), recorded per R10 |
| 1.8.1 | 2026-09-29 | User Preferences: Raspberry Pi first, and a very user-friendly deployer for each of Debian, Arch Linux, Windows, and Raspberry Pi; Arch added to the section 12 platforms in the dependency-record bullet. Audit checklist: checks for the admin console, the Raspberry Pi budgets, the four platform columns, and the reused ID S15. | The user's lasting instructions (S007 E046), recorded per R10 |
| 1.8.2 | 2026-09-29 | Documentation map: new folder `code-agent-docs/security/` for the threat model. Audit checklist: the threat model check (group H). | The user's approval of S03 with decision D-4 ("approve s03", S007 E050) |
| 1.9.0 | 2026-09-30 | Invariants I3 (the database is authoritative for ownership, access, and sharing; sidecars mirror them) and I4 (the typed search text may be embedded at query time) reworded as in plan 1.9.0; new rule **R14 Change intake** (the MVP scope is frozen after P008). The process itself is unchanged (D4). | The user's decisions D2, D3, D4, D6 on plan change request #8 (S007 E059) |
