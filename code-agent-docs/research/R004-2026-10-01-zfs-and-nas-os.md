# R004: ZFS, NAS operating systems, and hosting a NAS

| Field | Value |
|---|---|
| Date | 2026-10-01 (session S007, E076) |
| Trigger | The user: "updates: add support for: zfs, NAS OS, and everything related to this and hosting a NAS." and "also, approve the tasks and if you have questions, do what you deem best for that." |
| Used by | Plan 1.13.0: S15 (ZFS backend), releases R13 and R14, FR-383–FR-391, ADR-0027 revision 1, ADR-0051–ADR-0053 |
| Method | Web checks on 2026-10-01 (sources below); facts not checked are marked **Unverified** or "general knowledge" |

## 1. What "NAS OS" can mean, and what was decided

The request has three readings. The user delegated open questions to the agent (E076), so the agent decided, and the user can change any of it:

| Reading | Decision | Where |
|---|---|---|
| Run local-ai-nas **on** an existing NAS operating system (TrueNAS, Unraid, OpenMediaVault, Proxmox VE, Synology DSM, QNAP) | Yes: tested templates and guides; the host keeps managing drives, pools, and its own shares | New release **R13** |
| Ship local-ai-nas **as** a NAS operating system: a ready-to-flash image (Raspberry Pi first, then PCs) that manages the host (network, updates, power, services) | Yes: built on Raspberry Pi OS and Debian, not a new distribution | New release **R14** |
| "Everything related to hosting a NAS": the services a NAS host offers | Mapped feature by feature (section 4): most already have a place in the plan; the missing ones go to R14 or are excluded with a reason | R14, existing releases |

**ZFS** is decided separately (section 2): an optional second pool backend in the drives stage S15, next to mdadm, which stays the default.

## 2. ZFS facts (checked 2026-10-01)

| Fact | Source |
|---|---|
| OpenZFS 2.3 (January 2025) added **RAIDZ expansion** (adding a disk to an existing RAIDZ group without downtime), fast dedup, and Direct I/O; Linux 4.18–6.12 | Phoronix, "OpenZFS 2.3 Released" |
| OpenZFS 2.4 supports Linux 4.18–6.18, faster encryption with AVX2, default user, group, and project quotas; a 2.4.x line is current (2.4.4 listed on endoflife.date, August 2026) | Phoronix, 9to5Linux, endoflife.date |
| **Memory:** since OpenZFS 2.3 the Linux default ARC limit is **max(RAM − 1 GB, 5/8 of RAM)** (it was half of RAM). On a 4 GB Raspberry Pi that leaves the NAS very little unless the limit is set | OpenZFS commit "arc_default_max on Linux should match FreeBSD"; thalheim.io (2025) |
| **Debian / Raspberry Pi OS:** ZFS comes as `zfs-dkms` and `zfsutils-linux` (Debian contrib; backports for newer versions), built for the running kernel with DKMS; on Raspberry Pi OS the Pi kernel headers are needed and `--no-install-recommends` avoids pulling the Debian kernel. Forum reports show ZFS **failing after Raspberry Pi kernel updates** until the module is rebuilt, and the header package names changed on Trixie for the Pi 5 | Raspberry Pi forums (several threads), OpenMediaVault forum "ZFS on Raspberry Pi", a 2026 talk "Using ZFS on a souped-up Raspberry Pi 5" |
| **Arch Linux:** ZFS is not in the official repositories (license); it comes from the independent **ArchZFS** repository or the AUR | ArchWiki "ZFS", archzfs on GitHub |
| **Windows:** the OpenZFS port is in release candidates (zfswin-2.4.1rc1); not for production | search results on the Windows port; **Unverified** in its own documentation |
| **License:** CDDL-1.0, a separate kernel module; the project never ships it, the distribution's packages are installed with the operator's consent | ADR-0027 option C; general knowledge |

**Consequences for the design** (ADR-0051): ZFS is optional and Linux-only; mdadm stays the default (Raspberry Pi first: in-kernel, little memory, no rebuild after kernel updates); the NAS sets an ARC limit on small machines; the health page checks that the ZFS module matches the running kernel; ZFS stripe and mirror cover RAID 0 and RAID 1 under the user's capacity rule; RAIDZ waits with the other parity layouts (11a), now with a growth path.

## 3. NAS operating systems (checked 2026-10-01)

| Platform | What matters for local-ai-nas | Source |
|---|---|---|
| **TrueNAS** (Community Edition; "SCALE" until the 25.04 "Fangtooth" unification) | Apps run on Docker since 24.10 "Electric Eel" (Kubernetes before); custom apps from a Compose YAML; third-party catalogs did not migrate automatically in that change. Storage is ZFS | TrueNAS 24.10 and 25.04 release notes and blog |
| **Unraid** | Native ZFS pools since Unraid 7 (RAIDZ expansion in 7.2); Docker apps from templates in Community Applications | Unraid 7.0.0 release notes, Unraid blog, heise |
| **OpenMediaVault** | Debian-based; Docker through the compose plugin of omv-extras; ZFS through a plugin (with the Proxmox kernel); runs on a Raspberry Pi | omv-extras wiki (OMV7 compose and ZFS plugins) |
| **Proxmox VE** | ZFS is a standard option; services run in LXC containers or VMs | General knowledge (**Unverified** here) |
| **Synology DSM** | Docker through Container Manager on supported models; Btrfs volumes; system folders `@eaDir`, `#recycle`, `#snapshot` | General knowledge (**Unverified** here) |
| **QNAP QTS / QuTS hero** | Container Station; QuTS hero uses ZFS; system folders such as `@Recycle`, `.@__thumb`, `@Recently-Snapshot` | General knowledge (**Unverified** here) |

