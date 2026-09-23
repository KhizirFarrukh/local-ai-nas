# ADR-0005: Testing, linting, and CI toolchain

| Field | Value |
|---|---|
| Number | ADR-0005 |
| Status | **Accepted** |
| Date proposed | 2026-09-24 (session S003) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none |
| Superseded by | none |

## Context

NFR-014 requires tests, linting, formatting, and CI on every stage. NFR-023 requires security scanning, and NFR-013 license checks. P003 fixes most tools and leaves three choices to the agent: testify or not; Dependabot or Renovate; and CI host confirmation (GitHub, since the repository is hosted there).

## Options considered

| Area | Options | Chosen |
|---|---|---|
| Go assertions | standard `testing` only; + `go-cmp`; + testify | standard `testing` + **go-cmp** |
| Dependency updates | Dependabot; Renovate | **Dependabot** |
| Go license check | go-licenses; manual review | **go-licenses v2** |
| Web license check | `pnpm licenses list`; third-party checker | **`pnpm licenses list`** (built in) |

## Decision

| Area | Tool (verified 2026-09-24) | License | Notes |
|---|---|---|---|
| Go tests | Standard `testing` package; `go test -race` in CI (Linux) | BSD-3-Clause | |
| Go comparisons | `github.com/google/go-cmp` **v0.7.0** | BSD-3-Clause | Agent decided: readable diffs without testify's assertion DSL. testify v1.12.1 (MIT) was not needed ("prefer the standard library") |
| Go lint and format | **golangci-lint v2.13.2** (includes gofmt/goimports formatters, govet, staticcheck, gosec, depguard) | GPL-3.0 (**dev tool only**, never linked or distributed) | Pinned as a **binary**: the official `golangci/golangci-lint-action` with `version: v2.13.2` in CI, and the upstream binary install script with the same version locally. **Not** a `go.mod` tool directive, per upstream guidance (install docs: tool directives "aren't guaranteed to work") |
| Go vulnerabilities | **govulncheck** (`golang.org/x/vuln` **v1.8.0**) | BSD-3-Clause | Runs on every PR |
| Go dependency licenses | **go-licenses v2.0.1** (`github.com/google/go-licenses/v2`) | Apache-2.0 | Fails on disallowed licenses (policy from Q22 and NFR-029) |
| Web unit and component tests | **Vitest 5.0.1** | MIT | From S02 |
| Web end-to-end tests | **Playwright 1.63.0** (`@playwright/test`) | Apache-2.0 | From S02 |
| Web checks | **svelte-check 4.7.6**, **ESLint 10.11.0** (+ eslint-plugin-svelte 3.23.0), **Prettier 3.9.9** (+ prettier-plugin-svelte 4.1.1) | MIT | From S02 |
| Web audit and licenses | `pnpm audit`, `pnpm licenses list` | (pnpm MIT) | From S02 |
| Python | **pytest 9.1.1**, **Ruff 0.16.8** (lint + format) | MIT | From S12 |
| Container scanning | **Trivy v0.74.0** | Apache-2.0 | From the first image build (S01.1 dev image, S11 release images) |
| Dependency updates | **Dependabot** (ecosystems: gomod, npm, docker, github-actions; pip/uv for `ai-worker/` in S12) | GitHub service | Agent decided: native to GitHub, no app installation, config in the repository |
| CI | **GitHub Actions**, jobs on `ubuntu-latest` and `windows-latest` from S01.1 (macOS added for S09 client tests) | GitHub service | The repository is on GitHub (`origin`); confirms Q24 |
| Fixtures | `testdata/` with `testdata/SOURCES.md` recording the source and license of every fixture. Generated fixtures preferred | — | P003 |

## Consequences

- **Easier:** one lint entry point (golangci-lint), security and license gates on every PR, and automated update PRs.
- **Harder:** GPL-licensed golangci-lint must stay a development tool. It is never vendored into release artifacts (recorded in the dependency register).
- **Required:** CI workflow (S01.1-T07); Dependabot config (S01.1-T07); `testdata/SOURCES.md`; the license allow-list (after Q22).

## Approval record

> Decision made by the user in plan change request #3 (`code-agent-docs/prompts/P003-technology-stack.json`, `technology_stack.testing_and_quality`). testify vs go-cmp and Dependabot vs Renovate are "agent decides" items (listed for objection)
> (2026-09-24, session S003). Verified in S003 log E005/E007.

## Links

- **Related requirements:** NFR-013, NFR-014, NFR-023, NFR-029
- **Related ADRs:** ADR-0001, ADR-0004, ADR-0006, ADR-0009, ADR-0017
- **Related stages:** S01.1 onward (web from S02, Python from S12)
- **Plan version:** 0.3.0
