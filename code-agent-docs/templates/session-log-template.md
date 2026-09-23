# Session S<NNN>: <YYYY-MM-DD>

<!--
Template for a session log (RULES.md R2, R8).
File name: code-agent-docs/logs/sessions/<YYYY-MM-DD>_S<NNN>.md, with the session number sequential across all sessions.
Write entries AS YOU GO. Never delete or rewrite past entries; add a correction entry that references the earlier one.
Redact secrets (passwords, API keys, tokens) in every entry.
-->

| Field | Value |
|---|---|
| Session number | S<NNN> |
| Date | <YYYY-MM-DD> |
| Start time | <HH:MM +TZ, from the system clock> |
| Editor | <e.g. Claude Code (VS Code extension)> |
| Model | <model name and ID, if known> |
| Plan version at start | <x.y.z> |
| Active stage / task at start | <S<NN> / S<NN>-T<NN>, or none> |

## Startup

- **Documents read:** <RULES.md, CURRENT_STATE.md, plan.md, previous log(s), stage doc, ADRs>
- **Previous session ended cleanly?** <yes / no; if no, what was reconstructed>
- **Git state:** <branch, last commit, uncommitted or unrecorded changes found and reported>
- **Resume summary given to the user:**
  > <current stage and task, last completed action, next planned action, open questions or blockers>

## Entries

<!--
One entry per event. Header format: ### E<NNN> | <HH:MM or sequence> | <TYPE>
TYPE is one of:
  USER     : the user's message, verbatim (secrets redacted)
  AGENT    : summary of the agent's response: what was proposed, decided, or done
  ACTION   : files created/modified/deleted, commands run and outcome, tests run and results
  DECISION : a decision made (by whom), with a reference to an ADR or plan version if applicable
  NOTE     : observations, assumptions, corrections (reference the corrected entry ID)
-->

### E001 | <HH:MM> | USER
<verbatim>

### E002 | <HH:MM> | AGENT
<summary>

### E003 | <HH:MM> | ACTION
- Files: <created / modified / deleted paths>
- Commands: `<command>`: <success / failure, key output>
- Tests: <suite, pass/fail counts>

## Closing summary

<!-- Append when the session ends (R8). -->

- **What was done:**
- **Decisions made:**
- **State at end:** <plan version, active stage/task, what's committed or uncommitted>
- **Exact next step:** <concrete enough for a fresh agent to act on immediately>
