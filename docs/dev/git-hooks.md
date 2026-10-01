# Git hooks: secret scanning

Every commit and push in this repository is scanned for secrets (API keys, private keys, tokens, passwords) by [gitleaks](https://github.com/gitleaks/gitleaks), on your own computer, before anything reaches GitHub. This is the owner's decision D3 ([security policy](../security/security-policy.md), section 7.3). GitHub's own secret scanning with push protection is the backstop on the server.

## Turn the hooks on (once per clone)

| System | Command |
|---|---|
| Linux, macOS, or Git Bash on Windows | `scripts/install-git-hooks.sh` |
| Windows PowerShell | `powershell -ExecutionPolicy Bypass -File scripts\install-git-hooks.ps1` |

The script:
1. downloads the pinned gitleaks release (8.30.1) for your system, checks its SHA-256 against the value pinned in `scripts/install-gitleaks.sh` (or `.ps1`), and puts the binary in `./bin` (ignored by git);
2. runs `git config core.hooksPath .githooks` for this clone only.

## What runs

| Hook | Scans | Command |
|---|---|---|
| `pre-commit` | The staged changes | `gitleaks git --pre-commit --staged --redact` |
| `pre-push` | The commits being pushed that no remote branch already has | `gitleaks git --redact --log-opts=<range>` |

Both use `.gitleaks.toml` (gitleaks's default rules). Findings are printed **redacted**, so the secret itself is not shown again on screen or in logs. A scan of a normal commit takes well under a second.

## When the hook blocks you

- **A real secret:** remove it from the staged files, and if it was ever pushed or shared, **rotate it first** (removing it from history is not enough once it was public). Then keep it outside the repository (configuration file, environment variable, or a secrets store).
- **A false positive** (for example a deliberately fake value in a test): add the narrowest possible entry to `.gitleaks.toml` (one file or one value), with a comment explaining why, in its own commit.
- **Never** use `git commit --no-verify` or `git push --no-verify` to get past it (RULES R15).

## Scanning everything by hand

```
./bin/gitleaks git --redact --log-opts="--all" .     # the whole history
./bin/gitleaks dir --redact <folder>                  # files in a folder
```

The whole history and the tree were scanned on 2026-10-01 (S03.5-T07): no finding.
