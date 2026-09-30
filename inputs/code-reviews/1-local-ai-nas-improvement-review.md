# Local AI NAS — Architecture & Project Improvement Review

Repository: https://github.com/KhizirFarrukh/local-ai-nas

## Overall Assessment

The technical direction is strong, but the project is currently over-planned relative to how much has actually been implemented.

The core stack and most of the architecture should remain. However, several structural decisions should be tightened **before S01 begins**, because changing them later would be significantly more painful.

The project already has strong foundations:

- Separate `files/` and `photos/` areas
- Sidecars as durable photo metadata
- Rebuildable search indexes
- Centralized authorization
- Go core as the main writer
- Separate Python AI worker
- AI intentionally delayed until storage, security, metadata, search, and recovery are stable

The main risks are not the programming language, database, or framework choices. The biggest risks are:

1. Path-dependent identity
2. Metadata/sidecar collisions
3. Security state living in editable metadata
4. Replacing early infrastructure later instead of growing it
5. Permanent deletion before Trash exists
6. Excessive planning/process overhead
7. Contradictions in semantic-search requirements
8. A master plan that is too large to serve as hot context for every coding session

---

# Recommended Changes Before Coding

## 1. Introduce an Immutable `ItemID` / `MediaID`

This is the most important architectural improvement.

The current item model is effectively based on:

```text
OwnerID
Area
RelPath
Kind
Size
ModTime
MIME
ETag
```

but there is no permanent identity independent of the path.

That will eventually cause problems with:

- Renames
- Moves
- Albums
- Shares
- Search results
- Background jobs
- Face assignments
- Thumbnails
- External filesystem reconciliation
- Duplicate detection
- Audit logs

Every stored object should have a stable identifier, ideally something like UUIDv7 or ULID.

Example:

```text
ItemID: 0199e...
OwnerID: u0001
Area: photos
RelPath: 2026/Karachi/IMG_1204.jpg
```

If the file is moved, only `RelPath` changes.

For photos, the media ID should also exist in the application's sidecar metadata.

Albums, face mappings, jobs, shares, and other internal references should use IDs wherever possible instead of paths.

---

## 2. Change the Sidecar Filename Format

The proposed structure:

```text
IMG_0001.jpg
IMG_0001.jpg.json
```

is risky because tools such as Google Takeout already use similar `.json` sidecars.

The project also states that foreign JSON files must never be overwritten.

Those two requirements conflict.

Use a project-specific namespace instead:

```text
IMG_0001.jpg
IMG_0001.jpg.lainas.json
```

or:

```text
IMG_0001.jpg.local-ai-nas.json
```

Then a directory can safely contain:

```text
IMG_0001.jpg
IMG_0001.jpg.json
IMG_0001.jpg.lainas.json
```

where:

- `IMG_0001.jpg.json` may belong to Google Takeout or another tool
- `IMG_0001.jpg.lainas.json` belongs to Local AI NAS

This avoids filename collisions and avoids having to inspect JSON schemas just to determine file ownership.

---

## 3. Do Not Make Security/ACL Data Authoritative Inside Photo Sidecars

Portable metadata and security authority should be separated.

Sidecars are appropriate for descriptive metadata such as:

```text
capture time
tags
description
rating
GPS corrections
AI labels
face references
media ID
```

SQLite should be authoritative for:

```text
owner
ACL
shares
permissions
public links
revocation
security-related state
```

The sidecar can optionally contain a non-authoritative mirror of some access information for export or inspection.

However, the central authorization system should never trust an externally editable JSON file to decide whether access is allowed.

Otherwise, somebody with filesystem or WebDAV access could potentially alter security metadata outside the intended authorization boundary.

---

## 4. Introduce a Minimal Durable Job/Operation Foundation in S01

The early plan currently uses temporary cleanup/scheduling mechanisms that are intended to be replaced by the full job system later.

Since SQLite is already part of S01, it would be cleaner to introduce a very small durable job foundation immediately.

Example:

```text
jobs
├── id
├── type
├── state
├── payload
├── progress
├── attempts
├── created_at
└── updated_at
```

S01 does not need the complete future job system.

It only needs enough to safely handle operations such as:

```text
copying an 80 GB directory
deleting a large tree
cleaning abandoned uploads
indexing imported content
reconciling external filesystem changes
```

Then S04 can expand this system rather than replacing it.

This is also the right place to define crash/recovery behavior for operations touching multiple systems:

