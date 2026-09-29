# R001: Cloud storage and photo services — feature research

| Field | Value |
|---|---|
| Research ID | R001 |
| Date | 2026-09-28 (a dated snapshot; products change often) |
| Session | S007 (`logs/sessions/2026-09-28_S007.md`, E034–E036, E040) |
| Source | Plan change request #6: `prompts/P006-competitor-research-release-roadmap.json` (the research was done for the user and delivered in that prompt) |
| Plan version | Written against 1.4.2; compared with **1.5.0** here; plan **1.6.0** holds the requirements (FR-217–FR-327, NFR-040–NFR-043) and the roadmap (section 11b), and links to this file |
| Status | Research record. The destinations of the gaps are the planner's proposals (**[Planner addition]**) until the user confirms them (plan section 5, Q52–Q66) |

## Method

Web research on 2026-09-28 across commercial cloud storage, photo services, and self-hosted alternatives (the project's direct competitors). Features were taken from official product pages, help centers, release notes, and independent reviews. Enterprise-only features were noted but weighed low, because local-ai-nas targets households and small groups.

**Caveat:** Products change often. Before a release that depends on a competitor's behavior is planned in detail, re-check that behavior. Never copy a competitor's branding, UI, or text; features are ideas, not designs.

**Adaptation to plan 1.5.0:** the prompt was written against plan 1.4.2. Since then the user's walkthrough of S02 added a 30-day trash retention (FR-008), folder sizes (FR-214), the added date with sorting (FR-215), and finished uploads shown at once (FR-216). They are listed under "Already covered", and gaps G-025 and G-026 note what they already cover.

## Services

### Google Drive

*Commercial cloud storage*

- Share with people or by link with viewer, commenter, and editor roles
- Transferable ownership
- Shared drives owned by a team
- Version history (up to 100 versions / 30 days, named versions)
- Trash with retention
- Desktop client with stream (on-demand) or mirror modes
- Backup of computer folders
- Offline files on desktop and mobile
- Mobile document scanning with OCR
- Search by file type, owner, date, and content; OCR of images and PDFs
- Quick Access / recent files, starred files, descriptions
- Activity log
- Previews of 30+ file types including Office and archives
- Real-time collaborative editing (Docs, Sheets, Slides) and comments
- Block users and spam folder for shares
- 2FA; client-side encryption only on enterprise tiers
- AI summaries and natural-language search (Gemini)

Sources: <https://www.cloudwards.net/review/google-drive/>, <https://en.wikipedia.org/wiki/Google_Drive>

### Google Photos

*Commercial photo service*

- Automatic phone backup with Original, Storage saver, and Express quality
- Free up space on the phone
- People and pets grouping, auto-updating albums
- Shared albums with contributors, comments, and likes
- Partner sharing (share all photos, or from a date or of chosen people)
- Locked Folder (PIN/biometric protected, hidden)
- Memories carousel and recaps
- Map / places
- Editing: crop, adjust, filters, portrait blur, AI tools (Magic Eraser, Unblur, Moods, Remix)
- Redact tool to blur sensitive info in screenshots (Sept 2026)
- Collages, animations, movies, cinematic photos
- Archive
- TV casting and slideshows
- Print store and photo books
- Ask Photos natural-language AI search

Sources: <https://www.androidheadlines.com/google-photos>, <https://www.androidpolice.com/google-photos-explainer/>, <https://9to5google.com/2026/09/24/google-photos-redact-moods/>

### Microsoft OneDrive

*Commercial cloud storage*

- Files On-Demand
- PC folder backup (Desktop, Documents, Pictures)
- Personal Vault (identity-verified, auto-locking folder)
- Ransomware detection and 'Restore your OneDrive' to a point in the last 30 days
- Version history
- Sharing links with expiry and passwords
- Mobile scanning
- Photo auto-backup, albums, Memories
- New Photos app with people, AI slideshows, and photo stacks that group duplicates or blurry shots with cleanup suggestions
- Office co-authoring
- Copilot summaries, file comparison, natural-language photo agent

Sources: <https://www.microsoft.com/en-us/microsoft-365/onedrive/online-cloud-storage>, <https://www.thurrott.com/a-i/328079/onedrive-is-getting-new-copilot-features-photos-agent-more>

### Dropbox

*Commercial cloud storage*

- Block-level sync, selective sync, online-only (smart sync) files, bandwidth limits
- Share links with password, expiry, and download limits
- File requests (upload links)
- Dropbox Transfer: send large files with expiry, password, and custom branding; recipients need no account
- Rewind: restore a folder or the whole account to a point in time
- Version history and deleted-file recovery
- Document scanning, PDF editing, e-signature
- Replay: frame-accurate review comments on video, image, and audio, with versions and approvals
- Office and Google Docs integration
- Camera uploads
- WebDAV
- 2FA

Sources: <https://www.cloudwards.net/review/dropbox/>, <https://help.dropbox.com/delete-restore/rewind>, <https://help.dropbox.com/share/dropbox-transfer>, <https://help.dropbox.com/installs/dropbox-replay>

### Proton Drive

*Privacy-focused cloud storage*

- End-to-end encryption
- Photo backup and gallery
- Secure share links
- Large file transfers
- Collaborative Docs and Sheets
- Desktop apps (Windows, macOS), Linux client on the roadmap, CLI (launched June 2026)
- Document scanning on Android and iOS
- Folder upload from Android
- Shared SDK for consistent performance across platforms

Sources: <https://proton.me/drive/roadmap>, <https://proton.me/blog/drive-2026-q1-recap>

### iCloud Photos / iCloud Drive

*Commercial photo and storage service*

- iCloud Shared Photo Library for up to 6 people (add manually, by date, by people, or automatically from the camera)
- All participants can add, edit, caption, favorite, and delete; deleted items go to Recently Deleted and the contributor is notified
- Shared Albums (reduced resolution)
- Memories and Featured Photos
- Duplicate detection when merging libraries
- Hidden album, Recently Deleted

Sources: <https://www.macrumors.com/guide/icloud-shared-photo-library/>

### pCloud

*Commercial cloud storage*

- pCloud Drive virtual drive
- Crypto folder (client-side encryption add-on)
- Rewind (account-wide snapshot restore)
- Versioning (30 days, 1 year add-on)
- Share links with password, expiry, download limits, and branding
- Upload links
- Audio player with playlists and podcast speed
- Video player with quality selection
- Automatic photo upload
- Public folder hosting
- pCloud Transfer
- Kodi integration
- WebDAV

Sources: <https://www.cloudwards.net/review/pcloud/>

### MEGA

*Encrypted cloud storage*

- End-to-end encryption
- Desktop sync and backup for Windows, macOS, Linux
- CLI
- Camera uploads
- File requests
- Share links with password and expiry, QR code sharing
- Versioning and rubbish bin
- Encrypted chat and calls
- Password manager (MEGA Pass) and VPN bundles
- S4 object storage (S3-compatible)
- WebDAV
- Recovery key

Sources: <https://www.cloudwards.net/review/mega/>

### Sync.com

*Encrypted cloud storage*

- Zero-knowledge encryption
- Vault (cloud-only storage outside sync)
- Version history (180–365 days)
- Account rewind
- Share links with access management and tracking
- Team shared folders with permissions
- Event/activity log
- 2FA
- Custom branding

Sources: <https://cyberinsider.com/cloud-storage/reviews/sync-com/>

### Icedrive

*Cloud storage*

- Virtual drive for Windows, macOS, Linux; portable app
- Color-coded folders, favorites
- Share links with password, expiry, and download limits
- File requests
- Client-side encrypted folder
- 2FA including FIDO/U2F keys
- Version history and trash
- WebDAV, NAS sync

Sources: <https://www.cloudwards.net/review/icedrive/>

### Filen

*Encrypted cloud storage*

- End-to-end encryption for all accounts
- Five sync modes (two-way, one-way, backup) with .fileignore
- Unlimited versioning
- Built-in text editor
- Offline files
- Share links with password, expiry, download limits; upload links
- Network drive, WebDAV, S3
- 2FA, master key export
- Account data export and deletion
- REST API

Sources: <https://www.cloudwards.net/review/filen/>

### Koofr

*Cloud storage*

- Connect other clouds (Dropbox, Google Drive, OneDrive) and search across them
- Duplicate file finder
- Space usage analysis
- Advanced (batch) renaming
- Browser image editor (resize, crop, effects)
- Vault (client-side encrypted)
- Import from Facebook/Instagram
- WebDAV and rclone support
- 2FA with TOTP and passkeys

Sources: <https://koofr.eu/features/>

### Tresorit

*Encrypted business storage*

- End-to-end encryption
- File requests
- Link restrictions (password, expiry, download limits, email verification)
- Link access logs
- eSign
- Remote device wipe
- User groups and roles
- Version history
- Email integration for encrypted attachments
- Document scanner

Sources: <https://www.cloudwards.net/tresorit-review/>

### Box

*Business cloud storage*

- Customizable watermarks on files and folders (user email, IP, or custom text)
- Metadata and retention (enterprise)

Sources: <https://support.box.com/hc/en-us/articles/42817901880851-Box-is-launching-customizable-watermarking-for-files-and-folders>

### Amazon Photos

*Commercial photo service*

- Family Vault shared with up to 6 members
- Auto-save from phones and computers
- Recognition of people, things, and places with filtering and search
- Albums
- Viewing on TVs, smart displays, and tablets (Fire TV, Echo Show)

Sources: <https://www.pocket-lint.com/what-is-amazon-photos-and-how-do-you-use-it/>

### Nextcloud (Hub 26)

*Self-hosted platform (direct competitor)*

- Sync clients with selective sync and virtual files
- File locking
- Version history
- External storage (S3, SMB, SFTP, other clouds)
- Share links with password, expiry, download restrictions, one-time-password verification
- Server-side and per-folder end-to-end encryption
- Photos: adjustable grid, map, trip memories, slideshow, year recap, EXIF editor, video transcoding, Google Photos import
- Activity heatmap and filters
- Office (Collabora, Euro-Office), Text, Notes, Talk, Calendar, Contacts, Mail
- Flow workflow automation
- LDAP, SAML/OIDC SSO, TOTP, WebAuthn passwordless
- Brute-force protection, rate limiting, suspicious-login detection, breached-password checks
- Ransomware protection app, antivirus app
- Remote wipe

Sources: <https://nextcloud.com/blog/nextcloud-hub26-summer/>, <https://bestcloudstorageguide.com/blog/nextcloud-features-guide-2026>, <https://nextcloud.com/secure/>

### Seafile

*Self-hosted file sync*

- Encrypted libraries (client-side)
- Virtual drive (SeaDrive)
- Custom metadata properties and views (table, gallery, Kanban, map)
- Hierarchical tags
- Granular permissions (upload, download, preview-only)
- Share links with password, expiry, and email authentication
- SeaDoc co-authoring, OnlyOffice/Office Online integration
- Wiki
- Server-side virus scanning
- Remote wipe
- Audit logging

Sources: <https://www.seafile.com/en/features/>

### Synology (DSM, Photos, Drive, Snapshots, Hyper Backup, Active Backup)

*NAS vendor (direct competitor)*

- Photos: timeline or folder view, similar-photo stacking, conditional (rule-based) albums, people and object recognition, shared space, photo request (public upload) links, public folder links with password and expiry, Live Photos, 360° video, slideshow, map, mobile backup, widgets
- Btrfs snapshots with retention and immutable (WORM) snapshots against ransomware; users restore via a snapshot folder or Windows Previous Versions
- Snapshot replication to a second NAS
- Hyper Backup: backup to USB, remote NAS, file servers, or cloud with scheduling, compression, integrity checks, retention
- Active Backup: back up PCs, servers, and VMs to the NAS
- Drive: sync client, team folders, NAS-to-NAS sync

Sources: <https://www.synology.com/en-us/dsm/7.3/software_spec/synology_photos>, <https://ifeeltech.com/blog/synology-snapshots-explained>, <https://www.vinchin.com/nas-tips/synology-drive-vs-active-backup-for-business-vs-hyper-backup.html>

### Immich (v3)

*Self-hosted photo platform (direct competitor)*

- Mobile apps with background backup, automatic LAN/remote URL switching, free up space
- Timeline, folders view, archive, trash, stacking, tags, star ratings, recently added view
- Public links (custom slugs), shared albums (read-write or read-only), partner sharing
- CLIP smart search, map, explore, memories
- Faces with merge, duplicate and similar detection, OCR
- Live Photos, motion photos, RAW, 360°, HDR video
- External libraries with watching, custom storage structure, XMP sidecars
- Non-destructive editing (crop, rotate, mirror)
- Workflows automation (triggers, filters, actions)
- HLS real-time transcoding
- Integrity report (untracked, missing, checksum mismatch)
- Database backup and restore from the web UI
- Fine-grained API permissions
- Planned: user groups, automatic stacking, smart memories, smart albums, iCloud import

Sources: <https://immich.app/roadmap>, <https://linuxiac.com/immich-3-0-released-with-big-upgrades-for-self-hosted-photo-libraries/>, <https://chrislongros.com/2026/01/28/immich-v2-5-0-the-90000-stars-release-major-features-for-self-hosted-photo-management/>

### PhotoPrism

*Self-hosted photo platform*

- Search filters (color, quality, resolution, label, place)
- Places map, moments, calendar view
- Faces, labels, NSFW detection
- Batch metadata editing
- Live and motion photos, 360° panoramas, RAW, JPEG XL
- Private flag, archive, review queue
- YAML sidecars
- WebDAV sync, ZIP download, secret share links with expiry
- PWA install

Sources: <https://www.photoprism.app/features/>

### Ente Photos

*Encrypted photo service (open source)*

- End-to-end encryption with on-device ML (faces, semantic search)
- Shared albums and public links
- Collect: anyone with a link can add photos from a browser without an account
- Continuous incremental local export with metadata and album structure
- CLI
- Apps for iOS, Android, web, Windows, macOS, Linux

Sources: <https://ente.com/compare/ente-vs-google-photos/>

### Remote access (research note)

Many home connections are behind CGNAT, so port forwarding is often impossible. The self-hosted pattern that needs no third party is a WireGuard tunnel from the home server to a small public VPS the user controls, which relays traffic. Mesh VPNs (Tailscale, or self-hosted Headscale) give private access without exposing anything publicly. Third-party tunnels that terminate TLS let the provider see traffic, which conflicts with the project's privacy stance unless the user opts in knowingly.

Sources: <https://community.hetzner.com/tutorials/bypass-cgnat-with-a-wireguard-vps-relay/>

## Already covered by the plan

The research was compared with the plan, not only listed. These competitor features are already planned:

- Resumable uploads, range downloads, ZIP download, file operations (FR-004–FR-007)
- Previews of images, text, PDF, audio, video (FR-082)
- Per-user trash (FR-008); file versioning (FR-122, Could, pending Q34)
- WebDAV network drive; SMB proposed (FR-009, FR-124)
- Timeline, albums, favorites, hide/archive, lightbox viewer (FR-013–FR-016, FR-096)
- Video streaming with quality switching (FR-144–FR-148)
- HEIC (Should), RAW (Could) (FR-018, FR-020)
- Metadata editing: description, tags, date, location (FR-015, FR-099); offline place names (FR-062)
- Forgiving search with synonyms, typos, operators (FR-047–FR-063, FR-104–FR-110)
- Exact and near duplicates, look-alike stacks, bursts, file duplicates and shortcuts (FR-150–FR-175)
- Storage saver: resolution and quality reduction, upload policies (FR-176–FR-194)
- Multi-user, private namespaces, read sharing, shared with me (FR-112–FR-118)
- Admin dashboard, quotas, disk health, jobs, logs (FR-127–FR-130)
- Integrity scans, metadata backup, disaster recovery, local external backup (FR-119–FR-123)
- RAID 0 and RAID 1 pools (FR-195–FR-210)
- Face grouping, auto-classification, AI cleanup; OCR and semantic search as options (S16)
- Docker, per-platform setup scripts, safe updates (FR-131–FR-135, FR-149)
- HTTPS, sessions, API tokens, audit log, TOTP (Could, Q33) (FR-064–FR-091)
- **Added in plan 1.5.0** (the user's walkthrough, S007 E030–E032): trash items deleted automatically after 30 days (FR-008, S08.1); every item shows its size, folders too (FR-214, S02.3-T05); the date an item was added, shown with the modified date and sortable (FR-215, S02.3-T05); finished uploads appear at once, and a single uploaded item is shown with two blinks (FR-216, S02.4-T05).

## Gap list (coverage matrix)

Every feature below is missing from the plan or only partly covered. Each has exactly one destination: the MVP, a release (R01–R12, plan section 11b), the AI stage S16 (as an optional extension), or excluded. The FR IDs are in plan section 3.1 (MVP) and 3.3 (releases).

| Gap | Feature | In the plan before P006 | Seen in | Destination | FR |
|---|---|---|---|---|---|
| G-001 | Live Photos and Motion Photos | Missing | Immich, Synology, PhotoPrism, iCloud | MVP | FR-217, FR-218 |
| G-002 | Phone photo auto-backup (bridge until native apps) | Missing (native app is candidate in 11a) | Google Photos, iCloud, OneDrive, Dropbox, Amazon Photos, Immich, Synology, MEGA, pCloud | MVP | FR-219, FR-220 |
| G-003 | Alert delivery (email, webhook, ntfy) | Partial (alerts only in GUI and logs, FR-129) | Synology, Nextcloud, Dropbox (Rewind emails) | MVP | FR-221 |
| G-004 | File versioning | Partial (FR-122 Could, pending Q34) | Google Drive, OneDrive, Dropbox, pCloud, Sync.com, Filen, Icedrive, Nextcloud, Seafile | MVP (recommendation: answer Q34 'yes'; the user decides) | FR-122 (existing, Q34) |
| G-010 | Google Takeout import | Missing (Q11 mentions Takeout metadata) | Nextcloud (Google Photos import), Immich (tools) | R01 | FR-222 |
| G-011 | Apple Photos / iCloud export import | Missing | Immich (planned iCloud import) | R01 | FR-223 |
| G-012 | Import from other clouds (rclone remotes) | Missing | Koofr, Nextcloud external storage | R01 | FR-224 |
| G-013 | USB drive and camera card import | Missing | Synology (USB Copy) | R01 | FR-225 |
| G-014 | Export all my data and account deletion | Missing | Ente, Filen, MEGA, Google (Takeout) | R01 | FR-226 |
| G-015 | XMP sidecar export for other photo tools | Missing (Q19 asks) | Immich, PhotoPrism (YAML) | R01 | FR-227 |
| G-016 | External read-only libraries (index a host folder in place) | Superseded by the two-area design (A3) | Immich, Nextcloud | EXCLUDED unless the user decides otherwise (conflicts with I1 and A3; Q58) | none (excluded) |
| G-020 | Files area: stars, color labels, tags, descriptions | Missing (photos only) | Google Drive, Icedrive, Seafile, Filen | R02 | FR-228 |
| G-021 | Recent files and quick access | Missing | Google Drive | R02 | FR-229 |
| G-022 | Saved searches and recent searches | Missing | Google Drive, Seafile (custom views) | R02 | FR-230 |
| G-023 | Server-side archive create and extract | Missing | Synology File Station, Google Drive (archive preview) | R02 | FR-231 |
| G-024 | Batch rename with patterns | Missing | Koofr | R02 | FR-232 |
| G-025 | Personal storage analyzer and folder sizes | Missing (admin dashboard only). Folder sizes are already FR-214 (S02, plan 1.5.0); the release keeps the analyzer. | Koofr, Google Photos (storage management) | R02 | FR-233 |
| G-026 | Recently added view | Missing. The added date and its sort exist for the files area (FR-215, plan 1.5.0); the release keeps the photos view. | Immich | R02 | FR-234 |
| G-027 | Slideshow | Missing | Google Photos, Immich, Synology, Nextcloud, OneDrive | R02 | FR-235 |
| G-028 | Star ratings | Missing | Immich, Synology | R02 | FR-236 |
| G-029 | Smart (rule-based) albums | Missing | Synology (conditional albums), Immich (planned), Google Photos (auto-updating albums) | R02 | FR-237 |
| G-030 | Memories without AI (on this day, year recap, trips) | Missing | Google Photos, iCloud, OneDrive, Immich, Nextcloud | R02 | FR-238 |
| G-031 | Map view of photos | Missing (GPS and places exist; no map) | Google Photos, Immich, Synology, PhotoPrism, Nextcloud | R02 | FR-239 |
| G-032 | Folder view of photos | Missing (depends on Q40) | Synology, Immich, PhotoPrism | R02 | FR-240 |
| G-033 | Grid density, year/month scrubber, calendar view | Missing | Nextcloud, PhotoPrism, Synology | R02 | FR-241 |
| G-034 | Batch metadata editing (shift dates, set location) | Partial (single-item edits) | PhotoPrism, Synology | R02 | FR-242 |
| G-035 | Download as compatible format (HEIC to JPEG) | Missing | iCloud (export), Google Photos | R02 | FR-243 |
| G-040 | Edit (write) sharing and reshare control | Partial (FR-117 Could, pending Q30) | Google Drive, Dropbox, Nextcloud, Seafile, Immich (read-write albums) | R03 | FR-244 |
| G-041 | User groups and sharing with groups | Partial (FR-117 Could) | Immich (planned), Tresorit, Nextcloud | R03 | FR-245 |
| G-042 | Shared family/team folders owned by a group | Missing | Google shared drives, Nextcloud Teams, Synology shared space, Sync.com team folders | R03 | FR-246 |
| G-043 | Partner sharing / shared library | Missing | Google Photos, iCloud Shared Library, Immich, Amazon Family Vault | R03 | FR-247 |
| G-044 | Collaborative albums with contributors | Missing | Google Photos, Immich, Synology | R03 | FR-248 |
| G-045 | Comments and reactions on shared items | Missing | Google Photos, Google Drive, Nextcloud, Dropbox | R03 | FR-249 |
| G-046 | Activity feed and item history | Partial (admin audit log only) | Google Drive, Nextcloud, Seafile, Sync.com | R03 | FR-250 |
| G-047 | Notification center and per-user email notifications | Partial (transient GUI notifications) | Nextcloud, Dropbox, Google | R03 | FR-251 |
| G-048 | Ownership transfer | Missing | Google Drive | R03 | FR-252 |
| G-049 | Decline shares and block a user from sharing with you | Missing | Google Drive (block users, spam) | R03 | FR-253 |
| G-050 | Point-in-time restore (rewind) of a folder or account | Missing | Dropbox, OneDrive, pCloud, Sync.com | R04 | FR-254 |
| G-051 | Ransomware / mass-change detection | Missing | OneDrive, Nextcloud (apps), Synology (immutable snapshots) | R04 | FR-255 |
| G-052 | Filesystem snapshots with immutable retention | Missing (pools use ext4/XFS) | Synology | R04 (Could; needs a filesystem ADR) | FR-256 |
| G-053 | Off-site encrypted backup to user-owned targets | Missing (local targets only, S08.6) | Synology Hyper Backup, Nextcloud | R04 | FR-257 |
| G-054 | NAS-to-NAS replication | Missing | Synology (Snapshot Replication, ShareSync) | R04 | FR-258 |
| G-055 | UPS integration and graceful shutdown | Missing | Synology, TrueNAS (general NAS practice) | R04 | FR-259 |
| G-056 | Backup health dashboard and scheduled restore tests | Partial (restore tested in S08) | Synology Hyper Backup (integrity checks) | R04 | FR-260 |
| G-057 | Encrypted local backup targets | Missing | Synology Hyper Backup | R04 | FR-261 |
| G-060 | TOTP 2FA with recovery codes, enforceable | Partial (FR-091 Could, pending Q33) | all services | R05 (or MVP if Q33 is answered 'yes') | FR-262 |
| G-061 | Passkeys and security keys (WebAuthn/FIDO2) | Missing | Nextcloud, Koofr, Icedrive | R05 | FR-263 |
| G-062 | SSO with OIDC; LDAP (Could) | Missing | Nextcloud, Seafile, Filen (enterprise) | R05 | FR-264 |
| G-063 | Locked folder / private vault | Partial (hide/archive only) | Google Photos, OneDrive, Koofr, pCloud Crypto, Icedrive, Sync.com Vault | R05 | FR-265 |
| G-064 | Encryption at rest of the storage (full disk) | Missing | Nextcloud (server-side encryption), Synology (encrypted volumes) | R05 | FR-266 |
| G-065 | Malware scanning of uploads | Missing | Seafile, Nextcloud | R05 | FR-267 |
| G-066 | Password policy with strength and breached-password checks (offline) | Missing | Nextcloud | R05 | FR-268 |
| G-067 | New-login alerts, suspicious-login heuristics, per-user security page | Partial (sessions list) | Nextcloud, Google | R05 | FR-269 |
| G-068 | Admin security checklist | Missing | Nextcloud (security scan) | R05 | FR-270 |
| G-069 | Per-IP brute-force protection with trusted-proxy awareness | Partial (per-account lockout, FR-085) | Nextcloud | R05 | FR-271 |
| G-070 | Built-in WireGuard VPN for private remote access | Missing (11a candidate) | Synology (VPN Server), general NAS practice | R06 | FR-272 |
| G-071 | Mesh VPN guidance (Headscale self-hosted; Tailscale opt-in) | Missing | remote access research | R06 | FR-273 |
| G-072 | CGNAT relay through the user's own VPS | Missing | remote access research | R06 | FR-274 |
| G-073 | Dynamic DNS and IPv6 | Missing | Synology, general NAS practice | R06 | FR-275 |
| G-074 | Trusted TLS certificates via ACME with auto-renewal | Missing (self-signed or user-provided only, FR-088) | Synology, Nextcloud (deployment guides) | R06 | FR-276 |
| G-075 | Slow-link performance (WAN-aware thumbnails and transfers) | Partial (HLS Auto) | all mobile-first services | R06 | FR-277 |
| G-080 | Installable web app (PWA) with share target | Missing | PhotoPrism | R07 | FR-278 |
| G-081 | Native Android and iOS apps | Missing (NG2; 11a candidate) | all services | R07 | FR-279 |
| G-082 | Background automatic photo and video backup in the app | Missing | Google Photos, iCloud, OneDrive, Immich, Ente, Synology, Amazon | R07 | FR-280 |
| G-083 | Free up space on the phone | Missing | Google Photos, Immich | R07 | FR-281 |
| G-084 | Offline files and albums on mobile | Missing | Google Drive, Dropbox, Synology Photos | R07 | FR-282 |
| G-085 | Document scanner to PDF | Missing | Google Drive, Dropbox, OneDrive, Proton, Tresorit | R07 | FR-283 |
| G-086 | App lock (biometrics or PIN) | Missing | Google Drive iOS, Google Photos Locked Folder | R07 | FR-284 |
| G-087 | Widgets and OS share-sheet integration | Missing | Synology Photos, Google Photos | R07 | FR-285 |
| G-088 | LAN discovery and automatic LAN/remote address switching | Missing | Immich | R07 | FR-286 |
| G-089 | Push notifications without a vendor cloud where possible | Missing | all mobile apps | R07 | FR-287 |
| G-090 | Desktop sync client (two-way, selective, conflicts, bandwidth) | Missing | all services | R08 | FR-288 |
| G-091 | Files on demand / virtual files | Missing | OneDrive, Google Drive, Dropbox, Nextcloud, pCloud, Icedrive, Seafile | R08 | FR-289 |
| G-092 | PC folder backup | Missing | OneDrive, Google Drive, Synology Active Backup | R08 | FR-290 |
| G-093 | User CLI (upload, download, sync, share, search) | Missing (admin CLI only) | MEGA, Proton, Ente, Immich | R08 | FR-291 |
| G-094 | File locking for sync and co-editing | Partial (WebDAV locks, S09.1) | Nextcloud, Seafile | R08 | FR-292 |
| G-095 | Device list with remote sign-out and wipe of synced data | Partial (sessions) | Tresorit, Seafile, Nextcloud | R08 | FR-293 |
| G-100 | Public share links for people without an account | Missing (NG5; 11a) | all services | R09 | FR-294 |
| G-101 | File requests / upload-only links | Missing | Dropbox, pCloud, MEGA, Icedrive, Filen, Tresorit, Synology, Ente (Collect), Seafile | R09 | FR-295 |
| G-102 | Large file transfer (send files that expire) | Missing | Dropbox Transfer, pCloud Transfer, Proton, MEGA | R09 | FR-296 |
| G-103 | Guest verification by email code | Missing | Nextcloud (OTP), Seafile, Tresorit | R09 | FR-297 |
| G-104 | Link access logs and download notifications | Missing | Tresorit, Sync.com | R09 | FR-298 |
| G-105 | Watermarks on shared previews | Missing | Box | R09 (Could) | FR-299 |
| G-106 | Branded share pages | Missing | pCloud, Dropbox Transfer, Sync.com | R09 (Could) | FR-300 |
| G-107 | Invite-based registration, per-user bandwidth limits, abuse tools | Missing | MEGA (transfer quota), Nextcloud | R09 | FR-301 |
| G-108 | Internet-facing hardening | Missing (LAN-only posture) | Nextcloud, all public services | R09 | FR-302 |
| G-109 | Go-public readiness gate | Missing | Nextcloud (security scan) | R09 | FR-303 |
| G-110 | External security audit and staged beta before general availability | Missing | Proton, Ente (independent audits) | R09 | FR-304 |
| G-111 | UX overhaul for public use (recipients, phones, slow networks, languages) | Partial (accessibility NFR-015) | all public services | R09 | FR-305 |
| G-112 | Reverse proxy and tunnel support | Missing | Immich, Nextcloud, PhotoPrism (docs) | R09 | FR-306 |
| G-113 | Operator tools and guidance for hosting other people | Missing | Nextcloud | R09 | FR-307 |
| G-120 | Music library (tags, albums, playlists, speed) | Partial (audio preview) | pCloud | R10 | FR-308 |
| G-121 | Video library improvements (subtitles, resume, speed, audio tracks) | Partial (streaming) | pCloud, Immich (speed controls) | R10 | FR-309 |
| G-122 | DLNA/UPnP media server | Missing | Synology, general NAS practice | R10 | FR-310 |
| G-123 | Casting to TVs (Chromecast, AirPlay) | Missing | Google Photos, Amazon Photos, pCloud (Kodi) | R10 | FR-311 |
| G-124 | TV-friendly (10-foot) view | Missing | Amazon Photos, Google Photos | R10 | FR-312 |
| G-125 | Non-destructive photo editing | Missing (NG3) | all photo services, Immich, Koofr | R10 (needs the NG3 change, Q57) | FR-313 |
| G-126 | 360° panorama viewer | Missing | Immich, PhotoPrism, Synology | R10 | FR-314 |
| G-127 | Collages, burst animations, photo videos | Missing | Google Photos | R10 (Could) | FR-315 |
| G-128 | Manual redaction (blur/pixelate) tool | Missing | Google Photos (Sept 2026) | R10 (part of G-125) | FR-316 |
| G-130 | Office document previews | Missing (PDF and text only) | Google Drive, Dropbox, Filen, Nextcloud | R11 | FR-317 |
| G-131 | Optional online co-editing (Collabora or ONLYOFFICE) | Missing | Google, OneDrive, Nextcloud, Seafile, Proton, Koofr | R11 | FR-318 |
| G-132 | Text and Markdown editor, notes | Missing | Filen, Nextcloud, Seafile | R11 | FR-319 |
| G-133 | PDF tools (annotate, merge, split, forms); e-sign (Could) | Missing | Dropbox, Tresorit | R11 | FR-320 |
| G-134 | Full-text search inside documents | Partial (FR-111 Could, pending Q31) | Google Drive, Nextcloud | R11 (or S06 if Q31 is answered 'yes' for the MVP) | FR-321 |
| G-135 | Text version comparison | Missing | Nextcloud Collectives | R11 | FR-322 |
| G-140 | Rules and workflows engine | Missing | Immich (Workflows), Nextcloud (Flow) | R12 | FR-323 |
| G-141 | Signed webhooks for events | Missing | Nextcloud, general APIs | R12 | FR-324 |
| G-142 | API documentation portal and client SDKs | Partial (OpenAPI + Redoc) | Filen, Immich | R12 | FR-325 |
| G-143 | S3-compatible API | Missing | MEGA S4, Filen | R12 (Could) | FR-326 |
| G-144 | Extension / plugin system | Missing | Nextcloud apps, Synology packages | R12 (Could) | FR-327 |
| G-150 | Pet recognition and grouping | Missing | Google Photos, Amazon Photos | S16 (AI extension, Could) | S16.10 (Could) |
| G-151 | Smart memories using people, pets, and events | Missing | Google Photos, Immich (planned) | S16 (AI extension, Could) | S16.10 (Could) |
| G-152 | AI classification of documents in the files area | Missing (AI is photos-only) | Google Drive (AI search), OneDrive (Copilot) | S16 (AI extension, Could) | S16.10 (Could) |
| G-153 | Sensitive-content auto-hide suggestion | Missing | PhotoPrism (NSFW) | S16 (AI extension, Could) | S16.10 (Could) |
| G-154 | Blurry photo and screenshot cleanup suggestions | Partial (S16.11 smart cover) | OneDrive photo stacks, Google Photos storage management | S16 (extend S16.11) | S16.10 (Could) |
| G-155 | Local speech-to-text subtitles and search for videos | Missing | Nextcloud Assistant (transcription) | S16 (AI extension, Could) | S16.10 (Could) |
| G-156 | Sensitive-text redaction suggestions from OCR | Missing | Google Photos (Redact) | S16 (AI extension, Could; needs OCR, Q36) | S16.10 (Could) |

**Destinations:** EXCLUDED: 1, MVP: 4, R01: 6, R02: 16, R03: 10, R04: 8, R05: 10, R06: 6, R07: 10, R08: 6, R09: 14, R10: 9, R11: 6, R12: 5, S16: 7 (total 118).

## Considered and excluded

These are not requirements. The user can bring any of them back (Q65).

| ID | Feature | Seen in | Reason |
|---|---|---|---|
| X-01 | Generative AI: Ask Photos, Magic Eraser, Moods, Remix, AI summaries and chat over files | Google Photos, Google Drive, OneDrive, Nextcloud | Non-goal NG9 (generative AI) and I4 (no AI at query time). |
| X-02 | Chat, video calls, meetings | MEGA, Nextcloud Talk | Outside the purpose of a NAS. |
| X-03 | Password manager and general-purpose VPN service | MEGA Pass, MEGA VPN | Outside the purpose of a NAS. (The private-access VPN in R06 only reaches the NAS.) |
| X-04 | Calendar, contacts, and mail (CalDAV, CardDAV, webmail) | Nextcloud | Groupware, not storage. Kept as a not-scheduled candidate in 11a. |
| X-05 | Print store and photo books | Google Photos | A commercial fulfilment service, not software. |
| X-06 | Federated sharing between separate servers | Nextcloud | Large security surface; kept as a not-scheduled candidate in 11a. |
| X-07 | Enterprise governance: legal hold, eDiscovery, sensitivity labels, data rooms | Nextcloud Enterprise, Box, Tresorit | Outside the household and small-group scope. |
| X-08 | Professional media review workflow (frame-accurate review and approvals) | Dropbox Replay | Niche; comments (G-045) cover the basics. |
| X-09 | Hosting apps, containers, and virtual machines | Synology, TrueNAS | Outside scope; the plugin system (G-144) is the extension point. |
| X-10 | Vendor-operated relay or account service (QuickConnect-style) | Synology | Conflicts with I6 and NG1. The self-hosted relay (G-072) covers the need. |
| X-11 | Phone-number (SMS) two-factor authentication | Icedrive, Tresorit | Needs a paid SMS gateway and is weaker than TOTP and passkeys. |

## Verification (P006 step 3, 2026-09-28)

Only these checks were made now; everything else is verified when its release is planned in detail.

### G-001: Live Photos and motion photos

- **Apple still (HEIC or JPEG):** Apple MakerNotes tag `0x0011` **ContentIdentifier** (string; called MediaGroupUUID as an extended attribute). Source: <https://exiftool.org/TagNames/Apple.html>. **Verified.**
- **Apple video (MOV):** QuickTime Keys tag `'content.identifier'` → **ContentIdentifier**; related keys `live-photo-info`, `live-photo.auto`, `live-photo.vitality-score`, `live-photo.vitality-scoring-version`. Source: <https://exiftool.org/TagNames/QuickTime.html>. **Verified.** The pair is matched by equal ContentIdentifier values.
- **Google motion photos (format 1.0):** XMP namespace `http://ns.google.com/photos/1.0/camera/` with `Camera:MotionPhoto` (1 = motion photo), `Camera:MotionPhotoVersion` (1), `Camera:MotionPhotoPresentationTimestampUs`; a `Container:Directory` of `Item` entries with `Item:Mime`, `Item:Semantic` (`Primary`, `MotionPhoto` last, `GainMap`), `Item:Length`, `Item:Padding`. The video is appended to the JPEG, or wrapped in an `mpvd` box in HEIC/AVIF. The older MicroVideo tags (`Camera:MicroVideo`, `MicroVideoOffset`, …) are deprecated and must be ignored by writers; old files still carry them, so readers should accept them. Source: <https://developer.android.com/media/platform/motion-photo-format>. **Verified.**
- **Samsung motion photos:** a Samsung trailer after the image, with ExifTool tags `EmbeddedVideoType`, `EmbeddedVideoFile`, `MotionPhotoAutoPlayVideo`, `SamsungMotionPhotoVersion`. Source: <https://exiftool.org/TagNames/Samsung.html>. **Verified** (tag names; the trailer layout is read by ExifTool).

### G-002: phone apps with automatic WebDAV upload

- **PhotoSync** (touchbyte; iOS and Android; proprietary): automatic transfer ("Autotransfer") of new photos to WebDAV, SMB, or SFTP targets. On iOS, Autotransfer is a **PhotoSync Premium** feature and supports only location-based and when-charging triggers. Sources: <https://www.photosync-app.com/home>, <https://www.photosync-app.com/support/basics/answers/what-are-the-differences-between-the-free-version-and-premium>. **Verified for iOS; the Android pricing of Autotransfer is Unverified.**
- **FolderSync** (Tacit Dynamics; Android; proprietary, free with ads plus a paid full version): scheduled automatic sync of device folders with WebDAV and other servers. Sources: <https://foldersync.io/>, <https://foldersync.io/docs/faq/webdav/>. **Verified** (WebDAV and automatic sync); which features need the paid version is **Unverified**.
- **Autosync** (MetaCtrl, Android) also lists WebDAV; **Unverified** beyond its store listing. A search result also reported that Android 17 broke some network apps (FolderSync over SMB); **Unverified**.
- No app is recommended yet: S09 tests at least one Android and one iOS app, and the guide names only apps checked then (MVP addition G-002).

### Licenses of tools the releases would add (preliminary)

All would run as **separate programs**, as FFmpeg and ExifTool do (docs/licensing.md), so their licenses do not extend to the project's own code. Items for the user's attention are marked ⚠.

| Tool | Release | License | Source | Note |
|---|---|---|---|---|
| rclone | R01, R04 | MIT | GitHub API (`rclone/rclone`) | Permissive |
| ClamAV | R05 | GPL-2.0 | GitHub API (`Cisco-Talos/clamav`) | Separate program; the virus definitions have their own terms (check at R05) |
| Network UPS Tools | R04 | GPL-2.0-or-later (most files; some scripts GPL-3.0-or-later or Perl's terms) | `COPYING` in `networkupstools/nut` | Separate program |
| wireguard-tools | R06 | GPL-2.0 | GitHub API (`WireGuard/wireguard-tools`) | Separate program; the kernel module is part of Linux |
| LibreOffice | R11 | MPL-2.0 (`COPYING.MPL`); GitHub's detection reports GPL-3.0 from `COPYING` | `LibreOffice/core` | ⚠ The exact terms of the parts used for conversion are checked at R11 |
| Collabora Online | R11 | MPL-2.0 | `COPYING` in `CollaboraOnline/online` | Separate container |
| ONLYOFFICE Docs | R11 | ⚠ AGPL-3.0 | GitHub API (`ONLYOFFICE/DocumentServer`) | Compatible with the project's AGPL-3.0-or-later (Q22); runs as a separate container; its trademark and logo terms are checked at R11 |
| OpenStreetMap data | R02 | ⚠ ODbL 1.0 (data after September 2012) | <https://osmfoundation.org/wiki/Licence> | Attribution on the map; share-alike applies to derived databases (such as a tile or place file the project distributes). Checked in the map ADR (Q59) |

