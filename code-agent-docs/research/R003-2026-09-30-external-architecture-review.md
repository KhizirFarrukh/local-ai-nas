# R003: External architecture review (P008), record

| Field | Value |
|---|---|
| Sources | The review `code-reviews/1-local-ai-nas-improvement-review.md` (archived verbatim as `prompts/CR001-improvement-review-1.md`, S007 E056); the user's six inline suggestions and the planner's prompt `prompts/P008-external-architecture-review.json` (plan change request #8) |
| Date | 2026-09-30 (session S007, log E058–E061) |
| Plan | 1.9.0 |
| Earlier record | [R002](R002-2026-09-30-external-review-1.md) (the first check, superseded in part) |

## Context corrections (P008)

- The review was written as if S01 had not started; S01 and S02 are Done and S03 is In Progress, so its "before S01" changes became follow-up tasks built first in S03.
- The license is chosen (AGPL-3.0-or-later). The roadmap has 17 stages since 1.9.0, not twelve; AI stays last (I8). Direct video playback is planned (FR-144). The job system, the separate AI worker, and ONNX Runtime exist in the plan; Docker Compose with an AI profile is planned; multi-user is S07, sessions S03.

## Suggestions and verdicts

| ID | Suggestion | Verdict | Where |
|---|---|---|---|
| R-01 | Immutable item and media IDs | Adopted (highest priority) | FR-346–FR-349, ADR-0040; S01.3-T11, T12 |
| R-02 | Namespaced sidecar name | Adopted: `.lainas.json` (D1) | FR-356, FR-023, 8.6 |
| R-03 | Security and ACL data authoritative in SQLite | Adopted (D2); I3 refined | FR-359, A27, 8.8, 8.36 |
| R-04 | Minimal durable job and operation foundation; crash behavior | Adopted | FR-351–FR-353, NFR-053, ADR-0041, ADR-0011 amendment; S01.4-T08–T10 |
| R-05 | Trash earlier | Adopted | FR-354; S01.3-T13, S02.5-T05; RK-19 mitigated |
| R-06 | Browser-side protection before authentication | Adopted, urgent (without the per-install token: sessions arrive in S03) | FR-350, NFR-057, ADR-0042; S03.5-T02 first |
| R-07 | Semantic-search contradiction | Adopted (D3); I4 reworded | FR-054, 7.1, ADR-0043 |
| R-08 | Less process ceremony | Not applied: the user kept the process (D4) | — |
| R-09 | Split the plan with an index | Adopted (the user's request) | ADR-0044; `code-agent-docs/plan/` |
| R-10 | Keep the stack, no microservices | Confirmed | — |
| R-11 | Per-field analyzers, multilingual | Adopted | FR-357, 8.11, ADR-0014 amendment |
| R-12 | SQLite `synchronous=FULL` | Adopted (D5) | NFR-056, ADR-0007 amendment; S01.1-T12 |
| R-13 | README status, diagram, license; freeze the architecture | README adopted; change intake adopted (D6) | README update; RULES R14, plan 11d |
| I-01 | Asynchronous task queuing | Already the design; improved: OS priority, backpressure, responsiveness | NFR-054, NFR-055, 8.41, S04 notes |
| I-02 | Hardware-aware inference | Adopted with corrections (no Moondream, NG9; QuickSync is an encoder) | FR-361, ADR-0017 and ADR-0018 amendments |
| I-03 | Hybrid vector and full-text search | Adopted without new services | FR-358, FR-362, 8.40, ADR-0043 |
| I-04 | Docker Compose for everything | Already planned; hardening adopted | S14.1 note |
| I-05 | JWT or OAuth, scoped folders | Already planned; JWT not used for browser sessions (instant revocation) | S07.3 note; OIDC in R05 |
| I-06 | Dashboard resource metrics | Adopted | FR-360 |

## The user's decisions (S007 E059, verbatim option labels)

D1 ".lainas.json (Recommended)" · D2 "Database, sidecar mirror (Recommended)" · D3 "Yes, query text only (Recommended)" · D4 "Keep current process" · D5 "FULL, throttled writes (Recommended)" · D6 "Yes, freeze MVP (Recommended)" · D7 "Yes, internal alpha at M2 (Recommended)" · follow-ups "Approve all, F2 first (Recommended)".

## Added by the user during P008 (S007 E060)

- **New stage S13, dependency security review** (FR-363), before packaging (the user's choice); later stages renumbered (plan 10.18).
- **Google Takeout import** of Drive and Photos with everything Google recorded (FR-222 extended, R01, Must), with the user's **sample data as the prerequisite** (A28, 8.42).

## Verification (2026-09-30)

- `github.com/google/uuid` v1.6.0 has `NewV7` (BSD-3-Clause); a run produced version-7 UUIDs.
- modernc.org/sqlite v1.59.0 bundles SQLite 3.53.4; `synchronous=FULL` takes effect; `VACUUM INTO` works; the online backup API exists (`NewBackup`, `Step`, `Finish`). SQLite docs: NORMAL in WAL "might roll back following a power loss"; FULL is ACID in WAL mode.
- Bleve: vector search needs FAISS (C++) through cgo and the `vectors` build tag (docs/vectors.md, since v2.4.0); BM25 since v2.5.0 (release notes); latest v2.6.1.
- ONNX Runtime (onnxruntime.ai, PyPI; all MIT): `onnxruntime` and `onnxruntime-gpu` 1.30.0; `onnxruntime-openvino` 1.24.1; `onnxruntime-migraphx` 1.27.1; `onnxruntime-directml` 1.24.4 ("DirectML is in sustained engineering … new feature development has moved to WinML"); the ROCm provider is marked deprecated (`onnxruntime-rocm` 1.22.2.post3).
- Priority mechanisms: Linux I/O priorities are honored only by the bfq and mq-deadline schedulers (kernel ioprio docs); cgroup v2 `io.weight` needs the iocost controller or BFQ (cgroup-v2 docs); Windows `PROCESS_MODE_BACKGROUND_BEGIN` works only on the calling process, the priority classes lower CPU only (SetPriorityClass docs).
- CLIP ViT-B/32: the OpenAI CLIP repository is MIT; the exact weights license is **Unverified** (the Hugging Face card has no license tag). FastEmbed 0.8.1 (Apache-2.0) loads from a local folder without downloads (`specific_model_path`; `local_files_only`).
- Browsers send `Origin` on same-origin requests except GET and HEAD; certain referrer policies make it `null` for non-GET requests outside cors mode (MDN), so the app moves to `Referrer-Policy: same-origin` (ADR-0042).
- Google Drive's Takeout metadata: **Unverified** (no reliable description found); to be learned from the user's sample (A28).
