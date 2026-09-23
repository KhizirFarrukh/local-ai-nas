# ADR-0019: SMB network shares via Samba (optional, Linux-only)

| Field | Value |
|---|---|
| Number | ADR-0019 |
| Status | **Proposed** (deferred; decided in S09.1 or later) |
| Date proposed | 2026-09-24 (session S003) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none |
| Superseded by | none |

## Context

P003 decides **WebDAV first** (ADR-0015) and keeps **SMB via Samba** as a later, **Linux-only**, optional addition that "stays 'Proposed'". SMB gives the most native network-drive experience on Windows and macOS clients. However, Samba runs outside the core, so its writes **bypass** the core's policy, area separation, sidecar, index, and audit logic.

## Options considered

| Option | Pros | Cons |
|---|---|---|
| **Samba alongside the core (Linux hosts only)** | Native experience, fast | Needs a design that maps per-user NAS permissions and shares onto Samba users, shares, and ACLs. Changes arrive via the watcher (ADR-0016). Not available on Windows hosts. More attack surface |
| An in-process SMB server library in Go | Enforcement inside the core | Immature ecosystem; high risk |
| No SMB (WebDAV only) | Simplest | Less native experience |

## Decision

**Not decided.** Intended direction (per P003): optional Samba integration on Linux hosts. Its design, covering permission mapping, share generation from NAS users and shares, the photos-area exposure policy (Q32), audit, and the watcher dependency, is produced in **S09.1**, and this ADR is updated then.

## Consequences

- Until decided, network drives are WebDAV only (ADR-0015).
- If adopted, S09.4 (watcher) becomes essential for correctness, and S09.6 tests must cover SMB.

## Approval record

> _Not approved. Deferred per P003 ("SMB via Samba: Linux-only, later, and still 'Proposed'")._

## Links

- **Related requirements:** FR-009, FR-124, FR-125
- **Related ADRs:** ADR-0015, ADR-0016
- **Related stages:** S09.1
- **Plan version:** 0.3.0
