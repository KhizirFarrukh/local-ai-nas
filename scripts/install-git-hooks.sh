#!/usr/bin/env bash
# Turns on the repository's git hooks (S03.5-T07; the user's decision D3):
# installs the pinned gitleaks into ./bin and points git at .githooks/ for
# this clone only (a repository setting, not a global one).
#
# Usage: scripts/install-git-hooks.sh   (Linux, macOS, or Git Bash on Windows)
set -euo pipefail

cd "$(dirname "$0")/.."
scripts/install-gitleaks.sh
git config core.hooksPath .githooks
echo "Hooks on: pre-commit and pre-push scan for secrets with gitleaks (.gitleaks.toml)."
