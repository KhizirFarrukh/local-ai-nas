# ADR-0014: Search engine: Bleve, with query-time synonym expansion

| Field | Value |
|---|---|
| Number | ADR-0014 |
| Status | **Accepted** |
| Date proposed | 2026-09-24 (session S003) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none |
| Superseded by | none |

## Context

S06 delivers fast, forgiving search across both areas:
- FR-047–FR-059, FR-104–FR-110.
- Stemming and plurals, typo tolerance, prefix matching, synonyms, and operators (date, numeric, and keyword filters).
- Owner and ACL filtering from S07 (NFR-024).
- The index is a rebuildable cache in internal app data (I2, I3), and search reads only the index (I4).
- Embedded engines are preferred (few moving parts).

## Options considered

| Option | Pros | Cons |
|---|---|---|
| **Bleve (Go-native, embedded)** (chosen) | In-process, pure Go. Fuzzy (edit-distance), prefix, term, date-range, numeric-range, and boolean queries. Language analyzers. Scorch index format | Performance at 200k documents must be benchmarked (S06.8) |
| SQLite FTS5 | Already available through the database | No built-in edit-distance fuzzy matching |
| Meilisearch | Excellent typo tolerance and synonyms out of the box | A separate service with higher memory use |
| Tantivy | Excellent performance | Written in Rust, so it needs cross-language bindings |

## Decision

- **Bleve v2.6.1** (`github.com/blevesearch/bleve/v2`, **Apache-2.0**, released 2026-08-24, verified), embedded in the core. The scorch index lives at `<internal data>/index/` and can **always be rebuilt from disk and sidecars** (I3).
- **Capability verification** (Bleve `query.go` on master): `NewFuzzyQuery`, `NewPrefixQuery`, `NewTermQuery`, `NewDateRangeQuery`, `NewNumericRangeQuery`, `NewConjunctionQuery`, and `NewBooleanQuery` all exist. The English analyzer exists (`analysis/lang/en`). **All S06 needs are covered.**
- **Fuzzy and forms:** edit-distance queries (fuzziness by term length) and language analyzers (English stemming), plus prefix queries.
- **Operators:** parsed by the **project's own Go parser** (S06.3) into Bleve date-range, numeric-range, and term queries, combined with the text query in one `BooleanQuery`.
- **Access control:** `owner` and `acl` are **keyword fields used as filters on every query** from S07. They are reserved in the S06.1 mapping.
- **Synonyms:** expanded **at query time** from the project's own local, user-extendable synonym dictionary (S06.5), as a boosted disjunction.
- **Fallback:** if the S06.8 benchmarks miss their targets, revisit through a new ADR.

### Implementation details chosen by agent
| Detail | Choice | Reason |
|---|---|---|
| Text analyzer | A custom analyzer: unicode tokenizer → lowercase → ASCII folding (accents) → English possessive and stop filters → Porter stemmer. Also an **unstemmed** sub-field for exact-match boosting | Exact > stem > synonym > fuzzy ranking (FR-051) |
| Keyword fields | `area`, `owner`, `acl`, `type`, `ext`, `tag`, `face_group` (the last reserved for S17) | Exact filters |
| Typed fields | `taken_at` and `modified_at` (datetime); `size` (numeric) | Range operators |
| Native Bleve synonyms | Bleve v2.5+ supports synonym indexing (`docs/synonyms.md`). S06.5 may use it as the *mechanism* to apply our dictionary, but the dictionary itself stays project-owned and user-extendable as decided above | Keeps the user's decision; chooses the mechanism with evidence in S06.5 |

## Consequences

- **Easier:** one embedded engine for text and structured filters; pure Go.
- **Harder:** tuning fuzziness and ranking; making sure autocomplete and suggestions only use permission-scoped terms (plan 8.9).
- **Required:** the S06.8 benchmark at 100k photos + 100k files; the rebuild-equivalence test (S06.2); reserved fields in the S06.1 mapping.

## Approval record

> Decision made by the user in plan change request #3 (`code-agent-docs/prompts/P003-technology-stack.json`, `technology_stack.search`)
> (2026-09-24, session S003). Capabilities, version, and license verified in S003 log E005.

## Links

- **Related requirements:** FR-025, FR-047–FR-053, FR-055–FR-063, FR-104–FR-110, NFR-003, NFR-024
- **Related ADRs:** ADR-0007, ADR-0011, ADR-0013
- **Related stages:** S06 onward (S07.4 filters, S17.7 AI fields)
- **Plan version:** 0.3.0

## Amendment 1 (2026-09-30, plan 1.9.0)

- **Per-field analysis** (FR-357) instead of one analyzer everywhere: the field table of plan 8.11 (file names tokenized on separators, camel case, and digit boundaries, no stemming; people's names exact, prefix, typo-tolerant; stemming only on prose; keywords and structured filters). The complete mapping is recorded here in S06.1.
- **Unicode and multilingual:** NFC and case folding; Arabic-script normalization for Urdu and Arabic letter variants; no stemming of non-English text; Q15 decides extra language analyzers.
- **Scoring:** BM25 is available since Bleve **v2.5.0** (verified 2026-09-30 in the release notes; the pinned v2.6.1 has it); the choice between BM25 and TF-IDF is made in S06.1 with the golden queries.
- **No vectors in Bleve:** its vector search needs the FAISS C++ library through cgo and the `vectors` build tag (verified in Bleve's `docs/vectors.md`), against pure-Go builds (ADR-0001); hybrid search keeps vectors elsewhere (ADR-0043).
- **Documents keyed by item ID** (ADR-0040), with permission filtering once as a set of allowed IDs (FR-358).