**Lesson:** app systems on these platforms change (TrueNAS moved from Kubernetes to Docker in 2024), so the templates are tested at every release and versioned (RK-65).

**Data-integrity finding:** these platforms put **system folders inside shared folders**: ZFS's `.zfs` snapshot directory (when visible), Synology's `@eaDir` thumbnails and `#recycle` bin, QNAP's `@Recycle` and `.@__thumb`. If the storage root sits on such a share, the external-change watcher (S09.4) or a folder import (S04.2) would ingest snapshot copies and thumbnails as user files: duplicated libraries and leaked deleted files. **Fix (FR-385, MVP, R14 exception for data integrity):** these names are reserved, hidden, never indexed, ingested, followed, or written.

## 4. What a NAS host offers, mapped to the plan

| Feature (common to TrueNAS, OpenMediaVault, Unraid, Synology; general knowledge) | Plan |
|---|---|
| Pools: mdadm, ZFS, Btrfs | S15 (mdadm default; **ZFS optional, FR-383**) |
| Snapshots and replication | R04 (FR-256, FR-257; on ZFS: native snapshots, holds, `send`/`receive`) |
| Scrubs and integrity checks | S08.2, S15.7 (ZFS scrub results) |
| SMART and drive health | S10.3 |
| SMB, WebDAV | S09 (ADR-0019 for SMB) |
| **NFS, SFTP, Time Machine, rsync target** | **R14 (FR-389)** |
| FTP | Not planned (cleartext; SFTP covers it): X-12 |
| iSCSI (block storage) | Not planned (a block device bypasses the areas, the authorization, and the index): X-13 |
| Users, groups, ACLs, quotas | S07, R03, S10 |
| Directory services (LDAP, Active Directory) | R05 (FR-264; LDAP Could) |
| Apps, containers, VMs | **Stays excluded (X-09):** more attack surface than the security program allows (I12, R15 Tier 1), too heavy for the Raspberry Pi; host platforms (R13) already offer it |
| UPS | R04 (FR-259) |
| Alerts | S10 (FR-221, FR-378) |
| Cloud sync and backup tasks | R01, R04 |
| VPN, remote access, certificates, dynamic DNS | R06 |
| **Network settings, time, OS updates, SSH, firewall, power, Wake-on-LAN** | **R14 (FR-388)** |
| Disk spin-down | S16.6 (Q73) |
| Hardware monitoring | FR-360 (temperatures); fan control not planned |
| Antivirus | R05 (FR-267) |
| Media server, DLNA | R10 |
| Encryption | R05 (FR-266); ZFS native encryption as an option on ZFS pools (ADR-0051, later) |
| **SNMP, syslog forwarding** | **R14 (FR-390, Could)** |
| Mirrored boot drive | Not planned in the first version (11a candidate) |

## 5. Sources

- [OpenZFS 2.3 released](https://www.phoronix.com/news/OpenZFS-2.3-Released), [OpenZFS 2.4 released](https://www.phoronix.com/news/OpenZFS-2.4-Released), [9to5Linux on 2.4](https://9to5linux.com/openzfs-2-4-released-with-linux-6-18-lts-support-quotas-uncached-io-and-more), [endoflife.date OpenZFS](https://endoflife.date/openzfs)
- [OpenZFS ARC default commit](https://github.com/openzfs/zfs/commit/6a629f32344468ae81b264055916641480cb438d), [ZFS ate my RAM (2025)](https://blog.thalheim.io/2025/10/17/zfs-ate-my-ram-understanding-the-arc-cache/)
- [Raspberry Pi forum: ZFS after a kernel update](https://forums.raspberrypi.com/viewtopic.php?t=383630), [DKMS on Bookworm](https://forums.raspberrypi.com/viewtopic.php?t=357549), [OMV forum: ZFS on Raspberry Pi](https://forum.openmediavault.org/index.php?thread%2F53494-zfs-on-raspberry-pi-howto%2F=)
- [ArchWiki: ZFS](https://wiki.archlinux.org/title/ZFS), [ArchZFS](https://github.com/archzfs/archzfs)
- [TrueNAS 24.10 notes](https://www.truenas.com/docs/scale/24.10/gettingstarted/scalereleasenotes/), [TrueNAS 25.04 notes](https://www.truenas.com/docs/scale/25.04/gettingstarted/scalereleasenotes/)
- [Unraid 7.0.0 notes](https://docs.unraid.net/unraid-os/release-notes/7.0.0/), [Unraid 7.2.0](https://unraid.net/blog/unraid-7-2-0)
- [OMV7 compose plugin](https://wiki.omv-extras.org/doku.php?id=omv7%3Aomv7_plugins%3Adocker_compose), [OMV7 ZFS plugin](https://wiki.omv-extras.org/doku.php?id=omv7%3Aomv7_plugins%3Azfs)
- [rpi-image-gen (Raspberry Pi, 2025)](https://www.raspberrypi.com/news/introducing-rpi-image-gen-build-highly-customised-raspberry-pi-software-images/), [rpi-image-gen on GitHub](https://github.com/raspberrypi/rpi-image-gen)
