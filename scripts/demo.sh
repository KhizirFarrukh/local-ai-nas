#!/usr/bin/env bash
# The stage 1 API demo (S01.7-T06): manages files and folders end to end
# with curl, as docs/api/usage.md describes. It creates folders, uploads a
# file, uploads a larger file with tus, cuts that upload off and resumes
# it, lists, downloads a range, renames, moves, copies, and deletes. Every
# step checks the answer, and the script stops at the first unexpected one.
#
# Usage: scripts/demo.sh [server URL, default http://127.0.0.1:8080]
# The server must be running (README, "Development"). The demo works in a
# new folder /local-ai-nas-demo-<time> and deletes it at the end; set
# KEEP=1 to keep it. On Windows, run it from Git Bash, or use demo.ps1.
set -euo pipefail
# Paths on the server only appear inside URLs and JSON bodies, never as
# arguments of their own (such as path=/x), which Git Bash would turn
# into Windows paths.

api="${1:-http://127.0.0.1:8080}/api/v1"
root="/local-ai-nas-demo-$(date +%Y%m%d-%H%M%S)-$$"
work="$(mktemp -d)"
created=
trap 'rm -rf "$work"' EXIT

step() { printf '\n== %s\n' "$*"; }
ok() { printf '   ok: %s\n' "$*"; }
fail() {
	printf 'FAILED: %s\n' "$*" >&2
	[ -z "$created" ] || printf 'The demo folder %s is left on the server.\n' "$root" >&2
	exit 1
}

# call <status> <curl arguments...> runs curl, keeps the answer's body in
# $work/body and its headers in $work/headers, and stops unless the status
# is the expected one.
call() {
	local want="$1" got
	shift
	got="$(curl -sS -o "$work/body" -D "$work/headers" -w '%{http_code}' "$@")" || fail "curl failed ($?)"
	[ "$got" = "$want" ] || fail "expected HTTP $want, got $got: $(cat "$work/body")"
}

# json <status> <method> <path> <body> sends a JSON request.
json() {
	call "$1" -X "$2" "$api$3" -H 'Content-Type: application/json' -d "$4"
}

# header <name> prints a header of the last answer.
header() {
	tr -d '\r' <"$work/headers" | sed -n "s/^$1: //p" | tail -n 1
}

# has <text> checks that the last answer's body contains the text.
has() {
	grep -qF -- "$1" "$work/body" || fail "the answer lacks $1: $(cat "$work/body")"
}

b64() { printf %s "$1" | base64 | tr -d '\n'; }

sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | cut -d ' ' -f 1
	else
		shasum -a 256 "$1" | cut -d ' ' -f 1
	fi
}

step "Check the server at $api"
call 200 "$api/system/health"
has '"status":"ok"'
ok "healthy"

step "Create folders in $root"
json 201 POST /files/folders "{\"path\":\"$root/docs\",\"parents\":true}"
created=1
json 201 POST /files/folders "{\"path\":\"$root/archive\"}"
ok "$root/docs and $root/archive"

step "Upload a file"
printf 'hello, NAS\n' >"$work/report.txt"
call 201 -T "$work/report.txt" "$api/files/content?path=$root/docs/report.txt"
has '"size":11'
ok "report.txt (11 bytes)"
call 409 -T "$work/report.txt" "$api/files/content?path=$root/docs/report.txt"
has '"code":"conflict"'
ok "the same name again is refused (409 conflict)"
call 201 -T "$work/report.txt" "$api/files/content?path=$root/docs/report.txt&on_conflict=rename"
has '"name":"report (1).txt"'
ok "with on_conflict=rename it is stored as \"report (1).txt\""

step "Upload a 20 MiB file with tus, cut it off, and resume it"
head -c 20971520 /dev/urandom >"$work/big.bin"
size=20971520
sum="$(sha256 "$work/big.bin")"
call 201 -X POST "$api/files/uploads/" -H 'Tus-Resumable: 1.0.0' -H "Upload-Length: $size" \
	-H "Upload-Metadata: target_path $(b64 "$root/docs/big.bin"),sha256 $(b64 "$sum")"
