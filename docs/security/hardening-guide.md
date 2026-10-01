# Hardening guide for operators

| Field | Value |
|---|---|
| Version | 1.0 (2026-10-01, plan change request #10) |
| For | Whoever installs and runs local-ai-nas: on Debian or Ubuntu, Raspberry Pi OS, Arch Linux, Windows 11, or Docker |
| Plan requirements | NFR-073, NFR-076, NFR-077 (what the software and its deployers do); this guide covers what **you** do |
| Aligned with | CIS Benchmarks (Docker, Debian/Ubuntu, Windows) and ISO/IEC 27001:2022 controls 7.x, 8.9, 8.20–8.22 (based on public summaries) |
| Status | Guidance. Most automatic steps arrive with the deployers (S14.2); until then this is the checklist to follow by hand. |

The NAS protects itself as far as software can. The computer, the network, and the room around it are yours. This guide lists what makes the biggest difference, most important first.

## 1. Everywhere

- [ ] **Keep the operating system updated automatically** (unattended upgrades on Debian and Raspberry Pi OS; Windows Update on Windows).
- [ ] **Keep local-ai-nas updated.** Updates are signed and verified (FR-381); turn on the update check if you want to be told.
- [ ] **Run nothing else on the NAS computer** if you can. Every extra service is another way in.
- [ ] **Use the HTTPS address** on your network and check the certificate fingerprint once against the one the admin console shows.
- [ ] **Give every person their own account**; never share the admin account. Turn on two-factor authentication for admins.
- [ ] **Turn on disk encryption** (LUKS on Linux, BitLocker on Windows) for the system disk and the storage drives. It protects your data if a drive or the whole computer is stolen, and it makes deleted data unrecoverable when you destroy the key (see the [storage security guide](storage-security-guide.md)).
- [ ] **Keep backups**, including one copy that is not connected to the NAS.

## 2. Your router and network

- [ ] **No port forwarding** to the NAS until you deliberately set up private remote access (R06) or the public release (R09). Never forward the admin console.
- [ ] **Turn off UPnP** on the router, so no device can open ports by itself.
- [ ] Change the router's default password and keep its firmware updated.
- [ ] Optional: put the NAS and your own devices on a different network segment from guests and smart-home devices.

## 3. Linux (Debian, Ubuntu, Raspberry Pi OS, Arch Linux)

What the deployer does (S14.2; you can check it):
- The NAS runs as its **own unprivileged user**, which owns the storage root; files 0640, folders 0750, internal data 0700, secrets 0600 (NFR-073). It never runs as root; disk operations go through the small storage helper (S15.2).
- A **hardened systemd unit**: `NoNewPrivileges`, `ProtectSystem=strict` with write access only to the storage root, internal data, and configuration, `ProtectHome`, `PrivateTmp`, no capabilities, `RestrictSUIDSGID`, a system-call filter, `UMask=0027`; checked with `systemd-analyze security` against a target score (NFR-077).
- Where the NAS manages the drives (pools), the storage is mounted `noexec,nosuid,nodev`.
- With your consent, a **firewall rule** that allows the NAS port only from your local network.
- It never turns off SELinux or AppArmor.

What you do:
- [ ] **SSH:** only if you need it; keys only (`PasswordAuthentication no`), no root login.
- [ ] Set a **BIOS/UEFI password** and boot only from the internal disk, if the hardware allows.
- [ ] **Raspberry Pi:** use a good power supply (power cuts corrupt SD cards); prefer an SSD over an SD card for the system; keep the firmware updated. The media-tool sandbox uses the Landlock kernel feature when available; on some Raspberry Pi kernels it is not enabled by default (see section 7).

## 4. Windows 11

What the deployer does (S14.2):
- Runs the NAS as a **Windows service under a virtual service account**, not LocalSystem.
- Adds a **firewall rule for Private networks only**, never Public.
- **Never turns off Windows Defender** and never adds exclusions without asking and explaining why.

What you do:
- [ ] Mark your home network as **Private** in Windows settings, and public Wi-Fi as Public.
- [ ] Turn on **BitLocker** (or Device Encryption) for the system and storage drives.
- [ ] Use a standard (non-administrator) account for daily work on that computer.

## 5. Docker

What the image and the example Compose file do (S14.1, CIS Docker Benchmark):
- A **non-root user**, a **read-only root filesystem**, **all Linux capabilities dropped**, `no-new-privileges`, Docker's default seccomp profile, resource limits, a health check, and base images pinned by digest.
- **No Docker socket** is ever mounted into any container.
- Images are scanned (Trivy) with no unresolved High finding at release, and signed (Cosign).

What you do:
- [ ] Verify the image signature before running it (the command is in the install guide, S14.4).
- [ ] Do not add `--privileged`, extra capabilities, or the host network unless the guide says so for a specific feature.
- [ ] Keep Docker itself updated.

## 5a. ZFS and NAS platforms (plan 1.13.0; planned features)

- **ZFS pools** (optional, S15): install ZFS only from your distribution; on a Raspberry Pi use 8 GB of RAM or more; after a kernel update check the NAS's health page, which reports whether the ZFS module loaded and the pools are imported. On Arch Linux ZFS comes from a third-party repository (ArchZFS).
- **On TrueNAS, Unraid, OpenMediaVault, Proxmox VE, Synology, or QNAP** (R13): keep the host updated and its own admin interface off the internet; do not share the photos area for writing through the host's SMB or NFS; keep the host's snapshots on.

## 6. Physical security (ISO/IEC 27001:2022 controls 7.1–7.14; the software cannot do this for you)

- [ ] Keep the NAS where strangers cannot reach it, away from heat and water, with ventilation.
- [ ] Use a UPS if power cuts happen where you live (UPS support arrives in R04).
- [ ] Before you give away, sell, or throw away a drive, **retire it in the NAS** (secure erase, S15.11), or destroy the encryption key if it was encrypted.
- [ ] Lock the screen of any computer that is signed in to the NAS's admin console.

## 7. Known platform notes (checked 2026-10-01)

- **Landlock on Raspberry Pi:** the media-tool sandbox restricts each tool's file access with the Linux Landlock feature when the kernel offers it (NFR-064). A report for Ubuntu on the Raspberry Pi found Landlock missing from the default list of security modules; it was enabled by adding `lsm=landlock,lockdown,yama,integrity,apparmor` to `/boot/firmware/cmdline.txt`. Whether current Raspberry Pi OS kernels enable it is **Unverified**; the NAS reports in its health page whether Landlock is active, and the deployer will offer the change (with your consent) once it is verified.
- **Windows sandboxing is weaker** than on Linux: tools are limited by Job Objects (time, memory, processes) and a low-privileged account, but there is no equivalent of Landlock's per-file limits for them (ADR-0045).
