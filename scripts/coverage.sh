#!/usr/bin/env bash
# Runs the tests of internal/... with coverage and fails when the total
# statement coverage is below the threshold (S01.1-T04, docs/testing.md).
# Generated code (internal/api/gen) is left out of the figure: it is
# produced by oapi-codegen from the spec, and its behavior is tested
# through internal/api. Writes coverage.out and prints the per-function
# report.
#
# Usage: scripts/coverage.sh [extra go test flags, e.g. -race]
#        COVERAGE_MIN=80 scripts/coverage.sh
set -euo pipefail

min="${COVERAGE_MIN:-80}"

cd "$(dirname "$0")/.."

pkgs=$(go list ./internal/... | grep -v '/internal/api/gen$')
# shellcheck disable=SC2086 # the package list is split on purpose
go test -covermode=atomic -coverprofile=coverage.out "$@" $pkgs
go tool cover -func=coverage.out

total=$(go tool cover -func=coverage.out | awk '/^total:/ { sub(/%/, "", $NF); print $NF }')
summary="Total coverage of internal/...: ${total}% (minimum ${min}%)"
echo "$summary"
# In GitHub Actions, publish the figure on the run's summary page.
if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
	echo "$summary" >>"$GITHUB_STEP_SUMMARY"
fi
if awk -v t="$total" -v m="$min" 'BEGIN { exit !(t + 0 < m + 0) }'; then
	echo "FAIL: coverage ${total}% is below ${min}%" >&2
	exit 1
fi
