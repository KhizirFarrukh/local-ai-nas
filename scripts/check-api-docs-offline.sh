#!/usr/bin/env bash
# Checks that the offline API documentation renders without the network
# (S01.5-T06). It builds and starts the server on a loopback port with a
# temporary storage root, loads /api/docs/ in a headless Chromium-based
# browser (Edge or Chrome) whose resolver refuses every host except
# 127.0.0.1, and checks that Redoc rendered the operations.
#
# Usage: scripts/check-api-docs-offline.sh [path-to-browser]
# It needs a browser, so it is not part of CI; run it when the page or the
# bundle changes.
set -euo pipefail

browser="${1:-}"
if [ -z "$browser" ]; then
	for b in msedge microsoft-edge google-chrome chromium chromium-browser \
		"/c/Program Files (x86)/Microsoft/Edge/Application/msedge.exe" \
		"/c/Program Files/Google/Chrome/Application/chrome.exe"; do
		if command -v "$b" >/dev/null 2>&1 || [ -x "$b" ]; then
			browser="$b"
			break
		fi
	done
fi
if [ -z "$browser" ]; then
	echo "no Chromium-based browser found; pass its path" >&2
	exit 2
fi

work="$(mktemp -d)"
native() { if command -v cygpath >/dev/null 2>&1; then cygpath -m "$1"; else echo "$1"; fi; }
exe="$work/local-ai-nas"
case "$(uname -s)" in MINGW* | MSYS* | CYGWIN*) exe="$exe.exe" ;; esac
go build -o "$exe" ./cmd/local-ai-nas

port=$((18000 + RANDOM % 1000))
"$exe" serve --storage-root "$(native "$work/root")" --server-bind "127.0.0.1:$port" >"$work/server.log" 2>&1 &
pid=$!
cleanup() {
	kill "$pid" 2>/dev/null || true
	wait "$pid" 2>/dev/null || true
	rm -rf "$work"
}
trap cleanup EXIT

for _ in $(seq 1 60); do
	if curl -fsS "http://127.0.0.1:$port/api/v1/system/health" >/dev/null 2>&1; then
		break
	fi
	sleep 0.5
done

dom="$("$browser" --headless=new --disable-gpu --no-first-run --no-default-browser-check \
	--user-data-dir="$(native "$work/profile")" \
	--host-resolver-rules="MAP * ~NOTFOUND, EXCLUDE 127.0.0.1" \
	--virtual-time-budget=20000 --dump-dom "http://127.0.0.1:$port/api/docs/" 2>/dev/null)"

missing=0
for text in "local-ai-nas API" "Create a folder" "Download a file" "Upload a file (simple, streamed)" "Create a resumable upload (tus)"; do
	if ! grep -qF "$text" <<<"$dom"; then
		echo "not rendered: $text" >&2
		missing=1
	fi
done
if [ "$missing" -ne 0 ]; then
	echo "the API documentation did not render offline" >&2
	exit 1
fi
echo "the API documentation rendered with every host but 127.0.0.1 unreachable"