```text
filesystem
sidecar
SQLite
search index
```

An operation should remain resumable or safely retryable if:

```text
file write succeeds
rename succeeds
database commit fails
disk fills
application crashes
power disappears
```

The plan already treats atomic sidecar writes seriously. The same philosophy should apply to complete logical operations.

---

## 5. Move Trash Earlier

A NAS dealing with personal files and photos should not have permanent deletion as the normal early behavior.

The safer model is:

```text
Delete
   ↓
Trash
   ↓
Retention period
   ↓
Permanent deletion
```

instead of:

```text
Delete
   ↓
Gone
```

Move minimal Trash functionality from the later recovery stage into roughly S02 or S03.

Full backup, snapshot, and recovery functionality can remain later.

Minimal configuration/database backup could also be considered earlier.

---

## 6. Add Basic Browser-Side Security Before Full Authentication

It is reasonable to delay full authentication until S03.

However, even a localhost-only web application should protect destructive endpoints early.

Before the GUI can perform meaningful write/delete operations, add:

```text
strict Host validation
strict Origin validation
no permissive CORS
localhost-only bind
state-changing request protection
```

An unpredictable development/session token could be used before the complete authentication system exists.

Then S03 adds:

```text
user accounts
Argon2id
sessions
CSRF protection
authorization
rate limiting
share permissions
```

---

## 7. Resolve the Semantic Search Contradiction

The current architecture contains two requirements that conflict.

One part correctly acknowledges that semantic search requires a query/text model at search time.

Another part says AI should never be computed at query time.

True arbitrary semantic search requires something like:

```text
User query:
"photos of somebody standing beside a red car at night"

        ↓

text/query embedding

        ↓

vector similarity search
```

Image embeddings can be generated ahead of time.

The query embedding cannot be fully precomputed because the user can type arbitrary text.

A better invariant would be:

> No media inference or expensive AI pipeline runs in request paths. Lightweight local query embedding is allowed for semantic search.

This preserves the original intent while allowing real semantic search.

---

## 8. Reduce Project-Management Ceremony

The current agent-development process is extremely rigorous.

It includes things such as:

- Reading several large documents at session start
- Logging user messages
- Updating state frequently
- Task/status gates
- Branch discipline
- Commit-per-task expectations
- Documentation audits
- ADR approval gates
- Session logs
- Large master-plan context

Some of this is useful.

However, the process is approaching the point where maintaining the development process becomes a project of its own.

A simpler model would be:

```text
Every session reads:
  CURRENT_STATE.md
  active stage document
  relevant ADRs

Only when architecture context is required:
  relevant sections of plan.md

Every meaningful architectural decision:
  update/add ADR

Every completed task:
  commit

Every session:
  brief technical summary
```

Avoid storing complete conversation transcripts.

For a public repository especially, session records should contain concise technical decisions rather than verbatim user/agent messages.

---

## 9. Split `plan.md` Into Smaller Authoritative Documents

The master plan is useful as a complete architectural specification, but it is too large to be consumed repeatedly during every development session.

A better structure could be:

```text
docs/
├── ARCHITECTURE.md
├── REQUIREMENTS.md
├── ROADMAP.md
├── INVARIANTS.md
├── SECURITY.md
│
├── stages/
│   ├── S01.md
│   ├── S02.md
│   ├── S03.md
│   └── ...
│
├── adr/
│   └── ...
│
└── PLAN_INDEX.md
```

`PLAN_INDEX.md` could remain relatively short and contain:

- Stage status
- Stage dependencies
- Current milestone
- Links to detailed stage documents
- Links to relevant ADRs

The detailed specification should remain.

The improvement is simply to avoid using the entire master document as mandatory hot context.

---

## 10. Keep the Existing Core Stack

The current stack is appropriate:

```text
Go
SQLite
SvelteKit / Svelte
Bleve
ExifTool
libvips
FFmpeg
WebDAV
Python / ONNX AI worker
```

There is no strong reason to replace it.

Most importantly:

**Do not turn this into microservices.**

Logical components such as:

```text
FilesService
PhotosService
SearchService
JobService
AuthService
```

can remain inside one Go application.

That gives clean internal boundaries without adding network/service complexity.

The Python AI worker is a reasonable separate process because its runtime, dependencies, model loading, and hardware concerns are meaningfully different from the Go core.

---

# Search Architecture Improvements

Do not apply the same analyzer to every searchable field.

Suggested handling:

