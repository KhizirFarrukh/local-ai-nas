# Storage security guide

| Field | Value |
|---|---|
| Version | 1.0 (2026-10-01, plan change request #10; published at the owner's request, decision D4) |
| For | Operators and users who want to know how their stored data is protected |
| Aligned with | ISO/IEC 27040:2024 (storage security) and ISO/IEC 27001:2022 controls 7.10, 7.14, 8.10, 8.13 (based on public summaries; verify against the standards) |
| Status | Each section says what exists today and what is planned (with its stage). Nothing planned is claimed as working. |

## 1. What is stored where

| Data | Where | Notes |
|---|---|---|
| Your files | `<storage root>/files/<user>/` | Exactly as you uploaded them |
| Your photos and videos | `<storage root>/photos/<user>/` | Originals are never re-encoded by the NAS on its own (NG4); sidecar files beside them hold descriptive metadata (S05) |
| Internal data | `<storage root>/.local-ai-nas/` | `state/` (database, snapshots, trash, keys: durable, backed up), `cache/` (thumbnails, search index, transcodes: rebuildable, never backed up), `tmp/`, `logs/` |

The two areas never mix (invariant I1). Each user's data is private unless shared (I5), on every access path.

## 2. Protection at rest

| Protection | Today | Planned |
|---|---|---|
| **Nothing stored is ever executed** (invariant I12): files are stored without execute permission and are never run, imported, or loaded as code | The NAS has no code path that runs stored content | File modes 0640 and 0750 under the NAS's own account (S03.5-T05, NFR-073) |
| **Only the NAS can read its data on the computer**: the database, logs, keys, and configuration belong to the service account | Log files 0600 | Internal data 0700, secrets 0600, and a startup check that warns about (or, for secret files, refuses) permissions that are too open (S03.5-T05, NFR-073) |
| **`noexec,nosuid,nodev` mounts** where the NAS controls the mount (its own drives and pools) | — | S15 (NFR-073) |
| **Encryption at rest** | Use the operating system's disk encryption (LUKS, BitLocker); see the [hardening guide](hardening-guide.md) | Built-in full-disk encryption for pools with unlock at boot (R05, FR-266) |
| **Secrets in the database are encrypted** (SMTP and webhook credentials, two-factor secrets) with a key kept outside the database, so a database backup alone reveals nothing | — | S03 (NFR-074, ADR-0048) |
| **Passwords and tokens are never stored**, only their hashes (Argon2id for passwords, SHA-256 for tokens) | Argon2id built (S03.2-T02) | Tokens in S03.3 |

## 3. Integrity

| Protection | Today | Planned |
|---|---|---|
| Content hashes (SHA-256) of uploaded files | Yes (since S01.3-T10) | Used by duplicate detection (S11) and integrity scans |
| Crash safety: writes go to a temporary file, then an atomic rename; the database uses `synchronous=FULL`, so a committed change survives a power cut | Yes (S01, S01.1-T12) | The operation journal makes multi-step changes all-or-nothing (S01.4-T09) |
| Scheduled integrity scans that compare files with their hashes | — | S08.2 |
| Mirrored drives (RAID 1) with scrubs | — | S15 |
| A tamper-evident audit log | — | S03.6 (FR-379) |

## 4. Backups and recovery

| Protection | Today | Planned |
|---|---|---|
| Database snapshots (daily and before every migration) | — | S03.2-T06 (FR-355) |
| Metadata and configuration backup; disaster recovery steps | — | S08.3, S08.4 |
| Backups to an external drive | — | S08.6 |
| Encrypted off-site backup to a target you choose | — | R04 |

**What you should do now:** keep a copy of everything important outside the NAS. A copy on the same disk protects against mistakes and bad updates, not against a failed or stolen disk.

## 5. Deletion and sanitization

- **Trash** (planned: S01.3-T13 minimal trash, S08.1 full trash with retention; today a delete is permanent): deleting will move an item to the trash, where it can be restored until it is emptied or its retention period ends. Retention periods for trash, logs, and backups are documented and enforced (NFR-075).
- **Permanent deletion** removes the file from the filesystem. On SSDs and SD cards, the hardware may keep old data blocks for a while (wear levelling), so **permanent deletion cannot guarantee physical erasure on flash storage**. With disk encryption, data that is deleted (or a drive whose key is destroyed) cannot be recovered: this is called crypto-erase.
- **Retiring a drive:** the NAS can securely erase a drive it retires (S15.11), with a preview and a typed confirmation; for SSDs it uses the drive's own erase command where available. Planned.
- **Account deletion** removes a user's data after a confirmation and an undo window (I10; S07, R01 export first).

## 6. Who can reach the stored data

- **Through the NAS:** only signed-in users, each limited to their own data and what was shared with them (I5), checked on every request by one central authorization check (FR-089, S03.5) and tested across users (S07.7).
- **Through network shares** (WebDAV, S09): the same accounts and checks, over HTTPS only (NFR-072).
- **On the computer itself:** only the NAS's service account (and an administrator or root, who can always read everything; that is why the computer itself must be secured, see the hardening guide).
- **Media tools** (thumbnails, metadata, video): each runs in a sandbox that can read only the one file it works on and write only to its scratch folder (NFR-064, S04).

## 7. What you, the operator, must do

1. Encrypt the disks (hardening guide, section 1).
2. Keep a backup outside the NAS.
3. Keep the NAS and its operating system updated.
4. Retire drives through the NAS (or destroy their encryption key) before they leave your hands.
5. Do not give other programs or users on the NAS computer access to the storage root.
