# RULES: Permanent Operating Rules for AI Agents on local-ai-nas

**RULES.md version:** 1.1.0
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
3. Read `code-agent-docs/plan.md`.
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
- **I3:** Photo sidecar JSON files are the source of truth for photo metadata. The search index is a cache that can always be rebuilt from disk.
- **I4:** Search reads only the index. It never scans files or runs AI at query time.
- **I5:** A user cannot access another user's files or photos unless they have been explicitly shared. This is enforced server-side on every access path: API, downloads, previews, thumbnails, search, network shares, and background jobs.
- **I6:** Everything runs locally. No telemetry. No network calls at runtime except for features the user has explicitly enabled.
- **I7:** AI is optional and opt-in. The NAS must be fully functional with AI disabled.
- **I8:** AI work is always the last stage of the roadmap. Any stage added in the future is inserted before it, and the AI stage is renumbered.

---

## R1: Session start protocol

**Applies:** At the start of every session and every new chat, without exception, even if the user's first message asks for something else.

**Steps:**

1. Read `code-agent-docs/RULES.md` completely.
2. Read `code-agent-docs/CURRENT_STATE.md`.
3. Read `code-agent-docs/plan.md`.
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
- Write tests alongside or before the code (test-first where practical). No task is complete without its tests.
- Run the linter, formatter, and test suite before marking a task complete. Record results in the session log.
- No new dependency without a recorded justification and a compatible license.
- Every change to the sidecar JSON schema bumps `schemaVersion` and includes a migration plan for existing sidecar files.
- Privacy is non-negotiable: no telemetry, no cloud services, no outbound network calls at runtime unless the user explicitly enables a feature that requires one.
- Never write secrets into code, logs, or documentation.
- Keep code readable: consistent structure, meaningful names, comments where the reasoning is not obvious.
- When a bug is found, record it, write a failing test that reproduces it, then fix it.

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

## Documentation map

All agent documentation lives in `code-agent-docs/`:

| Path | Purpose |
|---|---|
| `code-agent-docs/RULES.md` | Permanent operating rules for the agent (this file). Read at the start of every session. |
| `code-agent-docs/CURRENT_STATE.md` | Short, always-current snapshot: where the project stands and what to do next. The first thing a resuming agent relies on. Keep under ~150 lines. |
| `code-agent-docs/plan.md` | The master development plan. Living document, versioned (R4). |
| `code-agent-docs/stages/` | One detailed plan document per development stage, e.g. `S00-foundation.md`. Created only when a stage is about to be planned in detail. |
| `code-agent-docs/decisions/` | Architecture Decision Records, e.g. `ADR-0001-backend-language.md`. |
| `code-agent-docs/logs/sessions/` | One log per working session: `<YYYY-MM-DD>_S<NNN>.md` with a sequential session number (S001, S002, ...). |
| `code-agent-docs/archive/plan-history/` | Every superseded version of `plan.md`, e.g. `plan_v0.1.0.md`. |
| `code-agent-docs/archive/sessions/` | Older session logs moved here, in `<YYYY-MM>/` folders with a monthly `SUMMARY.md` (R9). |
| `code-agent-docs/templates/` | `stage-template.md`, `session-log-template.md`, `adr-template.md`. |
| `code-agent-docs/bootstrap/` | The original bootstrap prompt, archived verbatim (`initial-prompt.json`). It stays there. It is effectively prompt P001. |
| `code-agent-docs/prompts/` | Later user prompts (change requests, instructions delivered as files), archived verbatim as `P<NNN>-<short-kebab-title>.<ext>`, e.g. `P002-staged-development-roadmap.json`. The session log records the USER entry as a pointer to the archived file. |

Empty folders contain a `.gitkeep` file so git tracks them.

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
- **Branching and pushing** (S001, 2026-09-23): use **feature branches + pull requests**. Create one branch per stage/task off `develop` (e.g. `feat/S01-T03-upload`; documentation-only work: `docs/<id>-<topic>`), push it to `origin`, and open a pull request into `develop`. Never commit directly to `main` or `develop`. Never merge PRs, force-push, rewrite history, or delete branches without explicit permission. User's answer: "Feature branches + PRs" (option text: "One branch per stage/task (e.g. feat/S01-T03-upload) off develop, pushed, with a pull request into develop.").

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
