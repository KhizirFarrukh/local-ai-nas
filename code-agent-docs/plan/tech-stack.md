# Chosen technology stack

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](PLAN_INDEX.md)). Moved unchanged from `plan.md` section 7 in version 1.9.0 (ADR-0044).

## 7. Chosen technology stack

> **Replaced in 0.3.0.** The v0.2.0 options and recommendations are archived in `archive/plan-history/plan_v0.2.0.md`. Each choice below is the user's decision from P003, recorded as an ADR, verified (version, license, maintenance) on 2026-09-24 (session S003, log E005/E007), and listed in `dependencies.md`. Alternatives and reasoning are in each ADR's "Options considered".

| Layer | Choice (verified version) | ADR | Status |
|---|---|---|---|
| Core server language | Go (go1.27.1, pinned in `go.mod`); pure-Go builds (`CGO_ENABLED=0`) | [ADR-0001](../decisions/ADR-0001-backend-language-framework.md) | Accepted |
| API | REST under `/api/v1`; OpenAPI 3 contract at `api/openapi.yaml` (spec-first, oapi-codegen v2.8.0); stdlib `net/http` router; RFC 9457 errors; Redoc 2.5.4 offline docs | [ADR-0002](../decisions/ADR-0002-api-style.md) | Accepted |
| Storage layout | Per-user namespaces `files/<ns>/`, `photos/<ns>/` from S01; internal data at `<root>/.local-ai-nas/` | [ADR-0003](../decisions/ADR-0003-storage-layout.md) | Accepted (S005, user: "Accept all (Recommended)") |
| Repository layout | Single repository; Go module at root; `cmd/`, `internal/`, `api/`, `web/`, `ai-worker/`, `deploy/`, `testdata/`, `docs/`, `scripts/` | [ADR-0004](../decisions/ADR-0004-repository-layout.md) | Accepted |
| Testing, linting, CI | Go `testing` + go-cmp; golangci-lint v2.13.2; govulncheck v1.8.0; go-licenses v2.0.1; Vitest 5.0.1, Playwright 1.63.0, svelte-check, ESLint, Prettier; pytest, Ruff; Trivy v0.74.0; Dependabot; GitHub Actions (Linux + Windows) | [ADR-0005](../decisions/ADR-0005-testing-linting-ci.md) | Accepted |
| Dev environment and packaging | Docker Compose primary; linux/amd64 + linux/arm64 images on `debian:trixie-slim`; AI via `--profile ai`; native Linux (binary + systemd) secondary | [ADR-0006](../decisions/ADR-0006-dev-environment-and-packaging.md) | Accepted (native Windows/macOS deferred) |
| Database | SQLite WAL via `modernc.org/sqlite` v1.59.0 (pure Go); goose v3.28.0 SQL migrations; **from S01** | [ADR-0007](../decisions/ADR-0007-database-sqlite.md) | Accepted |
| Resumable uploads | tus: tusd v2.10.1 embedded (hooks for auth); Uppy 6 + @uppy/tus in the browser; temp in internal data, atomic rename on finalize | [ADR-0008](../decisions/ADR-0008-resumable-uploads-tus.md) | Accepted |
| Web UI | SvelteKit 2 (Svelte 5) static SPA via adapter-static, TypeScript, Tailwind CSS 4; embedded with `go:embed`; pnpm; @tanstack/svelte-virtual | [ADR-0009](../decisions/ADR-0009-web-ui-sveltekit.md) | Accepted |
| Security | Argon2id (x/crypto, t=3, m=64 MiB, p=4); server-side sessions in SQLite; opaque HttpOnly/Secure/SameSite cookie; CSRF tokens; TOTP via pquerna/otp if S03.7 approved; `crypto/tls` | [ADR-0010](../decisions/ADR-0010-security-building-blocks.md) | Accepted |
| Background jobs | Custom persistent queue on SQLite inside the core (leases, retries with backoff, priorities, per-type limits, progress) | [ADR-0011](../decisions/ADR-0011-job-queue-sqlite.md) | Accepted |
| Media toolchain | ExifTool (stay_open, custom Go wrapper), libvips + libheif (WebP thumbnails), FFmpeg/ffprobe, all as subprocesses | [ADR-0012](../decisions/ADR-0012-media-toolchain.md) | Accepted (RAW conversion deferred; transcoding deferral superseded in part by ADR-0020) |
| Video streaming | HLS with a manual quality menu + Auto (hls.js 1.7.3); hybrid on-demand transcoding (FFmpeg, H.264/AAC, hardware encoder when present) with a size-capped cache; photos lightbox and files previews | [ADR-0020](../decisions/ADR-0020-video-streaming-quality-levels.md) | Accepted (user decision S004; details confirmed in the S04 stage document) |
| Reverse geocoding | GeoNames `cities500` + admin1/country tables, custom in-memory k-d tree, bundled at build time; CC BY 4.0 attribution | [ADR-0013](../decisions/ADR-0013-reverse-geocoding-geonames.md) | Accepted |
| Search | Bleve v2.6.1 embedded; English analyzers + fuzzy/prefix; own operator parser → range/term queries; owner/ACL keyword filters; query-time synonyms from own dictionary | [ADR-0014](../decisions/ADR-0014-search-engine-bleve.md) | Accepted (fallback: new ADR if S06.8 misses targets) |
| Network shares | WebDAV via `golang.org/x/net/webdav` with a custom FileSystem through policy, areas, and sidecars | [ADR-0015](../decisions/ADR-0015-network-shares-webdav.md) | Accepted |
| SMB | Samba, Linux-only, optional | [ADR-0019](../decisions/ADR-0019-smb-via-samba.md) | **Proposed** (deferred to S09.1) |
| File watching | fsnotify v1.10.1 + periodic reconciliation; inotify watch-limit guidance | [ADR-0016](../decisions/ADR-0016-file-watching.md) | Accepted |
| AI worker | Python 3.14 + ONNX Runtime 1.30.0 in a separate optional container; uv, Ruff, pytest; pulls jobs from a local-only internal API (token); media read-only; core validates and writes results (I9); no internet at runtime | [ADR-0017](../decisions/ADR-0017-ai-worker-architecture.md) | Accepted |
| Content hash (P005) | SHA-256 (Go standard library) at upload, BLAKE3 only if benchmarks justify it | [ADR-0021](../decisions/ADR-0021-content-hash-algorithm.md) | **Accepted** (S007: SHA-256) |
| Perceptual hash and similarity index (P005) | 64-bit dHash/pHash from a small rendition; BK-tree or multi-index hashing; goimagehash (BSD-2-Clause) or an in-house implementation | [ADR-0022](../decisions/ADR-0022-perceptual-hash-and-similarity-index.md) | **Proposed** (S04.4, S11.1) |
| Look-alike stacks and bursts (P005) | Stack data in each member's sidecar; cover and user-lock rules; burst identifiers first | [ADR-0023](../decisions/ADR-0023-look-alike-stacks.md) | **Proposed** (S11.4) |
| File shortcuts (P005) | App-level shortcut records by stable file ID (not symlinks or hard links); read-only view over shares | [ADR-0024](../decisions/ADR-0024-file-shortcuts.md) | **Proposed** (S11.6) |
| Video optimization codec (P005) | H.264 (libx264/CRF) by default; HEVC and AV1 optional | [ADR-0025](../decisions/ADR-0025-video-optimization-codec.md) | **Proposed** (S12.2) |
| Originals retention and revert (P005) | Own `originals/` store beside the trash, 30-day default; versioning (S08.5) used only if approved | [ADR-0026](../decisions/ADR-0026-originals-retention-and-revert.md) | **Proposed** (S12.5) |
| Drive pools (P005, S007) | mdadm (RAID 0 and RAID 1) orchestrated through the storage helper; parity and combined drives deferred (11a) | [ADR-0027](../decisions/ADR-0027-drive-pool-approach.md) | **Proposed** (S15.1) |
| Pool filesystem (P005) | ext4 (recommended) or XFS | [ADR-0028](../decisions/ADR-0028-pool-filesystem.md) | **Proposed** (S15.1) |
| Privileged storage helper (P005) | Separate Go service on the host as root; allow-listed operations over an authenticated Unix socket; audit log | [ADR-0029](../decisions/ADR-0029-privileged-storage-helper.md) | **Proposed** (S15.2) |
| Storage migration engine (P007) | Built-in Go copier with a journal, content-hash verification, throttling | [ADR-0030](../decisions/ADR-0030-storage-migration-engine.md) | Proposed (S08.4) |
| Switching the storage root (P007) | Stable mount point remounted by the helper on Linux; configuration change elsewhere; container restarted | [ADR-0031](../decisions/ADR-0031-switching-the-storage-root.md) | Proposed (S15.6) |
| Predictive drive health (P007) | Rules and trends over smartctl JSON (Linux) and the Windows reliability counters; four statuses | [ADR-0032](../decisions/ADR-0032-predictive-drive-health.md) | Proposed (S10.3) |
| Mirror conversion, hot replacement, growth (P007) | mdadm degraded RAID 1, add, `--replace`, `--grow --size=max`, online filesystem growth | [ADR-0033](../decisions/ADR-0033-mirror-conversion-and-hot-replacement.md) | Proposed (S15) |
| New-drive detection (P007) | Kernel events through the helper on Linux; re-scan everywhere | [ADR-0034](../decisions/ADR-0034-hot-plug-drive-detection.md) | Proposed (S15.3) |
| Secure erase (P007) | NVMe sanitize or format, ATA security erase, otherwise overwrite | [ADR-0035](../decisions/ADR-0035-secure-erase.md) | Proposed (S15.11) |
| SSD read cache (P007) | Application-level, content-addressed, TinyLFU admission, W-TinyLFU eviction, write-through | [ADR-0036](../decisions/ADR-0036-ssd-read-cache.md) | Proposed (S16.1) |
| Fast internal data (P007) | Derived data on the SSD by default; the database optionally | [ADR-0037](../decisions/ADR-0037-fast-internal-data-placement.md) | Proposed (S16.2) |
| Block-level SSD cache (P007) | lvmcache writethrough, only with an LVM layer from pool creation; not in the first version (Q70) | [ADR-0038](../decisions/ADR-0038-block-level-ssd-cache.md) | Proposed (S16.1) |
| Admin console (the user's requirement) | An admin section of the same web app at `/admin` with admin API routes under `/api/v1/admin` | [ADR-0039](../decisions/ADR-0039-admin-console.md) | **Accepted** (S007 E050, with the S03 approval) |
| Item identity (P008) | UUIDv7 IDs in an items table; paths are attributes | [ADR-0040](../decisions/ADR-0040-item-identity.md) | Proposed (S01 follow-up, built in S03) |
| Operation journal (P008) | Intent-first journal, idempotent steps, recovery at startup, failure matrix | [ADR-0041](../decisions/ADR-0041-operation-journal-and-crash-consistency.md) | Proposed (S01 follow-up, built in S03) |
| Local-origin protection (P008) | Host allow-list, Origin check, no CORS, before login | [ADR-0042](../decisions/ADR-0042-local-origin-protection.md) | Proposed (S03.5-T02) |
| Hybrid search (P008) | In-process vector index, rank fusion, permission pre-filter, query encoder | [ADR-0043](../decisions/ADR-0043-hybrid-search.md) | Proposed (S17.10) |
| Plan document structure (P008) | `code-agent-docs/plan/` with `PLAN_INDEX.md` | [ADR-0044](../decisions/ADR-0044-plan-document-structure.md) | **Accepted** (the user's request) |
| AI models | CLIP-family zero-shot (e.g. SigLIP, Apache-2.0); YuNet (MIT); SFace (Apache-2.0); HDBSCAN (scikit-learn); RapidOCR if S17.10 is approved; InsightFace excluded | [ADR-0018](../decisions/ADR-0018-ai-models.md) | Accepted direction (variants deferred to S17) |

### 7.1 Deferred items
- **Native Windows and macOS installs** (ADR-0006). Pending user decision (Q5).
- **SMB via Samba** (ADR-0019, Proposed). Design in S09.1.
- **Full RAW conversion** (e.g. LibRaw) (ADR-0012). _Video transcoding for streaming quality levels is in scope since 0.4.0 (ADR-0020)._
- **Exact AI model variants**, **GPU execution providers**, and **embedding storage** (ADR-0017/0018). Chosen in S17 by benchmark.
- **Semantic search** needs a text model at query time. _Decided in 1.9.0 (the user's decision D3, P008):_ allowed for the typed search text only, as I4 now says (ADR-0018 amendment, ADR-0043); whether semantic search is built at all is still Q36. _(The S005 default D-06, precomputed forms only, is superseded by D3.)_
- ~~Storage layout (ADR-0003)~~: accepted in S005 (D-01), no longer pending.
- **P005 decisions:** ADR-0021 to ADR-0029 (the table above). ADR-0021 (content hash) was **Accepted** in S007 (SHA-256); the others are Proposed.
- **P007 and the admin console (1.7.0):** ADR-0030 to ADR-0038 are **Proposed**, decided when their stages are planned in detail. **ADR-0039** (admin console) was **Accepted** with the S03 approval (S007 E050).
- **P008 (1.9.0):** new ADR-0040 (item identity), ADR-0041 (operation journal and crash consistency), ADR-0042 (local-origin protection before authentication), ADR-0043 (hybrid search), all **Proposed**; ADR-0044 (plan document structure) **Accepted** by the user's request. Amended: ADR-0007 (durability, D5), ADR-0011 (minimal job foundation now), ADR-0014 (per-field mapping, Arabic-script normalization, BM25, no vectors in Bleve), ADR-0017 and ADR-0018 (execution providers, quantized models, CLIP ViT-B/32, query encoder), ADR-0024 (targets by item ID).
- **Complex RAID** (parity, combined drives, nesting, SnapRAID with mergerfs) is deferred by the user's decision in S007 (section 11a).
- Implementation-time confirmations recorded as tasks:
  - Node.js LTS and TypeScript 7 / svelte-check compatibility (S02.1).
  - @tanstack/svelte-virtual on Svelte 5 (S02.3 prototype).
  - Debian FFmpeg build flags (S14.1).
  - ONNX Runtime on Python 3.14 (S17.1).
  - The synonym dictionary source and license (S06.5).

---
