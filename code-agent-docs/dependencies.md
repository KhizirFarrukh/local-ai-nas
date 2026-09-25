# Dependency register

**Purpose:** every dependency, external tool, dataset, and AI model the project uses or has evaluated, with its version, license, and verification status. This is required by P003 and by RULES.md **R6**: every new dependency is added here **in the same commit that introduces it**. **Section 12** lists what each platform needs at runtime; the per-platform setup scripts (S11.2, FR-149) install exactly that list (NFR-032).

**Last updated:** 2026-09-24 (session S005: section 12 added; S01.1-T01 created `go.mod`; S01.1-T02 license policy; S01.1-T05 CI actions; S01.1-T07 go-toml; S01.1-T10 SQLite and goose; S01.1-T06 dev image; S01.2-T04 x/sys direct; S01.6-T01 x/text direct; S01.5-T03 kin-openapi for tests; S01.3-T02 oapi-codegen runtime; audit A002: section 3 system tools for development, the scripts, and the API guide). Verification method and raw results: `logs/sessions/2026-09-24_S003.md` entries E005 and E007, and audit A001 group F (`audits/A001-2026-09-24-documentation-audit.md`).

**Licensing policy (P003, NFR-029):** every dependency and model must have a license that allows anyone to deploy and use this project. Nothing may be restricted to non-commercial or research-only use. The project's own license is **AGPL-3.0-or-later** (Q22, decided in S005; policy and allow-list in `docs/licensing.md` and `scripts/allowed-licenses.txt`). Items that need attention under it are marked ⚠.

**Status legend:**
- **Verified (date, source):** version and license checked against an official source on that date.
- **Unverified:** not yet checked, with the reason and the stage that checks it.
- **Candidate:** likely to be used, not yet decided.
- **Evaluated, not adopted** / **Rejected:** kept for traceability.

Versions are the **latest stable at verification**. Only **direct** dependencies are listed. Transitive Go dependencies (e.g. those of tusd and Bleve) are checked automatically by go-licenses and govulncheck in CI (ADR-0005). The exact pinned version is set when the dependency is first added (go.sum, pnpm-lock.yaml, uv.lock, image digests) and this register is updated in the same commit.

## 1. Toolchains and runtimes

