# ADR-0007: Database: SQLite in WAL mode, pure-Go driver, goose migrations

| Field | Value |
|---|---|
| Number | ADR-0007 |
| Status | **Accepted** |
| Date proposed | 2026-09-24 (session S003) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none |
| Superseded by | none |

## Context

The core needs persistent internal state:
- From S01: upload session index, settings, schema version.
- Later: users, sessions, the job queue, the audit log, and the database-side mirror of sharing and access data.

P003 principle: "one core server binary, one database file". Pure-Go builds without cgo (ADR-0001). Internal data lives outside `files/` and `photos/` (I2, ADR-0003).

## Options considered

| Option | Pros | Cons |
|---|---|---|
| **SQLite (WAL), pure-Go driver `modernc.org/sqlite`** (chosen) | One file, zero ops, easy backup. No cgo. WAL allows concurrent readers alongside one writer | Single writer; the pure-Go driver is somewhat slower than the cgo one |
| SQLite with `mattn/go-sqlite3` (cgo) | Fastest SQLite binding | Needs cgo, which breaks pure-Go cross-compilation. Fallback only |
| PostgreSQL | Powerful, concurrent writers | Another service to install, run, and back up. Overkill at household scale, and conflicts with "anyone can deploy" |

## Decision

- **SQLite in WAL mode** is the internal database, **from S01 onward**, stored at `<internal data>/db/nas.db` (ADR-0003).
- **Driver:** **`modernc.org/sqlite` v1.59.0** (BSD-3-Clause, verified 2026-09-24). Its documentation discusses WAL locking via the `-shm` file and the supported platforms. Switching to **`mattn/go-sqlite3` v1.14.52** (MIT, cgo) only happens if benchmarks show a real problem, and that needs a new ADR.
- **Stores:** users, sessions, settings, the job queue (ADR-0011), the audit log, and the mirror of sharing and access data. **Not** photo metadata: sidecars are the source of truth (I3), and the search index is Bleve (ADR-0014).

### Implementation details chosen by agent
| Detail | Choice | Reason |
|---|---|---|
| Migrations | **goose v3.28.0** (`github.com/pressly/goose/v3`, MIT, LICENSE file verified). Versioned **SQL** migration files embedded with `go:embed` from `internal/db/migrations/`. Run at startup (up only) and through a CLI subcommand | Plain SQL files, embeddable, works with any `database/sql` driver, maintained |
| Connection setup | `PRAGMA journal_mode=WAL; foreign_keys=ON; busy_timeout=5000; synchronous=NORMAL` on every connection. **One writer connection** (a `database/sql` pool with `MaxOpenConns=1`) plus a separate read-only pool | Avoids `SQLITE_BUSY` storms. `synchronous=NORMAL` is durable across application crashes in WAL mode, and at most the last transactions can be lost on OS crash, which is acceptable for caches and queues. Settings and users are covered by backups (S08.3) |
| Backups | `VACUUM INTO` for consistent online copies (S08.3) | Built into SQLite |
| Verification task | S01.1 includes a test that opens the database with the driver and asserts `journal_mode=wal`, a concurrent read during a write, and that migrations apply on Linux and Windows | Confirms the driver behavior in CI |

## Consequences

- **Easier:** zero-ops persistence; single-file backup; pure-Go builds for arm64.
- **Harder:** single-writer throughput. Mitigated by short transactions and by keeping the heavy data (index, media) out of SQLite.
- **Required:** migrations from the first table; a writer/reader pool split; the S01.1 driver test.

## Approval record

> Decision made by the user in plan change request #3 (`code-agent-docs/prompts/P003-technology-stack.json`, `technology_stack.database`). The migration tool is an "agent decides" item (goose)
> (2026-09-24, session S003). Verified in S003 log E005.

## Links

- **Related requirements:** FR-070, FR-095, FR-120, NFR-006, NFR-012, NFR-017
- **Related ADRs:** ADR-0003, ADR-0010, ADR-0011, ADR-0014
- **Related stages:** S01 onward (S03 users and sessions, S04.3 jobs, S07 sharing mirror, S08.3 backup)
- **Plan version:** 0.3.0
