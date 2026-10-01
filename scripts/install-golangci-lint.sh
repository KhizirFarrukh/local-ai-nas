#!/usr/bin/env bash
# Installs the pinned golangci-lint binary into ./bin (ADR-0005). The
# release archive is checked against the SHA-256 pinned below before
# anything is unpacked. Until S03 this script piped the upstream install
# script into sh; RULES R15 forbids running a download as a script
# (bug S03-B03, fixed in S03.5-T07).
#
# Usage: scripts/install-golangci-lint.sh   (on Windows, run it from Git Bash)
# Then:  ./bin/golangci-lint run  and  ./bin/golangci-lint fmt --diff
set -euo pipefail

VERSION=2.13.2

cd "$(dirname "$0")/.."

# SHA-256 of each release archive, from the digests GitHub reports for the
# v2.13.2 release assets (an immutable release).
case "$(uname -s)-$(uname -m)" in
	Linux-x86_64) platform=linux-amd64; ext=tar.gz; sum=2277d43b98ec0054280f2ac26b53268bae97682444678a59a657dd565da021d6 ;;
	Linux-aarch64 | Linux-arm64) platform=linux-arm64; ext=tar.gz; sum=a2a4e0065aa41be71f7c5ac90f271b61751331e5d04314e62afe4027855f0893 ;;
	Linux-armv7l) platform=linux-armv7; ext=tar.gz; sum=01648f24c70b37a6d2e240a2696d35e6ca3fe5bd1215815624e7cac90e6071f7 ;;
	Darwin-x86_64) platform=darwin-amd64; ext=tar.gz; sum=8a13aaf9cbbb1dee52824e862cf0d0720e5bb97c1f4260d1e51623a09492b57b ;;
	Darwin-arm64) platform=darwin-arm64; ext=tar.gz; sum=f4bf83f0b64f055c42b28fc9a38861839f69c096e61c788e72dfaae412011789 ;;
	MINGW*-x86_64 | MSYS*-x86_64 | CYGWIN*-x86_64) platform=windows-amd64; ext=zip; sum=4735fdc8e84a0cfb7a15a1c364a650942f88215e0d36c674ebc4024f7b554524 ;;
	MINGW*-aarch64 | MSYS*-aarch64) platform=windows-arm64; ext=zip; sum=2dbffbd1225d41ac5740f0b478a43b6517f3e3f702fe0ab3aec470bd6ec8e263 ;;
	*)
		echo "install-golangci-lint: no pinned build for $(uname -s) $(uname -m)" >&2
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

name="golangci-lint-${VERSION}-${platform}"
url="https://github.com/golangci/golangci-lint/releases/download/v${VERSION}/${name}.${ext}"
curl -sSfL -o "$tmp/archive" "$url"
actual="$(sha256 "$tmp/archive")"
if [ "$actual" != "$sum" ]; then
	echo "install-golangci-lint: checksum mismatch for $url" >&2
	echo "  expected $sum" >&2
	echo "  got      $actual" >&2
	exit 1
fi

mkdir -p "$tmp/x" bin
if [ "$ext" = zip ]; then
	unzip -q "$tmp/archive" -d "$tmp/x"
	cp "$tmp/x/$name/golangci-lint.exe" bin/golangci-lint.exe
else
	tar -xzf "$tmp/archive" -C "$tmp/x"
	cp "$tmp/x/$name/golangci-lint" bin/golangci-lint
	chmod +x bin/golangci-lint
fi
echo "golangci-lint ${VERSION} installed in ./bin"