| Name | Version | License | Purpose | Stage | ADR | Verification status |
|---|---|---|---|---|---|---|
| Go | go1.27.1 | BSD-3-Clause | Core server language, toolchain | S01+ | [ADR-0001](decisions/ADR-0001-backend-language-framework.md) | Verified 2026-09-24 (go.dev/dl). **Pinned in `go.mod`** (`go 1.27`, `toolchain go1.27.1`, S01.1-T01): an older local Go (e.g. winget's go1.27.0 on the Windows dev PC) downloads go1.27.1 automatically (`GOTOOLCHAIN=auto`) |
| Python | 3.14.7 | PSF-2.0 | AI worker runtime | S12 | [ADR-0017](decisions/ADR-0017-ai-worker-architecture.md) | Version verified 2026-09-24 (endoflife.date). License **unverified**: the PSF wording was not found on docs.python.org/3/license.html in A001; re-check at S12.1 |
| Node.js (LTS) | v24.21.0 ("Krypton" LTS, 2026-09-07) | MIT | Build-time only: SvelteKit/Vite build and tests | S02+ | [ADR-0009](decisions/ADR-0009-web-ui-sveltekit.md) | Verified 2026-09-24 (nodejs.org/dist/index.json) in audit A001. S02.1 pins the LTS current at that time **Installed** (S02.1-T01): 24.19.0 on the Windows development PC (winget `OpenJS.NodeJS.LTS`, whose newest is 24.19.0 as of 2026-09-25), and 24.21.0 in WSL (official Linux tarball in `~/.local/opt`, SHA-256 checked). `web/package.json` `engines`: `>=24.19.0 <25`. |
| pnpm | 12.6.0 | MIT | Web package manager (lockfile) | S02+ | [ADR-0009](decisions/ADR-0009-web-ui-sveltekit.md) | Verified 2026-09-24 (npm) **Installed** (S02.1-T01) with `npm install -g pnpm@12.6.0` on Windows and in WSL; `packageManager: pnpm@12.6.0` in `web/package.json`. Its minimum-release-age guard is on (`web/pnpm-workspace.yaml` lists the one exception). |
| uv | 0.12.18 | MIT OR Apache-2.0 | Python env and lockfile | S12 | [ADR-0017](decisions/ADR-0017-ai-worker-architecture.md) | Verified 2026-09-24 (PyPI) |
| Perl | 5.40.1-6+deb13u1 (Debian trixie) | Artistic-1.0-Perl OR GPL-1.0-or-later | Runtime for ExifTool (bundled in the image) | S04+ | [ADR-0012](decisions/ADR-0012-media-toolchain.md) | Version verified 2026-09-24 (sources.debian.org) in A001. License per Perl's standard terms (not re-fetched) |

## 2. Go libraries linked into the core binary

| Name | Version | License | Purpose | Stage | ADR | Verification status |
|---|---|---|---|---|---|---|
| github.com/pelletier/go-toml/v2 | v2.4.3 | MIT | Config file (TOML) parsing | S01.1 | [ADR-0001](decisions/ADR-0001-backend-language-framework.md) | Verified 2026-09-24 (proxy.golang.org, deps.dev). **In `go.mod`** and linked (S01.1-T07); `go-licenses report` = MIT |
| modernc.org/sqlite | v1.59.0 | BSD-3-Clause | Pure-Go SQLite driver (WAL) | S01+ | [ADR-0007](decisions/ADR-0007-database-sqlite.md) | Verified 2026-09-24 (proxy, deps.dev, pkg.go.dev WAL docs). **In `go.mod`** and linked (S01.1-T10). It links modernc.org/libc v1.75.7, mathutil v1.7.1, memory v1.12.1 (BSD-3-Clause), github.com/dustin/go-humanize v1.0.1, mattn/go-isatty v0.0.24, ncruces/go-strftime v1.0.0 (MIT), remyoudompheng/bigfft (BSD-3-Clause), golang.org/x/sys (BSD-3-Clause), per `go-licenses report` 2026-09-24 |
| github.com/pressly/goose/v3 | v3.28.0 | MIT | SQL schema migrations (embedded) | S01+ | [ADR-0007](decisions/ADR-0007-database-sqlite.md) | Verified 2026-09-24 (proxy, LICENSE file). **In `go.mod`** and linked (S01.1-T10). It links mfridman/interpolate v0.0.2 (MIT), sethvargo/go-retry v0.4.0 (Apache-2.0), go.uber.org/multierr v1.11.0 (MIT), golang.org/x/sync (BSD-3-Clause), per `go-licenses report` 2026-09-24 |
| github.com/tus/tusd/v2 | v2.10.1 | MIT | Embedded tus server (resumable uploads) | S01.4, S04.2 | [ADR-0008](decisions/ADR-0008-resumable-uploads-tus.md) | Verified 2026-09-24 (proxy, GitHub, handler/config.go). **In `go.mod` and linked** since S01.4-T01: only `pkg/handler`, `pkg/filestore`, and `pkg/filelocker` are imported, which link github.com/tus/lockfile v1.2.0 (MIT) and golang.org/x/exp (BSD-3-Clause); tusd's cloud stores (S3, GCS, Azure) are not imported, so their modules appear only in `go.sum`. `go-licenses` and `govulncheck` clean |
| golang.org/x/exp | v0.0.0-20260824195058-e88cd73687aa | BSD-3-Clause | The `slog` types of tusd's logger (the adapter in `internal/uploads`) | S01.4 | [ADR-0008](decisions/ADR-0008-resumable-uploads-tus.md) | **In `go.mod`** since S01.4-T01 (direct only for the adapter; tusd links it anyway). An experimental module without compatibility promises: pinned by `go.mod`; drop the adapter when tusd logs through `log/slog` |
| github.com/oapi-codegen/runtime | v1.7.0 | Apache-2.0 | Runtime helpers for generated API code | S01.5 | [ADR-0002](decisions/ADR-0002-api-style.md) | Verified 2026-09-24 (proxy, deps.dev). **In `go.mod` and linked** since S01.3-T02 (query binding of the generated server). It links github.com/apapsch/go-jsonmerge/v2 v2.0.0 (MIT) and github.com/google/uuid v1.6.0 (BSD-3-Clause), per `go-licenses report` 2026-09-24 |
| golang.org/x/crypto | v0.57.0 | BSD-3-Clause | Argon2id | S03 | [ADR-0010](decisions/ADR-0010-security-building-blocks.md) | Verified 2026-09-24 |
| github.com/pquerna/otp | v1.5.0 | Apache-2.0 | TOTP 2FA, **only if S03.7 is approved** | S03.7 | [ADR-0010](decisions/ADR-0010-security-building-blocks.md) | Verified 2026-09-24. **Low activity** (last push 2025-08); re-check at S03.7 |
| github.com/blevesearch/bleve/v2 | v2.6.1 | Apache-2.0 | Embedded full-text search | S06+ | [ADR-0014](decisions/ADR-0014-search-engine-bleve.md) | Verified 2026-09-24 (proxy, GitHub, query.go capabilities) |
| github.com/fsnotify/fsnotify | v1.10.1 | BSD-3-Clause | File system events | S05.7, S09.4 | [ADR-0016](decisions/ADR-0016-file-watching.md) | Verified 2026-09-24 |
| golang.org/x/sys | v0.48.0 | BSD-3-Clause | Free-space query (`unix.Statfs`, `windows.GetDiskFreeSpaceEx`) and, from S01.2-T05, volume/device identity (same-filesystem check) | S01.2 | [ADR-0001](decisions/ADR-0001-backend-language-framework.md) | Verified 2026-09-24 (proxy.golang.org, deps.dev). **In `go.mod` as a direct requirement** and linked (S01.2-T04); it was already present indirectly (via modernc and the tools) |
| golang.org/x/text (unicode/norm) | v0.42.0 | BSD-3-Clause | NFC normalization of file names in the path resolver | S01.6 | [ADR-0001](decisions/ADR-0001-backend-language-framework.md) (stage-level choice; not named in the ADR text) | Verified 2026-09-24 (proxy.golang.org, deps.dev) in A001. **Decided in S01.6-T01 (S005): normalize to NFC** (not reject). **In `go.mod` as a direct requirement** and linked |
| golang.org/x/net (webdav) | v0.59.0 | BSD-3-Clause | WebDAV server with a custom FileSystem | S09 | [ADR-0015](decisions/ADR-0015-network-shares-webdav.md) | Verified 2026-09-24 (FileSystem interface in webdav/file.go) |

## 3. Go development tools (not linked into the product)

**How `go.mod` records these tools:** Go writes the three `tool` modules as `// indirect` requires, plus their own dependencies (for example kin-openapi, cobra, golang.org/x/tools, and x/sys, x/net, x/text at the versions in section 2). These are **build-time only**: `go build ./cmd/local-ai-nas` does not link them, and go-licenses and govulncheck check the product binary's real imports (S01.1-T02, T05).

| Name | Version | License | Purpose | Stage | ADR | Verification status |
|---|---|---|---|---|---|---|
| github.com/oapi-codegen/oapi-codegen/v2 | v2.8.0 | Apache-2.0 | Generate Go server interfaces and types from `api/openapi.yaml` | S01.5 | [ADR-0002](decisions/ADR-0002-api-style.md) | Verified 2026-09-24. **In `go.mod`** as a `tool` directive (S01.1-T01); `go tool oapi-codegen -version` = v2.8.0 |
| github.com/google/go-cmp | v0.7.0 | BSD-3-Clause | Test comparisons and diffs | S01+ | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Verified 2026-09-24. **In `go.mod`** as a direct requirement (S01.1-T04); imported only by tests, so not linked into the binary |
| golangci-lint (github.com/golangci/golangci-lint/v2) | v2.13.2 (pinned binary; CI action + local binary install via `scripts/install-golangci-lint.sh` into `./bin`, not a go.mod tool) | ⚠ GPL-3.0 (dev tool only; never linked or distributed) | Lint and format (gofmt/goimports), depguard architecture rules | S01+ | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Verified 2026-09-24 |
| govulncheck (golang.org/x/vuln) | v1.8.0 | BSD-3-Clause | Vulnerability scan | S01+ | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Verified 2026-09-24. **In `go.mod`** as a `tool` directive (S01.1-T01); `go tool govulncheck -version` = v1.8.0 |
| github.com/google/go-licenses/v2 | v2.0.1 | Apache-2.0 | Dependency license check | S01+ | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Verified 2026-09-24. **In `go.mod`** as a `tool` directive (S01.1-T01). Run through `scripts/check-licenses.sh` with `scripts/allowed-licenses.txt` (S01.1-T02) |
| github.com/getkin/kin-openapi (openapi3) | v0.142.0 | MIT | **Tests only:** loads and validates `api/openapi.yaml` and checks responses against its schemas (the S01.5-T03 contract test). Not linked into the binary (`go-licenses report` does not list it) | S01.5 | [ADR-0002](decisions/ADR-0002-api-style.md) | Version from the module graph (it is oapi-codegen v2.8.0's own OpenAPI library); license MIT (GitHub repository LICENSE). **In `go.mod`** as a direct requirement since S01.5-T03 |

**System tools for development, the scripts, and the API guide** (added in audit A002, F-011). They are not linked into the product or shipped with it. A release install needs none of them (section 12).

| Name | Version | License | Purpose | Stage | ADR | Verification status |
|---|---|---|---|---|---|---|
| Git | Any current version; used: 2.55.0 (Git for Windows) | GPL-2.0-only (tool only) | Clone and version control; the README's first requirement | S01+ | [ADR-0004](decisions/ADR-0004-repository-layout.md) | Verified 2026-09-24 (`git --version` on the development PC) |
| Docker Engine or Docker Desktop, with the Compose plugin | Used: Docker 29.7.2, Compose v5.5.1 | Apache-2.0 (Engine, Compose); Docker Desktop has its own license terms | **Optional** development environment (`deploy/compose.dev.yaml`, S01.1-T06) and the CI image job. Docker mode for deployment is section 12 (S11.1) | S01.1+ | [ADR-0006](decisions/ADR-0006-dev-environment-and-packaging.md) | Verified 2026-09-24 (`docker --version`, `docker compose version`; CI `image` job green) |
| Bash | 3.2 or later (macOS ships 3.2); used: 5.3 (Git Bash, WSL Ubuntu) | GPL-3.0-or-later (tool only) | Runs `scripts/*.sh` (lint install, coverage, licenses, docs offline check, performance baseline, demo). On Windows through Git Bash | S01+ | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Verified 2026-09-24 on bash 5.3 (Git Bash and WSL). `demo.sh` avoids bash 4 features for macOS, but was not run on macOS |
| POSIX utilities (`head`, `mktemp`, `tr`, `sed`, `grep`, `cut`, `base64`; `sha256sum`, or `shasum` on macOS) | System versions | Varies by platform: GNU coreutils GPL-3.0-or-later; uutils coreutils MIT (Ubuntu in WSL); BSD on macOS (tools only) | Used by `scripts/*.sh` and the guide's examples | S01+ | – | Verified 2026-09-24: Git Bash (GNU) and WSL Ubuntu (uutils 0.8.0) |
| curl | 7.87 or later for the guide's `--url-query`; `demo.sh` needs no newer features. Used: 8.21.0 (Git for Windows and Windows `curl.exe`), 8.18.0 (WSL Ubuntu) | curl license (MIT-style) | `docs/api/usage.md`, `scripts/demo.sh`, `scripts/demo.ps1` (`curl.exe`), and the CI health and demo checks | S01.7 | [ADR-0002](decisions/ADR-0002-api-style.md) | Verified 2026-09-24: every block of the guide and both demos run with these versions (S005 E119, E121) |
| Windows PowerShell 5.1 / PowerShell 7 | 5.1 (part of Windows 10 and 11); 7.x | Part of Windows (5.1); MIT (PowerShell 7) | `scripts/demo.ps1` and the guide's PowerShell examples | S01.7 | – | Verified 2026-09-24 on 5.1.26100 (development PC, and `windows-latest` in the CI `demo` job). PowerShell 7 was not run, but the script avoids 5.1-only behavior |
| Microsoft Edge or Google Chrome (headless) | Any current version | Proprietary (the Chromium base is BSD-3-Clause) | `scripts/check-api-docs-offline.sh`: renders `/api/docs/` with the network cut off. Manual only, not in CI | S01.5 | [ADR-0002](decisions/ADR-0002-api-style.md) | Verified 2026-09-24 with Microsoft Edge on the development PC (S005 E109) |

## 4. Web UI (npm)

| Name | Version | License | Purpose | Stage | ADR | Verification status |
|---|---|---|---|---|---|---|
| svelte | 5.57.1 | MIT | UI framework (bundled) | S02+ | [ADR-0009](decisions/ADR-0009-web-ui-sveltekit.md) | Verified 2026-09-24 (npm) **Installed** in `web/package.json` (S02.1-T01). |
| @sveltejs/kit | 2.70.3 | MIT | App framework | S02+ | [ADR-0009](decisions/ADR-0009-web-ui-sveltekit.md) | Verified 2026-09-24 **Installed** in `web/package.json` (S02.1-T01). |
| @sveltejs/adapter-static | 3.0.10 | MIT | Static SPA build | S02+ | [ADR-0009](decisions/ADR-0009-web-ui-sveltekit.md) | Verified 2026-09-24 **Installed** in `web/package.json` (S02.1-T01). |
| vite | 8.3.1 | MIT | Build tool and dev server (proxy to core) | S02+ | [ADR-0009](decisions/ADR-0009-web-ui-sveltekit.md) | Verified 2026-09-25 (npm; 8.3.0 in S003). **Installed** in `web/package.json` (S02.1-T01). |
| tailwindcss / @tailwindcss/vite | 4.3.3 | MIT | CSS utility framework (build time) | S02+ | [ADR-0009](decisions/ADR-0009-web-ui-sveltekit.md) | Verified 2026-09-24 **Installed** in `web/package.json` (S02.1-T01). |
| typescript | **6.0.3** (pinned; 7.0.2 is the latest but unsupported by svelte-check) | Apache-2.0 | Type checking | S02+ | [ADR-0009](decisions/ADR-0009-web-ui-sveltekit.md) | Verified 2026-09-24. svelte-check 4.7.6 declares `typescript: ^5.0.0 \|\| ^6.0.0` (audit A001 F-021) **Installed** in `web/package.json` (S02.1-T01). |
| @uppy/core, @uppy/tus | 6.0.1, 6.0.0 | MIT | Upload UI and tus client | S02.4, S04.7 | [ADR-0008](decisions/ADR-0008-resumable-uploads-tus.md) | Verified 2026-09-24 **Installed** in `web/package.json` (S02.4-T01). |
| openapi-typescript | 7.13.0 | MIT | Generate TypeScript types from the spec (dev) | S02.1 | [ADR-0002](decisions/ADR-0002-api-style.md) | Verified 2026-09-24 **Installed** in `web/package.json` (S02.1-T03). |
| openapi-fetch | 0.17.0 | MIT | Typed API client (bundled) | S02.1 | [ADR-0002](decisions/ADR-0002-api-style.md) | Verified 2026-09-24 **Installed** in `web/package.json` (S02.1-T03). |
| @tanstack/svelte-virtual | 3.13.39 | MIT | List and grid virtualization | S02.3, S04.7 | [ADR-0009](decisions/ADR-0009-web-ui-sveltekit.md) | Verified 2026-09-24. Svelte 5 fit to be prototyped in S02.3 **Installed** in `web/package.json` (S02.3-T01; the prototype confirmed it on Svelte 5). |
| @lucide/svelte | 1.48.0 | ISC | Bundled icons, one module per icon (FR-083, I6) | S02+ | [ADR-0009](decisions/ADR-0009-web-ui-sveltekit.md) | Verified 2026-09-25 (npm). **Installed** in `web/package.json` (S02.1-T04); named in `web/pnpm-workspace.yaml` (published a day before) |
| hls.js | 1.7.3 | Apache-2.0 | HLS player: manual quality menu + Auto (adaptive bitrate); MSE/ManagedMediaSource | S04.8 | [ADR-0020](decisions/ADR-0020-video-streaming-quality-levels.md) | Verified 2026-09-24 (npm) in S004 |
| redoc (standalone bundle, vendored) | 2.5.4 | MIT | Offline API documentation page served by the core | S01.5 | [ADR-0002](decisions/ADR-0002-api-style.md) | Verified 2026-09-24. **Vendored** since S01.5-T06 as `internal/api/docs/redoc.standalone.js`, taken from the npm tarball `redoc-2.5.4.tgz`, which matched npm's integrity `sha512-M6jWhG1qoBnH6TFmzJnstyCZ87HmOY/UzDm78mHiYihEdlV/YcS9ogOo1NlElnJMeLsyxHFe2yFc4sNjHTrABQ==` (sha1 c2b77b368800f56f69d43697d0fc45738fd8f466). The bundle's **SHA-256 is `dcaf76612bc4a3fbcc923a8966dee2f6146a5f32e5ce1b6f02dd60cbbf89500b`** (1,103,471 bytes); `TestAPIDocs` checks it (`api.RedocSHA256`). Redoc's LICENSE and the bundle's third-party license file are vendored next to it; `.gitattributes` keeps all three byte for byte |
| svelte-check | 4.7.6 | MIT | Svelte/TypeScript checks (dev) | S02+ | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Verified 2026-09-24 **Installed** in `web/package.json` (S02.1-T01). |
| eslint / eslint-plugin-svelte | 10.11.0 / 3.23.0 | MIT | Lint (dev) | S02+ | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Verified 2026-09-24 **Installed** in `web/package.json` (S02.1-T01). |
| prettier / prettier-plugin-svelte | 3.9.9 / 4.1.1 | MIT | Format (dev) | S02+ | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Verified 2026-09-24 **Installed** in `web/package.json` (S02.1-T01). |
| @sveltejs/vite-plugin-svelte | 7.3.1 | MIT | Svelte compiler integration for Vite (SvelteKit's peer; `vitePreprocess`) | S02+ | [ADR-0009](decisions/ADR-0009-web-ui-sveltekit.md) | Verified 2026-09-25 (npm). **Installed** in `web/package.json` (S02.1-T01). |
| @eslint/js / typescript-eslint / globals | 10.0.1 / 8.70.1 / 17.12.0 | MIT | ESLint's recommended rules, TypeScript rules and parser, and browser/Node globals for the flat config (dev). typescript-eslint accepts TypeScript `<6.1.0` | S02+ | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Verified 2026-09-25 (npm). **Installed** in `web/package.json` (S02.1-T01). |
| vitest | 5.0.1 | MIT | Unit and component tests (dev) | S02+ | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Verified 2026-09-24 |
| @playwright/test | 1.63.0 | Apache-2.0 | End-to-end tests (dev) | S02+ | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Verified 2026-09-24 **Installed** in `web/package.json` (S02.1-T04, early: by-hand browser checks with the installed Edge; tests in S02.8). |
| pdfjs-dist | 6.3.289 | Apache-2.0 | PDF previews | S02.6 | [ADR-0009](decisions/ADR-0009-web-ui-sveltekit.md) | **Candidate**: decided in S02.6 (vs. the browser's built-in viewer) |
| axe-core / @axe-core/playwright | 4.13.0 / 4.13.0 | ⚠ MPL-2.0 (file-level copyleft; dev and test tool only, not shipped) | Automated accessibility checks | S02.7 | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) (plan S02.7, 12.1) | Verified 2026-09-24 (npm) in A001. **Candidate** |

**License scan of the installed tree (S02.1-T01, `pnpm licenses list`):** 166 packages. All carry licenses on `scripts/allowed-licenses.txt` except **BlueOak-1.0.0** (`minimatch`, a permissive license, pulled in by the dev tools only). MPL-2.0 appears for `lightningcss` (Tailwind's CSS processor, build time). S02.1-T05 adds the check to CI: the packages that ship are held to the allow-list, and the dev tools to it plus BlueOak-1.0.0 (reason in `docs/licensing.md`).

## 5. Python (AI worker, S12)

| Name | Version | License | Purpose | Stage | ADR | Verification status |
|---|---|---|---|---|---|---|
| onnxruntime | 1.30.0 | MIT | Model inference (CPU; GPU providers evaluated in S12.1) | S12 | [ADR-0017](decisions/ADR-0017-ai-worker-architecture.md) | Verified 2026-09-24 (PyPI). Python 3.14 wheels (cp314) **verified** in A001 |
| numpy | 2.5.3 | BSD-3-Clause AND 0BSD AND MIT AND Zlib AND CC0-1.0 | Array math | S12 | [ADR-0017](decisions/ADR-0017-ai-worker-architecture.md) | Verified 2026-09-24 |
| scikit-learn | 1.9.1 | BSD-3-Clause | HDBSCAN face clustering | S12.5 | [ADR-0018](decisions/ADR-0018-ai-models.md) | Verified 2026-09-24 |
| opencv-python-headless | 5.0.0.93 | Apache-2.0 | YuNet/SFace helpers, if ONNX Runtime alone is not enough | S12.4–S12.5 | [ADR-0018](decisions/ADR-0018-ai-models.md) | **Candidate**: S12.1 decides |
| pillow | 12.3.0 | MIT-CMU | Image decoding in the worker | S12 | [ADR-0017](decisions/ADR-0017-ai-worker-architecture.md) | **Candidate** |
| rapidocr | 3.9.2 | Apache-2.0 | OCR, **only if S12.10 is approved** | S12.10 | [ADR-0018](decisions/ADR-0018-ai-models.md) | Verified 2026-09-24 (package). Model weights: partially verified in A001 (the README states that converted models are under the Apache License 2.0; the MODEL_LICENSE file was not found). Re-check at S12.10 |
| pytest | 9.1.1 | MIT | Tests (dev) | S12 | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Verified 2026-09-24 |
| ruff | 0.16.8 | MIT | Lint and format (dev) | S12 | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Verified 2026-09-24 |

## 6. External tools (run as separate programs; bundled in the Docker image)

| Name | Version (upstream / Debian trixie) | License | Purpose | Stage | ADR | Verification status |
|---|---|---|---|---|---|---|
| ExifTool | 13.59 / 13.25 | Artistic-1.0-Perl OR GPL-1.0-or-later ("same terms as Perl itself") | Metadata (EXIF, XMP, IPTC, maker notes, HEIC, RAW, video); RAW preview extraction | S04.4, S05.3 | [ADR-0012](decisions/ADR-0012-media-toolchain.md) | Verified 2026-09-24 (exiftool.org/ver.txt, README; Debian sources) |
| libvips (`vips`, `vipsthumbnail`) | 8.18.6 / 8.16.1 | LGPL-2.1 | Thumbnails and previews | S04.4 | [ADR-0012](decisions/ADR-0012-media-toolchain.md) | Verified 2026-09-24 (LICENSE; Debian sources) |
| libheif (+ libde265 decoder plugin) | libheif 1.23.5 / 1.19.8; libde265 1.0.15-1+deb13u2 (trixie) | ⚠ libheif: LGPL (library), MIT (samples, wrappers). libde265: LGPL-3+ (library); Debian copyright also lists GPL-3+ and BSD-4-clause parts | HEIC decoding (via libvips). **No x265 encoder plugin** | S04.4 | [ADR-0012](decisions/ADR-0012-media-toolchain.md) | Verified 2026-09-24 (COPYING; Debian sources and libde265 debian/copyright, audit A001 F-023) |
| FFmpeg / ffprobe | 9.0.2 / 7.1.5 (trixie) | ⚠ **The Debian trixie build is GPL-3.0-or-later** (`--enable-gpl --enable-version3`; not `--enable-nonfree`). Upstream default is LGPL-2.1+ | Video poster frames and metadata; **HLS transcoding for streaming quality levels** (since 0.4.0) | S04.4, S05.3, S04.8 | [ADR-0012](decisions/ADR-0012-media-toolchain.md), [ADR-0020](decisions/ADR-0020-video-streaming-quality-levels.md) | Verified 2026-09-24 (ffmpeg.org/legal; Debian `debian/rules` for 7:7.1.5-0+deb13u1, audit A001 F-022). **D-04 decided in S005:** Debian's GPL build is accepted as a separate program for now, with a source offer in third-party notices; revisit at S11.1 |
| x264 / libx264 (inside Debian's FFmpeg build) | 2:0.164.3108 (Debian trixie) | ⚠ GPL-2.0-or-later (per upstream; not fetched) | H.264 CPU encoding for streaming levels | S04.8 | [ADR-0020](decisions/ADR-0020-video-streaming-quality-levels.md) | Version and build flag verified 2026-09-24 (sources.debian.org; `--enable-libx264` in FFmpeg `debian/rules`). The FFmpeg build also enables libvpl (Intel QSV). VAAPI, V4L2-M2M, and NVENC availability **unverified** (autodetect; check in S04.8) |
| debian:trixie-slim (base image) | Debian 13 | Various free-software licenses (DFSG) | Runtime base image | S01.1 (dev), S11.1 | [ADR-0006](decisions/ADR-0006-dev-environment-and-packaging.md) | Package versions verified 2026-09-24 (sources.debian.org). Digest pinned at image build |
| golang:1.27.1-trixie (build image) | Go 1.27.1 on Debian 13 | BSD-3-Clause (Go) + Debian packages (DFSG) | Build stage of `deploy/Dockerfile.dev`, and the `dev` stage that runs `go run` in the compose dev environment. Not in the runtime image | S01.1-T06 | [ADR-0006](decisions/ADR-0006-dev-environment-and-packaging.md) | Tag verified on Docker Hub 2026-09-24 (pushed 2026-09-19; amd64 and arm64/v8). Updated by Dependabot (`docker`, `/deploy`) |
| Trivy | v0.74.0 | Apache-2.0 | Container image vulnerability scanning (CI) | S01.1+ | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Verified 2026-09-24 (GitHub release) |
| Docker Engine (moby) | host-installed | Apache-2.0 | Container runtime for deployment and the dev environment | S01.1+, S11 | [ADR-0006](decisions/ADR-0006-dev-environment-and-packaging.md) | License verified 2026-09-24 (LICENSE file) in A001. Version: user's host |
| Docker Compose | host-installed | Apache-2.0 | Primary deployment method, including the `ai` profile | S01.1+, S11 | [ADR-0006](decisions/ADR-0006-dev-environment-and-packaging.md) | License verified 2026-09-24 (LICENSE file) in A001 |
| Docker buildx | host / CI | Apache-2.0 | Multi-architecture image builds (amd64, arm64) | S01.1+, S11 | [ADR-0006](decisions/ADR-0006-dev-environment-and-packaging.md) | License verified 2026-09-24 (LICENSE file) in A001 |
| LibRaw | not chosen | LGPL-2.1 OR CDDL-1.0 | Full RAW conversion (**deferred**) | Later | [ADR-0012](decisions/ADR-0012-media-toolchain.md) | License files verified 2026-09-24 in A001. **Deferred** |
| Samba | 4.22.11 (Debian trixie) | ⚠ GPL-3.0 | SMB network shares (**deferred**, Linux-only, optional) | S09.1 | [ADR-0019](decisions/ADR-0019-smb-via-samba.md) (Proposed) | Version verified (sources.debian.org) and license verified (samba.org GPL page), 2026-09-24, in A001. **Deferred** |

## 7. Datasets

| Name | Version | License | Purpose | Stage | ADR | Verification status |
|---|---|---|---|---|---|---|
| GeoNames `cities500.zip` (+ `admin1CodesASCII.txt`, `countryInfo.txt`) | Export date recorded at bundle time; 13.9 MB zipped (2026-09-24) | CC BY 4.0: **attribution required** (docs + About page) | Offline reverse geocoding | S05.4 | [ADR-0013](decisions/ADR-0013-reverse-geocoding-geonames.md) | Verified 2026-09-24 (download.geonames.org readme.txt, file sizes) |
| Synonym dictionary seed | TBD | TBD | Query-time synonym expansion | S06.5 | [ADR-0014](decisions/ADR-0014-search-engine-bleve.md) | **Unverified**: the source and license are chosen in S06.5 |

## 8. AI models (S12; exact variants deferred to S12 benchmarks)

| Name | Version | License | Purpose | Stage | ADR | Verification status |
|---|---|---|---|---|---|---|
| SigLIP (e.g. `google/siglip-base-patch16-224` or `google/siglip2-base-patch16-224`), exported to ONNX | Variant chosen in S12 | Apache-2.0 | Zero-shot photo classification; image embeddings | S12.3 | [ADR-0018](decisions/ADR-0018-ai-models.md) | License verified 2026-09-24 (Hugging Face model cards) |
| YuNet (OpenCV Zoo `face_detection_yunet`) | Variant chosen in S12 | MIT | Face detection | S12.4 | [ADR-0018](decisions/ADR-0018-ai-models.md) | License verified 2026-09-24 (OpenCV Zoo LICENSE/README). Training-data note in ADR-0018 |
| SFace (OpenCV Zoo `face_recognition_sface`) | Variant chosen in S12 | Apache-2.0 | Face embeddings | S12.5 | [ADR-0018](decisions/ADR-0018-ai-models.md) | License verified 2026-09-24 (OpenCV Zoo LICENSE/README). Training-data note in ADR-0018 |
| RapidOCR / PaddleOCR ONNX models | Chosen at S12.10 | Apache-2.0 (README statement) | OCR, only if S12.10 is approved | S12.10 | [ADR-0018](decisions/ADR-0018-ai-models.md) | **Partially verified** in A001 (README statement only). Deferred to S12.10 (F-032) |
| ONNX Runtime GPU execution providers (e.g. CUDA, OpenVINO packages) | chosen at S12.1 | MIT (onnxruntime family; per-package check at S12.1) | Optional GPU acceleration | S12.1 | [ADR-0017](decisions/ADR-0017-ai-worker-architecture.md) | **Deferred** (S12.1) |
| sqlite-vec | chosen at S12 if needed | MIT OR Apache-2.0 | Vector search for embeddings, **only if** brute-force cosine is too slow | S12 | [ADR-0018](decisions/ADR-0018-ai-models.md) | License files verified 2026-09-24 in A001. **Candidate.** Note: it is a C extension, so it is likely unusable from the pure-Go SQLite driver (ADR-0007); evaluate in the Python worker at S12 |

## 9. Services

| Name | Version | License | Purpose | Stage | ADR | Verification status |
|---|---|---|---|---|---|---|
| GitHub Actions | n/a | Service (GitHub terms) | CI (Linux + Windows; macOS for S09 client tests) | S01.1+ | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Repository hosted on GitHub (origin verified in S001) |
| Dependabot | n/a | Service (GitHub terms) | Automated dependency update PRs | S01.1+ | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Agent choice (vs. Renovate) |
| GitHub Action actions/checkout | v7.0.1 | MIT | CI checkout | S01.1-T05 | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Version verified 2026-09-24 (GitHub releases) in A001; license verified 2026-09-24 (GitHub license API) in S005. **In use** in `.github/workflows/ci.yml` (S01.1-T05) |
| GitHub Action actions/setup-go | v7.0.0 | MIT | CI Go toolchain (`go-version-file: go.mod`) | S01.1-T05 | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Version verified 2026-09-24 in A001; license verified 2026-09-24 (GitHub license API) in S005. **In use** (S01.1-T05) |
| GitHub Action golangci/golangci-lint-action | v9.3.0 | MIT | CI lint (runs golangci-lint v2.13.2 binary) | S01.1-T05 | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Version verified 2026-09-24 in A001; license verified 2026-09-24 (GitHub license API) in S005. **In use** (S01.1-T05) |
| GitHub Action docker/setup-buildx-action | v4.4.1 | Apache-2.0 | CI image builds (dev image now; multi-arch in S11.1) | S01.1-T06 (moved from T05) | [ADR-0006](decisions/ADR-0006-dev-environment-and-packaging.md) | Version verified 2026-09-24 in A001; release re-checked and license verified (LICENSE file at the tag) 2026-09-24 in S005. **In use** (`image` job, S01.1-T06) |
| GitHub Action aquasecurity/trivy-action | v0.36.0 | Apache-2.0 | CI image scan (Trivy; the action's default Trivy is v0.70.0, the workflow pins v0.74.0) | S01.1-T06 (moved from T05) | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Version verified 2026-09-24 in A001; release re-checked and license verified (LICENSE file at the tag) 2026-09-24 in S005. **In use** (`image` job, S01.1-T06) |
| GitHub Action actions/upload-artifact | v7.0.1 | MIT | Keep the cross-built binaries of each CI run (7 days) | S01.1-T05 | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | **New in S01.1-T05.** Latest release v7.0.1 (2026-04-10) and license MIT verified 2026-09-24 (GitHub API) in S005. **In use** |
| GitHub Action actions/download-artifact | v8.0.1 | MIT | The build job takes the web interface built by the `web` job | S02.1-T05 | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Version verified 2026-09-25 (GitHub releases, latest). In `.github/workflows/ci.yml` |
| GitHub Action pnpm/action-setup | v6.1.0 | MIT | Installs pnpm from `web/package.json` `packageManager` and caches its store | S02.1-T05 | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Version verified 2026-09-25 (GitHub releases, latest). In `.github/workflows/ci.yml` |
| GitHub Action actions/setup-node | v7.0.0 | MIT | Node.js 24.21.0 for the `web` job | S02.1-T05 | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Version verified 2026-09-25 (GitHub releases, latest). In `.github/workflows/ci.yml` |

## 10. Evaluated, not adopted, or rejected

| Name | Version | License | Why not adopted | ADR | Verification status |
|---|---|---|---|---|---|
| github.com/barasher/go-exiftool | v1.10.0 (2023-06-07) | ⚠ GPL-3.0 | **Rejected**: GPL-3.0 would apply to the linked core binary, and there has been no release since 2023. A custom wrapper is used instead | [ADR-0012](decisions/ADR-0012-media-toolchain.md) | Verified 2026-09-24 |
| github.com/go-chi/chi/v5 | v5.3.2 | MIT | Evaluated, not adopted: the stdlib ServeMux is sufficient. Documented fallback | [ADR-0002](decisions/ADR-0002-api-style.md) | Verified 2026-09-24 |
| github.com/mattn/go-sqlite3 | v1.14.52 | MIT | Evaluated, not adopted: needs cgo. Fallback only via a new ADR after benchmarks | [ADR-0007](decisions/ADR-0007-database-sqlite.md) | Verified 2026-09-24 |
| maragu.dev/goqite | v0.4.0 | MIT | Evaluated, not adopted: no priorities, per-type concurrency, or progress | [ADR-0011](decisions/ADR-0011-job-queue-sqlite.md) | Verified 2026-09-24 |
| github.com/BurntSushi/toml | v1.6.0 | MIT | Evaluated, not adopted: an equal alternative to go-toml/v2 | [ADR-0001](decisions/ADR-0001-backend-language-framework.md) | Verified 2026-09-24 |
| github.com/stretchr/testify | v1.12.1 | MIT | Not adopted: stdlib testing + go-cmp chosen | [ADR-0005](decisions/ADR-0005-testing-linting-ci.md) | Verified 2026-09-24 |
| swagger-ui-dist / @scalar/api-reference | 5.33.0 / 1.71.0 | Apache-2.0 / MIT | Not adopted: Redoc chosen for the offline docs page | [ADR-0002](decisions/ADR-0002-api-style.md) | Verified 2026-09-24 |
| InsightFace pretrained models | — | Non-commercial research only | **Excluded by policy** (NFR-029) | [ADR-0018](decisions/ADR-0018-ai-models.md) | Per P003 and S001 research |

## 11. Alternatives named in ADRs (not adopted; not verified)

These are listed so the register covers every technology named in the ADRs (audit A001, F-024). They are **not** dependencies, and no version or license was checked.

| Name | Named in | Why not adopted (see ADR) |
|---|---|---|
| Rust, Node.js/TypeScript, Python/FastAPI as core language | ADR-0001 | Go chosen by the user (P003) |
| GraphQL, gRPC/Connect, WebDAV as primary API | ADR-0002 | REST + OpenAPI chosen |
| Separate repositories, multi-module go.work | ADR-0004 | Single repository chosen |
| Renovate | ADR-0005 | Dependabot chosen |
| PostgreSQL | ADR-0007 | SQLite chosen |
| Custom chunked uploads, S3-style multipart | ADR-0008 | tus chosen |
| React + Vite, Electron/Tauri | ADR-0009 | SvelteKit web UI chosen |
| bcrypt, scrypt, JWT sessions | ADR-0010 | Argon2id and server-side sessions chosen |
| Redis/broker-based queues | ADR-0011 | SQLite queue chosen |
| cgo bindings (govips, libheif-go), Go-native metadata libraries | ADR-0012 | Subprocess tools chosen |
| Natural Earth / OSM boundaries, online geocoders (Nominatim) | ADR-0013 | GeoNames chosen (Natural Earth remains an optional later improvement) |
| SQLite FTS5, Meilisearch, Tantivy | ADR-0014 | Bleve chosen |
| In-process Go SMB server | ADR-0019 | Samba direction (deferred) |
| fanotify / OS-specific watchers | ADR-0016 | fsnotify chosen |
| PyTorch, onnxruntime-go via cgo | ADR-0017 | ONNX Runtime in a Python worker chosen |
| OpenCLIP (family member), fixed-label CNNs, captioning VLMs, SCRFD/RetinaFace, ArcFace/InsightFace, DBSCAN, Chinese Whispers, Tesseract | ADR-0018 | SigLIP-family, YuNet, SFace, HDBSCAN, RapidOCR chosen; InsightFace excluded by policy |

## 12. Deployment prerequisites per platform (input for the setup scripts)

**Why:** the user's requirement (S005 E015): "keep record of all dependencies needed, in the end you will have to make a setup script, a separate one for each platform, which when run, will automatically handle the deployment." The setup scripts are built in S11.2 (FR-149). This section is their source list (NFR-032).

**Rules:**
- Every stage that adds something the running NAS needs on the target machine adds or updates its row here **in the same commit** (R6).
- Package names are only marked verified after they were checked against the platform's package index. Everything else says when it is verified.
- Build-only tools (Go, Node.js, pnpm, linters) are **not** needed on the target when a release binary is installed. They are listed in sections 1 and 3, and at the end of this section for from-source installs.

**Platforms** (Q1, answered in S005): **Linux x86-64** (mini-PC or old PC; Debian 13 or a current Ubuntu LTS), **Raspberry Pi** (Raspberry Pi OS 64-bit, ARM64; 32-bit is not supported, NFR-030), **Windows 11** (x86-64; also the test PC). **Docker mode**: the host needs only Docker Engine and Compose; everything else is inside the image (Q41 decides the Linux default).

| Prerequisite | Needed for | First needed | Linux x86-64 | Raspberry Pi (ARM64) | Windows 11 | Docker image | Status |
|---|---|---|---|---|---|---|---|
| local-ai-nas binary (pure Go, `CGO_ENABLED=0`; SQLite compiled in) | The whole NAS | S01 | Release binary `linux/amd64` | Release binary `linux/arm64` | Release binary `windows/amd64` (`.exe`) | Built into the image | Built from source until release binaries exist (S11.7). **S01 needs nothing else on the target** |
| System service manager | Running as a service (FR-132) | S11.2 | systemd (part of the OS) | systemd (part of the OS) | Windows Service Control Manager (part of the OS); service support inside the binary | Docker restart policy | Unit file and service registration are written in S11.2 |
| ExifTool (+ Perl) | Photo and video metadata | S04.4, S05.3 | apt `libimage-exiftool-perl` (pulls in Perl) | apt `libimage-exiftool-perl` | Windows build from exiftool.org (bundles Perl) | Bundled (Debian package) | Debian package: section 6. Windows source: verify in S11.2 |
| libvips command-line tools (`vips`, `vipsthumbnail`) | Thumbnails and previews | S04.4 | apt `libvips-tools` | apt `libvips-tools` | Official libvips Windows binaries | Bundled (Debian package) | Debian package name: verify in S04.4. Windows source: verify in S11.2 |
| libheif + libde265 decoder plugin | HEIC decoding (via libvips) | S04.4 | apt `libheif1` + `libheif-plugin-libde265` | same | Included in the libvips Windows build (check) | Bundled | Package names: verify in S04.4 (Q26) |
| FFmpeg + ffprobe | Video poster frames, metadata, streaming quality levels | S04.4, S04.8 | apt `ffmpeg` | apt `ffmpeg` | A Windows FFmpeg build (source chosen in S11.2) | Bundled (Debian build, GPL, D-04) | Debian package: section 6. Windows source: verify in S11.2 |
| Hardware video encoder drivers (optional) | Faster transcoding (FR-148) | S04.8 | VAAPI / Intel QSV drivers | V4L2 M2M (if the model supports it) | Vendor GPU drivers | Device passthrough | Optional. Verified in S04.8 |
| GeoNames data (`cities500`) | Offline place names | S05.4 | Shipped with the NAS | Shipped with the NAS | Shipped with the NAS | Bundled | No install step. Attribution required (section 7) |
| Docker Engine + Compose plugin | **Docker mode only** | S11.1 | Docker's apt repository (`docker-ce`, `docker-compose-plugin`) | same (arm64) | Docker Desktop (only if Docker mode is chosen on Windows) | n/a | Mode default per Q41. Verify in S11.1 |
| Python 3.14 + uv + AI worker packages (section 5) | **Optional AI** (S12) | S12 | AI worker install (S12.1) | AI worker install (S12.1) | AI worker install (S12.1) | `ai` Compose profile | Decided in S12.1 |

**From-source installs only** (not needed with release binaries): Go go1.27.1 (S01+), Node.js LTS + pnpm (S02+, to build the web UI). See sections 1 and 3.
