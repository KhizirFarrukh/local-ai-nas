# S14: Packaging, deployment, and pre-AI release

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](../PLAN_INDEX.md)). Moved unchanged from `plan.md` section 10.14 in version 1.9.0 (ADR-0044).

### 10.14 S14: Packaging, deployment, and pre-AI release

- **Origin:** Planner-proposed
- **Stage ID:** S14 since 1.4.0 (was S11; see 10.18).
- **Reason added (quoted from P002):** "The project's goal is that anyone can deploy it. Hardening and packaging here produce a stable release before AI is added, so the AI stage builds on a proven base."
- **Goal:** A hardened, packaged, documented, stable release without AI.
- **User requirements:** none (planner-proposed).
- **Status:** Not started

#### S14.1: Container packaging
- **Goal:** One-command deployment.
- **Scope:** Docker image and Docker Compose setup.
- **Deliverables:** multi-arch images (amd64, arm64); Compose file; volume layout documentation.
- **Depends on:** S08–S12 (Done).
- **Requirements:** FR-131, NFR-008, NFR-009, NFR-030.
- **Acceptance criteria:**
  1. `docker compose up` starts a working NAS from the published images on amd64 and arm64.
  2. All data lives in mounted volumes, and recreating the containers loses nothing.
- **Risks/notes:** The image registry choice is the user's.
- **P008 (1.9.0):** one compose file, no database or queue containers (both live in the core); health checks, restart policies, a non-root user, a read-only root filesystem where possible, documented volumes (storage root, internal data, configuration), resource limits for the AI container; AI images per accelerator as compose profiles (CPU, NVIDIA, Intel, AMD), each with its ONNX Runtime build (FR-361).
- **Status:** Not started

