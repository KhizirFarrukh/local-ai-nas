#!/usr/bin/env bash
# Checks that every Go package linked into the product has an allowed license
# (NFR-029). The allow-list is scripts/allowed-licenses.txt; the policy is
# docs/licensing.md. Test-only imports are not checked because they are not
# distributed.
#
# Usage: scripts/check-licenses.sh   (on Windows, run it from Git Bash)
set -euo pipefail

cd "$(dirname "$0")/.."

allowed=$(grep -v -e '^[[:space:]]*#' -e '^[[:space:]]*$' scripts/allowed-licenses.txt | tr -d '\r' | paste -sd, -)

exec go tool go-licenses check ./... --allowed_licenses="$allowed"
