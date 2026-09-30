# ADR-0011: Background job queue on SQLite, inside the core server

| Field | Value |
|---|---|
| Number | ADR-0011 |
| Status | **Accepted** |
| Date proposed | 2026-09-24 (session S003) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none |
| Superseded by | none |

## Context

S04.3 builds the shared background job system (FR-095, NFR-012). S05, S06, S08, S09, and S17 reuse it. Required:
- Survives restarts.
- Retries with backoff.
- Priorities.
- Per-job-type concurrency limits.
- Progress reporting.
- A per-user job context (the hook for I5 checks in S07.4).
- The optional AI worker pulls AI jobs through an internal API (ADR-0017).

## Options considered

| Option | Pros | Cons |
|---|---|---|
| **Small custom queue in SQLite** (chosen) | Exactly the needed features. No extra dependency. Same database (ADR-0007). Transactional with other state | Code to write and test (~400–600 lines estimated) |
| `maragu.dev/goqite` **v0.4.0** (MIT, maintained, verified) | A small SQLite queue library | A message queue: no priorities, per-type concurrency limits, or progress. These would be built around it anyway |
| Redis- or broker-based queue | Mature | Another service, which violates "few moving parts" |

## Decision

A **persistent job queue stored in SQLite and run inside the core server** (user decision, P003).

### Implementation details chosen by agent
**Custom implementation** in `internal/jobs`, after evaluating goqite (above).
- **Table `jobs`:** `id`, `type`, `owner_id`, `payload` (JSON), `priority`, `state` (queued/running/succeeded/failed/cancelled), `attempts`, `max_attempts`, `run_after`, `lease_owner`, `lease_until`, `progress` (0–100 + message), `last_error`, `created_at`, `updated_at`.
- **Claiming:** in a write transaction, select the highest-priority due job of a type with free capacity, then set a lease (`UPDATE … RETURNING`). Workers heartbeat to extend leases. On startup, and periodically, expired leases are requeued, which gives restart survival.
- **Retries:** exponential backoff with jitter, capped, until `max_attempts`, then `failed` with the error kept.
- **Concurrency:** per-type worker pools with configured limits. A global I/O throttle for heavy jobs (NFR-012).
- **Priorities:** interactive (thumbnails for the current view) > ingest > indexing > integrity/backup > AI (plan 8.15).
- **External workers:** the AI worker claims AI job types through the internal API using the same lease protocol (ADR-0017).
- **Retention:** succeeded jobs are pruned after N days. Failed jobs are kept for the S10.4 monitor.

## Consequences

- **Easier:** one queue for everything; transactional enqueue with other state changes; no extra service.
- **Harder:** correctness of lease and heartbeat logic (thorough tests, including crash simulation, are required).
- **Required:** tests for restart survival, retry and backoff, concurrency limits, and lease expiry (S04.3 acceptance criteria). The S01.4 and S03.6 simple schedulers migrate to jobs.

## Approval record

> Decision made by the user in plan change request #3 (`code-agent-docs/prompts/P003-technology-stack.json`, `technology_stack.job_queue`). Custom vs. library is an "agent decides" item (custom chosen)
> (2026-09-24, session S003). goqite verified in S003 log E005.

## Links

- **Related requirements:** FR-067, FR-095, NFR-012
- **Related ADRs:** ADR-0007, ADR-0012, ADR-0014, ADR-0016, ADR-0017
- **Related stages:** S04.3 onward
- **Plan version:** 0.3.0

## Amendment 1 (2026-09-30, plan 1.9.0): a minimal job foundation now

- **P008 F3 (approved by the user, S007 E059):** the core of this ADR is built now, in S03, as S01 follow-up S01.4-T08, instead of the temporary S01.4 cleanup scheduler: a persistent `jobs` table (ID, type, state, payload, progress, attempts, lease, timestamps, error), workers with leases and retries with backoff, progress, and cancel.
- **Long operations become jobs:** copies and moves over the synchronous limits, large deletes and trash purges, abandoned-upload cleanup, the ID backfill (ADR-0040); later indexing and imports. The API answers `202 Accepted` with a job ID.
- **S04.3 grows it** (priorities, per-type limits, the full monitor, OS-level priority for external tools, NFR-055) instead of creating it (principle 4). A job may run one journaled operation (ADR-0041).
- **Built in S01.4-T08** (S007 E068): `internal/jobs` (migration `00005_jobs.sql`): two workers by default (a Raspberry Pi's four cores); a claim reads first, so an idle queue writes nothing; progress written at most once a second; a cancel watcher checks once a second (read-only); retries with backoff (2^attempt seconds, at most 5 minutes) only for unexpected errors, while expected files errors fail at once; running jobs re-queued at start; finished jobs pruned after 7 days; `schedule.Scheduler.Every` is now a periodic job (the upload and temporary-file cleanup runs as `uploads.cleanup`).
