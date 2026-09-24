<#
.SYNOPSIS
The stage 1 API demo (S01.7-T06) for Windows PowerShell 5.1 and PowerShell 7.

.DESCRIPTION
Manages files and folders end to end with curl.exe, as docs/api/usage.md
describes. It creates folders, uploads a file, uploads a larger file with
tus, cuts that upload off and resumes it, lists, downloads a range,
renames, moves, copies, and deletes. Every step checks the answer, and the
script stops at the first unexpected one.

The server must be running (README, "Development"). The demo works in a
new folder /local-ai-nas-demo-<time> and deletes it at the end, unless
-Keep is given.

.EXAMPLE
powershell -ExecutionPolicy Bypass -File scripts\demo.ps1
powershell -ExecutionPolicy Bypass -File scripts\demo.ps1 -BaseUrl http://127.0.0.1:8080 -Keep
#>
param(
  [string]$BaseUrl = 'http://127.0.0.1:8080',
  [switch]$Keep
)
$ErrorActionPreference = 'Stop'

$api = "$BaseUrl/api/v1"
$root = '/local-ai-nas-demo-' + (Get-Date -Format 'yyyyMMdd-HHmmss') + "-$PID"
$work = Join-Path ([IO.Path]::GetTempPath()) ('local-ai-nas-demo-' + [guid]::NewGuid())
New-Item -ItemType Directory $work | Out-Null
$bodyFile = Join-Path $work 'body'
$headerFile = Join-Path $work 'headers'
$jsonFile = Join-Path $work 'request.json'
$created = $false

function Step($text) { Write-Host "`n== $text" }
function Ok($text) { Write-Host "   ok: $text" }
function Fail($text) {
  if ($script:created) { $text += "`nThe demo folder $root is left on the server." }
  throw "FAILED: $text"
}
function Body { [IO.File]::ReadAllText($bodyFile) }

# Call <status> <curl arguments...> runs curl.exe, keeps the answer's body
# and headers, and stops unless the status is the expected one.
function Call($Want) {
  $got = & curl.exe -sS -o $bodyFile -D $headerFile -w '%{http_code}' @args
  if ($LASTEXITCODE -ne 0) { Fail "curl.exe failed ($LASTEXITCODE)" }
  if ($got -ne "$Want") { Fail "expected HTTP $Want, got ${got}: $(Body)" }
}

# Json <status> <method> <path> <hashtable> sends a JSON request. The body
# goes through a file: Windows PowerShell 5.1 would drop the quotes of a
# JSON argument.
function Json($Want, $Method, $Path, $Data) {
  [IO.File]::WriteAllText($jsonFile, ($Data | ConvertTo-Json -Compress))
  Call $Want -X $Method "$api$Path" -H 'Content-Type: application/json' --data-binary "@$jsonFile"
}

# Header <name> returns a header of the last answer.
function Header($Name) {
  $value = $null
  foreach ($line in [IO.File]::ReadAllLines($headerFile)) {
    if ($line -match "^${Name}: (.*)$") { $value = $Matches[1] }
  }
  $value
}

# Has <text> checks that the last answer's body contains the text.
function Has($Text) {
  if (-not (Body).Contains($Text)) { Fail "the answer lacks ${Text}: $(Body)" }
}

function B64($Text) { [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($Text)) }
function Sha256($Path) { (Get-FileHash $Path -Algorithm SHA256).Hash.ToLowerInvariant() }

