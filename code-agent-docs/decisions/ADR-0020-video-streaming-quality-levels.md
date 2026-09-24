# ADR-0020: Video streaming with live quality switching: HLS, on-demand transcoding with cache

| Field | Value |
|---|---|
| Number | ADR-0020 |
| Status | **Accepted** (approach chosen by the user in S004; implementation details below are proposed by the agent and confirmed in the S04 stage document) |
| Date proposed | 2026-09-24 (session S004) |
| Date of last status change | 2026-09-24 (session S004) |
| Supersedes | **ADR-0012 in part**: its "video transcoding deferred" clause, for transcoding into streaming quality levels |
| Superseded by | none |

## Context

- **The user's request** (S004 E005): "If not present, add a photo viewer and video playback (streaming with quality adjustment option live in video playback) in the gui part".
- **Already in the plan:** the photo viewer (FR-014, S04.7; image previews in S02.6) and basic video playback (FR-019; S02.6 range-streamed previews; S04.7 lightbox playback).
- **Missing:** **switching quality live during playback.** That requires transcoding into lower-quality renditions. Transcoding had been deferred by P003 ("Transcoding browser-incompatible video is deferred and not part of the S04 baseline"; ADR-0012) and excluded by plan NG4.
- **Constraints:**
  - Modest, CPU-only hardware by default (NFR-004).
  - Derived data in internal app data (I2).
  - Every access path authorized (I5).
  - Originals never modified (NFR-006).
  - External tools as subprocesses (P003; ADR-0012).
  - Licenses that allow anyone to deploy (NFR-029).

## Options considered

### Transcoding strategy (the user's choice, S004 E008)
| Option | Pros | Cons |
|---|---|---|
| **Hybrid: on-demand + cache** (chosen) | The original needs no work. Lower levels cost CPU only when someone asks for them. Cached levels replay instantly | The first switch to a new level waits a few seconds. A cache must be sized and evicted |
| Pre-generate at upload | Instant switching; no waiting | +50–80% disk per video; long background CPU time on weak hardware |
| Live transcoding only | No extra disk | Needs a strong CPU or hardware encoder; likely stutters above 480p on Raspberry Pi-class CPUs |

### Streaming format and player
| Option | Pros | Cons |
|---|---|---|
| **HLS + hls.js** (chosen) | Native on Safari/iOS; hls.js elsewhere via Media Source Extensions (including ManagedMediaSource on newer iOS). Supports manual level selection and automatic bitrate adaptation. FFmpeg writes HLS natively | Segment handling and playlist generation in the core |
| MPEG-DASH + dash.js | Codec-agnostic standard | No native iOS playback |
| Separate progressive MP4 per quality | Simplest server | No seamless live switching (reload + seek); no adaptive "Auto" |

### Controls (the user's choice)
- **Manual + Auto** (chosen): a quality menu with Auto, Original, 1080p, 720p, 480p, 360p. Auto adapts to network throughput.
- Manual only.

### Scope (the user's choice)
- **Photos + files previews** (chosen): the photos lightbox and video previews in the files browser.
- Photos area only.

## Decision

