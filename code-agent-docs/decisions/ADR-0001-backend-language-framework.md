# ADR-0001: Core server language: Go

| Field | Value |
|---|---|
| Number | ADR-0001 |
| Status | **Accepted** |
| Date proposed | 2026-09-24 (session S002) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none |
| Superseded by | none |

> **History:** first drafted in S002 as *Proposed*, recommending Python/FastAPI. Updated in S003 to the user's decision (P003). The earlier recommendation is kept under "Options considered". Because the ADR was never Accepted before, updating it in place is allowed (R5).

## Context

The core server is a long-running NAS process. Its work is mostly disk and network I/O: streaming uploads and downloads with ranges, resumable uploads (ADR-0008), directory operations, the job queue (ADR-0011), the embedded search index (ADR-0014), and WebDAV (ADR-0015). Constraints:
- NFR-004: runs on modest hardware, including arm64 boards.
- NFR-009: amd64 and arm64.
- NFR-021: streaming I/O.
- P003 guiding principles: few moving parts (one core binary), prefer the standard library, and call external media tools as subprocesses instead of cgo.

## Options considered

### Option A: Go (chosen)
- **Pros:**
  - Excellent streaming I/O and concurrency.
  - A single static binary with low memory use.
  - Trivial cross-compilation for linux/amd64 and linux/arm64 (e.g. Raspberry Pi).
  - A strong standard library (`net/http`, `io`, `crypto/tls`, `log/slog`).
  - Pure-Go builds (`CGO_ENABLED=0`).
- **Cons:**
  - A second language (Python) for the optional AI worker (ADR-0017).
  - Fewer native imaging and metadata libraries, which is mitigated by external tools (ADR-0012).
- **License:** Go is BSD-3-Clause.

### Option B: Rust
- **Pros:** fastest; memory-safe; the Tantivy search library.
- **Cons:** much slower development; CPU speed is not the bottleneck for a NAS.

### Option C: Node.js / TypeScript
- **Pros:** one language across the whole stack.
- **Cons:** a heavier runtime; weaker for streaming large files; harder to package as one artifact.

### Option D: Python 3.12+ with FastAPI (the S002 Proposed recommendation)
- **Pros:** the best AI ecosystem; shared models with the AI worker; fast development.
- **Cons:** a weaker fit for a long-running file server (memory, packaging, CPU-bound work). Under this decision, Python is used **only** for the AI worker.

## Decision

**Go** is the core server language, per the user's decision (P003).
- **Version:** the latest stable release at setup time, **go1.27.1** at verification (2026-09-24, go.dev/dl). It is pinned in `go.mod` with a `go 1.27` line plus a `toolchain go1.27.1` line. Developer tools are pinned with `tool` directives in `go.mod`.
- **Build:** `CGO_ENABLED=0` for release builds (a pure-Go static binary). External media tools run as subprocesses (ADR-0012).

### Implementation details chosen by agent
| Detail | Choice | Reason |
|---|---|---|
| Logging | Standard library `log/slog` with a JSON handler; a request-ID middleware | No dependency; structured (NFR-016) |
| Config file format | **TOML** via `github.com/pelletier/go-toml/v2` **v2.4.3** (MIT, verified), with strict decoding (unknown keys rejected). Environment variable overrides via a small in-house layer (no Viper) | Human-editable with comments; one small, maintained dependency. BurntSushi/toml v1.6.0 (MIT) was the equal alternative |
| Config precedence | Built-in defaults < config file < environment variables (`LOCALAINAS_*`) < CLI flags. The config file path comes from `--config`, then `LOCALAINAS_CONFIG`, then an OS default path (ADR-0003) | Predictable, documented, testable |
| Error model | Typed domain errors mapped to RFC 9457 problem details in one place (ADR-0002) | One error format (S01.1) |
| OS-specific calls | `golang.org/x/sys` **v0.48.0** (BSD-3-Clause, verified) for free-space queries and volume/device identity on Windows (S01.2 same-filesystem check). Unix uses the standard `syscall` package | Official Go sub-repository; no cgo |
| Traversal-resistant file access | Standard library **`os.Root`** (Go 1.24+; OpenFile, Mkdir/MkdirAll, Rename, Remove/RemoveAll, Stat/Lstat verified on pkg.go.dev) for every file operation inside a namespace | Defense in depth behind the path resolver (S01.6) |

## Consequences

- **Easier:** single-binary distribution; low RAM; cross-compiling for arm64; streaming performance; embedding the web UI (`go:embed`, ADR-0009).
- **Harder:** two languages once S12 arrives. The contract between the core and the AI worker must be specified (JSON, versioned) and contract-tested (ADR-0017).
- **Required:**
  - Go toolchain pinning.
  - `CGO_ENABLED=0` in release builds. The race detector (`-race`) needs cgo, so it runs only in CI test jobs.
  - External tools are documented for native installs (ADR-0006).

## Approval record

> Decision made by the user in plan change request #3: "The technology choices here are the user's decisions." (`code-agent-docs/prompts/P003-technology-stack.json`, `technology_stack.core_server`)
> (2026-09-24, session S003). Verified in S003 log E005: Go go1.27.1 current, BSD-3-Clause, no blocking problem.

## Links

- **Related requirements:** NFR-004, NFR-008, NFR-009, NFR-014, NFR-016, NFR-021, NFR-029, NFR-030
- **Related ADRs:** ADR-0002 (API), ADR-0004 (repository layout), ADR-0005 (toolchain), ADR-0006 (packaging), ADR-0012 (media tools), ADR-0017 (AI worker)
- **Related stages:** S01.1 onward
- **Plan version:** 0.3.0
