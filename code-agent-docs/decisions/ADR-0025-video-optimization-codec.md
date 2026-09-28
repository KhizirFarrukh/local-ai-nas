# ADR-0025: Video codec for storage optimization

| Field | Value |
|---|---|
| Number | ADR-0025 |
| Status | Proposed |
| Date proposed | 2026-09-28 (session S007) |
| Date of last status change | 2026-09-28 (session S007) |
| Supersedes | none (ADR-0012 and ADR-0020 are extended, not superseded; see Context) |
| Superseded by | none |

## Context

- **The user's requirement (P005):** reduce the resolution and quality of existing videos, and of future uploads by policy, with a live preview (FR-176, FR-179, FR-181, FR-192).
- The optimized file **replaces the original** after confirmation, with the original kept for the retention period (ADR-0026). It must play in the target browsers and in the S04.8 player, and keep its creation time, GPS, and rotation (FR-190).
- **Existing decisions:** FFmpeg is the video tool, run as a separate program (ADR-0012). ADR-0020 transcodes streaming levels into a cache. ADR-0012 deferred transcoding browser-incompatible video for playback; this ADR does not change that deferral, so it supersedes nothing. Plan NG4 was revised in 1.4.0 to allow user-started optimization.
- **Licensing (verified in S007, ffmpeg.org/legal.html):** FFmpeg is LGPL-2.1-or-later, "however … if those parts [GPL] get used the GPL applies to all of FFmpeg"; libx264 and libx265 are GPL-2.0-or-later. AV1: SVT-AV1 is BSD-3-Clause-Clear with the AOMedia Patent License 1.0; libaom is BSD-2-Clause with the same patent licence. The project already accepts Debian's GPL FFmpeg as a separate program (D-04, RK-30).
- Weak hardware: CPU-only boards (NFR-004, RK-29).

## Options considered

### Option A: H.264 (libx264), CRF
- **Pros:** plays everywhere (every target browser and TV); fast to encode; already used by ADR-0020, so no new component.
- **Cons:** larger files than HEVC or AV1 at the same quality.
- **License / cost:** GPL-2.0-or-later (libx264), in Debian's FFmpeg; separate program.

### Option B: H.265/HEVC (libx265)
- **Pros:** about half the size of H.264 at similar quality.
- **Cons:** slow on CPUs; browser playback depends on the platform (hardware decoding); patent pools.
- **License / cost:** GPL-2.0-or-later (libx265).

### Option C: AV1 (SVT-AV1)
- **Pros:** the smallest files; royalty-free intent; modern browsers decode it.
- **Cons:** the slowest to encode on weak CPUs; older devices lack hardware decoding.
- **License / cost:** BSD-3-Clause-Clear plus the AOMedia Patent License 1.0.

### Option D: Hardware encoders (VAAPI, QSV, NVENC, V4L2 M2M on the Pi)
- **Pros:** fast and light on the CPU.
- **Cons:** lower quality per bit than software encoders; availability differs per machine.
- **License / cost:** depends on the driver.

## Decision

**Recommended:**

- **H.264 (libx264) with CRF by default**, in an MP4 container with `+faststart`; audio copied unless the user chooses to re-encode it (AAC at a set bitrate).
- **HEVC and AV1 offered as opt-in choices** only if the user wants them (Q48), with a clear note on encoding time and playback support; they are checked against the S04.8 player before being offered.
- **Hardware encoding** used when detected and enabled (as ADR-0020), with a quality check in the preview.
- **Metadata:** creation time, GPS, and rotation copied (FFmpeg `-map_metadata 0`, rotation kept as display matrix; ExifTool for fields FFmpeg drops), verified after encoding (NFR-039).
- Every encoder runs inside the separately installed FFmpeg program, never linked.

## Consequences

- **Easier:** optimized videos play everywhere the NAS is used; no new tool.
- **Harder:** H.264 files are larger than HEVC or AV1; users who want the smallest files must choose a slower codec.
- **Required (follow-up work, constraints this imposes):** S12.2 builds the pipeline; `dependencies.md` records the encoder licences (x264, x265, SVT-AV1); the S13.6 licence audit checks the FFmpeg build.

## Approval record

_Pending: put to the user with the P005 report (S007). Q48 is asked at the same time._

## Links

- **Related requirements:** FR-176, FR-180, FR-190, FR-192, NFR-039
- **Related ADRs:** ADR-0012 (media toolchain), ADR-0020 (video streaming), ADR-0026 (originals retention)
- **Related stages:** S12.2, S12.4, S12.5
- **Plan version:** 1.4.0 (RK-36)
