# ADR-0043: Hybrid lexical and semantic search

| Field | Value |
|---|---|
| Number | ADR-0043 |
| Status | Proposed (decided in the AI stage, S17.10) |
| Date proposed | 2026-09-30 (session S007) |
| Date of last status change | 2026-09-30 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

Search by meaning ("tax receipt from last month") needs the typed text embedded at query time and compared with image or text embeddings (FR-362; the user's decision D3 allows query-text encoding, I4 reworded). Whether semantic search is wanted at all is still Q36.

## Options considered

| Option | Pros | Cons |
|---|---|---|
| **(a) Embeddings in internal data, an in-process index in the core, rank fusion with Bleve** | No new service; pure Go; rebuildable | Brute force must be fast enough (expected at 100,000 photos) |
| (b) Bleve's own vector search | One index | Needs the FAISS C++ library through cgo and the `vectors` build tag (verified 2026-09-30, Bleve docs), against pure-Go builds (ADR-0001) |
| (c) pgvector, Qdrant, ChromaDB | Mature | A new database or service (against ADR-0007 and the one-program design) |
| (d) SQLite FTS5 for full text | Built in | Rejected in ADR-0014; Bleve covers full text with fuzzy matching and BM25 (since v2.5.0, verified) |

## Decision (proposed)

**(a).** Embeddings per model, keyed by item ID (ADR-0040), never mixing two models' vectors; brute-force similarity over compact vectors, a pure-Go approximate index only if benchmarks demand it; reciprocal rank fusion with tunable weights; operators as hard filters on both sides; permission pre-filtering before ranking (NFR-024); a similarity threshold; meaning-matched results labeled. The query-text encoder runs in the AI worker (ADR-0017) with a strict time limit and a small cache; on timeout or with the worker off, lexical results only (I7). Ships only if the golden queries show it beats lexical search.

## Consequences

- **Required:** S06.1 hooks (FR-358); S17.10 builds it if Q36 says so.

## Approval record

_Proposed with plan 1.9.0 (P008); decided when S17.10 is planned in detail, after Q36._

## Links

plan FR-054, FR-358, FR-362, 8.40 · [ADR-0014](ADR-0014-search-engine-bleve.md) · [ADR-0017](ADR-0017-ai-worker-architecture.md) · [ADR-0018](ADR-0018-ai-models.md) · P008 AI-B, AI-C
