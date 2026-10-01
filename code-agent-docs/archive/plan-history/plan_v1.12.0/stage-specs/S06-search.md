# S06: Search

> Part of the local-ai-nas development plan (see [PLAN_INDEX.md](../PLAN_INDEX.md)). Moved unchanged from `plan.md` section 10.7 in version 1.9.0 (ADR-0044).

### 10.7 S06: Search

- **Origin:** User-defined
- **Goal:** Fast, forgiving search across both files and photos.
- **User requirements (quoted):**
  > "Stage 6 is searching, in both files and photos."
  > "From the README: search covers file name, description, place (from GPS), date and time, and metadata."
  > "Search is not exact-match. It works like a search engine, with near matches and different word forms, and synonyms work too (searching 'receipts' also returns items with 'receipt', 'invoice', or 'voucher' in their metadata)."
  > "Operator-style searches work, e.g. 'before:2026' returns all items dated before 2026."
- **Status:** Not started
- **Security (1.12.0, P010; rule R15):** the search query parser is fuzzed in the stage gate; results are filtered per user before ranking (I5); both **Tier 1**.

#### S06.1: Search architecture and engine
- **Goal:** One embedded, rebuildable search index for both areas, ready for access control and AI fields.
- **Scope:** engine decided in 0.3.0: **Bleve** embedded (ADR-0014). S06.1 designs the index mapping; the index is a rebuildable cache (I3); one query path covering both areas; an index schema with reserved fields for owner and access list (S07) and AI tags and face groups (S17).
- **Deliverables:** Bleve index mapping (analyzers, keyword, date, and numeric fields); `SearchEngine` interface; query API skeleton.
- **Depends on:** S05 (Done).
- **Requirements:** FR-025, FR-047, FR-052, FR-104, FR-109.
- **Acceptance criteria:**
  1. Bleve runs embedded in the core with no extra service (ADR-0014), and the mapping matches the documented index schema.
  2. The index can be deleted and fully rebuilt from disk and sidecars.
  3. One query API searches files, photos, or both.
  4. The reserved fields exist: `owner` populated now, `acl` and AI fields later.
- **Risks/notes:** Performance risk if the embedded engine underperforms. The interface allows a swap.
- **P005 change (1.4.0):** the index schema reserves fields for the stack ID, the cover flag, and a shortcut flag. Search behavior for stacks and shortcuts is defined in S11.
- **P008 (1.9.0):** per-field analysis and Unicode/Arabic-script normalization (FR-357, 8.11 table); documents keyed by item ID and a pluggable semantic retriever that returns nothing yet (FR-358); vectors never in Bleve (FAISS through cgo, verified); BM25 scoring available since Bleve v2.5.0 (verified), recorded in the ADR-0014 amendment; multilingual golden queries.
- **Status:** Not started

#### S06.2: Indexing pipeline
- **Goal:** The index always reflects what is on disk.
- **Scope:**
  - Initial full indexing; incremental updates through the job system on every create, update, move, and delete.
  - A full rebuild command; index consistency checks.
  - A design hook for document-content indexing (Q31).
- **Deliverables:** index jobs wired to service hooks; CLI rebuild; consistency checker.
- **Depends on:** S06.1, S04.3.
- **Requirements:** FR-068, FR-108, FR-111.
- **Acceptance criteria:**
  1. Every create, update, move, and delete in either area is searchable within the delay set in the stage document (proposed ≤ 5 s).
  2. A rebuilt index equals the incrementally maintained one (equivalence test).
  3. The consistency check detects and repairs deliberately introduced drift.
- **Risks/notes:** Document-content search stays a hook unless Q31 approves it.
- **Status:** Not started

#### S06.3: Query language and parser
- **Goal:** A precise, documented query language that never fails badly.
- **Scope:**
  - Free text plus operators. Baseline set: `before:`, `after:`, `on:`, `place:`, `tag:`, `type:`, `in:files`, `in:photos`, `ext:`, `size:`.
  - Quoting and combinations; `face:` reserved for S17.
  - Clear errors for malformed queries; a documented formal grammar.
- **Deliverables:** EBNF grammar document; tokenizer and parser; operator semantics (Q14); error and hint messages.
- **Depends on:** S06.1.
- **Requirements:** FR-056, FR-057, FR-058, FR-059, FR-060, FR-061, FR-063, FR-105, FR-106.
- **Acceptance criteria:**
  1. The EBNF grammar and the parser agree (grammar-driven tests).
  2. Each operator works alone and combined with free text and other operators.
  3. Malformed queries return a clear error or hint, never a server error (property-based tests).
  4. `face:` parses and returns a "not available until AI is enabled" hint.
