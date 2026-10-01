# Architecture Decision Records

Every significant technical decision (RULES R5). An ADR becomes **Accepted** only with the user's approval; an Accepted decision is never rewritten, only amended with the user's approval or superseded by a new ADR. This index lists every ADR with its status and is updated whenever an ADR is added or its status changes (audit checklist, group F). Created in session S007 (the user's decision "Rename + ADR index (Recommended)").

| ADR | Decision | Status |
|---|---|---|
| [ADR-0001](ADR-0001-backend-language-framework.md) | Core server language: Go | Accepted |
| [ADR-0002](ADR-0002-api-style.md) | API style: REST with an OpenAPI contract; router and code generation | Accepted |
| [ADR-0003](ADR-0003-storage-layout.md) | Storage layout, per-user namespaces, and internal data location | Accepted · amended |
| [ADR-0004](ADR-0004-repository-layout.md) | Repository layout | Accepted · amended |
| [ADR-0005](ADR-0005-testing-linting-ci.md) | Testing, linting, and CI toolchain | Accepted |
| [ADR-0006](ADR-0006-dev-environment-and-packaging.md) | Development environment and packaging: Docker Compose, multi-architecture | Accepted (parts deferred) |
| [ADR-0007](ADR-0007-database-sqlite.md) | Database: SQLite in WAL mode, pure-Go driver, goose migrations | Accepted · amended |
| [ADR-0008](ADR-0008-resumable-uploads-tus.md) | Resumable uploads: the tus protocol (tusd embedded, Uppy in the browser) | Accepted |
| [ADR-0009](ADR-0009-web-ui-sveltekit.md) | Web UI: SvelteKit static SPA embedded in the Go binary | Accepted |
| [ADR-0010](ADR-0010-security-building-blocks.md) | Security building blocks: Argon2id, server-side sessions, CSRF tokens, optional TOTP, TLS | Accepted |
| [ADR-0011](ADR-0011-job-queue-sqlite.md) | Background job queue on SQLite, inside the core server | Accepted · amended |
| [ADR-0012](ADR-0012-media-toolchain.md) | Media toolchain: ExifTool, libvips with libheif, FFmpeg (as subprocesses) | Accepted (parts deferred) |
| [ADR-0013](ADR-0013-reverse-geocoding-geonames.md) | Offline reverse geocoding with GeoNames | Accepted |
| [ADR-0014](ADR-0014-search-engine-bleve.md) | Search engine: Bleve, with query-time synonym expansion | Accepted · amended |
| [ADR-0015](ADR-0015-network-shares-webdav.md) | Network shares: WebDAV first, via golang.org/x/net/webdav | Accepted (parts deferred) |
| [ADR-0016](ADR-0016-file-watching.md) | File watching: fsnotify plus a periodic reconciliation scan | Accepted |
| [ADR-0017](ADR-0017-ai-worker-architecture.md) | AI worker architecture: Python + ONNX Runtime in a separate optional container; core as single writer | Accepted · amended |
| [ADR-0018](ADR-0018-ai-models.md) | AI models: CLIP-family classification, YuNet, SFace, HDBSCAN; license exclusions | Accepted (parts deferred) · amended |
| [ADR-0019](ADR-0019-smb-via-samba.md) | SMB network shares via Samba (optional, Linux-only) | Proposed (parts deferred) |
| [ADR-0020](ADR-0020-video-streaming-quality-levels.md) | Video streaming with live quality switching: HLS, on-demand transcoding with cache | Accepted |
| [ADR-0021](ADR-0021-content-hash-algorithm.md) | Content hash algorithm for uploads and duplicate detection | Accepted |
| [ADR-0022](ADR-0022-perceptual-hash-and-similarity-index.md) | Perceptual hash and similarity index | Proposed |
| [ADR-0023](ADR-0023-look-alike-stacks.md) | Look-alike stacks and bursts | Proposed |
| [ADR-0024](ADR-0024-file-shortcuts.md) | File shortcuts after duplicate resolution | Proposed |
| [ADR-0025](ADR-0025-video-optimization-codec.md) | Video codec for storage optimization | Proposed |
| [ADR-0026](ADR-0026-originals-retention-and-revert.md) | Retention and revert of originals replaced by optimization | Proposed |
| [ADR-0027](ADR-0027-drive-pool-approach.md) | Drive pool approach (RAID 0 and RAID 1) | Proposed · revised (ZFS optional) |
| [ADR-0028](ADR-0028-pool-filesystem.md) | Filesystem on a drive pool | Proposed |
| [ADR-0029](ADR-0029-privileged-storage-helper.md) | Privileged storage helper | Proposed |
| [ADR-0030](ADR-0030-storage-migration-engine.md) | Storage migration engine | Proposed |
| [ADR-0031](ADR-0031-switching-the-storage-root.md) | Switching the storage root after a migration | Proposed |
| [ADR-0032](ADR-0032-predictive-drive-health.md) | Predictive drive health rules | Proposed |
| [ADR-0033](ADR-0033-mirror-conversion-and-hot-replacement.md) | Mirror conversion, hot replacement, and pool growth with mdadm | Proposed |
| [ADR-0034](ADR-0034-hot-plug-drive-detection.md) | Detecting newly installed drives | Proposed |
| [ADR-0035](ADR-0035-secure-erase.md) | Secure erase of retired drives | Proposed |
| [ADR-0036](ADR-0036-ssd-read-cache.md) | SSD read cache of original files | Proposed |
| [ADR-0037](ADR-0037-fast-internal-data-placement.md) | Fast internal-data placement on an SSD | Proposed |
| [ADR-0038](ADR-0038-block-level-ssd-cache.md) | Block-level SSD cache (lvmcache) and an optional LVM layer | Proposed |
| [ADR-0039](ADR-0039-admin-console.md) | Admin console architecture | Accepted |
| [ADR-0040](ADR-0040-item-identity.md) | Item identity: UUIDv7 IDs in an items table | Accepted |
| [ADR-0041](ADR-0041-operation-journal-and-crash-consistency.md) | Operation journal and crash consistency | Accepted · amended |
| [ADR-0042](ADR-0042-local-origin-protection.md) | Protection against other websites before login (Host and Origin checks) | Accepted |
| [ADR-0043](ADR-0043-hybrid-search.md) | Hybrid lexical and semantic search | Proposed |
| [ADR-0044](ADR-0044-plan-document-structure.md) | Plan document structure: a plan folder with an index | Accepted |
| [ADR-0045](ADR-0045-media-processing-sandbox.md) | Media processing sandbox | Proposed |
| [ADR-0046](ADR-0046-user-content-origin-isolation.md) | User content origin isolation | Proposed |
| [ADR-0047](ADR-0047-tls-and-certificate-policy.md) | TLS and certificate policy (amends the TLS part of ADR-0010 once accepted) | Proposed |
| [ADR-0048](ADR-0048-secrets-management.md) | Secrets management | Proposed |
| [ADR-0049](ADR-0049-release-signing-and-provenance.md) | Release signing and provenance (direction decided by the user, D2) | Proposed |
| [ADR-0050](ADR-0050-service-account-permissions-and-mounts.md) | Service account, file permissions, and mount options per platform | Proposed |
| [ADR-0051](ADR-0051-zfs-pool-backend.md) | ZFS as an optional pool backend | Proposed |
| [ADR-0052](ADR-0052-appliance-os-image.md) | The appliance OS image ("local-ai-nas OS") | Proposed |
| [ADR-0053](ADR-0053-host-management.md) | Host management through the privileged helper | Proposed |
