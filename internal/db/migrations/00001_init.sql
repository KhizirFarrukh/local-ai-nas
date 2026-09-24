-- Initial schema (S01.1-T10, ADR-0007).
-- Times are Unix milliseconds (UTC). Tables are STRICT, so a value of the
-- wrong type is an error instead of being stored as-is.

-- +goose Up
CREATE TABLE settings (
    key        TEXT    PRIMARY KEY,
    value      TEXT    NOT NULL,
    updated_at INTEGER NOT NULL
) STRICT;

-- Upload sessions (tus, S01.4). One row per upload that has been created
-- and not yet finished or expired.
CREATE TABLE uploads (
    id            TEXT    PRIMARY KEY,           -- tus upload ID
    namespace     TEXT    NOT NULL,              -- owner namespace, such as u0001 (ADR-0003)
    target_path   TEXT    NOT NULL,              -- path in the files area, relative to the namespace
    on_conflict   TEXT    NOT NULL CHECK (on_conflict IN ('fail', 'rename', 'overwrite')),
    declared_size INTEGER NOT NULL CHECK (declared_size >= 0),
    sha256        TEXT    CHECK (sha256 IS NULL OR length(sha256) = 64), -- optional expected checksum, lowercase hex
    created_at    INTEGER NOT NULL,
    expires_at    INTEGER NOT NULL
) STRICT;

CREATE INDEX uploads_expires_at ON uploads (expires_at);

-- +goose Down
DROP TABLE uploads;
DROP TABLE settings;
