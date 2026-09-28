# ADR-0021: Content hash algorithm for uploads and duplicate detection

| Field | Value |
|---|---|
| Number | ADR-0021 |
| Status | Proposed |
| Date proposed | 2026-09-28 (session S007) |
| Date of last status change | 2026-09-28 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- **P005** (plan 1.4.0) asks for a content hash of every file, computed while it streams during upload and stored with the item (FR-211). Duplicate detection uses it: at ingest in the photos area (S04.2, FR-022), for photo duplicates (S11.2), and for file duplicates (S11.5). A backfill covers files that have none (S11.1).
- P005 prefers **standard-library SHA-256 unless benchmarks justify BLAKE3** (A23).
- The hash identifies duplicates **within one user's library** (FR-153). It is not a security boundary, but collisions must be practically impossible, because an exact match can lead to a deletion (I10, RK-32).
- **Resumable uploads** (tus, ADR-0008) arrive in chunks, possibly over several server runs. Plan 8.26: the hash state is serialized between chunks; if it cannot be restored, the assembled file is hashed once before the atomic rename.
- **Hardware:** x86-64 mini-PCs and old PCs, Raspberry Pi (ARM64), Windows 11 for testing (Q1). Uploads are bounded by the network: gigabit Ethernet carries about 118 MB/s.
- Already in the code: tus uploads can carry an expected SHA-256 for verification (`internal/uploads`, S01.4 "optional checksum verification").
- This ADR must be Accepted before the S01 follow-up tasks are built (Q50, answered in S007: now).

## Options considered

### Option A: SHA-256 from the Go standard library (`crypto/sha256`)
- **Pros:** no new dependency; the hash state can be serialized (`encoding.BinaryMarshaler`), which the resumable-upload rule needs; hardware-accelerated on x86-64 (SHA-NI) and on ARMv8 CPUs with the crypto extensions; the same algorithm as the existing tus checksum verification; widely understood, and users can check a file with standard tools (`sha256sum`, `Get-FileHash`).
- **Cons:** slower than BLAKE3 in software. The Raspberry Pi 4 (Cortex-A72 in the BCM2711) has no ARMv8 crypto extensions, so SHA-256 runs in plain software there. The Pi 5 (Cortex-A76) has them.
- **License / cost:** BSD-3-Clause (Go).

### Option B: BLAKE3 (`lukechampine.com/blake3` v1.4.1)
- **Pros:** much faster on x86-64 with SIMD; parallel over large inputs.
- **Cons:** a new dependency; its SIMD code is for amd64, so on ARM64 it runs as generic Go (speed on the Pi not measured); fewer users can verify a file with standard tools; serializing the hash state between chunks needs checking (not verified in S007).
- **License / cost:** MIT (verified in S007: pkg.go.dev, v1.4.1).

### Option C: A non-cryptographic hash (xxHash) for grouping, a cryptographic one to confirm
- **Pros:** very fast first pass.
- **Cons:** two hashes to store and keep in sync; the confirmation still costs a full cryptographic read; the size and partial-hash steps of S11.5 already make the first pass cheap.
- **License / cost:** BSD-2-Clause (xxHash Go ports).

### Benchmark (S007, 2026-09-28)

Go 1.27.0, windows/amd64, AMD Ryzen 5 7640HS; 1 MiB writes into one hash (`go test -bench`, 3 runs, `scratchpad/hashbench`, not committed):

| Hash | Throughput |
|---|---|
| SHA-256 (`crypto/sha256`) | 1,196–1,918 MB/s |
| BLAKE3 (`lukechampine.com/blake3` v1.4.1) | 5,503–5,658 MB/s |

Both are 10 to 48 times faster than gigabit Ethernet, so neither slows an upload on this class of machine. The Raspberry Pi is **not measured** (Unverified); it is measured when the device is available, like the S01.7 throughput deviation.

## Decision

**Recommended: Option A, SHA-256 from the standard library.**

- Stored as text with the algorithm name: `sha256:<64 lowercase hex digits>`. A later change of algorithm (a new ADR) can then coexist with old hashes and be backfilled, without ambiguity.
- Computed while the upload streams (simple uploads and each tus chunk), with the hash state saved with the upload session between chunks. If the state cannot be restored (e.g. an upload from before this change, or a lost state file), the assembled file is hashed once before the atomic rename.
- The existing optional tus checksum check compares against this hash, so the file is still read only once.

**Why not BLAKE3:** the benchmark shows SHA-256 far faster than the network on x86-64, and P005 asks for SHA-256 unless benchmarks justify BLAKE3. The one open question is the Raspberry Pi 4 (no crypto extensions), where both run in software; if its measured speed limits uploads or the S11 backfill, a new ADR can switch to BLAKE3 thanks to the algorithm prefix.

## Consequences

- **Easier:** no new dependency; exact-duplicate checks at ingest (S04.2) and in S11 are simple equality lookups; users can verify files with standard tools.
- **Harder:** hashing on the Raspberry Pi 4 is software-only (to be measured).
- **Required (follow-up work, constraints this imposes):**
  - S01 follow-up task (stage document `stages/S01-basic-nas.md`): hash at upload in both paths, store it in the internal database, expose it in item details, and keep the hash state between tus chunks.
  - S04.2 reuses it for photos; S05.1 reserves `hashes.content` in the sidecar; S11.1 backfills items without a hash.
  - The Raspberry Pi measurement is added to the S01 follow-up notes.

## Approval record

_Pending: put to the user with the P005 report (S007)._

## Links

- **Related requirements:** FR-211, FR-022, FR-150, FR-169, NFR-006
- **Related ADRs:** ADR-0008 (tus uploads), ADR-0007 (SQLite), ADR-0022 (perceptual hash)
- **Related stages:** S01 (follow-up), S04.2, S11.1, S11.2, S11.5
- **Plan version:** 1.4.0 (A23, concern 8.26)
