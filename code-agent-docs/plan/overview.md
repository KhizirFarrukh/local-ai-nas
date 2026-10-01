# Overview, goals, and principles

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](PLAN_INDEX.md)). Moved unchanged from `plan.md` section 1, 2, 2b in version 1.9.0 (ADR-0044).

## 1. Project overview

**local-ai-nas** is a self-hosted NAS. It runs on the user's own hardware and serves files to devices on the local network. It is built around a **two-area design**. The storage root holds exactly two user-data areas:

- **`files/`**: a general-purpose file store (documents, archives, anything), managed like a classic NAS.
- **`photos/`**: a separate photo and video library with Google Photos–style management: timeline, albums, viewer, video streaming with live quality switching, place names, and later AI tagging and face groups.

The areas never intersect. Content moves between them only when the user explicitly copies or moves it (invariant I1). Internal application data (database, search index, thumbnails, trash, configuration) lives outside both areas (I2).

Every photo has a **sidecar JSON file** next to it (`IMG_0001.jpg.lainas.json`, a suffix only this project uses, since 1.9.0) that is the **source of truth** for its descriptive metadata; ownership and sharing are authoritative in the database (I3, 1.9.0). The search index is a cache that can always be rebuilt from disk (I3).

**Search** covers both areas through one query path. It reads only the index (I4) and is forgiving: word forms (`receipts` → `receipt`), synonyms (`receipts` → `invoice`, `voucher`), typos (`reciept`), and operators such as `before:2026`, `place:karachi`, `in:photos`, and `type:video`.

The system grows into a **multi-user** NAS: each user has private files and photos, and items can be shared explicitly with ownership and read-access data (I5). Later stages find duplicates and look-alike photos (S11), shrink the library on request (S12), and combine drives into pools (S15). **Optional, fully local AI** (I6, I7) comes last (I8). It classifies photos (so unlabelled receipts are found by searching `receipts`) and groups faces. Results are stored in the sidecars, so search never runs a model.

**Staged approach.** The work is divided into 17 stages. Each stage is split into substages and, just in time, into tasks (section 9). Every stage ends with a working, tested, demonstrable system (section 2b).

| Stage | Name | Origin |
|---|---|---|
| S01 | Basic NAS implementation (API only, files area, two-area layout) | User-defined |
| S02 | NAS GUI | User-defined |
| S03 | Security | User-defined |
| S04 | Media management (photos area) | User-defined |
| S05 | Media metadata (sidecars) | User-defined |
| S06 | Search (both areas) | User-defined |
| S07 | Multi-user and sharing | User-defined |
| S08 | Data protection and recovery | Planner-proposed |
| S09 | Network file access and external change sync | Planner-proposed |
| S10 | Administration, monitoring, and quotas | Planner-proposed |
| S11 | Duplicate and look-alike management (photo duplicates, look-alike stacks and bursts, file duplicates and shortcuts) | User-defined (P005, S007) |
| S12 | Storage optimization (smaller photos and videos, now and on upload) | User-defined (P005) |
| S13 | Dependency security review: every first-party and third-party component checked for known issues, vulnerabilities, and exploits; upgrade, else downgrade, else mitigate _(new in 1.9.0)_ | User-defined (S007) |
| S14 | Packaging, deployment, and pre-AI release (was S11, then S13 until 1.9.0) | Planner-proposed |
| S15 | Drives, pools, and drive lifecycle: new drives, upgrades, mirrors, replacements; RAID 0 and RAID 1 pools (complex RAID deferred, 11a) | User-defined (P005, P007; scope and position set by the user in S007) |
| S16 | SSD caching (optional) | User-defined (P007) |
| S17 | AI features (always last; was S12, then S15 until 1.7.0, then S16 until 1.9.0) | User-defined |

**Beyond the MVP (1.6.0, P006).** The product goal is to make most commercial cloud storage and photo services unnecessary for the people who use local-ai-nas (G13): first on the home network with the MVP, then from anywhere through private remote access, and finally through a hardened public release. Features that competitors have and the plan lacked are planned as **releases R01–R12** after the MVP, one fixed set of features at a time (section 11b; research in `research/R001-2026-09-28-cloud-storage-feature-research.md`). The public internet release (R09) ships only when its security and UI/UX gates pass (section 11c).

