# Using the API with curl

This guide goes through every file operation of stage 1 with `curl`: first for Linux and macOS (`sh`, `bash`, `zsh`), then the same step for Windows PowerShell. The rules behind each request are in [conventions.md](conventions.md), and the error codes in [errors.md](errors.md). The server also serves the full reference at `http://127.0.0.1:8080/api/docs/`.

Stage 1 has no login yet. The server only listens on this computer (`127.0.0.1`), and everything below goes to your own files area, shown as `/`.

## Before you start

Start the server as described in the [README](../../README.md) (section "Development"). Then, in a second terminal, set a variable for the API's address:

```sh
API=http://127.0.0.1:8080/api/v1
```

```powershell
$API = 'http://127.0.0.1:8080/api/v1'
```

**In PowerShell:**

- Type `curl.exe`, not `curl`. In Windows PowerShell 5.1, `curl` is another name for `Invoke-WebRequest`. `curl.exe` comes with Windows 10 (version 1803 and later) and Windows 11.
- Windows PowerShell 5.1 removes the double quotes from a JSON argument, so `-d '{"path":"/docs"}'` reaches the server broken. Define this helper once per session instead. It builds the JSON from a hashtable and sends it through a temporary file (UTF-8 without a byte-order mark), which works in every PowerShell version:

  ```powershell
  function Send-Json($Method, $Url, $Body) {
    $file = New-TemporaryFile
    try {
      [IO.File]::WriteAllText($file.FullName, ($Body | ConvertTo-Json -Compress))
      curl.exe -s -X $Method $Url -H 'Content-Type: application/json' --data-binary "@$($file.FullName)"
    } finally {
      Remove-Item $file
    }
  }
  ```

- To show names with non-ASCII characters correctly, run `[Console]::OutputEncoding = [Text.Encoding]::UTF8` first.

**In Git Bash on Windows:** run `export MSYS_NO_PATHCONV=1` first. Otherwise Git Bash turns arguments such as `path=/docs` into Windows paths (`C:/Program Files/Git/docs`).

Every answer is JSON on one line. `jq` (Linux and macOS) or `ConvertFrom-Json` (PowerShell) makes it easier to read. The answers below are shortened.

## Check the server

```sh
curl -s "$API/system/health"
```

```powershell
curl.exe -s "$API/system/health"
```

The answer is `"status":"ok"`, with one entry for each check (configuration, storage, free space, database). If a check fails, the status is `503` and the entry says why.

## Create folders

```sh
curl -s -X POST "$API/files/folders" -H 'Content-Type: application/json' -d '{"path":"/docs"}'
curl -s -X POST "$API/files/folders" -H 'Content-Type: application/json' -d '{"path":"/archive/2026","parents":true}'
```

```powershell
Send-Json POST "$API/files/folders" @{ path = '/docs' }
Send-Json POST "$API/files/folders" @{ path = '/archive/2026'; parents = $true }
```

Each answers `201` with the new folder:

```json
{"kind":"dir","mod_time":"2026-09-24T18:02:05.9978821Z","name":"docs","path":"/docs","size":0}
```

`parents` also creates the missing folders above it (`/archive`). If the folder already exists, the answer is `409 conflict`; with `"on_conflict":"overwrite"`, it is `200` with the existing folder.

## Upload a file

A simple upload sends the file as the body of a `PUT`. `-T` does that, and it also sets the `Content-Length` header, which the server needs.

```sh
printf 'hello, NAS\n' > report.txt
curl -s -T report.txt "$API/files/content?path=/docs/report.txt"
curl -s -T report.txt "$API/files/content?path=/docs/report.txt&on_conflict=rename"
```

```powershell
Set-Content report.txt 'hello, NAS'
curl.exe -s -T report.txt "$API/files/content?path=/docs/report.txt"
curl.exe -s -T report.txt "$API/files/content?path=/docs/report.txt&on_conflict=rename"
```

The first upload answers `201` with the file:

