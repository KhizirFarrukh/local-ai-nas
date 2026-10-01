# ADR-0048: Secrets management

| Field | Value |
|---|---|
| Number | ADR-0048 |
| Status | Proposed |
| Date proposed | 2026-10-01 (session S007) |
| Date of last status change | 2026-10-01 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

The NAS keeps some secrets it must use again: two-factor secrets (S03.7), SMTP and webhook credentials (FR-221), later remote-access and backup credentials (R04, R06). Others it generates: session keys, the internal API token for the AI worker and the storage helper. Passwords and tokens are stored only as hashes (ADR-0010). A database backup or a stolen database file must not reveal the reusable secrets (threat T-46; NFR-074; the user's requirement UR-4).

## Options considered

### Option A: Store secrets in the database in plain text, protected by file permissions
- **Cons:** every database backup and snapshot reveals them.

### Option B: Encrypt them with a key kept outside the database and outside its backup set (recommended)
- **Algorithm:** AES-256-GCM (Go standard library) or XChaCha20-Poly1305 (`golang.org/x/crypto`, already a dependency); a random nonce per value; the record ID as associated data, so a value cannot be moved to another record.
- **Key:** 256 bits from `crypto/rand` at first start; kept in the configuration folder with 0600 (Linux) or a restricted ACL (Windows), not in `.local-ai-nas/state/`, which is backed up; optionally the OS key store where available (to decide).
- **Backups:** the key is never in the database backup; the console offers to export it once as a recovery key the owner stores separately; without it, restored secrets must be entered again (the backup itself stays usable).
- **Rotation:** a command that re-encrypts every value with a new key; part of incident response.

### Option C: Derive the key from the admin's password
- **Cons:** the NAS could not send alerts or run jobs after a restart until someone signs in.

## Decision

**Option B (recommendation while Proposed).** To decide on acceptance: the exact key location per platform (Linux packages, Docker volume, Windows service), and whether to use the OS key store.

## Consequences

- **Easier:** backups and snapshots are safe to keep.
- **Harder:** losing the key means re-entering the stored secrets; the recovery-key export must be explained well.
- **Required:** the secrets store package in S03 before S03.7; documentation in the crypto policy and the storage security guide.

## Approval record

_Not yet approved._

## Links

- **Related requirements:** NFR-074, NFR-080, FR-091, FR-221, FR-355
- **Related ADRs:** ADR-0010 (passwords and sessions), ADR-0003 (internal data layout)
- **Related stages:** S03, S08, S10
- **Plan version:** 1.12.0
