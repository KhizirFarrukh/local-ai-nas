# S13: Dependency security review (new in 1.9.0, the user's requirement)

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](../PLAN_INDEX.md)). Moved unchanged from `plan.md` section 10.13a in version 1.9.0 (ADR-0044).

### 10.13a S13: Dependency security review (new in 1.9.0, the user's requirement)

- **Origin:** User-defined (S007)
- **Goal:** Every first-party and third-party component the NAS is built from or runs with is checked for known issues, vulnerabilities, and exploits, and moved to a safe version before the first release is packaged; the check then becomes a repeatable procedure.
- **User requirement (quoted, S007):**
  > "Also add a stage that alone goes over every first party and third party library used and sees if it has any known issues, vulnerabilities or exploits or anything similar and then sees if there is a fix in newer update then does that else if a previous version of it didn't have the issue and it's usable then it downgrades to that"
- **Position:** before packaging (S14), so the MVP (M3) ships reviewed: the user's choice ("Before packaging (Recommended)", S007). AI stays last (I8).
- **Status:** Not started
- **Security (1.12.0, P010; rule R15):** the continuous dependency policy (NFR-079) between reviews; the SBOM in SPDX (ISO/IEC 5962) or CycloneDX (FR-382); OpenChain ISO/IEC 5230 and 18974:2023 alignment; the stage is **Tier 1**.

