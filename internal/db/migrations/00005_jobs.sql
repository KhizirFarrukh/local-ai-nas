-- The durable job foundation (S01.4-T08, ADR-0011 amendment 1, FR-351).
-- Long operations and periodic work run as jobs that survive a restart:
-- a job that was running when the server stopped is queued again at the
-- next start. S04.3 grows this table (priorities, per-type limits) instead
-- of replacing it.

-- +goose Up
CREATE TABLE jobs (
    id               TEXT    PRIMARY KEY CHECK (length(id) = 36), -- UUIDv7
    type             TEXT    NOT NULL,                            -- such as files.copy
    owner            TEXT    NOT NULL,                            -- the namespace it works for; '' for system jobs
    state            TEXT    NOT NULL CHECK (state IN ('queued', 'running', 'succeeded', 'failed', 'canceled')),
    payload          TEXT    NOT NULL CHECK (json_valid(payload)),
    result           TEXT    CHECK (result IS NULL OR json_valid(result)),
    progress_done    INTEGER NOT NULL DEFAULT 0,
    progress_total   INTEGER NOT NULL DEFAULT 0,
    attempts         INTEGER NOT NULL DEFAULT 0,
    max_attempts     INTEGER NOT NULL CHECK (max_attempts >= 1),
    cancel_requested INTEGER NOT NULL DEFAULT 0 CHECK (cancel_requested IN (0, 1)),
    run_after        INTEGER NOT NULL,                            -- not before (retry backoff)
    lease_until      INTEGER,                                     -- while running
    error            TEXT,                                        -- the last failure, safe to show
    created_at       INTEGER NOT NULL,
    updated_at       INTEGER NOT NULL,
    finished_at      INTEGER
) STRICT;

CREATE INDEX jobs_runnable ON jobs (state, run_after);
CREATE INDEX jobs_owner ON jobs (owner, created_at);

-- +goose Down
DROP TABLE jobs;