- **Risks/notes:** Date semantics depend on Q14.
- **Status:** Not started

#### S06.4: Fuzzy matching
- **Goal:** Near matches and word forms find the right items.
- **Scope:** normalization of case, accents, and Unicode; stemming and plurals; typo tolerance; prefix matching.
- **Deliverables:** normalization and stemming pipeline; typo candidate generator; prefix support.
- **Depends on:** S06.2.
- **Requirements:** FR-048, FR-049, FR-107.
- **Acceptance criteria:**
  1. `receipts` matches `receipt`, and `Café` matches `cafe`.
  2. `reciept` finds `receipt` items.
  3. `rece` finds `receipt` by prefix.
  4. Typo expansion respects edit-distance limits by term length (no absurd matches).
- **Risks/notes:** The vocabulary must become permission-scoped in S07 (8.9).
- **Status:** Not started

#### S06.5: Synonyms and related terms
- **Goal:** Related words find each other, offline.
- **Scope:** a local synonym dictionary (e.g. receipt, invoice, voucher, bill) that users can extend, applied at query time, with no network access.
- **Deliverables:** shipped dictionary (with license verified); user extension file; query-time expansion.
- **Depends on:** S06.4.
- **Requirements:** FR-050.
- **Acceptance criteria:**
  1. `receipts` returns items containing `invoice`, `voucher`, or `bill`.
  2. User-added synonym sets take effect without a restart.
  3. Synonym expansion makes no network calls.
- **Risks/notes:** Over-expansion noise (RK-10) is managed by ranking weights.
- **Status:** Not started

#### S06.6: Ranking
- **Goal:** The best results come first, predictably.
- **Scope:** relevance scoring with per-field weights (name, tags, description, place); recency boosting; stable tie-breaking.
- **Deliverables:** scoring function; golden query set with expected rankings.
- **Depends on:** S06.4, S06.5.
- **Requirements:** FR-051.
- **Acceptance criteria:**
  1. The golden query set passes.
  2. Exact matches outrank stem, synonym, and typo matches.
  3. Equal scores return in a stable order across calls.
- **Risks/notes:** None.
- **Status:** Not started

#### S06.7: Search API and GUI
- **Goal:** Search is available everywhere and easy to use.
- **Scope:** a global search bar; results filterable by area; photo results as thumbnails and file results as a list; operator hints and autocomplete; filter chips; helpful no-result states.
- **Deliverables:** search API endpoints; search UI components.
- **Depends on:** S06.3–S06.6, S02, S04.7.
- **Requirements:** FR-053, FR-055, FR-104, FR-110.
- **Acceptance criteria:**
  1. The search bar is on every page, and results filter to files, photos, or both.
  2. Typing an operator shows hints and completions.
  3. Active operators show as removable filter chips.
  4. No-result states suggest fixes (typo corrections, removing filters).
- **Risks/notes:** Suggestions must draw only from items the user can access (ready for S07, 8.9).
- **Status:** Not started

#### S06.8: Performance and stage review
- **Goal:** Search meets its latency targets at scale, then the stage closes.
- **Scope:** benchmarks at scale (100,000 photos plus 100,000 files, Q18) against the latency targets in NFR-003; documentation; completion record; user sign-off.
- **Deliverables:** benchmark suite and report; search user guide (including the operator reference); completion record.
- **Depends on:** S06.1–S06.7.
- **Requirements:** NFR-003.
- **Acceptance criteria:**
  1. At 100k photos + 100k files, search p95 meets NFR-003 on reference hardware.
  2. A full rebuild from disk meets NFR-003.
  3. The completion record is written and the user's sign-off is recorded.
- **Risks/notes:** Reference hardware (Q1, S005): an x86-64 mini-PC or old PC, a Raspberry Pi, and the Windows 11 PC.
- **Status:** Not started

**Design notes (S06):**
- Search reads only the index (I4).
- Full-text search inside document contents in the files area is an open question (Q31). The pipeline has a hook for it.
- The reserved `owner`, `acl`, `ai_tags`, and `face_groups` fields avoid a reindex design change later.
- Permission pre-filtering is designed now and activated in S07.4.

**Exit criteria (quoted):** "Searching 'receipts' finds items tagged with 'invoice', typos still find results, operators work alone and combined, and queries meet the latency targets."

---