#### S14.2: Native installation
- **Goal:** Deploy the NAS on each supported platform by running one setup script (user requirement, S005 E015).
- **Scope:**
  - A **separate deployer for each focus platform** (FR-149, the user's requirement in S007): **Debian** and Ubuntu (x86-64 and ARM64), **Arch Linux**, **Raspberry Pi OS** (64-bit, ARM64), and **Windows 11** (Q1, Q5). Others through Docker Compose; macOS only if Q5 adds it.
  - **User-friendly by requirement (NFR-052):** one command (Linux) or a double-click (Windows); a guided flow with defaults (storage location, port, Docker or native per Q41); prerequisite checks with plain explanations and consent before installing anything; progress and a log; a summary at the end with the address to open; re-running repairs or upgrades; uninstall keeps data. The **form** of each deployer (a guided script, or a native package such as a `.deb`, an Arch package, or a Windows installer) is decided in the S14 stage document (Q76).
  - **Raspberry Pi checks:** 64-bit OS, memory, the NAS drive on USB 3 or NVMe and never the SD card for data or the database, a power-supply warning (8.33).
  - Each script checks the platform, installs or verifies every prerequisite listed for it in `dependencies.md` section 12 (NFR-032), installs the NAS, creates the configuration and the storage root, registers and starts the system service (systemd; Windows service), and finishes with a health check.
  - The default mode of the Linux scripts (Docker Compose or native service) is decided by Q41.
- **Deliverables:** one deployer per focus platform (e.g. `deploy/setup/setup-debian.sh`, `deploy/setup/setup-arch.sh`, `deploy/setup/setup-raspberry-pi.sh`, `deploy/setup/setup-windows.ps1` with a double-click launcher; names and forms fixed in the S14 stage document); service definitions; a matching uninstall path; a CI check that each script's prerequisite list matches `dependencies.md` section 12.
- **Depends on:** S14.1.
- **Requirements:** FR-132, FR-149, NFR-009, NFR-032, NFR-052.
- **Acceptance criteria:**
  1. On a clean machine of each focus platform (Debian, Arch Linux, Windows 11, Raspberry Pi OS), running only its deployer gives a running NAS that passes the health check, runs as a service, and survives a reboot; a usability test with someone who is not technical succeeds (NFR-052).
  2. The script installs or verifies every prerequisite listed for its platform in `dependencies.md` section 12, and stops with a clear message when one cannot be installed.
  3. Running the script again is safe (idempotent) and never touches user data.
  4. Uninstalling leaves user data untouched.
- **Risks/notes:** Package sources and names differ per platform (ExifTool, libvips, FFmpeg); the Windows sources are verified in S14.2. Scripts run with administrator rights, so they are reviewed in S14.6.
- **P005 change (1.4.0):** the Linux setup scripts accept optional components; the storage helper is added as one in S15.2, with the pool documentation in S15.13.
- **Status:** Not started

#### S14.3: Updates and migrations
- **Goal:** Safe upgrades.
- **Scope:** an update mechanism with automatic data and schema migrations and a backup before every update.
- **Deliverables:** upgrade procedure; pre-update backup; migration runner; rollback.
- **Depends on:** S08.3, S05.7.
- **Requirements:** FR-133, NFR-017.
- **Acceptance criteria:**
  1. An update runs migrations automatically after a backup.
  2. A failed migration rolls back to the backup.
  3. The upgrade path from the previous release is tested.
- **Risks/notes:** There is no automatic update download (I6). Updates are user-initiated.
- **Status:** Not started

#### S14.4: First-run polish and documentation
- **Goal:** A new user can succeed without help.
- **Scope:** polished first-run experience; install guide (built around the S14.2 setup scripts), admin guide, user guide, hardware requirements.
- **Deliverables:** guides; improved first-run wizard.
- **Depends on:** S14.1.
- **Requirements:** FR-134.
- **Acceptance criteria:**
  1. A new user goes from install to first upload using only the docs.
  2. Install, admin, and user guides and the hardware requirements are published.
- **Risks/notes:** None.
- **P007 addition (1.7.0):** guides "Moving to a new or bigger drive" and "My drive is failing" for each platform, built around the migration engine (S08.4), and a Raspberry Pi guide (drives on USB 3 or NVMe, never the database on the SD card).
- **Status:** Not started

#### S14.5: Full-system performance testing
- **Goal:** Confirm performance at target scale.
- **Scope:** load and performance tests at target library sizes and user counts.
- **Deliverables:** load test suite; performance report on reference hardware.
- **Depends on:** S14.1.
- **Requirements:** NFR-003, NFR-004.
- **Acceptance criteria:**
  1. At the target library sizes and user counts (Q1, Q18), NFR-003 targets are met on reference hardware.
  2. The results are recorded in the stage document.
- **Risks/notes:** None.
- **1.7.0 (the user's requirement):** the reference measurements are made on the Raspberry Pi first (NFR-051, A24).
- **Status:** Not started

#### S14.6: Final security review
- **Goal:** No known high-risk issues ship.
- **Scope:** re-review of the threat model and a full dependency audit. _(1.9.0: the audit re-runs the S13 procedure for everything that changed since S13.)_
- **Deliverables:** updated threat model; audit report (vulnerabilities and licenses).
- **Depends on:** S13 (1.9.0), S14.1–S14.5.
- **Requirements:** FR-084, NFR-013, NFR-023, NFR-029.
- **Acceptance criteria:**
  1. The threat model is re-reviewed, including the S07–S12 additions. (The storage helper is reviewed in S15.2 and S15.13.)
  2. The dependency audit has no unresolved high-severity vulnerability or license finding.
- **Risks/notes:** None.
- **Status:** Not started

#### S14.7: Release
- **Goal:** Publish the first stable release without AI, and close the stage.
- **Scope:**
  - Release process (versioning, changelog, tagged release) and the first stable release without AI. The version label is the user's decision.
  - _Added scope (planner):_ the stage review. Final integration test run, documentation updates, completion record, and user sign-off, because every stage must end with testing and review (section 2b).
- **Deliverables:** release process document; changelog; tagged release; completion record.
- **Depends on:** S14.1–S14.6.
- **Requirements:** FR-135.
- **Acceptance criteria:**
  1. The release process is documented and followed.
  2. The first stable pre-AI release is tagged with the user-chosen version label.
  3. A final integration test run passes on the release artifacts.
  4. The completion record is written and the user's sign-off is recorded.
- **Risks/notes:** Flagged in 10.17: scope extended to include the stage review.
- **Status:** Not started

**Design notes (S14):** The Compose file reserves an optional `ai` profile for S17.12. Update and migration machinery reuses S05.7 and S08.3.

**Exit criteria (planner-proposed):** Anyone can install the NAS with the documentation, upgrade it safely, and it meets performance and security targets. A stable pre-AI release is tagged.

---
