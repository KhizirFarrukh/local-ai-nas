# R002: External review #1, disposition

| Field | Value |
|---|---|
| Source | `code-reviews/1-local-ai-nas-improvement-review.md` (the user's file), archived verbatim as `prompts/CR001-improvement-review-1.md` |
| Date | 2026-09-30 (session S007, log E056–E057) |
| The user's instructions | "go through code-reviews/1-local-ai-nas-improvement-review.md … there are things that you need to see, understand and record accordingly"; "if something is already covered then skip it" |
| Result | Plan 1.8.1 (questions Q77–Q81, concerns 8.35–8.37, a search note in 8.11, risk RK-50); threat model 1.2 (T-58, T-59); S03 execution order adjusted; README proposal R-17 |

## Context

The review was written against an earlier state of the project: it speaks of changes "before S01 begins", a twelve-stage roadmap, and a README that shows "Planning / pre-alpha". On the review date S01 and S02 are Done and S03 is In Progress (plan 1.8.0, 16 stages). Every point was checked against the current plan, rules, ADRs, and code. Points already covered are **skipped**, as the user asked; they are listed here only with where they are covered, so the check can be followed.

## Disposition

| # | Review point | Covered? | Where it is covered, or where it is recorded |
|---|---|---|---|
| 1 | Immutable `ItemID` / `MediaID`; references by ID, not path | **No.** The plan names a "stable file ID" for shortcuts (ADR-0024, S11.6) and re-association by file ID or content hash (8.6), but defines no identity model. Filesystem file IDs change when data is copied, restored, or moved to another drive, which the migration engine does (ADR-0030) | **Recorded:** plan 8.35, **Q77**, RK-50 |
| 2 | Namespaced sidecar names (`.lainas.json`) | Yes | Q11 (sidecar naming vs. other tools, needed by S05.1), the identifying marker (8.6), FR-030. Skipped; the review's naming is one possible answer to Q11 |
| 3 | Ownership and ACL authority in SQLite, not in sidecars | **No.** Plan 8.8 keeps photo access data in the sidecar `access` section (the user's specification) and recommends hidden access sidecars for files; nothing makes the database the authority | **Recorded:** plan 8.36, **Q78**, threat T-58 |
| 4 | A minimal durable job foundation early; crash behavior of operations | Yes | S04.3 job system (ADR-0011: restart survival, crash simulation, retries); principle 4 (the S01.4 and S03.6 schedulers move onto it); atomic writes (NFR-006); the reconciler (FR-102). "In S01" no longer applies. Skipped |
| 5 | Trash earlier | Yes, by the user's decision | S007: "if trash is set to be implemented in later stage then follow that plan" (Trash stays in S08). Skipped |
| 6 | Host and Origin validation, no CORS, before full authentication | Yes | S03.5-T02 (the Host allow-list added by the threat model, T-19; bug S03-B01), CSRF and Origin checks, no CORS. **Scheduling adjusted:** the Host allow-list and the Origin check move to the front of the remaining S03 work, because bug S03-B01 is open now (stage document, execution order) |
| 7 | Semantic search needs a query-time text model (I4 contradiction) | Yes | Plan 7.1 and ADR-0018 / FR-054: a query-time text model is an explicit exception to I4 that needs the user's approval; default precomputed forms only (S005 D-06); Q36. Skipped |
| 8 | Less process ceremony; no verbatim conversation in a public repository | **No** | **Recorded:** **Q80** (RULES R1 and R2 are the user's to change) |
| 9 | Split `plan.md`; a short plan index as hot context | **No** | **Recorded:** **Q81** |
| 10 | Keep the stack; no microservices | Yes | Plan section 7 and ADR-0001–ADR-0020: one Go program with internal services, a separate Python AI worker (ADR-0017). Skipped |
| S1 | Per-field search analyzers (no stemming on names and filenames) | Partly: an unstemmed sub-field and keyword fields exist (8.11, S06.1), but one English analyzer with Porter stemming is applied to all text | **Recorded:** a note in plan 8.11 for the S06 stage document (a design detail, not a question) |
| S2 | Unicode and multilingual metadata (Urdu, Roman Urdu, Arabic script) | Yes | Q15 (languages for search), 11c (internationalization with right-to-left, Urdu recommended); the Arabic-script normalization detail goes with the 8.11 note |
| D | SQLite `synchronous=FULL` for users, sessions, shares, audit | **No.** ADR-0007 sets `synchronous=NORMAL` for the whole database; with WAL a power cut can roll back the last committed transactions, and S03 now puts sessions, revocations, password changes, and the audit trail there | **Recorded:** plan 8.37, **Q79**, threat T-59 |
| R1 | README status section | Yes | README proposal R-13 (A003) |
| R2 | README architecture diagram | **No** | **Recorded:** README proposal **R-17** |
| R3 | Choose a license before outside contributions | Yes | AGPL-3.0-or-later (Q22, S005; README, LICENSE) |
| RM | Roadmap adjustments (IDs in S01, Trash in S03, direct video playback, namespaced sidecars in S05, first usable release at S06) | Yes | The current 16-stage plan and the user's decisions: S01 and S02 Done; Trash in S08 (user); direct play first (ADR-0020); sidecar naming Q11; the first usable release is M3 = S13 (the user, S005; Q51). Only the ID point is new (row 1) |
| F | "Freeze the architecture and begin implementation" | Yes | Implementation is under way (S01, S02 Done; S03 In Progress) |
