# ADR-0041: Operation journal and crash consistency

| Field | Value |
|---|---|
| Number | ADR-0041 |
| Status | **Accepted** (2026-09-30, session S007) |
| Date proposed | 2026-09-30 (session S007) |
| Date of last status change | 2026-09-30 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

A logical operation (move, copy, delete to trash, restore, purge, upload finalize, later ingest and imports) can touch the filesystem, a sidecar, the database, and the search index. They cannot share one transaction. A crash, a power cut, or a full disk between two steps must never lose user data or leave a visible, inconsistent item (FR-352, NFR-053, P008 F3).

## Options considered

| Option | Pros | Cons |
|---|---|---|
| **(a) Intent-first journal in SQLite with idempotent steps and startup recovery** | One mechanism for every operation; testable step by step; uses the existing database with FULL durability (ADR-0007 amendment) | Every operation writes its intent first |
| (b) Best effort plus the reconciler | Simple | Inconsistent states are visible until the next scan; data can be lost in edge cases |
| (c) Filesystem-only markers (temp files) | No database write | Hard to query; spreads state across folders |

## Decision

**(a).**
- **The `operations` table:** ID, type, state (`pending`, `running`, `done`, `failed`, `rolled_back`), intent (JSON: item IDs, source and target paths, expected content hashes), current step, timestamps, error.
- **Rule:** the intent is committed before any side effect; each step is idempotent (safe to run twice); the operation is marked `done` at the end; a job (ADR-0011 amendment) may run one operation.
- **Recovery at startup:** every unfinished operation is rolled forward where its remaining steps can complete, and rolled back otherwise; temporary files (`.local-ai-nas-tmp-*`) are cleaned.
- **Failure matrix** (completed in S01.4-T09):

| Failure | Outcome |
|---|---|
| The file write succeeds, the rename fails | The temporary file is removed at recovery; the operation is marked failed; the original is untouched |
| The rename succeeds, the database commit fails | Recovery sees the file at the target and completes the database step (roll forward) |
| The disk fills mid-write | The free-space guard refuses early (S01.2); a partial temporary file is removed; nothing visible changes |
| The application crashes at any step | Recovery at the next start rolls forward or back |
| Power is lost | As a crash; with `synchronous=FULL` every committed step survives |

- **Rule of thumb:** never lose user data; never leave a visible, inconsistent item; an orphaned file is adopted by the reconciler, never deleted.
- **Crash injection:** named crash points in every step, driven by the S03.10 test harness (NFR-053); the S05.8 sidecar crash tests join the same harness.

## Consequences

- **Easier:** crash behavior is defined and tested instead of hoped for.
- **Harder:** every multi-place operation is written as steps; one extra database write per operation.
- **Required:** S01.4-T09 (P008 follow-up, built in S03); the S03.10 harness.

## Approval record

> "Accept all three (Recommended)"
> (2026-09-30, session S007; the user's answer to: "Accept the three ADRs the approved follow-ups are built on?", naming ADR-0040, ADR-0041, and ADR-0042 with a one-line summary of each)

Proposed with plan 1.9.0 (P008), whose follow-up tasks the user had approved ("Approve all, F2 first (Recommended)", S007 E059).

## Links

plan FR-352, NFR-053, 8.38 · [ADR-0011](ADR-0011-job-queue-sqlite.md) · [ADR-0007](ADR-0007-database-sqlite.md) · [ADR-0040](ADR-0040-item-identity.md) · P008 F3
