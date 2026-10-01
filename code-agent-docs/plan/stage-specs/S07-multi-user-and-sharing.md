# S07: Multi-user and sharing

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](../PLAN_INDEX.md)). Moved unchanged from `plan.md` section 10.8 in version 1.9.0 (ADR-0044).

### 10.8 S07: Multi-user and sharing

- **Origin:** User-defined
- **Goal:** Multiple users, each with private files and photos by default, and explicit sharing between them.
- **User requirements (quoted):**
  > "Stage 7 is multi-user implementation."
  > "Each user has separate files and photos access."
  > "One user cannot access another user's files or photos unless they are shared with them."
  > "Users can share files with each other."
  > "A shared file then carries data about which user owns it and who has read access to it."
- **Status:** Not started
- **Security (1.12.0, P010; rule R15):** authorization and sharing are **Tier 1**; the authorization matrix test with ownership (own, shared, foreign) joins S07.7 (NFR-081).

#### S07.1: User and role model
- **Goal:** Manage multiple accounts with clear roles.
- **Scope:** users; roles (admin and standard user); the admin creates, disables, and deletes users; profiles; admin-initiated password reset. Whether the admin can see users' data is an open question (Q29).
- **Deliverables:** user and role model; admin user API; defined outcome for deleting a user's data.
- **Depends on:** S06 (Done), S03.
- **Requirements:** FR-065, FR-112.
- **Acceptance criteria:**
  1. The admin can create, disable, re-enable, and delete users. Disabled users cannot log in and their sessions end.
  2. Standard users cannot reach any admin endpoint (tested).
  3. An admin password reset forces a password change at next login.
  4. Deleting a user follows a documented data policy and never deletes data silently.
- **Risks/notes:** None.
- **Status:** Not started

#### S07.2: Per-user data separation
- **Goal:** Each user has private areas and private views.
- **Scope:** per-user namespaces inside both `files/` and `photos/` (layout per ADR-0003, confirmed or superseded here); migration of existing single-user data to the first admin; per-user timelines, albums, and search scope.
- **Deliverables:** namespace provisioning; migration or binding of existing data; scoped queries.
- **Depends on:** S07.1.
- **Requirements:** FR-071, FR-113.
- **Acceptance criteria:**
  1. Each new user gets private namespaces in both areas.
  2. Existing data belongs to the first admin without file moves (per ADR-0003), or through a tested migration if the ADR changed.
  3. Timelines, albums, and search show only the user's own items (plus shared items after S07.5).
- **Risks/notes:** RK-21 is mitigated by ADR-0003.
- **Status:** Not started

#### S07.3: Ownership and access data model
- **Goal:** Ownership and read access travel with the item, as the user specified.
- **Scope:**
  - Owner and read-access list stored with each item's metadata. For photos: the sidecar's reserved section. For the files area: an ADR with trade-offs, defaulting to the user's stated approach (Q28, section 8.8).
  - Access data mirrored into the index for fast checks.
  - Folder-sharing inheritance, and what happens to ownership on copy and move.
- **Deliverables:** files-area access ADR; access data writers; index mirroring; inheritance and copy/move rules document.
- **Depends on:** S07.2, S05.1, S06.1.
- **Requirements:** FR-114, FR-115.
- **Acceptance criteria:**
  1. Owner and access list are stored with the item as the ADR defines, and mirrored into the index.
  2. Rebuilding the index from disk restores all access data.
  3. The inheritance and copy/move rules are documented and tested.
- **Risks/notes:** Hidden files are visible over SMB unless filtered (S09).
- **P008 (1.9.0):** ownership and access are authoritative in the database, with mirrors in sidecars (FR-359, D2); JWT is not used for browser sessions, because server-side sessions can be revoked at once (ADR-0010); OIDC sign-in stays in R05.
- **Status:** Not started

#### S07.4: Authorization enforcement
- **Goal:** I5 holds on every access path.
- **Scope:** extend the S03 central policy check to every access path: API, downloads, previews, thumbnails, search results, and background jobs (I5); default deny; no information leaks about items the user cannot access.
- **Deliverables:** ownership and ACL rules in `authorize()`; search permission pre-filter; job permission context; route and job inventory tests.
- **Depends on:** S07.3, S03.5.
- **Requirements:** FR-089, NFR-024.
- **Acceptance criteria:**
  1. Inventory tests prove every route and job type calls the policy check.
  2. Requesting another user's unshared item looks exactly like requesting a non-existent one.
  3. Search results, counts, facets, and autocomplete never reveal inaccessible items.
  4. Background jobs run with the owning user's permissions.
- **Risks/notes:** Leak channels are subtle (RK-17). See 8.9.
- **P008 (1.9.0):** only the policy layer decides, from the database; mirrors are never read for access decisions (FR-359).
- **Status:** Not started

#### S07.5: Sharing
- **Goal:** Users share items explicitly and can take shares back.
- **Scope:** share files, folders, photos, and albums with specific users (read access per the user's requirement; write access is an open question, Q30); revoke; "Shared with me" and "Shared by me" views; copying a shared item into one's own space.
- **Deliverables:** share API; share views; copy-to-own-space action.
- **Depends on:** S07.4.
- **Requirements:** FR-116, FR-117.
- **Acceptance criteria:**
  1. A shared item becomes readable by the recipient and nobody else.
  2. Revoking removes access immediately, including from search.
  3. "Shared with me" and "Shared by me" list exactly the current shares.
  4. Copying a shared item creates a new item owned by the recipient.
- **Risks/notes:** None.
- **Status:** Not started

#### S07.6: Multi-user GUI
- **Goal:** All multi-user features are usable from the GUI.
- **Scope:** admin user-management page; share dialog; shared views; owner and access details on each item.
- **Deliverables:** GUI pages and dialogs.
- **Depends on:** S07.5, S02, S04.7.
- **Requirements:** FR-116, FR-118.
- **Acceptance criteria:**
  1. The admin manages users entirely from the GUI.
  2. Users share and revoke from both the files and photos GUIs.
  3. Every item's details show its owner and who has access.
- **Risks/notes:** None.
- **1.7.0:** the admin user-management page is the console's Users and groups section (6.6, ADR-0039).
- **Status:** Not started

#### S07.7: Access control testing and stage review
- **Goal:** Prove isolation and sharing, then close the stage.
- **Scope:** cross-user tests for every endpoint (a non-owner without a share must always be refused); search leak tests; audit logging of share and revoke events; documentation; completion record; user sign-off.
- **Deliverables:** cross-user suite; leak suite; audit event checks; multi-user documentation; updated threat model; completion record.
- **Depends on:** S07.1–S07.6.
- **Requirements:** FR-090, NFR-024.
- **Acceptance criteria:**
  1. For every endpoint, a test proves that a non-owner without a share is refused.
  2. Search leak tests pass (no inaccessible item affects results, counts, or suggestions).
  3. Share and revoke events appear in the audit log.
  4. The completion record is written and the user's sign-off is recorded.
- **Risks/notes:** None.
- **Status:** Not started

**Design notes (S07):**
- `owner` has been in the model since S01, and namespaces exist since S01.2.
- The sidecar `access` section is reserved since S05.1, and the index `acl` field since S06.1.
- `authorize()` is extended, not replaced, and audit event types are extended.

**Exit criteria (quoted):** "No user can reach another user's data by any path unless it is shared, and sharing and revoking work in both the API and the GUI."

---
