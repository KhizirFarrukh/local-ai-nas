#!/usr/bin/env bash
# Runs the S01 performance baseline (S01.7-T04, NFR-003) on this machine
# and prints a Markdown table for docs/perf/: the listing benchmark of the
# files service, then TestPerfBaseline over real HTTP (listing latency,
# raw disk and loopback baselines, and upload and download throughput).
#
# Usage: scripts/perf-baseline.sh [transfer size, default 1GiB]
# The transfers write the size several times in the test's temporary
# folder; make sure there is room.
set -euo pipefail
size="${1:-1GiB}"
out="$(mktemp)"
trap 'rm -f "$out"' EXIT

echo "## Listing benchmark (files service, 10,000 entries, first page)"
go test -run '^$' -bench BenchmarkListLargeFolder -benchtime 20x ./internal/files | grep '^Benchmark'
echo
echo "## Over HTTP (transfer size $size)"
native="$out"
if command -v cygpath >/dev/null 2>&1; then native="$(cygpath -m "$out")"; fi
LOCALAINAS_PERF_SIZE="$size" LOCALAINAS_PERF_OUT="$native" \
	go test -count=1 -timeout 60m -run TestPerfBaseline ./internal/api >/dev/null
cat "$out"