| Field | Suggested Handling |
|---|---|
| Filename | normalized exact + token/prefix |
| File path | normalized exact/prefix |
| Description | language-aware full text |
| AI caption | full text + semantic |
| Tags | exact normalized |
| Person name | exact + prefix + typo tolerance |
| Camera model | exact/token |
| Location | exact + hierarchy |
| MIME/type | keyword |
| Dates | structured filters |

For example, Porter stemming should not be applied to person names or filenames.

Also design for Unicode and multilingual metadata from the beginning.

Potential content may include:

```text
English
Urdu
Roman Urdu
Arabic-script names
mixed-language metadata
```

A good foundation would be:

- Unicode-safe normalization
- Exact keyword fields where appropriate
- Prefix searching
- Language analyzers only on fields where they make sense

This is much easier to design correctly now than retrofit later.

---

# SQLite Durability Consideration

The current design uses:

```text
WAL mode
synchronous=NORMAL
```

This is reasonable for rebuildable or non-critical data such as:

```text
jobs
thumbnail state
index queue
cache state
```

Later, however, the same database may contain:

```text
users
sessions
shares
settings
audit/security state
```

These have different durability expectations.

Before the authentication/sharing stages, explicitly review whether:

```text
synchronous=FULL
```

is more appropriate.

A home NAS is unlikely to be bottlenecked by account/share-setting transactions.

The durability benefit may be worth the small performance cost.

---

# README Improvements

The README should make the current project status immediately obvious.

Suggested section:

```text
Current status: Planning / pre-alpha

Working today:
❌ File server
❌ Photo library
❌ Search
❌ User accounts
❌ AI

Architecture/specification:
✅ Master plan
✅ ADRs
✅ S01 specification

Next milestone:
S01 — Basic NAS API
```

Add a small architecture diagram:

```text
              Browser / PWA
                    │
                 REST API
                    │
              ┌──── Go Core ────┐
              │                 │
          FilesService      PhotosService
              │                 │
           files/            photos/
                                │
                             sidecars
              │                 │
              └───── SQLite ────┘
                     │
                  Job Queue
                     │
              ┌──────┴──────┐
              │             │
          Indexer       AI Worker
                         Python/ONNX
```

Also choose the project's own software license before meaningful outside contributions are accepted.

---

# Suggested Roadmap Adjustment

There is no need to redesign the entire twelve-stage roadmap.

A few foundations should simply move earlier.

| Stage | Suggested Change |
|---|---|
| S01 | Add immutable IDs, operation/job foundation, basic browser-origin protection |
| S02 | GUI |
| S03 | Full authentication/security + Trash |
| S04 | Photo pipeline; prioritize direct video playback |
| S05 | Namespaced sidecars such as `.lainas.json` |
| S06 | Search; good candidate for first genuinely usable release |
| Later | Sharing, backups, WebDAV, administration, packaging |
| S12 | AI remains last |

Keeping AI late is a good decision.

AI is the exciting part of the project, but storage integrity, authorization, recovery, metadata identity, and backup architecture are much harder to change once the NAS already contains a large photo library.

---

# Highest-Priority Changes

If only three architectural changes are made before implementation starts, they should be:

## 1. Add Immutable `ItemID` / `MediaID`

Do not make paths the identity of stored objects.

## 2. Rename Application Sidecars

Use something namespaced such as:

```text
.lainas.json
```

instead of generic:

```text
.json
```

## 3. Keep Ownership and ACL Authority in SQLite

Portable metadata can live in sidecars.

Security authority should remain in the application's controlled database and central authorization layer.

These three decisions affect almost every future stage and are cheapest to solve before real implementation and user data exist.

---

# Final Recommendation

The project does **not** need a new language, new database, or radically different architecture.

The overall direction is good.

The main risk is that the planning and agent-management framework becomes so elaborate that implementation progresses too slowly.

Before starting S01:

1. Add stable resource IDs.
2. Namespace sidecar filenames.
3. Make SQLite authoritative for security/ACL data.
4. Pull minimal Trash functionality earlier.
5. Introduce a tiny durable job/operation foundation.
6. Resolve the semantic-search query-time AI rule.
7. Simplify session/process requirements.
8. Split the master plan into smaller hot-context documents.
9. Update the README.
10. Freeze the architecture and begin implementation.

After these changes, avoid another broad architecture-planning cycle unless implementation exposes a real issue.

The project is ready to move from planning into execution.
