# ADR-0004: Repository layout

| Field | Value |
|---|---|
| Number | ADR-0004 |
| Status | **Accepted** |
| Date proposed | 2026-09-24 (session S003) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none |
| Superseded by | none |

## Context

S01.1 creates the repository structure. P003 gives a proposed layout as the starting point and asks the agent to refine it and record the reasoning. Folders are created only when their stage begins. Constraints:
- A Go core (ADR-0001), a SvelteKit UI embedded into the binary (ADR-0009), an optional Python AI worker (ADR-0017), and an OpenAPI contract (ADR-0002).
- Docker packaging (ADR-0006).

## Options considered

### Option A: Single repository with a Go module at the root (P003 layout, refined) (chosen)
- **Pros:** one clone; the Go module sees `web/` for `go:embed`; one CI; the contract lives next to the code.
- **Cons:** several toolchains (Go, pnpm, uv) in one tree, which is mitigated by per-folder tooling.

### Option B: Separate repositories (core, web, ai-worker)
- **Cons:** cross-repository versioning of the API contract; embedding the UI needs artifact passing. Too much overhead for one maintainer.

### Option C: Monorepo with multiple Go modules / go.work
- **Cons:** unneeded complexity. There is only one Go program.

## Decision

A single repository with the Go module at the root (module path `github.com/KhizirFarrukh/local-ai-nas`), using the P003 layout with these refinements:

```
local-ai-nas/
├── go.mod, go.sum                  Go module; tool directives pin oapi-codegen, govulncheck, go-licenses (golangci-lint is pinned as a binary; see below)
├── cmd/local-ai-nas/               main package: flags, config load, wiring, server start
├── internal/                       Go packages; each created when its stage begins
│   ├── config/  logging/  apperr/  health/          (S01.1–S01.2)
│   ├── storage/ (root layout, namespaces, resolver, locks, free space)   (S01.2, S01.6)
│   ├── files/   (FilesService + local backend + hooks)                   (S01.3)
│   ├── uploads/ (tusd integration, finalize, cleanup)                    (S01.4)
│   ├── api/     (handlers) + api/gen/ (oapi-codegen output, committed)   (S01.5)
│   ├── db/      (SQLite open, pragmas) + db/migrations/*.sql (embedded)  (S01.1/S01.2)
│   ├── auth/  policy/  audit/                                            (S03)
│   ├── photos/  jobs/  media/  transfer/                                 (S04)
│   ├── sidecar/  geocode/  reconcile/                                    (S05)
│   ├── search/                                                           (S06)
│   ├── users/  sharing/                                                  (S07)
│   ├── trash/  backup/  integrity/                                       (S08)
│   ├── webdav/  watch/                                                   (S09)
│   └── admin/  quota/                                                    (S10)
├── api/openapi.yaml                the API contract (ADR-0002)
├── web/                            SvelteKit UI (S02) + web/embed.go (package web, //go:embed all:build)
├── ai-worker/                      Python AI worker (S12): pyproject.toml, uv.lock, src/, tests/
├── deploy/                         Dockerfile(s), compose.yaml, compose.dev.yaml, systemd unit (S11), dev files from S01.1
├── testdata/                       test fixtures + testdata/SOURCES.md (origin and license of every fixture)
├── docs/                           user, admin, install documentation (+ third-party notices)
├── scripts/                        development helper scripts (cross-platform: Go or POSIX shell + PowerShell pairs)
├── .github/workflows/, .github/dependabot.yml
├── code-agent-docs/                agent planning and memory (exists)
└── README.md, LICENSE, AGENTS.md, CLAUDE.md, .gitattributes, .editorconfig, .gitignore
```

### Implementation details chosen by agent
| Refinement | Reason |
|---|---|
| `web/embed.go` inside `web/` | `go:embed` can only embed files at or below the directory of the Go source file, so the embed package must live in `web/` |
| Generated code committed under `internal/api/gen/` | Builds work without generators. CI checks drift (ADR-0002) |
| SQL migrations under `internal/db/migrations/` and embedded | The binary carries its own migrations (ADR-0007) |
| Tools pinned with Go 1.24+ `tool` directives in `go.mod` (oapi-codegen, govulncheck, go-licenses). **golangci-lint is the exception:** it is pinned as a **binary** (v2.13.2) through the official GitHub Action in CI and a documented binary install locally | One lockfile (`go.sum`) for most dev tools, and "pin every version" (P003). golangci-lint's official install docs warn that `go install`, the tools pattern, and `tool` directives "aren't guaranteed to work" and recommend binary installation (verified S003 E011) |
| **No task runner** (no Makefile, Taskfile, or just). Commands are documented plain `go`, `pnpm`, and `uv` invocations, and CI runs the same commands | Fewer moving parts. `make` is not native on Windows, which is the developer's platform |
| `testdata/SOURCES.md` | P003 requires fixture license records |
| Package boundaries enforce I1 | `internal/files` and `internal/photos` never import each other. Only `internal/transfer` imports both, checked by an architecture test (depguard rule in golangci-lint) |

## Consequences

- **Easier:** a single CI and clone; embedding the UI; clear per-stage package creation.
- **Harder:** three toolchains in one repository (Go always; pnpm from S02; uv from S12).
- **Required:** the depguard/architecture rule (S01.1-T05); `.gitattributes` for line endings (S001 finding).

## Approval record

> Decision made by the user in plan change request #3 (`code-agent-docs/prompts/P003-technology-stack.json`, `proposed_repository_layout`, status "Accepted (layout above, refined by agent)"). Refinements are listed above for the user to object to
> (2026-09-24, session S003).

## Links

- **Related requirements:** NFR-014, NFR-025, NFR-026
- **Related ADRs:** ADR-0001, ADR-0002, ADR-0005, ADR-0006, ADR-0007, ADR-0009, ADR-0017
- **Related stages:** S01.1 onward
- **Plan version:** 0.3.0