```json
{"etag":"\"5019dc6a2390a2b2672027d8\"","kind":"file","mime":"text/plain; charset=utf-8","mod_time":"2026-09-24T18:02:06.1744911Z","name":"report.txt","path":"/docs/report.txt","size":11}
```

The second one finds the name taken. Because of `on_conflict=rename`, it stores the file as `report (1).txt`. `on_conflict=overwrite` would replace the file (`200`), and without `on_conflict` the answer is `409 conflict`.

A file only appears when all of it has arrived, so a broken upload leaves nothing behind. For large files or an unreliable network, use a [resumable upload](#resumable-upload-tus).

## List a folder

```sh
curl -s "$API/files/items?path=/docs"
curl -s "$API/files/items?path=/docs&sort=size&order=desc&limit=1"
```

```powershell
curl.exe -s "$API/files/items?path=/docs"
(curl.exe -s "$API/files/items?path=/docs" | ConvertFrom-Json).items | Format-Table name, kind, size
```

The answer has `item`, the folder itself, and `items`, one page of what is in it:

```json
{"item":{"kind":"dir","name":"docs","path":"/docs",…},"items":[{"kind":"file","name":"report (1).txt",…},{"kind":"file","name":"report.txt",…}]}
```

- A page has at most `limit` items (default 100, at most 1000). When more follow, the answer also has `next_cursor`. Add `&cursor=<next_cursor>` to the same request to get the next page.
- `sort` is `name` (default), `size`, `mod_time`, or `kind`, and `order` is `asc` or `desc`.
- For a file, such as `?path=/docs/report.txt`, the answer is just `item`: the file's details.

## Download

```sh
curl -s -o copy.txt "$API/files/content?path=/docs/report.txt"
curl -s -r 0-4 "$API/files/content?path=/docs/report.txt"
curl -s -I "$API/files/content?path=/docs/report.txt"
```

```powershell
curl.exe -s -o copy.txt "$API/files/content?path=/docs/report.txt"
curl.exe -s -r 0-4 "$API/files/content?path=/docs/report.txt"
curl.exe -s -I "$API/files/content?path=/docs/report.txt"
```

- The first command saves the file as `copy.txt`.
- `-r 0-4` asks for a range: bytes 0 to 4, `hello`. The answer is `206`.
- `-I` asks for the headers only: the size, `ETag`, `Last-Modified`, and the file's name in `Content-Disposition`.
- An interrupted download continues where it stopped with `-C -`: `curl -s -C - -o copy.txt "…"`.

## Rename, move, and copy

```sh
curl -s -X POST "$API/files/operations/rename" -H 'Content-Type: application/json' -d '{"path":"/docs/report.txt","new_name":"notes.txt"}'
curl -s -X POST "$API/files/operations/move" -H 'Content-Type: application/json' -d '{"from":"/docs/notes.txt","to":"/archive/2026/notes.txt"}'
curl -s -X POST "$API/files/operations/copy" -H 'Content-Type: application/json' -d '{"from":"/archive","to":"/archive-copy"}'
```

```powershell
Send-Json POST "$API/files/operations/rename" @{ path = '/docs/report.txt'; new_name = 'notes.txt' }
Send-Json POST "$API/files/operations/move" @{ from = '/docs/notes.txt'; to = '/archive/2026/notes.txt' }
Send-Json POST "$API/files/operations/copy" @{ from = '/archive'; to = '/archive-copy' }
```

- Each answers with the item at its new path.
- `new_name` is one name in the same folder.
- `to` is the full new path, so a move can rename at the same time. The folder it goes into must exist.
- A copy of a folder copies everything in it. One request copies at most 1000 items or 1 GiB (`422 too_large_for_sync` above that).
- All three take `on_conflict`, as the upload does.

## Delete

```sh
curl -s -X DELETE "$API/files/items?path=/archive/2026/notes.txt"
curl -s -i -X DELETE "$API/files/items?path=/archive-copy"
curl -s -X DELETE "$API/files/items?path=/archive-copy&recursive=true"
```

```powershell
curl.exe -s -X DELETE "$API/files/items?path=/archive/2026/notes.txt"
curl.exe -s -i -X DELETE "$API/files/items?path=/archive-copy"
curl.exe -s -X DELETE "$API/files/items?path=/archive-copy&recursive=true"
```

- A delete answers `204` with no body.
- A folder with something in it needs `recursive=true`. Without it, the answer is `409 conflict`, as the second command shows.
- Deleting is permanent. A trash comes in a later stage.

## Resumable upload (tus)

For large files and unreliable connections, uploads use the [tus 1.0.0](https://tus.io/protocols/resumable-upload) protocol. It can continue after a broken connection without sending the whole file again. Apps normally use a tus client library (listed at tus.io), which does the steps below by itself. Here they are by hand with curl.

- Every tus request carries `Tus-Resumable: 1.0.0`.
- One request can carry at most `uploads.max_chunk_size` (64 MiB by default). The commands below send the file in one request, so they work for files up to that size. A client library splits larger files.

**1. Create the upload.** This makes a 20 MB test file and asks the server for an upload with the file's size and its target path. Metadata values are base64-encoded, as tus requires.

```sh
head -c 20000000 /dev/urandom > big.bin
LOC=$(curl -s -o /dev/null -D - -X POST "$API/files/uploads/" \
  -H 'Tus-Resumable: 1.0.0' \
  -H "Upload-Length: $(wc -c < big.bin)" \
  -H "Upload-Metadata: target_path $(printf %s /docs/big.bin | base64 | tr -d '\n')" |
  tr -d '\r' | sed -n 's/^Location: //p')
echo "$LOC"
```

```powershell
$bytes = [byte[]]::new(20MB); [Random]::new().NextBytes($bytes)
[IO.File]::WriteAllBytes("$PWD\big.bin", $bytes)
$meta = 'target_path ' + [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes('/docs/big.bin'))
$headers = curl.exe -s -o NUL -D - -X POST "$API/files/uploads/" -H 'Tus-Resumable: 1.0.0' -H "Upload-Length: $((Get-Item big.bin).Length)" -H "Upload-Metadata: $meta"
$LOC = ($headers | Select-String '^Location: (.+)').Matches[0].Groups[1].Value
$LOC
```

- The answer is `201`, and its `Location` header is the upload's address, such as `http://127.0.0.1:8080/api/v1/files/uploads/4l5qqhwc5m5g76tbb25cxrrs6v`.
- The server checks the target at once: the name, the folder, the free space, and whether the name is taken. So a refused upload fails before any data is sent.
- More metadata goes after a comma, each value base64-encoded:
  - `on_conflict` works as for the simple upload;
  - `sha256` is the file's checksum, which the server verifies before the file appears.

**2. Send the file.** To show a resume, this transfer is slowed to 2 MB/s and cut off after 3 seconds, as a dropped connection would be:

```sh
curl -s -X PATCH "$LOC" -T big.bin \
  -H 'Tus-Resumable: 1.0.0' -H 'Upload-Offset: 0' \
  -H 'Content-Type: application/offset+octet-stream' \
  --limit-rate 2M --max-time 3
```

```powershell
curl.exe -s -X PATCH $LOC -T big.bin -H 'Tus-Resumable: 1.0.0' -H 'Upload-Offset: 0' -H 'Content-Type: application/offset+octet-stream' --limit-rate 2M --max-time 3
```

curl stops with error 28 (timeout). The server keeps what arrived.

**3. Resume.** Ask the server how much it has (`HEAD` gives `Upload-Offset`), then send the rest. `-C` makes curl skip the bytes the server already has:

```sh
OFF=$(curl -s -I "$LOC" -H 'Tus-Resumable: 1.0.0' | tr -d '\r' | sed -n 's/^Upload-Offset: //p')
echo "$OFF"
curl -s -i -X PATCH "$LOC" -T big.bin -C "$OFF" \
  -H 'Tus-Resumable: 1.0.0' -H "Upload-Offset: $OFF" \
  -H 'Content-Type: application/offset+octet-stream'
```

```powershell
$OFF = (curl.exe -s -I $LOC -H 'Tus-Resumable: 1.0.0' | Select-String '^Upload-Offset: (\d+)').Matches[0].Groups[1].Value
$OFF
curl.exe -s -i -X PATCH $LOC -T big.bin -C $OFF -H 'Tus-Resumable: 1.0.0' -H "Upload-Offset: $OFF" -H 'Content-Type: application/offset+octet-stream'
```

The answer is `204` with `Upload-Offset` equal to the file's size and `Item-Path: /docs/big.bin`. The request that sends the last byte also puts the file in place, and `Item-Path` says where. With `on_conflict` set to `rename`, that can be a different name.

**4. Check it.** The downloaded file and the local file have the same checksum:

```sh
curl -s "$API/files/content?path=/docs/big.bin" | sha256sum    # on macOS: shasum -a 256
sha256sum big.bin
```

```powershell
curl.exe -s -o got.bin "$API/files/content?path=/docs/big.bin"
(Get-FileHash big.bin).Hash; (Get-FileHash got.bin).Hash
```

To cancel an unfinished upload, send `curl -s -X DELETE "$LOC" -H 'Tus-Resumable: 1.0.0'` (`204`). An upload that nobody continues is removed after `uploads.expiry` (24 hours by default).

## Errors

Every error is a JSON "problem" with a stable `code`. `-i` shows the status line and headers as well:

```sh
curl -s -i "$API/files/items?path=/nowhere"
```

```text
HTTP/1.1 404 Not Found
Content-Type: application/problem+json
X-Request-Id: 11ade03cd623fdd30f5f895f2921a52d

{"type":"about:blank","title":"Not Found","status":404,"detail":"no item at /nowhere","code":"not_found","correlation_id":"11ade03cd623fdd30f5f895f2921a52d"}
```

A name that breaks a [name rule](errors.md#name-rules) also says which rule it breaks. For example, creating the folder `/docs/aux.txt` gives `400` with `"code":"invalid_name","rule":"reserved_name"`. The `correlation_id` is the request's ID in the server log. [errors.md](errors.md) lists every code.

## Names with spaces and other characters

In the URL, a path must be percent-encoded: a space is `%20`, and `é` is `%C3%A9`. curl refuses a URL with a plain space. In a JSON body, write names as they are.

```sh
curl -s -X POST "$API/files/folders" -H 'Content-Type: application/json' -d '{"path":"/My Docs"}'
curl -s -T report.txt --url-query 'path=/My Docs/Café.txt' "$API/files/content"
curl -s -G --data-urlencode 'path=/My Docs' "$API/files/items"
curl -s -X DELETE -G --data-urlencode 'path=/My Docs/Café.txt' "$API/files/items"
```

- For an upload (`-T`), use `--url-query`, which needs curl 7.87 or later. With an older curl, encode the path yourself: `?path=/My%20Docs/Caf%C3%A9.txt`.
- `-G --data-urlencode` works for `GET` and `DELETE`.
- In Git Bash on Windows, curl receives non-ASCII characters in the wrong encoding, and the server refuses them (`invalid_utf8`). Encode them yourself there, or use PowerShell.

```powershell
Send-Json POST "$API/files/folders" @{ path = '/My Docs' }
$p = [uri]::EscapeDataString('/My Docs/Café.txt')
curl.exe -s -T report.txt "$API/files/content?path=$p"
curl.exe -s "$API/files/items?path=$([uri]::EscapeDataString('/My Docs'))"
curl.exe -s -X DELETE "$API/files/items?path=$p"
```

In a PowerShell script (`.ps1`) with non-ASCII names, save the file as UTF-8 with a byte-order mark. Otherwise Windows PowerShell 5.1 reads the names wrongly.
