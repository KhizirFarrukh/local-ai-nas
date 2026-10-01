# Turns on the repository's git hooks on Windows (S03.5-T07; the user's
# decision D3): installs the pinned gitleaks into .\bin and points git at
# .githooks\ for this clone only. Git for Windows runs the hooks with its
# own sh.
#
# Usage: powershell -ExecutionPolicy Bypass -File scripts\install-git-hooks.ps1
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$Root = Split-Path -Parent $PSScriptRoot
& (Join-Path $PSScriptRoot 'install-gitleaks.ps1')
git -C $Root config core.hooksPath .githooks
if ($LASTEXITCODE -ne 0) { throw 'install-git-hooks: git config failed' }
Write-Output 'Hooks on: pre-commit and pre-push scan for secrets with gitleaks (.gitleaks.toml).'
