# ADR-0039: Admin console architecture

| Field | Value |
|---|---|
| Number | ADR-0039 |
| Status | **Accepted** (2026-09-29, session S007, with the S03 approval) |
| Date proposed | 2026-09-29 (session S007) |
| Date of last status change | 2026-09-29 (session S007) |
| Supersedes | none |
| Superseded by | none |

## Context

- The user's requirement (S007 E044): the storage and drive features are "part of the admin console (gui based) app", a console "where sysadmin can manage everything related to storage management and system settings and drives management and all the admin stuff"; "it is very crucial" (FR-342–FR-345).
- The plan had admin pages spread over S03.8, S07.6, S08.7, S09.5, S10.1–S10.5, and S14 with no common structure.
- R09 (the public release) requires the admin interface to be reachable only from the LAN or the VPN by default (11c), so admin traffic must be easy to tell apart.
- The NAS will run on a Raspberry Pi (S007 E046): one binary, small memory footprint.

## Options considered

### Option A: An admin section of the same web app
The console lives at `/admin/…` in the same SvelteKit app (ADR-0009), shown only to admins, backed by admin API routes under `/api/v1/admin/…`.
- **Pros:** one binary and one design system; the admin sees the same app; a clear path prefix makes network restrictions (R09) and route-inventory tests simple; no extra memory on the Pi.
- **Cons:** the admin code ships to every client (it is only data-less code; every admin API is refused for non-admins on the server).

### Option B: A separate admin app on its own port
- **Pros:** can be firewalled separately.
- **Cons:** two apps to build and keep consistent; more memory; the same restriction is achievable with the path prefix and a bind setting.

## Decision

**Option A** (recommended, and chosen by the user with the S03 approval).
- **Structure (the console map, plan 6.6):** Overview; Storage and drives (health, drives, pools, migrations, SSD cache); Users and groups; Sharing; Security (sessions, 2FA policy, audit log); Network shares; Backups and recovery; Jobs; Logs and alerts; System settings (network and bind, time, notifications, updates, advanced); About and diagnostics. Each section is filled by the stage that builds its feature.
- **Rules:** every admin API route checks the admin role on the server (default deny; route inventory test); sensitive actions need recent re-authentication (S03.3); every admin action is audit-logged (S03.6); destructive actions show a preview and a typed confirmation (I10); one layout, component set, and wording; usable on a phone; accessible (NFR-015).
- **Optional:** a setting to allow `/admin` and `/api/v1/admin` only from given networks (LAN, VPN), required before R09.

## Consequences

- **Required:** S03.9 builds the console shell and its rules; every later admin feature is a console page (principle 4); S10.6 checks that every admin function is in the console and writes the admin guide; Q74 asks the user to confirm.

## Approval record

> "approve s03"
> (2026-09-29, session S007, log E050)

The user approved the S03 stage document, whose decision **D-1** (Q74) put option A (recommended) to the user: "admin console as `/admin` in the same web app (ADR-0039 option A, recommended), or a separate app on its own port". No change was asked, so option A is chosen and this ADR is Accepted. Proposed with the P007 report (S007 E047).

## Links

plan FR-342–FR-345, NFR-050, 6.6 · [ADR-0009](ADR-0009-web-ui-sveltekit.md) · the user's requirement (S007 E044)
