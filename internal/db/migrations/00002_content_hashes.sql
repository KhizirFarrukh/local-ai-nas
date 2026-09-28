-- Content hashes of files (S01.3-T10, FR-211, ADR-0021). One row per file
-- in a namespace of the files area. A row counts only while the file still
-- has the ETag it was hashed at (its size, modification time, and file
-- ID), so a file changed outside the app is never reported with a stale
-- hash. Paths are slash-separated and relative to the namespace, never
-- absolute, so the storage root can move (NFR-036).

-- +goose Up
CREATE TABLE content_hashes (
    namespace TEXT    NOT NULL,             -- owner namespace, such as u0001
    path      TEXT    NOT NULL,             -- relative path, such as docs/a.txt
    etag      TEXT    NOT NULL,             -- the file version that was hashed
    hash      TEXT    NOT NULL CHECK (hash LIKE 'sha256:%' AND length(hash) = 71),
    hashed_at INTEGER NOT NULL,             -- Unix milliseconds
    PRIMARY KEY (namespace, path)
) STRICT;

-- +goose Down
DROP TABLE content_hashes;
