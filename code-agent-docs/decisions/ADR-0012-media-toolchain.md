# ADR-0012: Media toolchain: ExifTool, libvips with libheif, FFmpeg (as subprocesses)

| Field | Value |
|---|---|
| Number | ADR-0012 |
| Status | **Accepted** (full RAW conversion and video transcoding **deferred**) |
| Date proposed | 2026-09-24 (session S003) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none |
| Superseded by | none |

## Context

- **Needs:** metadata extraction for images and video (S05.3: EXIF, XMP, IPTC, maker notes, HEIC, RAW, video); thumbnails and previews (S04.4) including HEIC and RAW previews; video poster frames and metadata.
- **P003 principle:** call external media tools as **subprocesses**, not through cgo, so the Go build stays pure and easy to cross-compile. Any exception needs its own ADR.

## Options considered

| Option | Pros | Cons |
|---|---|---|
| **External tools as subprocesses** (chosen) | Best-in-class coverage. Pure-Go build. License separation (separate programs) | Tools must be installed (bundled in the Docker image; documented for native installs). Process management |
| Go-native libraries (e.g. `goexif`, `image/*`) | No external tools | Poor coverage (no HEIC, RAW, XMP/IPTC depth, video) |
| cgo bindings (govips, libheif-go) | Faster, no process spawn | Breaks pure-Go cross-compilation; violates the P003 principle |

## Decision

| Tool | Role | Upstream version (verified) | License (verified) | Stages |
|---|---|---|---|---|
| **ExifTool** | Metadata: EXIF, XMP, IPTC, maker notes, HEIC, RAW, video. Extracting the embedded JPEG preview from RAW files for thumbnails | **13.59** (exiftool.org) | "Same terms as Perl itself (either the Perl Artistic License or GPL)". Needs a **Perl** runtime, which the Docker image bundles | S04.4 (RAW preview), S05.3 |
| **libvips** CLI (`vipsthumbnail`, `vips`) | Thumbnails and previews, orientation handling, format conversion | **8.18.6** | LGPL-2.1 | S04.4 |
| **libheif** (via libvips) | HEIC/HEIF decoding | **1.23.5** | Library LGPL; samples and wrappers MIT | S04.4 |
| **FFmpeg / ffprobe** | Video poster frames; duration, codec, and creation metadata | **9.0.2** | LGPL-2.1+ by default; **GPL if built with `--enable-gpl`** | S04.4, S05.3 |

- ExifTool runs as a **persistent process in `-stay_open` mode** (commands via `-@ -`, JSON output), avoiding Perl start-up cost per file.
- All tools run with argument arrays (never a shell), timeouts, and bounded concurrency through the job system (ADR-0011).
- **Deferred:** full RAW conversion (e.g. LibRaw); transcoding of browser-incompatible video (not in the S04 baseline).

### Implementation details chosen by agent
| Detail | Choice | Reason |
|---|---|---|
| ExifTool wrapper | **Small custom Go wrapper** in `internal/media/exiftool`. It manages one or more `-stay_open` processes, `-execute`/`{ready}` framing, JSON decoding, restart on crash, and a per-call timeout | The only notable Go wrapper, **`barasher/go-exiftool` v1.10.0, is GPL-3.0 and has no release since 2023-06-07** (verified). Linking it would impose GPL on the core binary. A custom wrapper is about 200 lines with no dependency |
| Thumbnail format | **WebP** (default quality 80; configurable, with JPEG as an option) | Smaller files. Supported by all target browsers: Chrome, Edge, Firefox, and Safari 14+/iOS 14+ (NFR-027) |
| Thumbnail sizes | 256 px (grid) and 1440 px (viewer), long edge. Keyed by content hash in `<internal>/thumbnails/` | Plan 8.15 |
| Bundled versions in Docker | Debian trixie packages: ExifTool 13.25, vips 8.16.1, FFmpeg 7.1.5, libheif 1.19.8 (verified). Newer upstream versions only via a later ADR if a needed feature is missing | Security-maintained distribution builds (ADR-0006) |
| HEIC codec plugins | Decoder only (libde265, LGPL). **No x265 encoder plugin (GPL)** | Encoding is not needed |

## Consequences

- **Easier:** broad format coverage; license separation; a pure-Go build.
- **Harder:** native installs need these tools installed (install guide, S11.4). Process lifecycle and timeout handling. Distribution versions lag upstream.
- **Required:**
  - Third-party notices (`docs/third-party-notices.md`), including **FFmpeg's build configuration**. Debian's build is likely GPL-enabled; check it in S11.1 (flagged for the user, Q22).
  - Health check reporting tool availability and versions (S04/S05).

## Approval record

> Decision made by the user in plan change request #3 (`code-agent-docs/prompts/P003-technology-stack.json`, `technology_stack.metadata_extraction`, `image_processing`, `video_processing`). The wrapper and thumbnail format are "agent decides" items
> (2026-09-24, session S003). Versions and licenses verified in S003 log E005/E007.

## Links

- **Related requirements:** FR-010, FR-012, FR-018, FR-019, FR-020, FR-098, NFR-029
- **Related ADRs:** ADR-0001, ADR-0006, ADR-0011
- **Related stages:** S04.4, S05.3
- **Plan version:** 0.3.0