**Admin console (1.7.0, the user's requirement).** Everything an administrator does (storage and drives, users, sharing, security, network shares, backups, jobs, logs and alerts, system settings, updates) happens in one **admin console** in the GUI, built from S03 on and completed in S10 (section 6.6, ADR-0039).

**Raspberry Pi first (1.7.0, the user's requirement).** The user will run the NAS on a **Raspberry Pi**, so every stage designs and measures for it (NFR-051, A24): bounded memory, background work sized for four cores, drives on USB 3 or PCIe NVMe, never the database on the SD card, and no reliance on a hardware video encoder (the Pi 5 has none). Until the user has a Pi, CI tests on ARM64 runners and in a Pi-sized resource profile (section 12).

**Foundations and review (1.9.0, P008 and the user's S007 requirements).** Stable item IDs, protection against other websites before login, a durable job and operation foundation with crash consistency, minimal trash, and full database durability are built now, in S03, as follow-ups of S01 and S02. Ownership and access are authoritative in the database (I3); the typed search text may be embedded at query time (I4). The new **S13 reviews every dependency** before packaging (the user's requirement). **Google Takeout import** of Drive and Photos with everything Google recorded is in release R01, after the user's sample data has been analyzed. An internal alpha at M2; the MVP scope is frozen (R14).

---

## 2. Goals and non-goals

### Goals
- **G1:** A reliable NAS with a storage root containing two separate user-data areas, `files/` and `photos/`. It is managed first through a versioned API (S01), then through a GUI (S02).
- **G2:** Security suitable for a home LAN: authentication, HTTPS, hardening, audit trail, localhost-only until secure (S03).
- **G3:** A photo library comparable to the core of Google Photos: timeline, albums, viewer, favorites, video playback with live quality switching, and explicit transfer to and from `files/` (S04).
- **G4:** Portable, human-readable, versioned sidecar metadata per photo that stays correct through every operation (S05).
- **G5:** Fast, forgiving search across both areas, with word forms, synonyms, typo tolerance, and operators (S06).
- **G6:** Multiple users with private data by default and explicit sharing, enforced on every access path (S07).
- **G7:** Protection against data loss (trash, integrity checks, backups, recovery), network-drive access, and admin control (S08–S10).
- **G8:** Easy deployment for anyone, and a stable pre-AI release (S14).
- **G9:** Optional local AI for auto-classification and face grouping on CPU-only hardware, with results persisted to sidecars (S17).
- **G10:** Privacy by design at every stage: no telemetry, no cloud, no runtime network calls unless the user enables a feature that needs one.
- **G11:** _New in 1.4.0 (P005):_ duplicate and look-alike photos, bursts, and duplicate files are found in each user's own library and resolved safely (preview, confirmation, trash), and the library can be shrunk on request with its metadata intact (S11, S12).
- **G12:** _New in 1.4.0 (P005, S007):_ several drives combined into RAID 0 or RAID 1 pools from the GUI, with honest capacity and fault-tolerance figures (S15). Parity layouts and combined drives are planned for later (11a). _Extended in 1.7.0 (P007):_ the whole **drive lifecycle**: a newly installed drive is detected and the admin is guided to upgrade capacity, add a RAID 1 mirror, replace a failing drive, grow a pool, or use it as a backup drive, with the data moved and verified (S15; predictive health and a command-line migration already in the MVP); and optional **SSD caching** of the most-used, large files (S16).
- **G14:** _New in 1.7.0 (the user's requirement, S007):_ an **admin console** in the GUI where the administrator manages everything: storage and drives, users, sharing, security, network shares, backups, jobs, logs and alerts, and system settings (S03.9, S10, 6.6).
- **G15:** _New in 1.7.0 (the user's requirement, S007):_ the NAS runs **well on a Raspberry Pi**, the user's production machine (NFR-051).
- **G13:** _New in 1.6.0 (P006, the user's message):_ local-ai-nas replaces most commercial cloud storage and photo services for its users: first on the home network (MVP), then from anywhere through private remote access (R06), and finally through a hardened public release (R09).

- **G16:** _New in 1.12.0 (P010, the user's requirement):_ **security by design across the whole lifecycle, aligned with ISO/IEC 27001:2022 Annex A and OWASP ASVS (self-assessed; not certified).** The user: "to implement a system of avoiding security vulnerabilities, having secure connections between client (app or webpage) and server, secure file transfers, storing files securely (transferring of files should not allow the attacker to exploit it in any way or execute any kind of scripts on the server or gain unauthorized access to the system), securing the server (to avoid backdoors, scripts execution by attacker or giving unauthorized access to the server to the attacker) and following cybersecurity protocols and especially cybersecurity ISO standards". Split into UR-1 (a secure development lifecycle), UR-2 (secure connections between every client and the server), UR-3 (secure file transfers that cannot exploit the server, run scripts on it, or give unauthorized access), UR-4 (secure file storage), UR-5 (a hardened server: no backdoors, no script execution by an attacker, no unauthorized access), and UR-6 (follow cybersecurity protocols, especially the ISO/IEC standards). The program: rule R15, invariant I12, the documents in `docs/security/`, and NFR-059–NFR-081, FR-377–FR-382 (section 3).

### Non-goals (at least for now)
- **NG1:** Cloud sync, cloud backup, or any hosted/SaaS component. _Change proposed in 1.6.0 (P006), pending Q55:_ "No hosted service operated by the project and no dependency on one; connections to third-party or user-owned remote services only as explicit opt-in features (I6)." R01 imports and R04 off-site backup depend on it.
- **NG2:** _Changed in 1.11.0 (P009, the user's messages):_ native client apps and automatic backup from phones and computers are no longer non-goals. They are planned after the MVP in the existing client releases **R07** (mobile apps: Android, iOS) and **R08** (desktop sync: Windows, macOS, Linux), with a platform matrix for all five (roadmap 11b.3). Until then the MVP bridges phone backup through WebDAV auto-upload apps (FR-219, FR-220). The user: "Add a to-do feature of android app client which will connect to the nas and help in auto backups like auto media backup and device folders sync into drive and ability to allow Android able to "share" files/media to the nas using share option and they will upload onto it." and "Wait, there should also be a windows app client too for the same, so add both windows and android into the plan with future expansion to Mac, ios and Linux" _(First proposed in 1.6.0, P006, as Q56; Q56's first part is answered by these messages.)_
- **NG3:** Photo editing, and writing metadata back into original media files (Q19). _Change proposed in 1.6.0 (P006), pending Q57:_ "no **destructive** editing": R10 adds non-destructive editing that never changes originals.
- **NG4:** _Changed in 0.4.0:_ video transcoding **for streaming quality levels** is now in scope (S04.8, ADR-0020, the user's request in S004). _Changed in 1.4.0 (P005):_ the app never re-encodes or converts original files **on its own**. The only exception is storage optimization (S12), which the user starts or enables explicitly: it replaces originals only after a preview and a confirmation, and keeps them for an undo window (I10).
- **NG5:** _Changed in 1.6.0 (P006, the user's message):_ remote access and public share links are no longer non-goals. They are planned after the MVP as **R06** (private remote access, nothing exposed publicly) and **R09** (the public internet release, gated by strict security and UI/UX requirements, 11c). Until then the NAS stays LAN-only. The user: "my project of nas should make every other (or atleast most of them) cloud storage services useless except that my project is currently limited to local hosting, add a future release of public release too but that is very crucial too due to severe UI/UX reasons and also severe security reasons"
- **NG6:** _Withdrawn in 0.2.0._ The v0.1.0 non-goal "no app-provided SMB/WebDAV" is reversed by stage S09.
- **NG7:** Running AI at search time, and any cloud AI API.
- **NG8:** _Changed in 1.4.0 (P005, the user's decision in S007):_ drive pools are in scope as **RAID 0 and RAID 1** (S15), built by orchestrating mature Linux tools (never RAID written in the app). Parity layouts, combined (virtual) drives, and nesting are planned for later (11a). Snapshots and general volume management remain out of scope. _1.7.0 (P007):_ the drive lifecycle (new drives, upgrades, mirror conversion, replacement, growth, retirement) is in scope in S15, with the same Linux-only rule for anything that needs the storage helper (A22).
- **NG9:** Generative AI features. (OCR and semantic search are optional S17.10 extensions, pending Q36.)
- **NG10:** Any automatic or implicit syncing between `files/` and `photos/` (I1).
- **NG11:** _New in 1.4.0 (P005):_ near-duplicate **documents** (e.g. two versions of a report). File duplicates are exact content matches only (S11.5).

---

## 2b. Cross-cutting principles

Every stage must follow these:

1. **Every stage ends with a working, tested, demonstrable system.** Nothing is left half-built between stages.
2. **Forward compatibility.** Design each stage with later stages in mind so they plug in without rewrites. Examples:
   - The storage layout in S01 allows per-user namespaces for S07.
   - The API has a central authorization hook from S03 that S07 extends.
   - The sidecar schema in S05 reserves sections for ownership and access (S07), AI results (S17), and, since 1.4.0 (P005), hashes, stacks, duplicate decisions, merged-metadata provenance, and optimization history (S11, S12).
   - The search index in S06 reserves fields for owner, access list, AI tags, and face groups, and since 1.4.0 for stack and shortcut flags (S11).
   - Content hashes are stored at upload and perceptual hashes with thumbnails (P005), so duplicate detection (S11) needs no extra reads.
3. **Secure-by-default baseline from S01**, even though the dedicated security stage is S03. The server binds to localhost only by default, path traversal is impossible, and all input is validated. The NAS must not be exposed on the network without authentication.
4. **Shared infrastructure is built once**, in the first stage that needs it, and reused later. Examples: the background job system (S04.3) is reused by S05, S06, S08, S09, S11, S12, and S17. The disk health collector (S10.3) is reused by the pools (S15). The authorization policy check (S03.5) is extended by S07.
5. **S01 is API-only and S02 is the GUI stage.** From S03 onward, any stage that adds user-facing features includes its own GUI substage.
6. **The last substage of every stage** writes and runs the stage's tests (unit, integration, and system/application; S006), then covers documentation updates, the completion record, and user sign-off.
7. **Every stage includes its security gate** _(1.12.0, P010; rule R15)_: the stage document lists its Tier 1 parts; security-relevant tasks record their threat model change, ASVS IDs, secure-coding self-review, and abuse-case tests; the final testing substage runs the gate (testing 12.5) and no stage is Done with a blocking finding open (the user's decision D1).

---
