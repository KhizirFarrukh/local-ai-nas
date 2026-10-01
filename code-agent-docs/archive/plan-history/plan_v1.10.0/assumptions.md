# Assumptions

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](PLAN_INDEX.md)). Moved unchanged from `plan.md` section 4 in version 1.9.0 (ADR-0044).

## 4. Assumptions

Assumptions are numbered permanently. Ones overturned by the 0.2.0 design are marked superseded.

- **A1:** One household on a trusted LAN. S01–S06 have a single owner (the admin from S03). Multiple users arrive in S07.
- **A2:** The GUI is a browser-based web UI served by the NAS: SvelteKit, embedded in the Go binary (Q25 answered by P003; ADR-0009). No native mobile apps (NG2).
- **A3:** _Superseded in 0.2.0._ Was: the app works in place on existing folders. Now: one storage root with `files/` and `photos/` (FR-069). Existing collections are brought in by upload, transfer, or server-side import (Q39).
- **A4:** The app never modifies original media except through explicit user operations. Metadata edits go to sidecars (Q19).
- **A5:** The app has write access to the whole storage root (sidecars, uploads, trash).
- **A6:** Derived data (thumbnails, index, job queue, embeddings, logs, models) is a rebuildable cache in internal app data (I2). The exceptions are users, sessions, sharing data, and settings, which are backed up (S08.3).
- **A7:** _Superseded in 0.2.0._ Was: library-level metadata in `.local-ai-nas/` inside library roots. That would violate I1/I2. Albums and face-group registries now live in internal app data (Q13).
- **A8:** Photo sidecars exist only for media in `photos/`. Items in `files/` have no photo sidecars. How owner and access data is stored for files is decided in S07.3 (Q28).
- **A9:** Capture times are stored in ISO 8601 with the original UTC offset when known. Otherwise a configured default timezone applies and the sidecar marks it as assumed.
- **A10:** English for the UI, the taxonomy, and the synonym dictionary in the first releases (Q15).
- **A11:** AI model weights are either bundled in the optional AI image or downloaded once at opt-in with explicit consent, and are checksum-verified (FR-032, ADR-0017). They are never fetched at runtime otherwise.
- **A12:** Performance planning targets 100,000 photos + 100,000 files per installation. **Confirmed by the user (Q18, S005).**
- **A13:** The primary deployment target is containers on Linux (x86-64, ARM64). Native installs come in S14.2 (Q5). Windows 11 is a supported development and test platform, and the server must run natively there (Q1, S005).
- **A14:** The README sidecar draft is a starting point. The schema is finalized by ADR in S05.1.
- **A15:** The project lives on GitHub (`origin`: `KhizirFarrukh/local-ai-nas`), with CI on GitHub Actions (Q24).
- **A16:** Development happens on Windows 11, so all tooling must work on Windows and on Linux CI.
- **A17:** Until S03 is Done, the NAS is used only on the machine it runs on (localhost, no authentication).
- **A18:** The storage root, including internal temp uploads and trash, sits on a single filesystem, so atomic renames work between them. The startup health check verifies this. _Changed in 1.4.0 (P005):_ several disks can be pooled by the host OS, or, on Linux, by the NAS itself as RAID 0 or RAID 1 (S15); a pool is still one filesystem, so this assumption holds. _1.7.0 (P007):_ a migration moves the whole root together; the SSD cache and fast internal data (S16) are rebuildable or backed-up internal data, not part of this rule.
- **A19:** Internal app data defaults to `<storage root>/.local-ai-nas/`, with an optional separate location for the database, index, and caches (ADR-0003, Accepted in S005). The configuration file lives outside the storage root, because it is what tells the app where the root is.
- **A20:** Within a stage, the task execution order may differ from substage numbering when dependencies require it. The stage document records the order.
- **A21:** The target browsers play HLS natively or through Media Source Extensions / ManagedMediaSource (hls.js). Where only native HLS is available, the quality menu offers Auto only (ADR-0020).
- **A22:** _(1.4.0, P005)_ Drive pools are **Linux only** (they need root and Linux tools). On Windows 11 and in any setup without the storage helper, the pool feature is hidden, and single-path storage works as before (Q47). _Refined in 1.7.0 (P007), not reversed:_ pools, mirror conversion, drive preparation, and secure erase stay Linux-only; predictive health (where SMART is readable) and migration to a drive the admin prepared work on every platform (platform matrix in 10.15).
- **A23:** _(1.4.0, P005; decided in S007, ADR-0021 Accepted)_ Content hashes use SHA-256 from the Go standard library. Perceptual hashes are 64-bit (ADR-0022).
- **A24:** _(1.7.0, the user's requirement)_ The reference production machine is a **Raspberry Pi 5** (64-bit Raspberry Pi OS; RAM per Q75) with the NAS drives on USB 3 or PCIe NVMe and the OS on the SD card or its own drive. It has four Cortex-A76 cores and **no hardware video encoder** (H.264 is decoded in software, HEVC in hardware). Performance targets (NFR-003) are set for it first.
- **A25:** _(1.7.0, the user's requirement)_ The four focus platforms for deployment are **Debian** (and Ubuntu), **Arch Linux**, **Windows 11**, and **Raspberry Pi OS**; others are served by Docker Compose and later deployers (FR-149).
- **A26:** _(1.9.0, P008)_ Items are identified by stable IDs, not by their paths (FR-346, ADR-0040).
- **A27:** _(1.9.0, the user's decision D2)_ The database is authoritative for ownership, access, and sharing; sidecars and hidden files only mirror it (FR-359).
- **A28:** _(1.9.0, the user's instruction)_ The user will provide **Google Takeout sample data** (Drive and Photos). Before any Takeout import work, the agent analyzes it, records the format facts it needs in a research record without personal data, keeps the sample itself out of git (in the git-ignored `dev/` folder), builds synthetic fixtures from the findings, and only then writes the R01 stage document (FR-222, 8.42).

---