**User-chosen (S004 E008):**
1. **Hybrid on-demand transcoding with a cache.** The original quality streams without transcoding. Lower quality levels are transcoded the first time they are requested and **cached** in internal app data, with a size cap. A hardware encoder is used when present.
2. **Manual quality menu + Auto** (adaptive bitrate), delivered with **HLS**: hls.js in browsers, native HLS where hls.js cannot run.
3. **Available in both** the photos-area lightbox (S04.7) and the files-area video preview (S02.6's player, upgraded in S04.8).

**Supersession:**
- This ADR supersedes ADR-0012's "video transcoding deferred" clause **for streaming quality levels**. Originals are still never re-encoded or modified.
- Full RAW conversion stays deferred (ADR-0012).
- **For the user's confirmation (side effect):** if a browser cannot play the original codec (e.g. HEVC in some browsers), the menu hides "Original", and Auto plays the transcoded H.264 levels. Browser-incompatible videos therefore become playable, which the P003 deferral had left out of S04.

### Implementation details proposed by agent (confirmed or changed in the S04 stage document)
| Detail | Proposal | Reason |
|---|---|---|
| Levels | Original (if browser-playable), then 1080p, 720p, 480p, 360p, **only up to the source height** | Covers phone to TV; no upscaling |
| Codecs | H.264 video + AAC stereo 128 kb/s in fMPEG-4 or MPEG-TS HLS segments | Plays on every target browser (NFR-027) |
| Segments | 4 s, with forced keyframes aligned across levels | Seamless switching between levels |
| On-demand sessions | Requesting a level starts or reuses one FFmpeg session per (video, level). It writes segments ahead of the playhead into the cache. A segment request waits briefly for production. A seek beyond the produced range restarts the session at the seek point. Finished levels stay cached | Standard approach in self-hosted media servers |
| Cache | `<internal>/transcode-cache/<content-hash>/<level>/`. Size cap (default 10% of free space, configurable absolute limit), LRU eviction job. Keyed by content hash, so there are no duplicates across areas or users | I2; derived data that can always be rebuilt |
| Concurrency and priority | Concurrent sessions capped (default 1 with the CPU encoder, 2–4 with a hardware encoder; configurable). Idle sessions end after a timeout. Background jobs (thumbnails, AI, indexing) yield to active playback | Keeps the NAS responsive (NFR-004, NFR-031) |
| Encoder selection | Detected at startup (`ffmpeg -encoders`/`-hwaccels`, device presence such as `/dev/dri`). Prefer a hardware encoder; fall back to libx264 with a fast preset. Docker docs cover device passthrough | Hardware varies. The Debian FFmpeg build explicitly enables libx264 and libvpl (Intel QSV); VAAPI, V4L2-M2M, and NVENC are FFmpeg autodetect features whose availability is **unverified** until S04.8 |
| Player | **hls.js 1.7.3** (Apache-2.0, verified 2026-09-24). Levels API for the manual menu (`currentLevel`), `-1` for Auto. On browsers where hls.js cannot run, native HLS (Auto only) | The manual menu works wherever hls.js runs; the limitation is documented |
| Security | Master playlists, level playlists, and segments go through authentication and `authorize()` like any download (I5). `Cache-Control: private`. No listing of the cache | Invariant I5 |

## Consequences

- **Easier:** watching large or high-bitrate videos on slow Wi-Fi or phones; playing originals in codecs some browsers lack.
- **Harder:**
  - CPU load on weak hardware (mitigated by the concurrency cap, the hardware encoder, and caching).
  - More moving parts in the core: session management, playlist generation, cache eviction.
  - Cross-browser testing, including iOS.
- **Required:**
  - New substage **S04.8** (video streaming and quality levels); the testing substage becomes S04.9.
  - New requirements FR-144–FR-148 and NFR-031.
  - Register rows for hls.js and x264.
  - **Licensing:** libx264 is GPL-2.0-or-later, and it comes with the Debian GPL FFmpeg build (see audit A001 D-04 and Q22).
  - Documentation of hardware-encoder passthrough for Docker.

## Approval record

> User's answers (S004 E008), options selected verbatim: "Hybrid: on-demand + cache (Recommended)"; "Manual + Auto (Recommended)"; "Photos + files previews (Recommended)". This follows the request in S004 E005 quoted above.
> (2026-09-24, session S004)
>
> The option descriptions the user selected included: "Lower qualities (e.g. 1080p/720p/480p) are transcoded into HLS segments the first time they're requested and cached in internal data (size-capped)… Uses a hardware encoder if present" and "Uses HLS (hls.js in browsers, native on Safari/iOS)." The implementation-detail table above is **not** separately approved. It is confirmed when the S04 stage document is approved.
>
> **Side effect confirmed (S005, decision D-14):** browser-incompatible originals may play through the transcoded levels. User's answer: "Accept all (Recommended)" (2026-09-24, session S005, log E007).

## Links

- **Related requirements:** FR-014, FR-019, FR-082, FR-144, FR-145, FR-146, FR-147, FR-148, NFR-004, NFR-024, NFR-027, NFR-031
- **Related ADRs:** ADR-0012 (**superseded in part** by this ADR), ADR-0006 (Docker, device passthrough), ADR-0009 (web UI, player), ADR-0011 (job system; cache eviction job), ADR-0010 (auth)
- **Related stages:** S02.6 (player hook), S04.7, **S04.8**, S04.9
- **Plan version:** 0.4.0
