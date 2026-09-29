# ADR-0022: Perceptual hash and similarity index

| Field | Value |
|---|---|
| Number | ADR-0022 |
| Status | Proposed |
| Date proposed | 2026-09-28 (session S007) |
| Date of last status change | 2026-09-28 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- **P005** asks for resolution variants of the same photo to be found (FR-151) and for look-alike photos to be grouped into stacks (FR-160). Both need a **perceptual hash**: a short fingerprint that stays close when an image is resized, re-compressed, or converted (FR-212).
- Lookups must scale to 100,000+ photos per user without comparing every pair (FR-213, NFR-003 scale from Q18).
- **Where the hash is computed:** P005 said "while generating thumbnails, since the image is already decoded". libvips runs as a **separate process** (ADR-0012), so the Go core never holds the decoded image. The hash is therefore computed in Go from a **small rendition** that the thumbnail job already produces with libvips (orientation applied), so the original is not read again (plan S04.4 note).
- Thresholds and false positives: plan concern 8.22, NFR-034, RK-32.

## Options considered

### Hash algorithm

| Option | Pros | Cons |
|---|---|---|
| **dHash** (difference hash, 64-bit from a 9×8 grayscale image) | Very cheap; robust to resizing and re-compression; tens of lines of code | Weak against crops and larger edits |
| **pHash** (DCT of a 32×32 grayscale image, 64-bit from the low frequencies) | More robust to compression and small edits; the usual choice for resolution variants | A little more code (a 32×32 DCT); still cheap |
| Both (two 64-bit values) | Each covers the other's weak cases; a match on both is strong evidence | 16 bytes per image instead of 8 |
| AI embeddings | Best for look-alikes not shot together | Needs AI, which is optional and last (I7, I8); this is S16.11 |

### Implementation

| Option | Pros | Cons | License |
|---|---|---|---|
| `github.com/corona10/goimagehash` | Ready-made dHash, pHash, and aHash | Last release v1.1.0 on 2022-05-26 (verified in S007): a maintenance concern; it takes an `image.Image`, so the decoding question stays | BSD-2-Clause (verified in S007) |
| **In-house** (`internal/similarity`) | No dependency; the algorithms are small and well documented; tested against published examples | Code to maintain and test | AGPL-3.0-or-later (project) |

### Similarity index

| Option | Pros | Cons |
|---|---|---|
| Linear scan per user | Simplest | 100,000 photos give 5×10⁹ pairs for a full scan: too slow |
| **BK-tree** on Hamming distance, per user, in memory | Simple; fast at the small distances used for duplicates (≤ 10 of 64 bits) | Slower at loose distances |
| Multi-index hashing (split the hash into blocks, exact lookups per block) | Fast at moderate distances; scales well | More code; memory for the tables |
| Sorting by time first (bursts and look-alikes are close in time) | Look-alike candidates come from a time window, which is small | Not for resolution variants, which can be years apart |

## Decision

**Recommended:**

- **Both dHash and pHash, 64 bits each,** stored as 16-digit hexadecimal text in the sidecar (`hashes.dhash`, `hashes.phash`, S05.1) and the index.
- **In-house implementation** in Go (`internal/similarity`), computed from a small grayscale rendition (32×32 for pHash, 9×8 for dHash, taken from one ≥ 32 px rendition) that the thumbnail job produces with libvips. No new dependency. goimagehash may be used only as a test oracle during development, not shipped (if used, it is recorded in `dependencies.md` as a test tool).
- **Index:** a per-user in-memory **BK-tree** built at startup from the stored hashes and updated on every change; if the S11.1 benchmark misses its target, multi-index hashing replaces it behind the same interface. Look-alike candidates are first narrowed by the time window (S11.4).
- **Thresholds:** starting points only (resolution variants: pHash distance ≤ 6 and dHash ≤ 8, plus the same aspect ratio within 1% and the same date taken where present; look-alikes: looser, within the time window). The real values are calibrated in S11.1 on the labelled fixture set (NFR-034) and recorded in the S11 stage document.

## Consequences

- **Easier:** no network, no dependency, cheap to compute and store; the index is a cache and is rebuilt from sidecars (I3).
- **Harder:** perceptual hashes give false positives on similar but different photos (e.g. two shots of a white wall); S11 must always show the evidence and never delete without confirmation (I10).
- **Required (follow-up work, constraints this imposes):** S04.4 produces the small rendition and stores both hashes; S05.1 reserves the fields; S06.1 reserves the index fields; S11.1 benchmarks the index and calibrates the thresholds.

## Approval record

_Pending: put to the user with the P005 report (S007)._

## Links

- **Related requirements:** FR-151, FR-160, FR-212, FR-213, NFR-034
- **Related ADRs:** ADR-0012 (media toolchain), ADR-0021 (content hash), ADR-0023 (stacks)
- **Related stages:** S04.4, S05.1, S06.1, S11.1, S11.2, S11.4
- **Plan version:** 1.4.0 (concern 8.22)
