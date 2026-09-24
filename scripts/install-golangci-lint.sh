#!/usr/bin/env bash
# Installs the pinned golangci-lint binary into ./bin (ADR-0005). The
# upstream install script comes from the same release tag, and it verifies
# the downloaded archive against the release checksums.
#
# Usage: scripts/install-golangci-lint.sh   (on Windows, run it from Git Bash)
# Then:  ./bin/golangci-lint run  and  ./bin/golangci-lint fmt --diff
set -euo pipefail

VERSION=v2.13.2

cd "$(dirname "$0")/.."

curl -sSfL "https://raw.githubusercontent.com/golangci/golangci-lint/${VERSION}/install.sh" |
	sh -s -- -b ./bin "${VERSION}"
