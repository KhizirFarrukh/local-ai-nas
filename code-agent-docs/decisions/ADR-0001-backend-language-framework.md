# ADR-0001: Backend language, framework, and toolchain

| Field | Value |
|---|---|
| Number | ADR-0001 |
| Status | Proposed |
| Date proposed | 2026-09-24 (session S002) |
| Date of last status change | 2026-09-24 (session S002) |
| Supersedes | none |
| Superseded by | none |

## Context

S01.1 needs a backend language, web framework, and toolchain before any code is written. Constraints:
- **Workload:** a long-running file server doing mostly disk and network I/O: streaming uploads and downloads with ranges, resumable uploads, directory listing. Background jobs come later (S04.3), and so does the search index (S06).
- **Requirements:**
  - NFR-004: runs on a 4-core CPU with 4 GB RAM.
  - NFR-009: Linux amd64/arm64 via containers, native installs later, development on Windows.
  - NFR-021: streaming I/O.
  - NFR-014: tests, linting, and CI from S01.
  - NFR-001: no network at runtime.
- **Later stages:** the AI worker (S12) is Python-centric regardless of this choice, because the AI and ONNX ecosystem lives there.
- **Development:** the project is built by an AI agent with user review. Clarity and a mainstream toolchain matter.

## Options considered

### Option A: Python 3.12+ with FastAPI
- **Pros:**
  - The same language as the S12 AI worker, so schemas and job contracts can be shared.
  - Fast development; OpenAPI is generated from typed models (Pydantic).
  - Rich imaging and metadata ecosystem.
- **Cons:**
  - Higher memory than compiled options.
  - No native single-binary packaging.
  - CPU-bound code needs process pools.
  - The async and file I/O model needs care (threads for file I/O).
- **License / cost:** FastAPI MIT, Starlette BSD-3, Pydantic MIT, Uvicorn BSD-3 (to verify at adoption).

### Option B: Go
- **Pros:**
  - A single static binary with easy cross-compilation (amd64/arm64).
  - Excellent streaming and concurrency for file serving; low memory.
  - A strong standard library (`net/http`, `io`).
- **Cons:**
  - Two languages once the AI worker arrives.
  - Fewer imaging and metadata libraries, so external tools are called as subprocesses.
  - OpenAPI is spec-first or needs a code generator.
- **License / cost:** Go BSD-3.

### Option C: TypeScript / Node.js (Fastify)
- **Pros:** the same language as the web UI; a large ecosystem.
- **Cons:** a heavier runtime; weaker for streaming very large files; two languages with AI; harder packaging.
- **License / cost:** MIT (Node, Fastify).

### Option D: Rust (Axum)
- **Pros:** best performance and memory safety; the Tantivy search library.
- **Cons:** much slower development; CPU is not the bottleneck for a NAS.
- **License / cost:** MIT/Apache-2.0.

## Decision

**Recommendation (Proposed, not decided): Option A, Python 3.12+ with FastAPI**, with this toolchain:
- `uv` for environments and lockfiles.
- `ruff` for linting and formatting; `mypy` in strict mode for types.
- `pytest` with `hypothesis` for tests.
- `pre-commit`; GitHub Actions on Linux and Windows.

Reasoning at the time of this draft: one language across the core and the AI worker, the fastest route to a tested product, and performance dominated by I/O. **Go (Option B) is the strongest alternative** if a single static binary and low memory are priorities.

## Consequences

- **Easier:** sharing Pydantic models with the S12 worker; fast iteration; generated OpenAPI.
- **Harder:** native single-binary distribution (S11.2); keeping memory low on small boards.
- **Required:** file I/O runs off the event loop; the process model is documented; the API docs assets are self-hosted (the FastAPI defaults use a CDN, which violates I6).

## Approval record

> _Pending user review (S002)._

## Links

- **Related requirements:** NFR-001, NFR-004, NFR-008, NFR-009, NFR-013, NFR-014, NFR-016, NFR-021, NFR-025
- **Related ADRs:** ADR-0002 (API style), ADR-0003 (storage layout)
- **Related stages:** S01.1
- **Plan version:** 0.2.0
