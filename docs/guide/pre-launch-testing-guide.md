# local-ai-nas: Testing Before Going Live

**Version 2** — updated for the current plan: MVP, drive pools and SSD caching, releases R01–R12, client apps, the public release, and AI last.

A practical guide for proving the NAS is safe to trust with real files and photos before anyone depends on it, and for doing the same before every later release.

> **The golden rule:** never test with your only copy of anything. Until a release has passed every step below, the NAS holds **copies** of your data. The originals stay where they are today.

---

## How this guide is organised

local-ai-nas does not go live once. It ships in **releases**, and each release adds features that need their own testing before you rely on them.

- **Part 1** is the full procedure for the **first release (the MVP)**, the first time real data goes on the NAS.
- **Part 2** lists the **extra tests for each later release**. Part 1's procedure still applies; Part 2 adds what is new.
- **Part 3** covers **going live**, the watch period, rollback, and what to do when something goes wrong.

### The releases (as currently planned)

Version labels are suggestions; you choose the final ones. Some items depend on decisions still open in the plan.

| Release | What it adds | Risk level |
|---|---|---|
| **MVP (v1.0)** | Files and photos areas, GUI, security, metadata sidecars, search, multi-user and sharing, trash, backups, network drive, admin, duplicates, storage optimization, packaging. Plus early additions: stable item IDs, Live Photos, phone backup through WebDAV apps, alerts, drive health warnings, move-to-new-drive command | High (first real data) |
| **Drives and storage** | RAID 0/1 pools, new-drive wizard, upgrades, mirrors, drive replacement, SSD caching | **Very high** (erases drives) |
| **R01 Migration** | Google Takeout and iCloud import, other clouds, full data export | High (bulk writes) |
| **R02 Everyday essentials** | Map, memories, slideshow, ratings, smart albums, zip/unzip, batch rename | Medium |
| **R03 Family sharing** | Groups, partner sharing, shared albums, comments, activity | High (privacy) |
| **R04 Data safety** | Rewind, ransomware detection, off-site backup, UPS | High |
| **R05 Security** | 2FA, passkeys, SSO, locked folder, disk encryption, virus scanning | High |
| **R06 Private remote access** | VPN, relay for CGNAT, certificates | High (network) |
| **R07 Mobile apps** | Android and iOS: auto backup, folder sync, share to NAS | High |
| **R08 Desktop apps** | Windows, macOS, Linux: folder sync, photo backup, share to NAS, CLI | High (sync can delete) |
| **R09 Public release** | Reachable from the internet, public links, file requests | **Highest** |
| **R10–R12** | Media center, documents and office, automation | Medium |
| **AI (always last)** | Auto-classification, face grouping, smart search | Medium (privacy) |

Keep a **test log** (spreadsheet or markdown): date, release, what you tested, result, bug found. Every bug is fixed and re-tested before a phase counts as passed.

---

# Part 1: Testing the first release (MVP)

| Phase | Where | Data | Goal | Typical duration |
|---|---|---|---|---|
| 1. Automated checks | CI / dev PC | Test fixtures | The code does what the plan says | Hours |
| 2. Clean install | Spare machine or VM | None | Anyone can install it | 1 day |
| 3. Functional walkthrough | Test machine | Sample data | Every feature works by hand | 3–4 days |
| 4. Failure drills | Test machine | Throwaway copies | Nothing is lost when things break | 3 days |
| 5. Security testing | Test machine | Sample data | Nobody gets in who shouldn't | 2–3 days |
| 6. Scale and endurance | Real hardware | Full-size copy | Fast and stable for weeks | 2–4 weeks |
| 7. Real-use pilot | Real hardware | Copy of real data | It works in daily life | 2–4 weeks |

## Phase 1: Automated checks

Confirm on the **exact release build**, not a developer build:

- [ ] All unit, integration, and system tests pass on Linux and Windows.
- [ ] Coverage is at or above the 80% target.
- [ ] **Crash-injection tests** pass: every multi-step operation (upload, move, delete, trash, sidecar write) recovers correctly when interrupted at each step.
- [ ] **Local-origin tests** pass: requests with a wrong Host header or from another website are refused.
- [ ] **Search golden queries** pass, including Urdu, Roman Urdu, and mixed-script names.
- [ ] **Responsiveness under load** passes: browsing and search stay fast while a 10,000-file import runs.
- [ ] Dependency vulnerability and license scans show no open high-severity findings.
- [ ] Container images build for `amd64` and `arm64`.
- [ ] If you used an internal alpha build: **upgrade from the alpha** to the release with data, and nothing is lost.

