-- Accounts, sessions, API tokens, and the audit trail (S03.2-T01, S03.3,
-- S03.6; ADR-0010, ADR-0003). Times are Unix milliseconds (UTC). Secrets
-- are never stored: passwords only as Argon2id hashes (PHC strings),
-- sessions and API tokens only as the SHA-256 of their secret, so a copy
-- of the database reveals no password, cookie, or token (threats T-06,
-- T-13).

-- +goose Up
CREATE TABLE users (
    id                  INTEGER PRIMARY KEY,
    -- Lowercase ASCII (the auth package normalizes it); NOCASE keeps the
    -- uniqueness case-insensitive even for a row written by hand.
    username            TEXT    NOT NULL UNIQUE COLLATE NOCASE CHECK (length(username) BETWEEN 1 AND 64),
    -- The owner namespace of the user's areas, such as u0001 (ADR-0003).
    -- The first admin gets u0001, which S01 already created on disk.
    namespace           TEXT    NOT NULL UNIQUE CHECK (length(namespace) >= 5 AND namespace GLOB 'u[0-9]*' AND substr(namespace, 2) NOT GLOB '*[^0-9]*'),
    role                TEXT    NOT NULL CHECK (role IN ('admin', 'user')),
    password_hash       TEXT    NOT NULL CHECK (password_hash GLOB '$argon2id$*'),
    password_changed_at INTEGER NOT NULL,
    created_at          INTEGER NOT NULL,
    updated_at          INTEGER NOT NULL,
    disabled_at         INTEGER                               -- NULL while the account is active
) STRICT;

-- Server-side sessions (ADR-0010). The cookie holds a random 256-bit
-- value; only its SHA-256 is stored.
CREATE TABLE sessions (
    id           INTEGER PRIMARY KEY,                         -- shown in the session list
    secret_hash  BLOB    NOT NULL UNIQUE CHECK (length(secret_hash) = 32),
    user_id      INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    csrf_token   TEXT    NOT NULL CHECK (length(csrf_token) >= 32),
    scheme       TEXT    NOT NULL CHECK (scheme IN ('http', 'https')), -- the listener it was created on
    source_addr  TEXT    NOT NULL,                            -- for the session list
    user_agent   TEXT    NOT NULL,                            -- shortened, for the session list
    created_at   INTEGER NOT NULL,
    last_seen_at INTEGER NOT NULL,                            -- written at most once a minute
    expires_at   INTEGER NOT NULL,                            -- the absolute lifetime
    reauth_at    INTEGER                                      -- the last re-authentication
) STRICT;

CREATE INDEX sessions_user ON sessions (user_id);
CREATE INDEX sessions_expires_at ON sessions (expires_at);

-- API tokens for scripts (FR-087): lan_<id>_<secret>; only the SHA-256 of
-- the secret is stored.
CREATE TABLE api_tokens (
    id           TEXT    PRIMARY KEY CHECK (length(id) BETWEEN 16 AND 64),
    user_id      INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name         TEXT    NOT NULL CHECK (length(name) BETWEEN 1 AND 100),
    secret_hash  BLOB    NOT NULL CHECK (length(secret_hash) = 32),
    scopes       TEXT    NOT NULL,                            -- space-separated, such as "files:read files:write"
    created_at   INTEGER NOT NULL,
    last_used_at INTEGER,                                     -- written at most once a minute
    expires_at   INTEGER,                                     -- NULL: no expiry
    revoked_at   INTEGER
) STRICT;

CREATE INDEX api_tokens_user ON api_tokens (user_id);

-- The security audit trail (FR-090, S03.6). Append-only: a trigger
-- refuses every change, the API has no route to change or delete events,
-- and only the retention job deletes old ones. No foreign key to users:
-- an event outlives its actor.
CREATE TABLE audit_events (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,          -- never reused, even after retention
    at            INTEGER NOT NULL,
    type          TEXT    NOT NULL CHECK (type GLOB '[a-z]*.[a-z]*'), -- dotted, such as auth.login.failed
    outcome       TEXT    NOT NULL CHECK (outcome IN ('success', 'failure')),
    actor_user_id INTEGER,                                    -- NULL for the CLI and anonymous callers
    actor         TEXT    NOT NULL,                           -- a username, "cli", or "anonymous"
    source_addr   TEXT    NOT NULL,
    user_agent    TEXT    NOT NULL,
    target        TEXT    NOT NULL,
    details       TEXT    NOT NULL CHECK (json_valid(details)), -- never secrets
    request_id    TEXT    NOT NULL
) STRICT;

CREATE INDEX audit_events_at ON audit_events (at);
CREATE INDEX audit_events_type_at ON audit_events (type, at);

-- +goose StatementBegin
CREATE TRIGGER audit_events_append_only BEFORE UPDATE ON audit_events
BEGIN
    SELECT RAISE(ABORT, 'audit events cannot be changed');
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER audit_events_append_only;
DROP TABLE audit_events;
DROP TABLE api_tokens;
DROP TABLE sessions;
DROP TABLE users;
