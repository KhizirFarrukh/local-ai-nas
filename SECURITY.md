# Security policy

## Reporting a vulnerability

Please report security problems **privately**, never in a public issue:

- Use GitHub's private vulnerability reporting: open the repository's **Security** tab and choose **Report a vulnerability**.
- If that option is not shown, open an issue that only asks for a private contact, with no details of the problem.
- Include the version or commit, what you found, how to reproduce it, and what an attacker could do with it.

What to expect:

- An acknowledgement within 7 days.
- Updates while the problem is investigated and fixed.
- Credit in the published advisory, if you want it.
- Target times for a fix after triage: Critical 7 days, High 30 days, Medium within 90 days. This is a small project with one maintainer, so these are best-effort targets, not guarantees.

Details of a problem are published only after a fixed release is available. The full process is in [docs/security/vulnerability-management.md](docs/security/vulnerability-management.md).

## Supported versions

There is no release yet; only the `develop` branch is maintained. The supported versions will be listed here from the first release.

## How the project handles security

Designed in alignment with ISO/IEC 27001:2022 Annex A controls (self-assessed; not certified). The security documents, including the alignment matrix and the threat model, are in [docs/security/](docs/security/README.md).
