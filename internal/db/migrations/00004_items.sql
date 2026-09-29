-- Stable item IDs (S01.3-T11, ADR-0040, FR-346). Every file and folder the
-- app creates, or first sees, gets a permanent ID; rename and move keep it,
-- a copy gets a new one, a delete retires it. Paths are attributes that
-- change; the ID never does. Other tables refer to items by ID. Paths are
-- relative to the namespace, never absolute, so the storage root can move
-- (NFR-036).

-- +goose Up
CREATE TABLE items (
    id            TEXT    PRIMARY KEY CHECK (length(id) = 36),  -- UUIDv7, canonical text form
    namespace     TEXT    NOT NULL,                             -- owner namespace, such as u0001
    area          TEXT    NOT NULL CHECK (area IN ('files', 'photos')),
    path          TEXT    NOT NULL CHECK (path <> '' AND path <> '.'), -- relative path, such as docs/a.txt
    kind          TEXT    NOT NULL CHECK (kind IN ('file', 'dir')),
    status        TEXT    NOT NULL CHECK (status IN ('present', 'missing', 'trashed', 'retired')),
    size          INTEGER,                                      -- last seen (reconciliation, S05.7)
    mod_time      INTEGER,                                      -- last seen, Unix milliseconds
    file_id       TEXT,                                         -- filesystem identity, only to re-associate external moves
    missing_since INTEGER,                                      -- when status became missing
    created_at    INTEGER NOT NULL,
    updated_at    INTEGER NOT NULL
) STRICT;

-- One present item per path; missing, trashed, and retired rows keep their
-- last path for history and references.
CREATE UNIQUE INDEX items_present_path ON items (namespace, area, path) WHERE status = 'present';
CREATE INDEX items_status ON items (status);

-- +goose Down
DROP TABLE items;
