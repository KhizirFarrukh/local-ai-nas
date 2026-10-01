# Incident response

| Field | Value |
|---|---|
| Version | 1.0 (2026-10-01, plan change request #10) |
| Plan requirement | NFR-062 |
| Aligned with | ISO/IEC 27035-1:2023 and 27035-2:2023 (incident management), based on public summaries; ISO/IEC 27001:2022 controls 5.24–5.28 |
| Status | Documented; not yet exercised. Several steps use features that are still planned (marked with their stage). |

## Part A: for operators (you run a NAS and something looks wrong)

### A1. Signs of compromise

- Sign-ins you do not recognize, new sessions, devices, API tokens, or admin accounts (security page, S03.8; alerts, FR-378).
- Files changed, renamed, or encrypted in bulk; unexpected deletions.
- Settings changed that you did not change (LAN access, certificates, shares, remote access).
- An audit-log integrity warning (FR-379), or a media tool crashing or being stopped repeatedly on uploads (FR-378).
- The NAS using much more CPU, memory, or network than usual.

### A2. Immediate steps (in this order)

1. **Cut exposure:** turn off remote and public access (R06, R09), or unplug the NAS from the internet-facing network. Keep it running if you can, so evidence is not lost.
2. **Preserve evidence before changing anything:** export the audit log and the application logs (`.local-ai-nas/logs/`), and note the time.
3. **Revoke access:** sign out every session, revoke all API tokens and device tokens, and disable public links (one action per kind in the console; S03.3, S03.9, R09).
4. **Rotate secrets:** change the admin password and the passwords of affected users; regenerate the TLS certificate; rotate the secrets encryption key and the SMTP and webhook credentials ([crypto policy](crypto-policy.md), section 4).
5. **Check integrity:** run an integrity scan (S08.2) and compare against your backups.
6. **Restore** from a known-good backup taken before the incident (S08, R04) if data was changed.
7. **Update** to the latest release, which may contain the fix.
8. **Tell the people affected** (users of your NAS). Depending on where you live, data-protection law may require you to tell an authority (ISO control 5.5); this guide cannot give legal advice.
9. **Tell the project** privately ([`SECURITY.md`](../../SECURITY.md)): the version, what you saw, the time, and logs with personal data removed.

## Part B: for the project (a vulnerability is being exploited, or a release is compromised)

1. **Assess:** reproduce, score (CVSS), identify affected versions ([vulnerability management](vulnerability-management.md)).
2. **Contain:** if a release or update is compromised, withdraw it, and stop the updater from offering it (the signed update manifest, FR-381); revoke and rotate the affected signing key and announce the new public key in a release signed with the old key while it is still trusted, or through the repository if it is not.
3. **Fix and release** an emergency patch release; publish the advisory with clear steps for operators.
4. **Review after the incident** (ISO/IEC 27035 "learning from incidents"): what happened, why the controls missed it, what changes (tests, threat model entries, gate checks, documents). The review is recorded in the session log and the plan.

## Part C: what makes this possible (prepared in advance)

| Need | Where |
|---|---|
| A tamper-evident audit log that can be exported | FR-090, FR-379 (S03.6) |
| Security event alerts | FR-378 (S03.6, S10) |
| One-click revocation of sessions, tokens, devices, links | S03.3, S03.9, FR-365, R09 |
| Documented secret rotation | [Crypto policy](crypto-policy.md), section 4 |
| Backups and restore tests | S08, FR-355, R04 |
| Signed releases and updates, with key rotation | FR-381, FR-382, ADR-0049 |