try {
  Step "Check the server at $api"
  Call 200 "$api/system/health"
  Has '"status":"ok"'
  Ok 'healthy'

  Step "Create folders in $root"
  Json 201 POST /files/folders @{ path = "$root/docs"; parents = $true }
  $script:created = $true
  Json 201 POST /files/folders @{ path = "$root/archive" }
  Ok "$root/docs and $root/archive"

  Step 'Upload a file'
  $report = Join-Path $work 'report.txt'
  [IO.File]::WriteAllText($report, "hello, NAS`n")
  Call 201 -T $report "$api/files/content?path=$root/docs/report.txt"
  Has '"size":11'
  Ok 'report.txt (11 bytes)'
  Call 409 -T $report "$api/files/content?path=$root/docs/report.txt"
  Has '"code":"conflict"'
  Ok 'the same name again is refused (409 conflict)'
  Call 201 -T $report "$api/files/content?path=$root/docs/report.txt&on_conflict=rename"
  Has '"name":"report (1).txt"'
  Ok 'with on_conflict=rename it is stored as "report (1).txt"'

  Step 'Upload a 20 MiB file with tus, cut it off, and resume it'
  $big = Join-Path $work 'big.bin'
  $size = 20MB
  $bytes = [byte[]]::new($size)
  [Random]::new().NextBytes($bytes)
  [IO.File]::WriteAllBytes($big, $bytes)
  $sum = Sha256 $big
  Call 201 -X POST "$api/files/uploads/" -H 'Tus-Resumable: 1.0.0' -H "Upload-Length: $size" `
    -H "Upload-Metadata: target_path $(B64 "$root/docs/big.bin"),sha256 $(B64 $sum)"
  $loc = Header 'Location'
  if (-not $loc) { Fail 'no Location header' }
  Ok "created $loc"
  & curl.exe -s -o NUL -X PATCH $loc -T $big -H 'Tus-Resumable: 1.0.0' -H 'Upload-Offset: 0' `
    -H 'Content-Type: application/offset+octet-stream' --limit-rate 2M --max-time 2
  if ($LASTEXITCODE -ne 28) { Fail "the cut-off transfer ended with curl.exe exit code $LASTEXITCODE, not 28 (timeout)" }
  Ok 'the transfer was cut off after 2 seconds'
  # The server may still be closing the cut-off request (423 locked).
  for ($i = 0; $i -lt 10; $i++) {
    $got = & curl.exe -s -o NUL -D $headerFile -w '%{http_code}' -I $loc -H 'Tus-Resumable: 1.0.0'
    if ($got -ne '423') { break }
    Start-Sleep -Seconds 1
  }
  if ($got -ne '200') { Fail "HEAD on the upload answered $got" }
  $offset = [long](Header 'Upload-Offset')
  if ($offset -le 0 -or $offset -ge $size) { Fail "offset $offset is not part of $size" }
  Ok "the server has $offset of $size bytes"
  Call 204 -X PATCH $loc -T $big -C $offset -H 'Tus-Resumable: 1.0.0' -H "Upload-Offset: $offset" `
    -H 'Content-Type: application/offset+octet-stream'
  if ((Header 'Upload-Offset') -ne "$size") { Fail "offset after the resume: $(Header 'Upload-Offset')" }
  if ((Header 'Item-Path') -ne "$root/docs/big.bin") { Fail "Item-Path: $(Header 'Item-Path')" }
  Ok "resumed from byte $offset; the file is $(Header 'Item-Path')"
  Call 200 "$api/files/content?path=$root/docs/big.bin"
  if ((Sha256 $bodyFile) -ne $sum) { Fail 'the downloaded file differs' }
  Ok "downloaded again: SHA-256 matches ($sum)"

  Step 'List the folder'
  Call 200 "$api/files/items?path=$root/docs"
  Has '"name":"report.txt"'
  Has '"name":"report (1).txt"'
  Has '"name":"big.bin"'
  Ok 'report.txt, report (1).txt, big.bin'

  Step 'Download a range'
  Call 206 -r 0-4 "$api/files/content?path=$root/docs/report.txt"
  if ((Body) -ne 'hello') { Fail "range body: $(Body)" }
  if ((Header 'Content-Range') -ne 'bytes 0-4/11') { Fail "Content-Range: $(Header 'Content-Range')" }
  Ok 'bytes 0-4 are "hello" (206, Content-Range: bytes 0-4/11)'

  Step 'Rename, move, and copy'
  Json 200 POST /files/operations/rename @{ path = "$root/docs/report.txt"; new_name = 'notes.txt' }
  Has "`"path`":`"$root/docs/notes.txt`""
  Ok 'report.txt renamed to notes.txt'
  Json 200 POST /files/operations/move @{ from = "$root/docs/notes.txt"; to = "$root/archive/notes.txt" }
  Has "`"path`":`"$root/archive/notes.txt`""
  Ok "notes.txt moved to $root/archive"
  Json 201 POST /files/operations/copy @{ from = "$root/archive"; to = "$root/archive-copy" }
  Call 200 "$api/files/content?path=$root/archive-copy/notes.txt"
  if ((Body) -ne "hello, NAS`n") { Fail "copied content: $(Body)" }
  Ok "$root/archive copied with its content to $root/archive-copy"

  Step 'Delete'
  Call 204 -X DELETE "$api/files/items?path=$root/archive-copy/notes.txt"
  Ok 'a file'
  Call 409 -X DELETE "$api/files/items?path=$root/archive"
  Ok 'a folder with something in it needs recursive=true (409)'
  Call 204 -X DELETE "$api/files/items?path=$root/archive&recursive=true"
  Call 404 "$api/files/items?path=$root/archive"
  Ok 'the folder with everything in it; it is gone (404)'

  if ($Keep) {
    Write-Host "`nKept the demo folder $root."
  } else {
    Call 204 -X DELETE "$api/files/items?path=$root&recursive=true"
    Write-Host "`nRemoved the demo folder $root."
  }
  Write-Host 'Demo finished: every step passed.'
} finally {
  Remove-Item -Recurse -Force $work
}
