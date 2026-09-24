# S01 performance baseline

The first measurement of local-ai-nas against its performance requirement, NFR-003 (S01.7-T04, 2026-09-24). It covers what stage 1 has: listing folders, moving file content, and memory during large transfers. The photo timeline and search come with later stages.

## Targets

NFR-003 sets these targets for S01, for a library of up to 100,000 photos and 100,000 files per installation (confirmed by the user, Q18):

| Target | Value |
|---|---|
| Listing a folder of 10,000 entries | p95 ≤ 500 ms |
| Upload and download throughput | ≥ 80% of the raw disk or network, whichever is slower |
| Memory during a 10 GB transfer (S01.4) | heap growth < 256 MB |

## How it was measured

- `scripts/perf-baseline.sh [size]` runs everything below and prints the tables.
- **Listing:** `BenchmarkListLargeFolder` (the files service), and `TestPerfBaseline` over real HTTP: 40 requests for the first page of 100 in a folder of 10,000 files, for each sort key, reported as p95.
- **Throughput:** `TestPerfBaseline` sends 1 GiB over real HTTP on loopback, with the median of three runs of each:
  - a simple upload (`PUT /api/v1/files/content`), a download, and a tus upload in 64 MiB chunks;
  - the baselines on the same machine: a raw disk write in 1 MiB writes with an fsync (like `dd conv=fsync`), a raw read of the same file (from the cache, as the download reads its file from the cache), and a raw loopback TCP transfer in 1 MiB writes.
  - Each transfer is compared with the slower of its disk baseline and the loopback baseline.
- **Memory:** `TestMemoryBound` samples the heap every 20 ms (`runtime/metrics`) while 10 GiB (Linux) or 1 GiB (Windows) go through the simple upload, tus, a copy, and a download. It runs in the CI job `memory`, which fails if the heap grows by 256 MiB or more.

## Results

### Windows 11 development PC

AMD Ryzen 5 7640HS (6 cores, 12 threads), 15.2 GiB RAM, WD PC SN810 NVMe SSD, Windows 11 Home 10.0.26200, Go 1.27.1.

| Measurement | Result | Target |
|---|---|---|
| List benchmark (files service), first page, sort by name / size / time | 26 / 18 / 24 ms | (information) |
| List 10,000 entries over HTTP, first page, sort by name: p95 | 35 ms | ≤ 500 ms: **met** |
| … sort by size: p95 | 44 ms | ≤ 500 ms: **met** |
| … sort by time: p95 | 36 ms | ≤ 500 ms: **met** |
| List all 10,000 entries in pages of 1,000 (10 requests) | 292 ms | (information) |
| Raw disk write, 1 MiB writes + fsync | 649 MB/s | baseline |
| Raw disk read (from the cache) | 1,537 MB/s | baseline |
| Raw loopback TCP | 2,231 MB/s | baseline |
| Simple upload over HTTP | 406 MB/s = 63% of the raw disk | ≥ 80%: **not met** |
| Download over HTTP | 532 MB/s = 35% of the raw disk read | ≥ 80%: **not met** |
| tus upload, 64 MiB chunks | 400 MB/s = 62% of the raw disk | ≥ 80%: **not met** |
| Heap growth, 1 GiB through upload / tus / copy / download | 1.6 / 1.3 / 1.0 / 0.0 MiB | < 256 MiB: **met** |

### Linux (WSL2 on the same PC)

The same test in WSL2 (Linux 6.x, 12 CPUs). Its temporary folder is in memory (tmpfs), so the "disk" figures here are memory speed.

| Measurement | Result | Target |
|---|---|---|
| List 10,000 entries over HTTP, first page, p95 (name / size / time) | 40 / 35 / 30 ms | ≤ 500 ms: **met** |
| Raw write + fsync / raw read / raw loopback TCP | 1,252 / 5,114 / 5,040 MB/s | baselines |
| Simple upload over HTTP | 993 MB/s = 79% of the raw write | ≥ 80%: not met by 1 point |
| Download over HTTP (sendfile) | 6,564 MB/s = 130% of the loopback baseline | ≥ 80%: **met** |
| tus upload, 64 MiB chunks | 653 MB/s = 52% of the raw write | ≥ 80%: **not met** |

### CI (GitHub-hosted runners)

The `memory` job passes with 10 GiB on Linux and 1 GiB on Windows: the heap stays under 256 MiB for every transfer.

### Raspberry Pi and x86-64 mini-PC

Not measured yet: these machines were not available (Q1). `scripts/perf-baseline.sh` is the procedure to run on them, over the real network as well as on loopback.

## Deviations for the user's review

1. **Throughput on a fast NVMe disk over loopback is below 80% of the raw disk** for uploads (about 62%) and, on Windows, for downloads (35%). Why:
   - On loopback the network costs almost nothing, so the per-byte CPU work of HTTP is what shows. An upload receives each 1 MiB block and then writes it, one after the other.
   - On Windows the download goes through `TransmitFile`, which is slower on loopback than Linux's `sendfile` (the Linux download meets the target at 130%).
   - The absolute figures are 400 to 530 MB/s on Windows and 650 to 6,500 MB/s on Linux: 3 to 50 times the ~117 MB/s of gigabit Ethernet. On the target hardware, where the network or a USB disk is the slower side, the 80% target is expected to hold, but that has to be measured there.
2. **tus uploads are slower than simple uploads** (52–62% of raw): tusd's own chunk handling (its request body reader and the file store) adds cost that the project does not control. For large files the resumability is worth it; for LAN transfers of files under `uploads.max_file_size`, the simple upload is the fast path.

**The user's review (S01 sign-off, 2026-09-24):** accepted as recorded ("Accept, measure later").

**Next steps:** measure on the Raspberry Pi and the mini-PC over gigabit Ethernet when they are available. If uploads fall short there, overlap receiving and writing (two buffers in flight) in the simple upload; it is a small, contained change in `files.copyExactly`.

## What was improved while measuring

- Downloads now keep the fast path to the connection: the access log and the download's status writer pass `ReadFrom` through, so Go can use `sendfile` (Linux) or `TransmitFile` (Windows). Before, every download was copied through a 32 KiB buffer.
