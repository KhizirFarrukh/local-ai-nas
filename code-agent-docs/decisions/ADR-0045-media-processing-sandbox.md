# ADR-0045: Media processing sandbox

| Field | Value |
|---|---|
| Number | ADR-0045 |
| Status | Proposed |
| Date proposed | 2026-10-01 (session S007) |
| Date of last status change | 2026-10-01 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

ExifTool, libvips with libheif, and FFmpeg read files that attackers can craft. Media parsers have had code-execution bugs: ExifTool CVE-2021-22204 (DjVu annotations passed to Perl's `eval`, fixed in 12.24) and CVE-2022-23935 (a file name ending in `|` was opened as a pipe and run as a command, fixed in 12.38), both verified 2026-10-01. These tools are the most realistic way an upload could run code on the NAS, so they need their own defenses besides patching (the user's requirements UR-3 and UR-5; invariant I12; NFR-064, NFR-065; threats T-50, T-60, T-64, T-65, T-68, T-72). They are Tier 1 (the user's decision D1). The decision must be made before any tool processes an upload (S04.3, S04.4).

Constraints: the Raspberry Pi is the production machine (NFR-051); Windows 11 is a supported platform; pure-Go builds are preferred (ADR-0001); no root for the core (NFR-037).

**Verified on 2026-10-01:**
- Go's `os.Root` (since Go 1.24) refuses `..` and links that escape the root and, on Windows, device names such as `NUL`; it does not stop bind mounts (which need root to create). The project uses Go 1.27.
- Landlock (Linux 5.13 and later) restricts a process's own file access without privileges. `github.com/landlock-lsm/go-landlock`: MIT, with a best-effort mode that degrades on older kernels.
- Landlock is **not** in the default security-module list of at least some Raspberry Pi kernels: a Launchpad report for Ubuntu on the Raspberry Pi showed `CONFIG_LSM` without `landlock`; adding `lsm=landlock,…` to the kernel command line enabled it. Raspberry Pi OS: **Unverified**. Third-party sources say Debian 12 and later enable it: **Unverified** against Debian's own sources.
- `github.com/elastic/go-seccomp-bpf`: Apache-2.0, pure Go (no libseccomp), system-call tables for 386, amd64, arm, and arm64.
- libvips 8.13 and later can block operations marked untrusted (`VIPS_BLOCK_UNTRUSTED`, `vips_block_untrusted_set`).
- FFmpeg accepts `-protocol_whitelist` and an explicit input format; a crafted playlist can otherwise make it read other local files or URLs.
- ExifTool's argument files (also used by `-stay_open`) take one argument per line, strip leading white space, and treat `#` lines as comments; `-charset filename=UTF8` controls file-name encoding (ExifTool 13.59 documentation).

## Options considered

### Option A: Process limits only (timeouts, rlimits or Job Objects, no shell)
- **Pros:** simple; works on every platform; low cost on a Pi.
- **Cons:** a compromised tool can still read every file the service can read and use the network.
- **License / cost:** none.

### Option B: Process limits plus OS isolation where available (recommended)
- One wrapper package (planned `internal/extexec`), the only code allowed to start programs (lint-enforced).
- **Per call:** the binary resolved to an absolute path at startup with its version checked; an argument list, never a shell; `--` before file arguments where supported; **core-generated file names** in a per-job scratch folder (a hard link where the filesystem allows, else a copy), so tools never see user-chosen names; a wall-clock timeout; CPU, memory, and file-size limits; no core dumps; lower CPU and I/O priority (NFR-055); the scratch folder deleted afterwards.
- **Linux:** the binary re-executes itself in a small sandbox mode (Go cannot run code between fork and exec), which applies to itself: rlimits, `no_new_privs`, a seccomp filter that denies network sockets (or a fresh network namespace where unprivileged namespaces are allowed), and Landlock limited to read access to the tool's own installation and the scratch folder; then it starts the tool. The `-stay_open` ExifTool process is restricted the same way to its scratch folder.
- **Windows:** a Job Object (time, memory, active processes, no breakaway) and the least-privileged account available; no per-file isolation (documented as weaker).
- **Docker:** the container's own restrictions (non-root, read-only root, no capabilities, default seccomp) apply on top.
- **Reporting:** the health page shows the sandbox level actually in force (full, no Landlock, Windows).
- **Pros:** strong containment on Linux at little cost; degrades safely; no new privileged component.
- **Cons:** more code; Landlock is missing on some Pi kernels; must be tested on Raspberry Pi OS, Debian, and x86.
- **License / cost:** go-landlock (MIT), go-seccomp-bpf (Apache-2.0); both compatible with AGPL-3.0-or-later.

### Option C: Run the tools in a separate container or VM per job (e.g. gVisor, Firecracker)
- **Pros:** the strongest isolation.
- **Cons:** heavy on a Raspberry Pi; needs a container runtime on every platform; complex.
- **License / cost:** varies; operational cost high.

## Decision

**Option B (recommendation while Proposed).** Every tool call goes through the wrapper with process limits everywhere and OS isolation where the platform offers it. Tool-specific restrictions (NFR-065): libvips with untrusted operations blocked and only the plan's formats allowed; FFmpeg with `-protocol_whitelist file` and an explicit `-f` placed before every input so nothing can override them; ExifTool 12.38 or newer with core-generated names and `-charset filename=UTF8`; ExifTool writes only in S12 on files the core created. Limits against resource bombs: images over 250 megapixels (configurable, read from the header first), and videos over duration, resolution, or stream-count limits are stored but not processed, and the owner is told why. A hostile media corpus runs through every tool in the S04.9 and S05.8 tests.

To verify when S04 is planned: the Landlock ABI version on Raspberry Pi OS and Debian kernels and how to enable it with the user's consent; whether unprivileged network namespaces are allowed there; the libvips mechanism in the version shipped in the Docker image and on each platform; the FFmpeg options in the packaged version; the cost per job on the Pi profile.

## Consequences

- **Easier:** one place to audit every external program; a compromised parser is contained.
- **Harder:** the re-exec sandbox mode and its tests on several kernels; Windows stays weaker.
- **Required:** the wrapper and the forbidden-API lint rule before S04.3/S04.4 process uploads; register entries for go-landlock and go-seccomp-bpf if adopted; the hostile corpus (license-clean, generated); threat model and Statement of Applicability updates.

## Approval record

_Not yet approved._

## Links

- **Related requirements:** NFR-064, NFR-065, NFR-055, NFR-051, NFR-063; I12
- **Related ADRs:** ADR-0012 (media toolchain), ADR-0020 (video streaming), ADR-0029 (storage helper)
- **Related stages:** S04, S05, S12, S17
- **Plan version:** 1.12.0
