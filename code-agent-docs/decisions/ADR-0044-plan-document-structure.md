# ADR-0044: Plan document structure: a plan folder with an index

| Field | Value |
|---|---|
| Number | ADR-0044 |
| Status | **Accepted** (the user's request in P008, S007) |
| Date proposed | 2026-09-30 (session S007) |
| Date of last status change | 2026-09-30 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

`plan.md` has grown to about 4,000 lines (over 300 KB). External review #1 (CR001) and plan change request #8 (P008 P-A, adopted by the user's request) ask to split it into small documents with a short index.

## Decision

- **Layout:** `code-agent-docs/plan/` with `PLAN_INDEX.md` (version, stage status, milestone, active stage, next actions, links to every file, and a table mapping every old `plan.md` section number to its new file, so references such as "plan 8.8" still resolve) and one file per group: `overview.md` (sections 1, 2, 2b), `invariants.md` (2a), `requirements.md` (3), `assumptions.md` (4), `questions.md` (5), `architecture.md` (6 and 8), `tech-stack.md` (7), `roadmap.md` (9, 10.1, 10.17, 10.18, 11, 11a–11d), `stages/SNN-<slug>.md` (one per stage, from 10.2–10.16), `testing.md` (12), `risks.md` (13), `changelog.md` (14). `plan.md` becomes a short stub pointing to the index.
- **The move is mechanical:** nothing is reworded; a check verifies that every non-heading line of the old plan appears in exactly one new file and that every ID is defined exactly once.
- **One version** for the whole plan set, kept in `PLAN_INDEX.md`. R4 archiving copies the **whole plan folder** per version (the user kept the current process, D4).
- **Reading (R1):** unchanged in scope: the session reads the whole plan, now through `PLAN_INDEX.md` and the files it lists (the user's decision D4 keeps the process; P008's leaner startup is not applied).

## Consequences

- **Easier:** a stage's specification can be opened on its own; smaller diffs.
- **Harder:** cross-references point into several files; the index's section map keeps old references working.

## Approval record

> "I got these suggestions, go through them, improve if possible, and then make a detailed json prompt for it so the ai coding agent can go through them and add them into the plan so these can also be implemented along the way"
> (the user's request behind P008, `prompts/P008-external-architecture-review.json`; P-A marked "Adopted by the user's request"; applied in S007 E058–E061)

## Links

RULES R1, R4, documentation map · plan 14 (revision 1.9.0) · P008 P-A
