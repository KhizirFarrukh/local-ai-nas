# A001: README change proposal

| Field | Value |
|---|---|
| Audit | A001 (P004), session S004, 2026-09-24 |
| Finding | F-018 (README out of date with approved decisions) |
| Status | **Decided (S005, D-05):** R-01–R-08 and R-10 **approved and applied** to `README.md` in S005; R-09 (optional technology section) not applied. User's answer: "Accept all (Recommended)" |

**Basis rule:** only decisions that are already approved are used. That means:
- the user's own instructions in P002/P003;
- invariants in RULES.md (pre-approved through P002/P003);
- Accepted ADRs.

Draft plan content is **not** used. For example, the planner-proposed stages S08–S11, the MVP proposal, and ADR-0003 (Proposed) are left out.

---

### R-01: Overview paragraph (two-area design)

**Current (line 5):**
> **local-ai-nas** is a network-attached storage platform that anyone can deploy on their own hardware. Beyond storing and serving files, it organizes your photos, understands what's in them, and lets you search your library with natural, forgiving queries, all without sending a single byte to the cloud.

**Proposed:**
> **local-ai-nas** is a network-attached storage platform that anyone can deploy on their own hardware. It keeps two separate areas: **Files**, a general-purpose file store, and **Photos**, a Google Photos–style library. It organizes your photos, understands what's in them, and lets you search both areas with natural, forgiving queries, all without sending a single byte to the cloud.

**Reason:** invariant I1 (RULES.md; P002): "The storage root contains exactly two user-data areas: files/ and photos/." P002 S06 user requirement: "Stage 6 is searching, in both files and photos."

---

### R-02: "NAS Core" features

**Current (lines 13–16):**
```
### 📁 NAS Core
- Store, browse, upload, and download files over your local network
- Self-hosted on your own machine: your data never leaves your system
- Designed to be simple to deploy on any system
```

**Proposed:**
```
### 📁 NAS Core
- Store, browse, upload, and download files over your local network, with resumable uploads for large files
- Two separate areas, **Files** and **Photos**, that never mix: content moves between them only when you explicitly copy or move it
- Multiple users, each with private files and photos, and explicit sharing between users (planned)
- Self-hosted on your own machine: your data never leaves your system
- Designed to be simple to deploy on any system
```

**Reason:**
- I1 (RULES.md; P002 S04 user requirements: "Content moves between them only when the user selects a file to copy or move to files or photos").
- P002 S07 user requirements ("Each user has separate files and photos access… Users can share files with each other").
- ADR-0008 (Accepted: resumable uploads with tus).

---

### R-03: Photo Management, sidecar sentence

**Current (line 21):**
> - Each photo gets its own **sidecar JSON metadata file** stored right next to it

**Proposed:**
> - Each photo in the Photos area gets its own **sidecar JSON metadata file** stored right next to it

**Reason:** I1 and I3 (RULES.md). Sidecars belong to the photos area.

---

### R-04: Optional Local AI, intro sentence

**Current (line 25):**
> When enabled, an AI model runs **locally** alongside the NAS software. Nothing is sent to external services.

**Proposed:**
> When enabled, AI runs **locally** in a separate, optional component alongside the NAS software. It has read-only access to your photos, and the NAS itself writes all results. Nothing is sent to external services, and only models whose licenses let anyone use them are included.

**Reason:**
- ADR-0017 (Accepted): separate optional container, read-only media mount.
- Invariant I9 (RULES.md; P003): the core is the single writer.
- NFR-029 / ADR-0018 (Accepted): the P003 license policy.

---

### R-05: Smart Search, scope and fields

**Current (lines 32–39):**
```
Search runs over stored metadata, so it is fast and doesn't require the AI model to be running.

Searchable fields:
- File name
- Description
- Place (derived from GPS data)
- Date and time
- AI-generated tags and face groups
```

**Proposed:**
```
Search covers both the Files and Photos areas. It runs over stored metadata and a local search index, so it is fast and doesn't require the AI model to be running.

Searchable fields:
- File name (both areas)
- Description
- Place (derived from GPS data)
- Date and time
- AI-generated tags and face groups (when AI is enabled)
```

