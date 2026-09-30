# docs

Project documentation, grouped by who it is for (organized in session S007; ADR-0004 amendment 1).

## For users and admins: [`guide/`](guide/)

- [`guide/web-interface.md`](guide/web-interface.md): using the web interface: browsing, uploads, downloads, organizing, previews, shortcuts, accessibility (S02.8-T03).
- [`guide/storage-root.md`](guide/storage-root.md): moving the storage root to another folder or disk (S01.2-T07, NFR-036).
- Install and admin guides, and the rest of the user guide, follow in S14.4.

## The API: [`api/`](api/)

- [`api/usage.md`](api/usage.md): using the API with curl, for Linux and macOS and for Windows PowerShell, step by step (S01.7-T05).
- [`api/errors.md`](api/errors.md): the error format (RFC 9457 problem details), every error code, and the name rules.
- [`api/conventions.md`](api/conventions.md): the rules every endpoint follows, with the review checklist (S01.5-T01).
- [`api/versioning.md`](api/versioning.md): versioning and deprecation policy (S01.5-T02).

## For contributors: [`dev/`](dev/)

- [`dev/testing.md`](dev/testing.md): how tests, fuzzing, coverage, and lint are run (S01.1-T04).
- [`dev/licensing.md`](dev/licensing.md): the dependency license policy (S01.1-T02).

## Measurements and checks: [`reports/`](reports/)

- [`reports/S01-performance-baseline.md`](reports/S01-performance-baseline.md): performance baselines against NFR-003 (S01.7-T04).
- [`reports/S02-cross-browser.md`](reports/S02-cross-browser.md): the cross-browser check of the web interface in Chrome, Edge, and Firefox (S02.8-T03).

The agent's working documents (plan, rules, stage documents, logs) live in [`code-agent-docs/`](../code-agent-docs/). The user's own inputs to the agent (prompts and code reviews) live in [`inputs/`](../inputs/).
