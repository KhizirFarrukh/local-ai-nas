#!/usr/bin/env bash
# Installs the pinned gitleaks binary into ./bin for the git hooks
# (S03.5-T07; the user's decision D3). The release archive is checked
# against the SHA-256 pinned below before anything is unpacked; nothing
# downloaded is ever run as a script (RULES R15).
#
# Usage: scripts/install-gitleaks.sh   (Linux, macOS, or Git Bash on Windows)
# Usually run through scripts/install-git-hooks.sh.
set -euo pipefail

VERSION=8.30.1

cd "$(dirname "$0")/.."

# SHA-256 of each release archive, from gitleaks_8.30.1_checksums.txt,
# cross-checked with the digests GitHub reports for the release assets.
case "$(uname -s)-$(uname -m)" in
	Linux-x86_64) asset=linux_x64.tar.gz; sum=551f6fc83ea457d62a0d98237cbad105af8d557003051f41f3e7ca7b3f2470eb ;;
	Linux-aarch64 | Linux-arm64) asset=linux_arm64.tar.gz; sum=e4a487ee7ccd7d3a7f7ec08657610aa3606637dab924210b3aee62570fb4b080 ;;
	Linux-armv7l) asset=linux_armv7.tar.gz; sum=8d39f0d94ba0d774b2282187656fb039a2d82893ec1fd6be7d7121aae759a57d ;;
	Darwin-x86_64) asset=darwin_x64.tar.gz; sum=dfe101a4db2255fc85120ac7f3d25e4342c3c20cf749f2c20a18081af1952709 ;;
	Darwin-arm64) asset=darwin_arm64.tar.gz; sum=b40ab0ae55c505963e365f271a8d3846efbc170aa17f2607f13df610a9aeb6a5 ;;
	MINGW*-x86_64 | MSYS*-x86_64 | CYGWIN*-x86_64) asset=windows_x64.zip; sum=d29144deff3a68aa93ced33dddf84b7fdc26070add4aa0f4513094c8332afc4e ;;
	MINGW*-aarch64 | MSYS*-aarch64) asset=windows_arm64.zip; sum=b95f5e4f5c425cedca7ee203d9afd29597e692c4924a12ed42f970537c72cc0f ;;
	*)
		echo "install-gitleaks: no pinned build for $(uname -s) $(uname -m)" >&2
		exit 1
		;;
esac

sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | cut -d' ' -f1
	else
		shasum -a 256 "$1" | cut -d' ' -f1
	fi
}

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

url="https://github.com/gitleaks/gitleaks/releases/download/v${VERSION}/gitleaks_${VERSION}_${asset}"
curl -sSfL -o "$tmp/archive" "$url"
actual="$(sha256 "$tmp/archive")"
if [ "$actual" != "$sum" ]; then
	echo "install-gitleaks: checksum mismatch for $url" >&2
	echo "  expected $sum" >&2
	echo "  got      $actual" >&2
	exit 1
fi

mkdir -p "$tmp/x" bin
case "$asset" in
	*.zip)
		unzip -q "$tmp/archive" -d "$tmp/x"
		cp "$tmp/x/gitleaks.exe" bin/gitleaks.exe
		;;
	*)
		tar -xzf "$tmp/archive" -C "$tmp/x"
		cp "$tmp/x/gitleaks" bin/gitleaks
		chmod +x bin/gitleaks
		;;
esac
echo "gitleaks ${VERSION} installed in ./bin"