**Pass:** everything green on the release artifacts.

## Phase 2: Clean install test

On a machine that has never seen the project (spare PC, fresh Raspberry Pi card, or a virtual machine), follow **only** the install guide.

- [ ] Linux x86-64 setup script
- [ ] Raspberry Pi setup script
- [ ] Windows 11 setup script
- [ ] Docker Compose (`docker compose up -d`)

For each:
- [ ] The health check passes; missing tools (ExifTool, FFmpeg, libvips) are reported clearly if absent.
- [ ] The first-run wizard creates the admin account (no default password exists).
- [ ] The NAS restarts on its own after a reboot.
- [ ] Running the setup script again breaks nothing.
- [ ] Uninstalling leaves the data folder untouched.

**Tip:** have a non-programmer follow the guide while you watch without helping. Every point where they get stuck is a documentation bug.

## Phase 3: Functional walkthrough

Use a **sample data set** of a few hundred files and photos covering every case below.

### Files
- [ ] Upload files, many files, and whole folders (drag and drop and button).
- [ ] Upload a 10+ GB file and pull the network cable halfway; it resumes.
- [ ] Rename, move, copy, delete. **Delete goes to trash**; restore brings it back to the same place.
- [ ] Rename or move a file that is shared or has a shortcut; the share and shortcut still work (stable item IDs).
- [ ] Large copies and deletes run in the background with progress, and can be cancelled.
- [ ] Download a file, a folder as zip, and resume a large download.
- [ ] Previews: images, PDF, text, audio, video.
- [ ] File names with Urdu, Arabic, emoji, and very long names.

