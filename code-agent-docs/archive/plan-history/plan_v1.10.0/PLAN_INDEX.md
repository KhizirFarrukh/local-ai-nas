# local-ai-nas: Development Plan (index)

| Field | Value |
|---|---|
| **Version** | 1.10.0 |
| **Status** | **Approved baseline** (approved by the user in S005, 2026-09-24); 1.1.0 adds the user's setup-script requirement (S005 E015) |
| **Last updated** | 2026-09-30 (session S007) |
| **Source of vision** | `README.md` (repository root), the user's staged roadmap (`code-agent-docs/prompts/P002-staged-development-roadmap.json`), the user's technology stack (`code-agent-docs/prompts/P003-technology-stack.json`), and the user's feature additions (`code-agent-docs/prompts/P005-feature-additions.json`, with the user's chat decisions in S007) |
| **Previous version** | 1.9.0, archived at `code-agent-docs/archive/plan-history/plan_v1.9.0/` (a copy of the plan folder, the R4 practice since the split; the single-file versions 0.1.0–1.8.1 are `plan_v<version>.md` there) |
| **Stage IDs** | Changed in 1.4.0 (P005): packaging became S13 (was S11) and AI S15 (was S12). Changed in 1.7.0 (P007): S15 became SSD caching and AI S16; S03.9 → S03.10, S10.6 → S10.7, S14.9 → S14.12, S14.10 → S14.13. **Changed in 1.9.0 (the user's requirement): the new S13 is the dependency security review; packaging is S14, drives S15, SSD caching S16, AI S17.** Older documents use the old IDs; the table in **10.18** translates them. |

> **This is a living document.** It changes as the user gives feedback. Every change follows `code-agent-docs/RULES.md` **R4**: the old version is archived, the version is bumped, and a revision entry is added. While the plan is a pre-1.0 draft, restructurings bump the MINOR version. **When the user approves this plan as the baseline, it becomes version 1.0.0.**
>
> **Technology decisions** are recorded as ADRs in `code-agent-docs/decisions/` (R5). Since 0.3.0, the user-chosen stack (P003) is **Accepted** (section 7). Items marked *Proposed* or *deferred* in section 7 still need a decision. Every dependency is listed in `code-agent-docs/dependencies.md`.

---

## Plan files (version 1.10.0; one version for the whole plan set)

| File | Holds (old `plan.md` sections) |
|---|---|
| [overview.md](overview.md) | 1, 2, 2b: Overview, goals, and principles |
| [invariants.md](invariants.md) | 2a: Project invariants |
| [requirements.md](requirements.md) | 3 (3.1–3.3): Requirements |
| [assumptions.md](assumptions.md) | 4: Assumptions |
| [questions.md](questions.md) | 5: Open questions |
| [architecture.md](architecture.md) | 6 and 8: Architecture and technical concerns |
| [tech-stack.md](tech-stack.md) | 7: Chosen technology stack |
| [roadmap.md](roadmap.md) | 9, 10, 10.1, 10.17, 10.18, 11, 11a–11d: Roadmap, milestones, and releases |
| [stage-specs/S01-basic-nas.md](stage-specs/S01-basic-nas.md) | 10.2: S01: Basic NAS implementation |
| [stage-specs/S02-nas-gui.md](stage-specs/S02-nas-gui.md) | 10.3: S02: NAS GUI |
| [stage-specs/S03-security.md](stage-specs/S03-security.md) | 10.4: S03: Security |
| [stage-specs/S04-media-management.md](stage-specs/S04-media-management.md) | 10.5: S04: Media management |
| [stage-specs/S05-media-metadata.md](stage-specs/S05-media-metadata.md) | 10.6: S05: Media metadata |
| [stage-specs/S06-search.md](stage-specs/S06-search.md) | 10.7: S06: Search |
| [stage-specs/S07-multi-user-and-sharing.md](stage-specs/S07-multi-user-and-sharing.md) | 10.8: S07: Multi-user and sharing |
| [stage-specs/S08-data-protection-and-recovery.md](stage-specs/S08-data-protection-and-recovery.md) | 10.9: S08: Data protection and recovery |
| [stage-specs/S09-network-file-access.md](stage-specs/S09-network-file-access.md) | 10.10: S09: Network file access and external change sync |
| [stage-specs/S10-admin-console.md](stage-specs/S10-admin-console.md) | 10.11: S10: Admin console: monitoring, quotas, and system settings |
| [stage-specs/S11-duplicates-and-look-alikes.md](stage-specs/S11-duplicates-and-look-alikes.md) | 10.12: S11: Duplicate and look-alike management |
| [stage-specs/S12-storage-optimization.md](stage-specs/S12-storage-optimization.md) | 10.13: S12: Storage optimization |
| [stage-specs/S13-dependency-security-review.md](stage-specs/S13-dependency-security-review.md) | 10.13a: S13: Dependency security review (new in 1.9.0, the user's requirement) |
| [stage-specs/S14-packaging-and-release.md](stage-specs/S14-packaging-and-release.md) | 10.14: S14: Packaging, deployment, and pre-AI release |
| [stage-specs/S15-drives-and-pools.md](stage-specs/S15-drives-and-pools.md) | 10.15: S15: Drives, pools, and drive lifecycle (RAID 0 and RAID 1) |
| [stage-specs/S16-ssd-caching.md](stage-specs/S16-ssd-caching.md) | 10.15a: S16: SSD caching (new in 1.7.0, P007) |
| [stage-specs/S17-ai-features.md](stage-specs/S17-ai-features.md) | 10.16: S17: AI features |
| [testing.md](testing.md) | 12: Testing strategy |
| [risks.md](risks.md) | 13: Risks and mitigations |
| [changelog.md](changelog.md) | 14: Revision history |

## Stage status (from roadmap 10.1; keep both equal)

| Stage | Name | Status | Plan file | Task document |
|---|---|---|---|---|
| S01 | Basic NAS implementation | Done | [stage-specs/S01-basic-nas.md](stage-specs/S01-basic-nas.md) | [task document](../stages/S01-basic-nas.md) |
| S02 | NAS GUI | **Done** | [stage-specs/S02-nas-gui.md](stage-specs/S02-nas-gui.md) | [task document](../stages/S02-nas-gui.md) |
| S03 | Security | **Approved** (2026-09-29, S007 E050; `stage-specs/S03-security.md`); In Progress from S03.1-T01 | [stage-specs/S03-security.md](stage-specs/S03-security.md) | [task document](../stages/S03-security.md) |
| S04 | Media management | Not started | [stage-specs/S04-media-management.md](stage-specs/S04-media-management.md) | not written yet |
| S05 | Media metadata | Not started | [stage-specs/S05-media-metadata.md](stage-specs/S05-media-metadata.md) | not written yet |
| S06 | Search | Not started | [stage-specs/S06-search.md](stage-specs/S06-search.md) | not written yet |
| S07 | Multi-user and sharing | Not started | [stage-specs/S07-multi-user-and-sharing.md](stage-specs/S07-multi-user-and-sharing.md) | not written yet |
| S08 | Data protection and recovery | Not started | [stage-specs/S08-data-protection-and-recovery.md](stage-specs/S08-data-protection-and-recovery.md) | not written yet |
| S09 | Network file access and external change sync | Not started | [stage-specs/S09-network-file-access.md](stage-specs/S09-network-file-access.md) | not written yet |
| S10 | Admin console: monitoring, quotas, and system settings | Not started | [stage-specs/S10-admin-console.md](stage-specs/S10-admin-console.md) | not written yet |
| S11 | Duplicate and look-alike management | Not started | [stage-specs/S11-duplicates-and-look-alikes.md](stage-specs/S11-duplicates-and-look-alikes.md) | not written yet |
| S12 | Storage optimization | Not started | [stage-specs/S12-storage-optimization.md](stage-specs/S12-storage-optimization.md) | not written yet |
| S13 | Dependency security review _(new in 1.9.0)_ | Not started | [stage-specs/S13-dependency-security-review.md](stage-specs/S13-dependency-security-review.md) | not written yet |
| S14 | Packaging, deployment, and pre-AI release (was S11, then S13 until 1.9.0) | Not started | [stage-specs/S14-packaging-and-release.md](stage-specs/S14-packaging-and-release.md) | not written yet |
| S15 | Drives, pools, and drive lifecycle | Not started | [stage-specs/S15-drives-and-pools.md](stage-specs/S15-drives-and-pools.md) | not written yet |
| S16 | SSD caching | Not started | [stage-specs/S16-ssd-caching.md](stage-specs/S16-ssd-caching.md) | not written yet |
| S17 | AI features (was S12, then S15, then S16) | Not started | [stage-specs/S17-ai-features.md](stage-specs/S17-ai-features.md) | not written yet |

## Where things stand

- **Milestones:** M3, the first release, is S01–S14 (roadmap section 11); an internal alpha at M2 (S01–S06) for the user's own use (D7).
- **Active stage:** S03 (Security), In Progress; the approved P008 follow-ups come first (the S03 task document, section 5).
- **Next actions:** see `code-agent-docs/CURRENT_STATE.md`.
- **Reading at session start (RULES R1):** this index, `invariants.md`, `questions.md`, `tech-stack.md`, `roadmap.md` (10.1), and the complete file of the active or next stage; other files when needed.

## Map from old section numbers to files

References such as "plan 8.8" or "plan section 10.18" in logs, ADRs, and stage documents use the old section numbers, which the files keep in their headings.

| Old sections | File |
|---|---|
| 1, 2, 2b | [overview.md](overview.md) |
| 2a | [invariants.md](invariants.md) |
| 3, 3.1–3.3 | [requirements.md](requirements.md) |
| 4 | [assumptions.md](assumptions.md) |
| 5 | [questions.md](questions.md) |
| 6, 6.1–6.6, 8, 8.1–8.42 | [architecture.md](architecture.md) |
| 7, 7.1 | [tech-stack.md](tech-stack.md) |
| 9, 10, 10.1, 10.17–10.18, 11, 11a, 11b, 11b.1–11b.2, 11c, 11d | [roadmap.md](roadmap.md) |
| 10.2 | [stage-specs/S01-basic-nas.md](stage-specs/S01-basic-nas.md) |
| 10.3 | [stage-specs/S02-nas-gui.md](stage-specs/S02-nas-gui.md) |
| 10.4 | [stage-specs/S03-security.md](stage-specs/S03-security.md) |
| 10.5 | [stage-specs/S04-media-management.md](stage-specs/S04-media-management.md) |
| 10.6 | [stage-specs/S05-media-metadata.md](stage-specs/S05-media-metadata.md) |
| 10.7 | [stage-specs/S06-search.md](stage-specs/S06-search.md) |
| 10.8 | [stage-specs/S07-multi-user-and-sharing.md](stage-specs/S07-multi-user-and-sharing.md) |
| 10.9 | [stage-specs/S08-data-protection-and-recovery.md](stage-specs/S08-data-protection-and-recovery.md) |
| 10.10 | [stage-specs/S09-network-file-access.md](stage-specs/S09-network-file-access.md) |
| 10.11 | [stage-specs/S10-admin-console.md](stage-specs/S10-admin-console.md) |
| 10.12 | [stage-specs/S11-duplicates-and-look-alikes.md](stage-specs/S11-duplicates-and-look-alikes.md) |
| 10.13 | [stage-specs/S12-storage-optimization.md](stage-specs/S12-storage-optimization.md) |
| 10.13a | [stage-specs/S13-dependency-security-review.md](stage-specs/S13-dependency-security-review.md) |
| 10.14 | [stage-specs/S14-packaging-and-release.md](stage-specs/S14-packaging-and-release.md) |
| 10.15 | [stage-specs/S15-drives-and-pools.md](stage-specs/S15-drives-and-pools.md) |
| 10.15a | [stage-specs/S16-ssd-caching.md](stage-specs/S16-ssd-caching.md) |
| 10.16 | [stage-specs/S17-ai-features.md](stage-specs/S17-ai-features.md) |
| 12, 12.1–12.4 | [testing.md](testing.md) |
| 13 | [risks.md](risks.md) |
| 14 | [changelog.md](changelog.md) |

## Former table of contents of plan.md (moved)

## Contents

1. Project overview
2. Goals and non-goals
2a. Project invariants
2b. Cross-cutting principles
3. Requirements
4. Assumptions
5. Open questions for the user
6. Proposed high-level architecture
7. Chosen technology stack
8. Key technical concerns
9. Development methodology
10. Stage roadmap
11. MVP definition
11a. Not scheduled / future candidates
11b. Release roadmap after the MVP (new in 1.6.0)
11c. Public release specification (R09) (new in 1.6.0)
11d. Change intake and backlog (new in 1.9.0)
12. Testing strategy
13. Risks and mitigations
14. Revision history

---
