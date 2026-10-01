# Open questions

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](PLAN_INDEX.md)). Moved unchanged from `plan.md` section 5 in version 1.9.0 (ADR-0044).

## 5. Open questions for the user

Questions keep their numbers permanently. **★ = needed for S01**: none left (Q1, Q22, and Q18 were answered in S005). Answered or superseded questions stay listed for traceability.

**Answered by the user in S005 (approval stage, decisions D-01–D-14 of audit A001):**
- Q1 (platforms: x86-64 mini-PC/old PC, Raspberry Pi, and Windows 11 for testing).
- Q16 (closed).
- Q22 (**AGPL-3.0-or-later**).
- Q37 (no candidates added).
- Q38 (first usable release = **S01–S11** in the IDs of that time, i.e. through the pre-AI release; since 1.4.0 the pre-AI release is **S14**, so M3 is S01–S14: see section 11 and Q51).
- Q18 (library size: **100,000 photos + 100,000 files** per installation, confirming A12; before the S01.7 baseline).

**Answered by P003 (0.3.0):**
- Q25 (GUI): web UI (SvelteKit).
- Q32, first half (WebDAV first; SMB later, Linux-only).
- Q5, partly (Docker Compose primary; native Linux secondary; native Windows/macOS undecided).
- Q7, and Q6 partly (CPU by default, optional GPU).

**Answered by implication of P003:**
- Q4 (languages: Go, SvelteKit/TypeScript, Python for AI).
- Q24 (CI: GitHub Actions; the repository is on GitHub).
- Q16 (face models must be permissively licensed; InsightFace excluded; NFR-029).

**Remaining ★ for S01:** none. ADR-0003 was Accepted in S005 (D-01).

**Needed for plan baseline approval (1.0.0)** (grouped by audit A001, F-016): all items were resolved in S005: Q38, Q37, the 10.17 flags (D-02), ADR-0003 (D-01), and the other A001 decisions. **What remains is the user's explicit approval of this plan as the 1.0.0 baseline** (R4).

### New in 0.2.0

25. _Answered (P003, ADR-0009):_ a web UI served by the NAS, built with SvelteKit (static SPA) and embedded in the Go binary. No desktop app.
26. **Photos area content.** _Partly answered (S004 E005/E008):_ **videos are included**, with streaming and live quality switching (FR-144–FR-148). **Still open:** which image formats must be supported (HEIC, RAW)? _Needed by: S04.1._ _Recommendation: JPEG/PNG/WebP/GIF + HEIC; RAW later._
27. **Photo moved from `photos/` to `files/`.** What happens to its sidecar: delete it, keep it, or keep it hidden? _Needed by: S04.6, S05.6._ _Recommendation: move the sidecar into internal app data, keyed by content hash, so moving the photo back restores its metadata, and `files/` stays clean._
28. **Ownership and access data for the files area.** A visible sidecar per file, a hidden sidecar, or a central store? _Needed by: S07.3._ _Recommendation (per your stated approach): a hidden sidecar only for items that are actually shared, with the owner implied by the user's namespace for everything else. Full trade-offs in section 8.8._ _1.9.0 (D2): whichever storage is chosen, the database is authoritative and any hidden file is only a mirror (FR-359)._
29. **Admin visibility.** Can the admin see all users' files and photos, or only manage accounts? _Needed by: S07.1._ _Recommendation: only manage accounts (privacy by default)._
30. **Sharing scope.** Read-only as specified, or also write access? Should sharing with groups of users be possible? _Needed by: S07.5._
31. **Document content search.** Include full-text search inside PDF, Word, and text files in the files area? _Needed by: S06.2 (design), later stage for implementation._ _(1.6.0, P006: the answer also decides G-134: in S06 if "yes" for the MVP, else in R11.)_
32. **Network shares.** _Partly answered (P003, ADR-0015):_ **WebDAV first**. SMB via Samba later, Linux-only, optional (ADR-0019, Proposed). **Still open:** should the photos area be exposed over network shares, and if so, read-only? _Needed by: S09.3._ _Recommendation: photos read-only over shares._
33. **Two-factor authentication.** Wanted? _Needed by: S03.7._ _(1.6.0, P006: the answer also decides G-060: in S03.7 if "yes", else in R05.)_ _Answered (S007 E050, with the S03 approval): **yes**. S03.7 builds TOTP with recovery codes and the "require for admins" policy; the rest of FR-262 stays in R05._
34. **File versioning.** Wanted? _Needed by: S08.5._ _(1.6.0, P006: **recommended "yes"** (gap G-004): every competitor keeps versions, and from S09 files can be overwritten over WebDAV. R04 (rewind and ransomware recovery) and R11 depend on it.)_
35. **AI opt-in scope.** Per installation or per user? _Needed by: S17.1._
36. **Optional AI extensions.** Which are wanted: OCR for receipts and documents, semantic search? _Needed by: S17.10._ _(1.4.0, P005: duplicate and similar-photo detection is no longer an option here; its baseline is S11 and its AI enhancement S17.11.)_ _1.9.0 (D3): encoding the typed search text at query time is allowed (I4); whether semantic search and OCR are wanted stays open._
37. _Answered (S005, D-08):_ **none added now**. Mobile auto-backup, public share links, and remote access stay unscheduled (11a). Any later addition goes before S17 (I8).
38. _Answered (S005, D-08):_ the first usable release is **S01–S11** (milestone M3 in section 11), in the stage IDs of that time: through the pre-AI release. _Since 1.4.0 (P005)_ the pre-AI release is S14, and the new stages S11 and S12 come before it, so M3 is **S01–S14**; the user is asked to confirm this in Q51.
39. _(Planner-added)_ **Importing an existing collection.** Besides browser upload, should the admin be able to import a folder already on the host into `photos/` or `files/` (server-side copy or move)? _Needed by: S04.2._
40. _(Planner-added)_ **Organization inside `photos/`.** Store media by date taken (`photos/<user>/YYYY/MM/`), by import batch, or in user-created folders? _Needed by: S04.1 (layout ADR)._ _Answered in 1.10.0 (the user, S007 E070: "Hybrid (Recommended)"): **hybrid**: folders the user uploads (including Google Takeout exports) keep their structure; loose photos (single uploads, phone backup) are filed by date taken into `YYYY/MM/` (by upload date when the date taken is unknown). The app's timeline and albums do not depend on it; item IDs (ADR-0040) keep references stable when a photo is refiled. The S04.1 layout ADR records the details._
41. _(Planner-added, 1.1.0)_ **Setup script default mode on Linux** (x86-64 and Raspberry Pi): should the setup script deploy with **Docker Compose** (installs Docker if missing, then starts the stack; ADR-0006 primary) or as a **native service** (binary + systemd + distribution packages)? _Needed by: S14.2._ _Recommendation: offer both; Docker Compose by default on Linux, native service on Windows 11._