### Photos
- [ ] JPEG, PNG, HEIC, and videos from different phones.
- [ ] **iPhone Live Photos** show as one item with a Live badge, whether the photo or the video part arrives first.
- [ ] **Android motion photos** play their motion part.
- [ ] Timeline dates are correct, including photos from other time zones.
- [ ] Places show correct city names from GPS.
- [ ] Albums, favorites, archive, hide.
- [ ] Video plays; the quality selector switches live.
- [ ] The photo's metadata file sits next to it with the project's own name (e.g. `IMG_0001.jpg.lainas.json`, per the plan's decision), and a Google Takeout `.json` file in the same folder is left untouched.
- [ ] Move a photo from photos to files and back; metadata survives.

### Phone backup (MVP bridge)
- [ ] An auto-upload app on **Android** and on **iPhone**, set up with the NAS's camera-upload address and a device password, backs up new photos automatically.
- [ ] The app cannot see or change the rest of the library.
- [ ] Revoking that device's password stops it immediately.

### Search
- [ ] `receipts` finds `receipt`, `invoice`, `voucher`; `reciept` still works.
- [ ] `before:2026`, `after:2024`, `place:karachi`, `type:video`, `in:files` work alone and combined.
- [ ] Person and file names are not "stemmed" (searching `Ahmed` does not match unrelated words).
- [ ] Urdu text matches whether typed with Arabic or Urdu letter forms.
- [ ] A malformed query shows a helpful hint, never an error page.

### Multi-user and sharing
- [ ] A second user cannot see your items in any view, search, thumbnail, count, or suggestion.
- [ ] Share a folder read-only; they see it but cannot change it. Revoke; it disappears everywhere immediately.

### Duplicates, stacks, and storage optimization
- [ ] Exact duplicates and resized copies are found; different photos are not flagged.
- [ ] Burst shots group into a stack; you can choose the cover.
- [ ] Optimization preview matches the result; metadata survives; revert restores the original byte-for-byte.

### Network drive
- [ ] Connect over WebDAV from Windows, macOS (if available), and Linux; you see only your own and shared items.

### Administration, alerts, and drive health
- [ ] The dashboard shows storage, CPU, memory, and queue status.
- [ ] **Test alerts** arrive by email, webhook, and ntfy (whichever you configured).
- [ ] Each drive shows a health status (Healthy, Watch, Replace soon, Replace now) or "not available" with the reason.
- [ ] Quotas stop uploads with a clear message when full.

**Pass:** every item checked, every bug fixed and re-tested.

## Phase 4: Failure drills

The most important phase. Do every drill on the **test machine with throwaway data**.

| Drill | How | What must happen |
|---|---|---|
| **Power cut during upload** | Pull the plug during a large upload and a photo import | No half-written files; no damaged metadata; upload resumes; import continues |
| **Power cut during a big move** | Start moving a 50 GB folder, pull the plug | Every file is in the old place or the new place, never lost; the operation finishes or rolls back on restart |
| **Repeated power cuts** | Cut power several times in an hour (like load-shedding) | Each restart recovers cleanly; no growing mess of temp files |
| **Disk full** | Fill the disk, keep uploading | Clear "disk full" message; nothing damaged; works again after freeing space |
| **App crash** | Kill the NAS process mid-operation | It restarts and recovers by itself |
| **Accidental delete** | Delete a folder of photos | Everything restores from trash with albums and metadata |
| **Database lost** | Stop the NAS, delete the internal database | Restore from the automatic database snapshot or rebuild from disk; shares and users come back from the backup |
| **Full restore from backup** | Wipe the test machine | Fresh install plus external backup restore; file counts and checksums match |
| **External changes** | Add, rename, and move files directly on the disk or over the network drive | The NAS notices; renamed items keep their shares and album places |
| **Failing drive warning** | Use recorded SMART data from a failing drive (test fixtures) | Status changes to "Replace soon" and an alert arrives |
| **Move to a new drive** | Use the storage migrate command to copy everything to a second drive | Every file is verified before switching; switching back to the old drive works |
| **Network drop** | Disconnect Wi-Fi during phone backup | It resumes later with no duplicates and nothing missing |

**Pass:** in every drill, **zero data lost**, and the documented recovery steps worked exactly as written.

> A backup that has never been restored is not a backup. Do the full restore drill even if everything else passes.

## Phase 5: Security testing

Test as an attacker, from another device on your network.

- [ ] Every page and API call without login is refused (try API addresses in a private browser window).
- [ ] Wrong passwords lock the account after the set number of tries.
- [ ] HTTPS works, and plain HTTP redirects to it.
- [ ] "Log out everywhere" ends every session.
- [ ] User A cannot reach User B's items by changing IDs or paths in the address bar.
- [ ] A malicious `.html` or `.svg` upload does not run its script when previewed.
- [ ] **A hostile web page** open in another tab (you can make a tiny test page that tries to send a delete request to the NAS) is refused.
- [ ] **A request with a fake Host name** (DNS rebinding test) is refused.
- [ ] **Editing a photo's metadata file** by hand to add another user to its access list gives that user **no** access.
- [ ] A read-only share cannot be written over the network drive.

Free tools:
- **OWASP ZAP**: scans the web interface for common weaknesses (run it logged in as a normal user).
- **nmap**: `nmap -sV <nas-ip>`; only the expected ports should be open.
- **testssl.sh**: checks the HTTPS setup.
- **Browser developer tools**: confirm security headers are present.

**Pass:** no way found to access data without permission; all findings fixed.

## Phase 6: Scale and endurance testing

Move to the **real hardware** (mini-PC or Raspberry Pi) with a **full-size copy** of your library, around 100,000 photos and 100,000 files.

- [ ] Import the full copy. Note how long it takes; the NAS stays usable during import.
- [ ] Timeline and folders open fast (target under half a second).
- [ ] Search stays fast (target under 0.3 seconds).
- [ ] Rebuilding the search index finishes within the planned time.
- [ ] Video plays smoothly while a big import runs (background work stays at low priority).

**Soak test:** leave it running at least **two weeks** of normal use. Watch for:
- [ ] Memory that keeps growing.
- [ ] Disk space disappearing (logs, caches, temp files).
- [ ] Scheduled jobs (backups, integrity scans, trash cleanup, SMART tests, database snapshots) running on time.
- [ ] Normal drive temperatures.
- [ ] Errors in the logs.

**Several people at once:** one streaming video, one backing up a phone, one browsing. Nothing should stutter badly.

**Pass:** fast at full size, stable for two weeks, no leaks.

## Phase 7: Real-use pilot

Use the NAS for **real daily life** with safety nets in place.

1. **Keep your old services running.** Google Photos, OneDrive, and so on stay active and untouched.
2. Put a **copy** of your real data on the NAS.
3. Use the NAS as your main way to view photos and files for 2–4 weeks.
4. Turn on phone backup to the NAS (while the phone still backs up to the old service).
5. Invite one or two family members as users.
6. Write down every annoyance, bug, and confusing moment.

At the end, answer honestly:
- Did I reach for the old service because the NAS could not do something?
- Did family members manage without asking me?
- Did any photo or file ever look wrong, missing, or duplicated?
- Did alerts actually reach me?

> If the plan's **internal alpha** (after search is done) is used, run this pilot on the alpha first, then again on the MVP.

**Pass:** no data problems, no blocking annoyances, family can use it alone.

---

# Part 2: Extra tests for each later release

For **every** release below, always do these first:

- [ ] **Backup before upgrading.** Confirm the backup is recent and restorable.
- [ ] **Upgrade test:** on a test copy, upgrade from the previous release with real-size data; nothing is lost; everything from the previous release still works.
- [ ] **Rollback test:** you know how to go back to the previous version if the upgrade fails, and you have tried it.
- [ ] Repeat the Phase 1 automated checks and a short Phase 4 power-cut drill on the new version.
- [ ] Run a short pilot (1–2 weeks) before relying on the new features.

Then add the release's own tests.

## Drives and storage (pools, drive lifecycle, SSD caching) — very high risk

These features **erase drives**. Test with virtual disks first, then with spare drives that hold nothing you need.

- [ ] Every destructive step names the drive by model and serial and needs a typed confirmation.
- [ ] The OS drive and drives holding NAS data are never offered for erasing.
- [ ] **Capacity upgrade:** move to a bigger drive; every file is verified; switching back works during the rollback window.
- [ ] **Mirror conversion:** add a second drive to make RAID 1; the old drive is erased only after verification and a second confirmation; the resync completes.
- [ ] **Drive failure:** unplug one mirror drive (power off first if not hot-swappable); the NAS keeps working, shows "degraded", alerts you, and rebuilds onto a replacement.
- [ ] **Failing drive replacement:** replace a mirror member while it still works; protection never drops.
- [ ] **New drive detection:** plug in a drive; the wizard appears with sensible recommendations.
- [ ] **Secure erase** (on a spare drive only) completes and is recorded.
- [ ] **SSD cache:** repeated reads of large files get faster; **unplug the SSD mid-read** and the file still downloads correctly from the hard drive; change a file directly on disk and the cache never serves the old version; statistics never show other users' file names.
- [ ] Reboot with the pool; it assembles automatically.

## R01 Migration

- [ ] Import a real **Google Takeout** copy: dates, places, descriptions, albums, and favorites are correct; count the items against the Takeout.
- [ ] Import an **iCloud export**; Live Photos stay paired.
- [ ] Import from another cloud (if configured) resumes after a network drop.
- [ ] Every import shows a dry-run report first and can be undone within the trash window.
- [ ] **Round trip:** export all your data, import it into a fresh install, and compare; nothing lost.

## R02 Everyday essentials

- [ ] The map and memories work with the internet disconnected.
- [ ] Zip extraction refuses "zip bombs" and archives that try to write outside the target folder.
- [ ] Batch rename shows a preview and can be undone.
- [ ] Smart albums update when new matching photos arrive.

## R03 Family sharing

- [ ] Repeat all multi-user leak tests with groups, partner sharing, collaborative albums, comments, and activity feeds.
- [ ] Removing someone from a group removes their access immediately, everywhere.
- [ ] A contributor's photos in a shared album stay theirs; the album owner can remove them from the album but not delete them.

## R04 Data safety

- [ ] **Ransomware simulation:** on a test folder over the network drive, run a script that rewrites many files with random data. The NAS detects it, pauses that client, alerts you, and **rewind** restores everything.
- [ ] **Off-site backup:** restore everything from the off-site backup onto a fresh install.
- [ ] **UPS:** pull the wall plug; the NAS alerts and shuts down cleanly before the battery runs out.
- [ ] An overdue backup triggers an alert.

## R05 Security

- [ ] 2FA and passkeys work, and recovery codes work when the phone is "lost".
- [ ] Single sign-on (if used) cannot bypass 2FA or permissions.
- [ ] **Locked folder** items never appear in the timeline, search, memories, shares, network drive, or notifications.
- [ ] **Disk encryption:** unlock at boot works; the **recovery key** works on a test system (test this before trusting it!).
- [ ] Virus scanning: upload the standard **EICAR test file**; it is quarantined, not deleted, and you are alerted.
- [ ] Login from a new device sends an alert.

## R06 Private remote access

- [ ] From a phone on **mobile data**, reach the NAS through the VPN and through the relay (if you are behind CGNAT).
- [ ] An **external port scan** (from outside your home) shows nothing open to the public.
- [ ] Certificates renew automatically (test with a short-lived test certificate).
- [ ] Revoking a device's VPN access works immediately.

## R07 Mobile apps (Android, iOS)

Test on real phones, not just emulators, and include brands that are aggressive about stopping background apps.

- [ ] **Device matrix:** at least Samsung, Xiaomi, Infinix or Tecno, Oppo or Vivo, a Pixel, and an iPhone, across several Android versions.
- [ ] Pairing by QR code, LAN discovery, and manual address.
- [ ] **Auto backup** of 10,000 photos: no duplicates, nothing missing (compare counts).
- [ ] Backup survives: killing the app, rebooting the phone, airplane mode, switching between Wi-Fi and mobile data, and low battery.
- [ ] **Battery:** overnight drain with backup on is acceptable.
- [ ] **Share to NAS:** share one photo, 100 photos, and a 5 GB video from other apps; all upload in the background, even after closing the app; non-media sent to Photos is redirected to Files.
- [ ] **Folder sync:** Backup mode never deletes on the NAS; Mirror mode sends deletions to NAS trash; a sudden mass deletion on the phone pauses and asks.
- [ ] **Free up space** deletes only photos verified on the NAS.
- [ ] Signing out the device from the web stops it immediately.
- [ ] Deleting a photo on the phone never deletes it on the NAS.

## R08 Desktop apps (Windows, macOS, Linux)

Use test folders, never real ones, until the end of the pilot.

- [ ] **All four sync modes** (Backup, Mirror, Download, Two-way) on a test folder.
- [ ] **Conflict:** edit the same file on the PC and in the web UI; a conflict copy appears, nothing is silently overwritten.
- [ ] **Stress:** rename and move hundreds of files and folders quickly; both sides end up identical.
- [ ] Offline edits sync correctly when the PC reconnects.
- [ ] Sleep and hibernation: sync catches up afterwards.
- [ ] A **OneDrive online-only** folder is not downloaded by accident.
- [ ] Locked files (open in another program) are retried, not skipped silently.
- [ ] Very long paths and Unicode names work.
- [ ] **Share to NAS:** right-click "Upload to local-ai-nas", "Send to", and drag-and-drop onto the tray icon.
- [ ] Photo folder backup sends only media to the photos area.
- [ ] Uninstalling the app leaves NAS data untouched.
- [ ] The command-line tool works with a device password and cannot exceed its permissions.

## R09 Public release — highest risk

This is the release where the NAS faces the whole internet. The plan makes every item below **release-blocking**: if one fails, the release waits.

**Security**
- [ ] An **independent penetration test** by someone other than the developer, with no open critical or high findings.
- [ ] The **go-public wizard refuses** to enable exposure on an unprepared system (no admin 2FA, no valid certificate, no recent backup, outdated version).
- [ ] The **admin interface is unreachable** from the internet (test from mobile data).
- [ ] Brute-force protection and rate limits hold against a scripted login attack.
- [ ] **Public links:** expiry works, passwords work, download limits work, revoking is immediate, tokens cannot be guessed.
- [ ] **File requests:** uploaders cannot see other uploads; the EICAR file and oversized files are refused or quarantined.
- [ ] Automated fuzzing of every public endpoint passes.

**UI/UX**
- [ ] **Usability tests with at least five non-technical people:** opening a shared album on a phone over a slow connection, uploading through a file request, and an owner creating and revoking a link. Every blocking issue fixed.
- [ ] Public pages load quickly on a mid-range phone over a throttled mobile connection.
- [ ] An accessibility audit (WCAG 2.2 AA) of public pages passes.
- [ ] Error messages (expired link, wrong password, file too large) are clear to someone who has never heard of the NAS.

**Rollout**
- [ ] A **private beta** with invited outside users for about four weeks, with no data-loss or security incident.
- [ ] Then a staged rollout, not everyone at once.
- [ ] **"Go private again"** tested: one click removes all public exposure.
- [ ] Incident runbook written and rehearsed: revoke all links, force logout, rotate secrets, restore from backup.

## R10–R12 Media center, documents, automation

- [ ] **DLNA** serves only folders the admin marked for the household.
- [ ] **Casting** works and uses short-lived, authorised links.
- [ ] **Photo editing** never changes the original (its checksum stays the same); revert works.
- [ ] **Office previews** of deliberately malicious documents cannot harm the NAS.
- [ ] **Co-editing** respects file locks; versions are kept.
- [ ] **Workflows** show a dry run; **webhooks** are signed; **plugins** cannot bypass permissions.

## AI (always the last release)

- [ ] Classification accuracy meets the target on the labelled evaluation set (e.g. unlabelled receipts are found by searching `receipts`).
- [ ] Face groups are accurate; naming, merging, and splitting work; corrections survive reprocessing.
- [ ] Processing speed on your real hardware (Raspberry Pi or mini-PC) is acceptable; the NAS stays responsive while AI runs.
- [ ] The hardware is detected correctly (CPU, NVIDIA, Intel, AMD), with automatic fallback to CPU.
- [ ] **AI-off test:** stop the AI worker; every other feature, including normal search, still works.
- [ ] If semantic search is enabled: stopping the AI worker makes search fall back to normal results without errors.
- [ ] **Opting out** deletes all face and AI data.
- [ ] No AI data about one user is visible to another.

---

# Part 3: Going live

## Before switching (every release)
- [ ] Backups are running **and** a restore was tested for this release.
- [ ] Alerts reach your email or phone (send a test alert).
- [ ] Admin account has two-factor authentication (from R05 on; strongly recommended earlier).
- [ ] You have written down: where the admin password and recovery keys are, which drive is which, and how to restore.
- [ ] The installed version is exactly the version you tested.

## The switch (first release)
1. Do a final sync of new files and photos from the old services.
2. Compare counts of files and photos on the NAS against the source.
3. Point phone backup at the NAS.
4. Show family members how to use it.

## The watch period
- **Keep old services for at least one to three months** before deleting anything from them. This is your way back.
- Check the dashboard daily for the first week, then weekly.
- **Restore test** from backup one month after going live, and again after each release.
- Only then cancel subscriptions or delete old copies.

## Go / No-Go checklist

Go live only when every answer is **yes**.

| Question | Yes / No |
|---|---|
| Did every automated test (including crash-injection and security tests) pass on the release build? | |
| Did a clean install, or an upgrade from the previous release, work from the guide alone? | |
| Did every feature of this release pass the manual walkthrough? | |
| Did every failure drill lose zero data? | |
| Did a full restore from backup work? | |
| Were all security findings fixed? | |
| Was it fast and stable at full size for two weeks? | |
| Did the pilot pass with no data problems? | |
| Are backups and alerts running right now? | |
| Is there a tested way back (old services, or rollback to the previous release)? | |
| **Public release only:** did the independent security test, usability tests, and beta all pass? | |

## If something goes wrong after going live

1. **Stop and do not panic-fix.** Most data loss happens during rushed repairs.
2. Check the dashboard, alerts, and logs to understand what happened.
3. If data looks wrong, **stop writing to the NAS** (maintenance mode; for public installs, "go private" first).
4. Restore from trash, versions, rewind, or backup, following the recovery guide.
5. Fall back to the old service or roll back to the previous release if needed.
6. Write down what happened and add it to the test list, so the next release is tested for it.