#### S13.1: Inventory of every component (software bill of materials)
- **Goal:** A complete, machine-readable list of everything the NAS is built from and runs with.
- **Scope:** first-party components (the Go toolchain and standard library, the `golang.org/x` modules, Node.js and the web build tools, the project's own packages) and third-party components (Go modules, direct and indirect; npm packages, shipped and development; the system tools and libraries of every platform in register section 12, such as ExifTool, libvips, libheif, FFmpeg, smartmontools, mdadm; container base images; CI actions; bundled data and assets such as GeoNames, icons, pdf.js); exact versions as built; an SBOM in CycloneDX or SPDX generated from the build (the generator and its license chosen and verified in the stage document); reconciliation with the dependency register.
- **Deliverables:** an SBOM per release artifact (binary per platform, container image, web bundle); the register reconciled.
- **Depends on:** S08–S12 (every MVP feature built).
- **Requirements:** FR-363, NFR-023, NFR-032.
- **Acceptance criteria:**
  1. The SBOM lists every component of each artifact with its exact version, and the register lists every SBOM component and nothing unused.
  2. First-party and third-party components are marked; the system tools of every focus platform are included.
- **Risks/notes:** Tools inside container images (for example the FFmpeg build) come from the base image's distribution and follow its versions.
- **Status:** Not started

#### S13.2: Known vulnerabilities and exploits
- **Goal:** Know every published vulnerability and exploit that affects a component version in use.
- **Scope:** every SBOM component checked against OSV (Go, npm, PyPI, Debian, and other ecosystems), the Go vulnerability database with call reachability (govulncheck), GitHub security advisories, the NVD for native tools, the security trackers of the distributions (Debian, Arch Linux, Raspberry Pi OS), image scanning (Trivy), and the list of known exploited vulnerabilities (CISA KEV) to rank urgency; each finding records its severity, whether an exploit is known, whether the NAS is affected (reachable code, feature in use), and the fixed versions.
- **Deliverables:** a vulnerability report per component: affected or not, why, and the fixing versions.
- **Depends on:** S13.1.
- **Requirements:** FR-363, NFR-023.
- **Acceptance criteria:**
  1. Every component was checked in every listed source that covers its ecosystem, with the date of the check.
  2. Every finding states its severity, exploit status, whether the NAS is affected, and the versions that fix it.
- **Risks/notes:** Databases disagree on severity; the highest credible rating is used, with the source named.
- **Status:** Not started

#### S13.3: Known issues beyond vulnerabilities
- **Goal:** Catch problems that are not published vulnerabilities but can hurt users: data loss, corruption, crashes, abandoned projects.
- **Scope:** the issue trackers and changelogs of the versions in use (data-loss, corruption, crash, and security-relevant bugs); maintenance status (last release, open security issues, archived repositories, maintainer changes); license changes since the last check (NFR-029); deprecations (for example ONNX Runtime providers, P008).
- **Deliverables:** a known-issues report per component.
- **Depends on:** S13.1.
- **Requirements:** FR-363, NFR-029.
- **Acceptance criteria:**
  1. Every direct dependency and every system tool has a recorded check of its tracker and changelog, with the date.
  2. Unmaintained or relicensed components are flagged with a proposed action.
- **Risks/notes:** None.
- **Status:** Not started

#### S13.4: Remediation: upgrade, else downgrade, else mitigate
- **Goal:** No component version with a known relevant issue ships.
- **Scope:** for every finding, in the user's order: **(1) upgrade** to a version that fixes it (the smallest compatible fixed version, or the newest if it is compatible); **(2) if no fixed version exists, downgrade** to an earlier version that does not have the issue and is usable: compatible with the code, without another known relevant issue, still supported enough, passing every test; **(3) otherwise** mitigate (the affected code is not reachable, the feature is off, or configuration blocks it), replace the component, or record an accepted risk with the user's approval. Every change is its own commit with its reason; the register records the chosen version and why; the full test suites run on every platform.
- **Deliverables:** updated versions and lock files; the remediation log in the report; register entries.
- **Depends on:** S13.2, S13.3.
- **Requirements:** FR-363, NFR-023.
- **Acceptance criteria:**
  1. No high or critical finding that affects the NAS remains, unless the user accepted it in writing.
  2. Every downgrade names the issue it avoids, confirms the earlier version has no other relevant known issue, and lists the tests that passed.
  3. All test suites pass on Linux, Windows, and ARM64 after the changes.
- **Risks/notes:** A downgrade can bring back older bugs, so the target version is checked like any other; downgrades are re-checked at every later run and upgraded as soon as a fixed version exists.
- **Status:** Not started

#### S13.5: A repeatable procedure
- **Goal:** The review can be repeated at any time with little effort.
- **Scope:** a script that produces the SBOM and the reports of S13.1–S13.3; a CI job (with the stage-end CI and by hand, the user's CI rule) that fails on new high or critical findings; the written procedure and its triggers: before every release (R13; S14.6 runs it for what changed since S13), before the AI stage (S17), and whenever an advisory for a used component appears (Dependabot and similar alerts).
- **Deliverables:** the script and CI job; the procedure document; the report format.
- **Depends on:** S13.1–S13.4.
- **Requirements:** FR-363, NFR-023.
- **Acceptance criteria:**
  1. Running the procedure on a clean clone produces the SBOM and the reports without manual steps (apart from accepting risks).
  2. A deliberately vulnerable version (on a throwaway branch) is detected and fails the job.
- **Risks/notes:** None.
- **Status:** Not started

#### S13.6: Testing and stage review
- **Goal:** Evidence that S13 holds, then close the stage.
- **Scope:** the full test suites after remediation (unit, integration, system on Linux, Windows, and ARM64; the Raspberry Pi profile); the review report published in the repository; documentation audit (R12); completion record; the user's sign-off.
- **Deliverables:** test results; the published review report; audit report; completion record.
- **Depends on:** S13.1–S13.5.
- **Requirements:** FR-363, NFR-023, NFR-014.
- **Acceptance criteria:**
  1. All suites pass on every platform with the final versions.
  2. The review report lists every component, finding, and decision, and no Critical audit finding is open.
  3. The completion record is written and the user's sign-off is recorded.
- **Risks/notes:** None.
- **Status:** Not started

**Design notes (S13):** the continuous checks that exist since S01–S03 (govulncheck, gosec, `pnpm audit`, Trivy, Dependabot, license checks) stay; S13 adds the complete inventory, the sources they do not cover (system tools, distributions, known issues that are not vulnerabilities, maintenance status), and the user's upgrade-else-downgrade rule. **Exit criteria:** "No component with a known relevant issue ships in the first release, and the review can be repeated with one command."

---
