# Installs the pinned gitleaks binary into .\bin for the git hooks
# (S03.5-T07; the user's decision D3), on Windows PowerShell. The release
# archive is checked against the SHA-256 pinned below before it is
# unpacked; nothing downloaded is ever run as a script (RULES R15).
#
# Usage: powershell -ExecutionPolicy Bypass -File scripts\install-gitleaks.ps1
# Usually run through scripts\install-git-hooks.ps1.
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$Version = '8.30.1'
# From gitleaks_8.30.1_checksums.txt, cross-checked with GitHub's asset digests.
$Builds = @{
    'AMD64' = @{ Asset = 'windows_x64.zip';   Sum = 'd29144deff3a68aa93ced33dddf84b7fdc26070add4aa0f4513094c8332afc4e' }
    'ARM64' = @{ Asset = 'windows_arm64.zip'; Sum = 'b95f5e4f5c425cedca7ee203d9afd29597e692c4924a12ed42f970537c72cc0f' }
}

$Root = Split-Path -Parent $PSScriptRoot
$Arch = $env:PROCESSOR_ARCHITECTURE
if (-not $Builds.ContainsKey($Arch)) {
    throw "install-gitleaks: no pinned build for Windows $Arch"
}
$Build = $Builds[$Arch]

$Tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("gitleaks-" + [guid]::NewGuid())
New-Item -ItemType Directory -Path $Tmp | Out-Null
try {
    $Archive = Join-Path $Tmp 'archive.zip'
    $Url = "https://github.com/gitleaks/gitleaks/releases/download/v$Version/gitleaks_${Version}_$($Build.Asset)"
    Invoke-WebRequest -Uri $Url -OutFile $Archive -UseBasicParsing
    $Actual = (Get-FileHash -Algorithm SHA256 -Path $Archive).Hash.ToLowerInvariant()
    if ($Actual -ne $Build.Sum) {
        throw "install-gitleaks: checksum mismatch for $Url (expected $($Build.Sum), got $Actual)"
    }
    $Out = Join-Path $Tmp 'x'
    Expand-Archive -Path $Archive -DestinationPath $Out
    $Bin = Join-Path $Root 'bin'
    New-Item -ItemType Directory -Force -Path $Bin | Out-Null
    Copy-Item -Path (Join-Path $Out 'gitleaks.exe') -Destination (Join-Path $Bin 'gitleaks.exe') -Force
    Write-Output "gitleaks $Version installed in .\bin"
}
finally {
    Remove-Item -Recurse -Force -Path $Tmp
}
