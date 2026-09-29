# A003: README change proposal

| Field | Value |
|---|---|
| Audit | A003 (R12, the S02 final review, S02.8-T04), session S007, 2026-09-28 |
| Finding | F-004 (the README does not show the P005 features and stages, or stage 2) |
| Status | **Proposed**: for the user's decision at the S02 sign-off (S02.8-T05) |

> **Applied in P008 phase C (2026-09-30, pre-approved by the user's request, P-C):** R-13 (replaced by a "Current status" section), R-17 (the "How It Fits Together" diagram), and the part of R-14 that both options share (Stage 2 ticked, Stage 3 marked in progress). Still for the user: the rest of R-14 (option A or B), R-15, R-16.


**Basis rule (as in A001 and A002):**
- Only approved decisions and completed, verified work are used.
- The README is the source of the vision (plan, header).
- Tasks change only the sections their stage document names. In S02 that is the Development section only (S02.1-T05, S02.8-T03).
- Every other README change is proposed here and applied only with the user's approval.

IDs continue from A002 (R-01 to R-12).

---

### R-13: Status line

**Current (line 7):**
> ⚠️ **Status: early development.** Stage 1, the NAS core, is complete: the Files area can be managed through a REST API on the same computer (see Development). Everything else below is the planned scope and is not implemented yet.

**Proposed:**
> ⚠️ **Status: early development.** Stages 1 and 2 are complete: the Files area can be managed on the same computer through a web interface or a REST API (see Development). Everything else below is the planned scope and is not implemented yet.

**Reason:**
- The S02 exit criterion (plan 10.3): "A non-technical user can do everything from S01 through the GUI."
- The user's walkthrough in S02.8-T05 checks it.
- "On the same computer" stays: until login and HTTPS arrive (stage 3), the server answers only on this computer.

**When:** after the user signs off S02, because "complete" depends on the sign-off.

---

### R-14: Roadmap

**Current (Roadmap section):**
```
- [x] Stage 1: Basic NAS (storage service and API; Files and Photos areas)
- [ ] Stage 2: Web interface for the NAS
- [ ] Stage 3: Security (login, HTTPS, hardening)
- [ ] Stage 4: Media management (Photos area: timeline, albums, viewer)
- [ ] Stage 5: Media metadata (EXIF, sidecar JSON, offline place names)
- [ ] Stage 6: Search across Files and Photos (fuzzy, synonyms, operators)
- [ ] Stage 7: Multiple users and sharing
- [ ] Final stage: optional local AI (auto-classification, face grouping)
```

**Proposed (option A, recommended):**
```
- [x] Stage 1: Basic NAS (storage service and API; Files and Photos areas)
- [x] Stage 2: Web interface for the NAS
- [ ] Stage 3: Security (login, HTTPS, hardening)
- [ ] Stage 4: Media management (Photos area: timeline, albums, viewer)
- [ ] Stage 5: Media metadata (EXIF, sidecar JSON, offline place names)
- [ ] Stage 6: Search across Files and Photos (fuzzy, synonyms, operators)
- [ ] Stage 7: Multiple users and sharing
- [ ] Stages 8–10: Trash and backups, network drives, administration and monitoring
- [ ] Stage 11: Duplicate and look-alike photos (including bursts) and duplicate files
- [ ] Stage 12: Storage optimization (smaller photos and videos, on request)
- [ ] Stage 13: Dependency security review (every library checked; upgrade, else downgrade)
- [ ] Stage 14: Packaging and the first release (without AI)
- [ ] Stage 15: Drives: new drives, upgrades, mirrors, and storage pools (RAID 0 and RAID 1)
- [ ] Stage 16: SSD caching (optional)
- [ ] Final stage: optional local AI (auto-classification, face grouping)
```

**Option B:** only tick stage 2, and add stages 11, 12, and 14 (the stages the user asked for in P005), without stages 8–10 and 13.

**Reason:**
- Stage 2 is done once the user signs it off (S02.8-T05).
- The user asked for stages 11, 12, and 14 in P005 (plan 1.4.0; RAID 0 and RAID 1 only, the user's decision in S007 E008).
- A001 (R-08) left out stages 8–11 of that time because they were planner proposals that had not been approved yet. The user approved the whole plan as the baseline in S005 (plan 1.0.0), so they are approved now. Option A shows them; option B keeps the README to the stages the user named.
- The AI stage stays last (invariant I8).

**When:** together with R-13, after the S02 sign-off.

---

### R-15: Features

**Current:** the Features section has NAS Core, Photo Management, Optional Local AI, and Smart Search. It has no line for the P005 features.

**Proposed:** a new subsection after "Photo Management":
```
### 🧹 Duplicates and Storage
- Finds duplicate photos, including the same photo at a different resolution, and lets you choose which one to keep
- Groups look-alike photos and bursts into stacks: the grid shows one cover photo you choose, and you can keep one and delete the rest
- Finds duplicate files in the Files area; keep both, or keep one and optionally leave a shortcut in the other's place
- Shrinks existing photos and videos on request (a fixed size or a percentage, with a live preview), and optionally future uploads too
- Combines several drives into one storage pool (RAID 0 or RAID 1)
- Nothing is removed without a preview, a confirmation, and a way to undo it
```

**Reason:**
- These are the user requirements of P005 (`prompts/P005-feature-additions.json`, `feature_specifications`), in plan 1.4.0 as stages S11, S12, and S14.
- The last line is invariant I10 (P005, pre-approved).
- Parity layouts and combining smaller drives into a larger one are deferred with the complex RAID (plan 11a, the user's decision in S007 E008), so the pool line names RAID 0 and RAID 1 only.
- The status line already says that everything beyond the finished stages is "the planned scope and is not implemented yet", so the new lines do not claim to be built.

**When:** any time. It does not depend on the S02 sign-off.

---

**Recommendation:** approve R-13, R-14 option A, and R-15, and apply them right after the S02 sign-off.

---

### R-16: Deployment options and the Raspberry Pi _(added after A003, S007 E047)_

**Current (Installation section):**
```
Planned deployment options:
- [ ] Docker Compose (primary), with images for x86-64 and ARM64 (e.g. Raspberry Pi); optional AI component enabled with a Compose profile
- [ ] Native install on Linux (single binary + systemd)
- [ ] Native install on Windows and macOS (under consideration)
```

**Proposed:**
```
Planned deployment options, each with its own easy deployer (one command or a double-click):
- [ ] Debian and Ubuntu
- [ ] Arch Linux
- [ ] Raspberry Pi (Raspberry Pi OS 64-bit), the main target: the NAS is tuned to run well on a Pi
- [ ] Windows 11
- [ ] Docker Compose for any other system, with images for x86-64 and ARM64; optional AI with a Compose profile
```

**Reason:** the user's requirements in S007 (E046): "this project should be very optimized on a raspberry pi, i will run it on a pi" and "deployable on debian, arch, windows, pi and more (but these 4 mentioned must be focused) … a different deployment script … very user friendly" (plan FR-149, NFR-051, NFR-052).

**When:** any time; asked with R-13–R-15.

---

### R-17: Architecture at a glance _(added after A003, from external review #1, S007 E057)_

**Current:** the README has no architecture overview.

**Proposed:** a short section after "Features":
```
## 🧭 How it fits together

Browser (the web app) ──▶ local-ai-nas: one program on your NAS
                           ├─ Files area  → files/
                           ├─ Photos area → photos/ + metadata sidecars (stages 4–5)
                           ├─ Search index, rebuildable (stage 6)
                           ├─ Background jobs (stage 4)
                           └─ Database: accounts, sessions, settings, jobs
                                  ▲
                      Optional AI helper (local only, last stage)
```

**Reason:** external review #1 (`code-agent-docs/prompts/CR001-improvement-review-1.md`) suggests a small architecture diagram; the plan's diagrams (section 6) are too detailed for the README. The diagram shows the planned parts with their stages, so it does not claim they are built.

**When:** any time; asked with R-13–R-16.