### New in 1.4.0 (P005)

42. **Stack cover.** Is the "first photo" of a look-alike stack the one with the **earliest date taken** (proposed), or the earliest upload? _Needed by: S11.4._
43. **Cross-area duplicates.** Should the same photo in both `files/` and `photos/` be reported as a duplicate, or ignored? _Needed by: S11.2._ _Recommendation: report it for information only, never resolve it automatically, since the areas are separate by design (I1)._
44. **Retention of replaced originals** after optimization: how long (proposed: **30 days**)? _Needed by: S12.5._
45. **Exact duplicate photos at upload:** "Skip and report" (proposed), "Keep both", or "Ask"? _Needed by: S04.2 (default), S11.2 (policy)._
46. **File shortcuts over network shares:** shown as a read-only view of the target (proposed in ADR-0024), or hidden? _Needed by: S11.6, S09.2._
47. **Pools:** is mdadm acceptable for RAID 0 and RAID 1 (recommended, ADR-0027), and is **Linux-only** acceptable for pools (A22)? _Needed by: S15.1._ _(SnapRAID with mergerfs and parity layouts are deferred with the complex RAID, 11a; RAID 5 and RAID 6 are answered by the user's decision in S007: later.)_
48. **Video codec for optimization:** H.264 only (most compatible, recommended default), or also H.265/HEVC and AV1 (smaller files; slower encoding and weaker browser support)? _Needed by: S12.2 (ADR-0025)._
49. **Who applies optimization policies:** only each user to their own media (proposed), or can an admin apply policies to all users? _Needed by: S12.6._
50. _Answered (S007, E010): now, with the current stage ("make sure to work on the new stuff too if they were meant to be part of current or previous stages"). ADR-0021 was Accepted in E013, and the three follow-up tasks were done in S007 (E017–E019)._ _(Planner-added, from fixing P005)_ **When are the S01 follow-up tasks done** (content hash at upload in both areas; storage-root checks), now that S01 is Done? _Recommendation:_ at the start of S04, before S04.2 needs the hashes; files uploaded before then are covered by the S11.1 backfill. The content-hash ADR (ADR-0021) must be decided first. _Needed by: S04._
51. _(Planner-added)_ **Confirm the first usable release (M3)** as S01–S14, which now includes duplicates and look-alikes (S11) and storage optimization (S12), with drive pools (S15) after the release (section 11). _Needed by: S14._

### New in 1.6.0 (P006)

_(Planner-added from P006; Q54 and Q34 first, because they change the MVP, then Q52 and Q53.)_

52. AI position: keep 'AI always last' so S17 comes after all releases R01–R12 (default, the current invariant I8), or change I8 so AI comes right after the drive pools and the releases follow AI?
53. Drive pools (S15): keep them right after the MVP (default), or move them after the new releases, since you placed them 'in the end'?
54. Confirm the MVP additions: Live Photos and motion photos, the phone auto-backup bridge, and alert delivery (email, webhook, ntfy). And answer Q34 (file versioning), recommended 'yes'.
55. Reword NG1 so opt-in imports from other clouds and off-site backup to targets you choose are allowed?
56. Mobile apps (reverses NG2): ~~native apps (framework by ADR) or PWA only?~~ _Answered in 1.11.0 (P009, the user's messages 1 and 2): native client apps, Android and Windows first, with iOS, macOS, and Linux covered (roadmap 11b.3); the framework by ADR (8.43)._ Still open: publish on app stores, F-Droid, or both? Allow Apple and Google push services as an opt-in?
57. Reword NG3 to allow non-destructive photo editing (originals never changed)?
58. External read-only libraries (index a folder on the host without copying it): exclude (default, keeps I1) or include?
59. Map tiles: bundled low-detail offline tiles, a self-hosted tile file you download once, an opt-in online tile server, or a combination?
60. Public release accounts: only accounts the admin creates, admin-approved invitations (recommended), or open registration?
61. Public exposure methods to support: port forwarding with reverse proxy, your own VPS relay, both (recommended), and should third-party tunnels be documented?
62. Locked folder: hidden and re-authenticated only (simpler), or also client-side encrypted (no server search or thumbnails inside it)?
63. Release version labels: accept the suggested labels (v1.0.0 MVP, v1.1.0 pools, v1.2.0 R01, … v2.0.0 public release) or choose your own?
64. Office co-editing engine for R11: Collabora Online, ONLYOFFICE, or none?
65. Confirm the excluded features (X-01 to X-11); any to bring back?
66. Approve the proposed invariant I11 (internet exposure is always an explicit, checked choice; admin interface LAN/VPN-only by default)?

### New in 1.7.0 (P007, the admin console, and the Raspberry Pi)

67. _(Planner-added, P007)_ **Confirm the two MVP additions:** predictive drive health (S10.3, FR-332) and the storage migration engine with its command line and console page (S08.4, FR-333)? _Needed by: S08, S10._
68. **Rollback window** for the old drive after a migration: 7 days (proposed), longer, or "until I retire it"? _Needed by: S08.4._
69. **Burn-in of new drives:** off by default (proposed), a quick read test by default, or a full write test by default for blank drives? _Needed by: S15.9._
70. **An optional LVM layer at pool creation**, so a block-level SSD cache can be added later without recreating the pool (ADR-0038; recommended: no, for the first version)? _Needed by: S15.5._
71. **The database on an SSD:** allow it on a single SSD (a recent metadata backup required), or require a mirrored SSD pair? _Needed by: S16.2._
72. **SSD cache defaults:** minimum file size and access threshold (proposed: 8 MB and 3 reads in 7 days)? _Needed by: S16.4._
73. **Hard-drive spin-down:** include it as an option (Could), or leave it out? _Needed by: S16.6._
74. _(the user's requirement, S007)_ **Admin console form:** an admin section of the same web app at `/admin` (recommended, ADR-0039), or a separate app on its own port? _Needed by: S03.9._ _Answered (S007 E050, with the S03 approval): the same app at `/admin`; ADR-0039 Accepted._
75. _(the user's requirement, S007)_ **Which Raspberry Pi:** Raspberry Pi 5 with 4 GB, 8 GB, or 16 GB of RAM, and how the drives connect (USB 3 enclosures, or an NVMe HAT on PCIe)? The answer sets the memory and speed budgets (NFR-051). _Needed by: S03 (budgets), S14.5 (measurements)._
76. _(the user's requirement, S007)_ **Deployer form:** guided setup scripts for all four platforms first (recommended, simplest to keep in step), or native packages (a `.deb` repository, an Arch package, a Windows installer) as well, later? _Needed by: S14.2._
77. _(external review #1, CR001)_ **Stable item IDs:** give every item an app-assigned, immutable ID (UUIDv7 or ULID) kept in the database and, for photos, in the sidecar (`mediaId`), and let albums, shares, jobs, faces, shortcuts, and audit targets refer to IDs, with paths only locating items (**recommended**; 8.35)? In the files area an ID would be assigned when something first refers to the item, and re-associated after external moves by file ID and content hash. _Needed by: the S04 stage document (before photos are stored); affects ADR-0024, S07, S11, ADR-0030._ _Answered (S007 E059): adopted through P008 F1 with the user's approval: UUIDv7 in an items table (FR-346, ADR-0040)._
78. _(external review #1, CR001)_ **Authority for access data:** make the database the authority for owners, ACLs, shares, public links, and revocations, with the sidecar `access` section only a mirror for export and inspection (**recommended**; 8.36)? This changes 8.8, where photo access data lives in the sidecar as the user specified, and would drop the hidden per-file access sidecars recommended there. _Needed by: S05.1 (the schema reserves `access`), S07.3._ _Answered: yes (P008 D2, "Database, sidecar mirror (Recommended)")._
79. _(external review #1, CR001)_ **Database durability:** `synchronous=FULL` for the whole database from S03 (**recommended**: one flush per commit, negligible for this workload), `FULL` only for security writes, or keep `NORMAL` (8.37)? It changes the connection setup of ADR-0007 (Accepted). _Needed by: S03.3-T01 (sessions)._ _Answered: `synchronous=FULL` for the whole database, high-frequency writes throttled (P008 D5, NFR-056)._
80. _(external review #1, CR001)_ **Process weight:** keep the current process, or lighten it: at startup read `CURRENT_STATE.md`, the active stage document, and `RULES.md` (plan sections only when needed); session logs quote the user verbatim only for approvals and decisions and summarize the rest, since the repository is public? _Needed by: now (RULES R1, R2)._ _Answered: keep the current process (P008 D4, "Keep current process")._
81. _(external review #1, CR001)_ **Plan structure:** keep one `plan.md`, add a short plan index (stage status, dependencies, milestone, links) that sessions read instead of the whole plan (**recommended**, cheapest), or split the plan into several documents (architecture, requirements, roadmap, invariants, security)? _Needed by: now (RULES R1)._ _Answered: the plan is split into `code-agent-docs/plan/` with `PLAN_INDEX.md` (P008 P-A, the user's request); sessions still read the whole plan (D4)._
82. _(P008 D6)_ **Change intake:** freeze the MVP scope after P008, with new ideas going to a release backlog unless they are security or data-integrity fixes, foundations that are cheap only now, or the user puts them in the MVP? _Answered: yes ("Yes, freeze MVP (Recommended)"): rule R14, section 11d._
83. _(P008 D7)_ **Internal alpha:** an early build for the user's own daily use at M2 (S01–S06)? _Answered: yes ("Yes, internal alpha at M2 (Recommended)"); the first release stays M3 (section 11)._

### New in 1.11.0 (P009, client apps)

84. _(P009)_ **Two-way folder sync on Android:** include it in R07, or ship Backup, Mirror, and Download first and add two-way later (**recommended**: two-way sync on a phone has the most edge cases, and Backup covers the request)? _Needed by: the R07 stage document (FR-368)._
85. _(P009)_ **Minimum supported versions:** Android 8.0 or newer (**recommended**)? Windows 10 22H2 and Windows 11, or Windows 11 only (Windows 10 is past its end of support)? _Needed by: the client technology ADR (8.43) and the R07 and R08 stage documents._
86. _(P009)_ **Windows distribution:** a signed installer (a code-signing certificate costs money every year), an unsigned installer from GitHub (Windows warns on first run), and/or winget? The Windows 11 right-click menu and the Share window need package identity, which needs a signed package, so an unsigned installer leaves Send To, drag-and-drop, and the classic menu (11b.3). _Needed by: R08 (FR-374)._

### Carried over from 0.1.0

1. _Answered (S005, D-13):_ **multi-platform**. The NAS runs on x86-64 mini-PCs or old PCs and on Raspberry Pi (ARM64), and **this Windows 11 PC is used for testing**. User's words: "mini pc/old pc/raspberry pi/also this windows 11 pc (this one for testing) so multi platform compatibility". Performance targets are measured on the Windows 11 development PC and, when available, on a Raspberry Pi and an x86-64 mini-PC (exact models are recorded when benchmarking in S01.7). _Extended in 1.7.0 (S007 E046):_ the four focus platforms for deployment are Debian, Arch Linux, Windows 11, and Raspberry Pi OS; the Raspberry Pi is the user's production machine (NFR-051, A24, A25).
2. _Resolved by P002:_ single admin account from S03; multi-user in S07.
3. _Superseded by Q37_ (remote access is a not-scheduled candidate).
4. _Answered by P003:_ Go for the core (ADR-0001), REST + OpenAPI (ADR-0002), SvelteKit + TypeScript for the UI (ADR-0009), Python for the AI worker only (ADR-0017).
5. **Deployment method.** _Partly answered (P003, ADR-0006):_ Docker Compose primary (linux/amd64 + arm64); native Linux secondary (binary + systemd). _Informed by Q1 (S005):_ the server must run natively on Windows 11 for testing (already required by CI and S01.1). _Partly answered by the user (S005 E015):_ **a separate setup script for each platform** that deploys the NAS automatically (FR-149): Linux x86-64, Raspberry Pi, and Windows 11. **Still open:** macOS; and the default mode of the Linux scripts (Q41). _Needed by: S14.2._
6. **AI hardware and speed expectations.** _Partly answered (P003):_ CPU by default, optional GPU. **Still open:** what minimum machine and processing speed are acceptable (e.g. "backfill 50,000 photos overnight")? _Needed by: S17.1._
7. _Answered (P003, ADR-0017):_ CPU by default. Optional GPU acceleration through ONNX Runtime execution providers (e.g. CUDA, OpenVINO); which ones are supported is evaluated in S17.1.
8. _Superseded by Q26._
9. _Superseded by Q39_ (the "in place" library model was replaced by the two-area layout).
10. **External changes**: policy for files changed outside the app (re-associate by hash; orphaned sidecars quarantined, never deleted silently)? _Needed by: S05.7, S09.4._
11. **Sidecar naming vs. other tools** (e.g. Google Takeout also writes `<name>.json`). The risk is much lower now, because `photos/` accepts only media through the app. It remains for network shares (S09.3) and server-side import (Q39). Keep README naming plus an identifying marker (**recommended**)? Import Takeout metadata? _Needed by: S05.1._ _Answered in 1.9.0 (P008 D1, the user): `.lainas.json` (FR-356)._
12. _Resolved by P002:_ the files area has no photo sidecars and is searchable by name and file metadata in S06. Content search is Q31.
13. **Albums and face-group storage.** Under I2 they cannot live inside `photos/`. Options: (a) the internal database, backed up by S08.3; (b) JSON documents in internal app data, easy to back up and export (**recommended**). Also: keep `groupName` in each sidecar (README draft) or only a `groupId`? _Needed by: S04.5, S17.5._
14. **Date operator semantics.** Does `after:2025` mean "from 2026" (**recommended**) or include 2025? Compare on the photo's local capture time (**recommended**)? _Needed by: S06.3._
15. **Languages** for search, synonyms, and taxonomy: English only, or Urdu too? _Needed by: S06.5, S17.3._
16. _Closed (confirmed by the user in S005, D-11). Answered by implication of P003 (NFR-029, ADR-0018):_ only models whose licenses let anyone deploy and use the project. **InsightFace pretrained weights are excluded.** YuNet (MIT) + SFace (Apache-2.0) are chosen.
17. _Superseded by Q38._
18. ★ **Library size**: current and expected number of photos, files, and GB? _Needed by: S01.7 baseline, S04.9, S06.8 targets._
19. **Writing into originals / XMP export**: never write originals (**recommended**); XMP export later? _Needed by: S05._
20. _Superseded by Q32._
21. _Superseded by Q37._
22. _Answered (S005, D-12):_ **AGPL-3.0** (GNU Affero General Public License v3.0), in the **"or later"** form: SPDX `AGPL-3.0-or-later` (the user's answer in S005 E020: "AGPL-3.0-or-later (Recommended)"). The LICENSE file and the policy `docs/dev/licensing.md` were added in S01.1-T02.
    - Relevant facts from the dependency register:
      - Every linked Go library is MIT, BSD, or Apache-2.0.
      - External tools run as separate programs: ExifTool (Artistic/GPL), libvips (LGPL-2.1), libheif (LGPL), FFmpeg (LGPL, or GPL depending on the build).
      - golangci-lint (GPL-3.0) is a development tool only.
23. _Answered in S001:_ commit after each task; feature branch per stage/task off `develop`, pushed, PR into `develop` (RULES.md User Preferences).
24. _Answered by implication of P003 (ADR-0005):_ GitHub Actions. The repository is hosted on GitHub.

---