loc="$(header Location)"
[ -n "$loc" ] || fail "no Location header"
ok "created $loc"
rc=0
curl -s -o /dev/null -X PATCH "$loc" -T "$work/big.bin" -H 'Tus-Resumable: 1.0.0' -H 'Upload-Offset: 0' \
	-H 'Content-Type: application/offset+octet-stream' --limit-rate 2M --max-time 2 || rc=$?
[ "$rc" = 28 ] || fail "the cut-off transfer ended with curl exit code $rc, not 28 (timeout)"
ok "the transfer was cut off after 2 seconds"
# The server may still be closing the cut-off request (423 locked).
for _ in 1 2 3 4 5 6 7 8 9 10; do
	got="$(curl -s -o /dev/null -D "$work/headers" -w '%{http_code}' -I "$loc" -H 'Tus-Resumable: 1.0.0')" || true
	[ "$got" = 423 ] || break
	sleep 1
done
[ "$got" = 200 ] || fail "HEAD on the upload answered $got"
offset="$(header Upload-Offset)"
[ "$offset" -gt 0 ] && [ "$offset" -lt "$size" ] || fail "offset $offset is not part of $size"
ok "the server has $offset of $size bytes"
call 204 -X PATCH "$loc" -T "$work/big.bin" -C "$offset" -H 'Tus-Resumable: 1.0.0' -H "Upload-Offset: $offset" \
	-H 'Content-Type: application/offset+octet-stream'
[ "$(header Upload-Offset)" = "$size" ] || fail "offset after the resume: $(header Upload-Offset)"
[ "$(header Item-Path)" = "$root/docs/big.bin" ] || fail "Item-Path: $(header Item-Path)"
ok "resumed from byte $offset; the file is $(header Item-Path)"
call 200 "$api/files/content?path=$root/docs/big.bin"
[ "$(sha256 "$work/body")" = "$sum" ] || fail "the downloaded file differs"
ok "downloaded again: SHA-256 matches ($sum)"

step "List the folder"
call 200 "$api/files/items?path=$root/docs"
has '"name":"report.txt"'
has '"name":"report (1).txt"'
has '"name":"big.bin"'
ok "report.txt, report (1).txt, big.bin"

step "Download a range"
call 206 -r 0-4 "$api/files/content?path=$root/docs/report.txt"
[ "$(cat "$work/body")" = hello ] || fail "range body: $(cat "$work/body")"
[ "$(header Content-Range)" = "bytes 0-4/11" ] || fail "Content-Range: $(header Content-Range)"
ok "bytes 0-4 are \"hello\" (206, Content-Range: bytes 0-4/11)"

step "Rename, move, and copy"
json 200 POST /files/operations/rename "{\"path\":\"$root/docs/report.txt\",\"new_name\":\"notes.txt\"}"
has "\"path\":\"$root/docs/notes.txt\""
ok "report.txt renamed to notes.txt"
json 200 POST /files/operations/move "{\"from\":\"$root/docs/notes.txt\",\"to\":\"$root/archive/notes.txt\"}"
has "\"path\":\"$root/archive/notes.txt\""
ok "notes.txt moved to $root/archive"
json 201 POST /files/operations/copy "{\"from\":\"$root/archive\",\"to\":\"$root/archive-copy\"}"
call 200 "$api/files/content?path=$root/archive-copy/notes.txt"
[ "$(cat "$work/body")" = "hello, NAS" ] || fail "copied content: $(cat "$work/body")"
ok "$root/archive copied with its content to $root/archive-copy"

step "Delete"
call 204 -X DELETE "$api/files/items?path=$root/archive-copy/notes.txt"
ok "a file"
call 409 -X DELETE "$api/files/items?path=$root/archive"
ok "a folder with something in it needs recursive=true (409)"
call 204 -X DELETE "$api/files/items?path=$root/archive&recursive=true"
call 404 "$api/files/items?path=$root/archive"
ok "the folder with everything in it; it is gone (404)"

if [ "${KEEP:-}" = 1 ]; then
	printf '\nKept the demo folder %s.\n' "$root"
else
	call 204 -X DELETE "$api/files/items?path=$root&recursive=true"
	printf '\nRemoved the demo folder %s.\n' "$root"
fi
printf 'Demo finished: every step passed.\n'