**Reason:**
- P002 S06 user requirement ("searching, in both files and photos").
- I4 (RULES.md): search reads only the index.
- ADR-0014 (Accepted): Bleve embedded index.

---

### R-06: Search operators table

**Current (lines 48–56):** 7 operators: `before:`, `after:`, `on:`, `place:`, `tag:`, `face:`, `type:`.

**Proposed:** keep the 7 rows, change the `face:` description to "Photos containing a named face group (available when AI is enabled)", and add:
```
| `in:` | `in:photos` | Limit results to the Files or Photos area |
| `ext:` | `ext:pdf` | Filter by file extension |
| `size:` | `size:>10MB` | Filter by file size |
```

**Reason:** P002 S06.3 (the user's baseline operator set): "before:, after:, on:, place:, tag:, type:, in:files, in:photos, ext:, size: … face: is reserved for S12."

---

### R-07: Getting Started, planned deployment options

**Current (lines 139–141):**
```
Planned deployment options:
- [ ] Docker / Docker Compose
- [ ] Native install (Linux, Windows, macOS)
```

**Proposed:**
```
Planned deployment options:
- [ ] Docker Compose (primary), with images for x86-64 and ARM64 (e.g. Raspberry Pi); optional AI component enabled with a Compose profile
- [ ] Native install on Linux (single binary + systemd service)
- [ ] Native install on Windows and macOS (under consideration)
```

**Reason:** ADR-0006 (Accepted): Docker Compose primary, multi-architecture, native Linux secondary, native Windows/macOS deferred pending the user's decision.

---

### R-08: Roadmap order

**Current (lines 147–157):** an 11-item feature list (core storage and web UI … user accounts and permissions).

**Proposed:**
```
- [ ] Stage 1: Basic NAS (storage service and API; Files and Photos areas)
- [ ] Stage 2: Web interface for the NAS
- [ ] Stage 3: Security (login, HTTPS, hardening)
- [ ] Stage 4: Media management (Photos area: timeline, albums, viewer)
- [ ] Stage 5: Media metadata (EXIF, sidecar JSON, offline place names)
- [ ] Stage 6: Search across Files and Photos (fuzzy, synonyms, operators)
- [ ] Stage 7: Multiple users and sharing
- [ ] Final stage: optional local AI (auto-classification, face grouping)
```

**Reason:**
- P002 user-defined stages S01–S07 and S12, and invariant I8 ("AI work is always the last stage").
- The planner-proposed stages S08–S11 are **not** included because they are not yet approved.

---

### R-09: New section "Technology" (optional)

**Current:** none.

**Proposed (after "Privacy Principles"):**
```
## 🧱 Technology

- Core server: Go, as a single binary; embedded SQLite database; embedded search index (Bleve)
- Web interface: SvelteKit, served by the NAS itself
- Media tools: ExifTool, libvips (with HEIC support), FFmpeg
- Resumable uploads: the tus protocol
- Network drive access: WebDAV
- Optional AI: a Python worker using ONNX Runtime, running fully offline
```

**Reason:** Accepted ADR-0001, 0007, 0014, 0009, 0012, 0008, 0015, 0017 (P003).

---

### R-10: Photo Management, video playback (added after A001, same session)

**Current (lines 18–22):** no mention of video playback.

**Proposed:** add a bullet to "🖼️ Photo Management":
```
- Video playback streamed to any device, with a live quality menu (Auto, Original, 1080p, 720p, …); lower qualities are prepared on demand
```

**Reason:** ADR-0020 (Accepted from the user's decisions in S004 E005/E008); plan 0.4.0 FR-144–FR-147.

---

### Not proposed (and why)
- **License (line 169):** still undecided (Q22). Unchanged.
- **Sidecar schema example (lines 82–120):** the schema is decided in S05.1 (ADR pending). Unchanged.
- **Status line (line 7):** still accurate. Unchanged.
- **Planner-proposed stages (S08–S11), MVP milestones, storage-layout details (ADR-0003):** not approved yet.
