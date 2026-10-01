# Sourced by the hooks in this folder: sets GITLEAKS to the pinned binary in
# ./bin (installed by scripts/install-gitleaks.*), or else the one on PATH.
# $root must be set by the caller.
if [ -x "$root/bin/gitleaks" ]; then
	GITLEAKS="$root/bin/gitleaks"
elif [ -f "$root/bin/gitleaks.exe" ]; then
	GITLEAKS="$root/bin/gitleaks.exe"
elif command -v gitleaks >/dev/null 2>&1; then
	GITLEAKS="$(command -v gitleaks)"
else
	echo "gitleaks not found: run scripts/install-git-hooks.sh (or .ps1 on Windows)" >&2
	exit 1
fi
