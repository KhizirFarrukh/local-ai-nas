# Licensing

## Project license

local-ai-nas is licensed under the **GNU Affero General Public License v3.0 or later** (SPDX: `AGPL-3.0-or-later`). The full text is in [`LICENSE`](../LICENSE). The project owner chose AGPL-3.0 in session S005 (plan question Q22) and the "or later" form in the same session.

In short: anyone may use, study, change, and share the software. Anyone who distributes it, or **lets users interact with a modified version over a network**, must offer those users the complete corresponding source code under the same license.

The root `LICENSE` covers every file in this repository unless a file says otherwise. Vendored third-party files (for example the Redoc bundle, S01.5) keep their own license notices, and they are listed in the dependency register.

## Dependency license policy (NFR-029)

Every dependency, external tool, dataset, and AI model must have a license that lets **anyone deploy and use** this project, including commercially. Licenses restricted to non-commercial or research use are never allowed. Every dependency is recorded with its license in [`code-agent-docs/dependencies.md`](../code-agent-docs/dependencies.md) (RULES R6).

### Code linked into the binary or bundled into the web UI

These licenses are allowed. The machine-readable list for the Go check is [`scripts/allowed-licenses.txt`](../scripts/allowed-licenses.txt).

| License | Why it is allowed |
|---|---|
| MIT, BSD-2-Clause, BSD-3-Clause, ISC | Permissive. Compatible with AGPL-3.0; only the notice must be kept. |
| Apache-2.0 | Permissive with a patent grant. Compatible with (A)GPL-3.0 (not with GPL-2.0-only, which is irrelevant here). |
| MPL-2.0 | File-level copyleft. Compatible with the GPL family through MPL section 3.3. Changes to MPL files stay under MPL-2.0. |
| LGPL-2.1, LGPL-3.0 | Weak copyleft. The whole program's source is available under the AGPL, so the relinking requirement of static Go linking is met. |
| GPL-3.0, AGPL-3.0 | Strong copyleft of the same family. GPL-3.0 section 13 and AGPL-3.0 section 13 explicitly allow combining the two. |

Not allowed in linked or bundled code:
- **GPL-2.0-only**: incompatible with (A)GPL-3.0. go-licenses reports GPL-2.0 without telling "-only" and "-or-later" apart, so GPL-2.0 is not on the list. A GPL-2.0-or-later library needs a manual review and a recorded exception here.
- Non-commercial, research-only, or field-of-use licenses (for example CC BY-NC, the InsightFace model terms), and source-available licenses (SSPL, BUSL, Elastic License).
- Code with no license or an unrecognized license.
- Public-domain dedications (Unlicense, CC0) are not on the list yet. Adding one needs a reason recorded here.

### External programs (run as separate processes)

ExifTool, libvips, libheif, and FFmpeg are not linked into the binary. The core runs them as separate programs, so their licenses do not extend to the core. The Docker image ships them, so their license texts and source offers go into the image's third-party notices (S11.1). The Debian FFmpeg build is GPL-3.0-or-later (decision D-04, S005). Their licenses are listed in `dependencies.md` section 6.

### Development and CI tools

Tools that are never linked or distributed may have copyleft licenses. Examples: golangci-lint (GPL-3.0) and axe-core (MPL-2.0, tests only).

### Datasets and AI models

Datasets and models follow the same "anyone may deploy and use" rule. Attribution licenses are allowed, and the attribution is shown in the documentation and the About page. For example, GeoNames data is CC BY 4.0 (S05.4).

## How the policy is checked

- **Go:** [`scripts/check-licenses.sh`](../scripts/check-licenses.sh) runs `go tool go-licenses check ./...` with the allow-list. It fails on any linked package with a license outside the list or with no recognized license. CI runs it on every pull request (S01.1-T05). On Windows, run it from Git Bash.
- **Web UI (from S02):** `pnpm licenses` with the same allow-list (ADR-0005).
- **Release (S11.6):** a full audit of dependencies, external tools, datasets, and models.

## Adding a dependency

1. Check its license against the lists above. If it is not allowed, do not add it.
2. Record it in `code-agent-docs/dependencies.md` with version, license, purpose, and verification, **in the same commit** (RULES R6). Anything the running NAS needs on the target machine also gets a row in section 12 (per-platform prerequisites, NFR-032).
3. Run `scripts/check-licenses.sh`.
