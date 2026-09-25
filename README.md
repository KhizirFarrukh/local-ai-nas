# local-ai-nas

> A self-hosted NAS with Google Photos–style photo management and optional, fully local AI.

**local-ai-nas** is a network-attached storage platform that anyone can deploy on their own hardware. It keeps two separate areas: **Files**, a general-purpose file store, and **Photos**, a Google Photos–style library. It organizes your photos, understands what's in them, and lets you search both areas with natural, forgiving queries, all without sending a single byte to the cloud.

> ⚠️ **Status: early development.** Stage 1, the NAS core, is complete: the Files area can be managed through a REST API on the same computer (see Development). Everything else below is the planned scope and is not implemented yet.

---

## ✨ Features

### 📁 NAS Core
- Store, browse, upload, and download files over your local network, with resumable uploads for large files
- Two separate areas, **Files** and **Photos**, that never mix: content moves between them only when you explicitly copy or move it
- Multiple users, each with private files and photos, and explicit sharing between users (planned)
- Self-hosted on your own machine: your data never leaves your system
- Designed to be simple to deploy on any system

### 🖼️ Photo Management
- Timeline and album views, similar to Google Photos
- Video playback streamed to any device, with a live quality menu (Auto, Original, 1080p, 720p, …); lower qualities are prepared on demand
- Automatic extraction of EXIF data (date/time, camera, GPS location)
- Each photo in the Photos area gets its own **sidecar JSON metadata file** stored right next to it
- Your metadata stays portable: it lives with your files, not locked inside a database

### 🤖 Optional Local AI (opt-in)
When enabled, AI runs **locally** in a separate, optional component alongside the NAS software. It has read-only access to your photos, and the NAS itself writes all results. Nothing is sent to external services, and only models whose licenses let anyone use them are included.

- **Auto-classification:** photos are analyzed and tagged by content (e.g. `receipt`, `document`, `food`, `landscape`, `pet`, `screenshot`). Upload a random receipt and it becomes findable by searching `receipts`, with no manual tagging needed.
- **Face detection & grouping:** detects human faces and clusters similar faces together, so you can browse all photos of the same person. A single photo can contain multiple faces, and each face is grouped independently.
- **Results are persisted:** all classifications and face groups are written into the photo's JSON metadata file. The AI runs once per photo at processing time, **not** at search time.

### 🔍 Smart Search
Search covers both the Files and Photos areas. It runs over stored metadata and a local search index, so it is fast and doesn't require the AI model to be running.

Searchable fields:
- File name (both areas)
- Description
- Place (derived from GPS data)
- Date and time
- AI-generated tags and face groups (when AI is enabled)

Search is **fuzzy and meaning-aware**, not exact-match:
- **Plurals / word forms:** `receipts` → `receipt`
- **Synonyms / related terms:** `receipts` → `invoice`, `voucher`, `bill`
- **Typo tolerance:** `reciept` → `receipt`

#### Search operators

| Operator | Example | Description |
|---|---|---|
| `before:` | `before:2026` | Photos taken before a date |
| `after:` | `after:2024-06` | Photos taken after a date |
| `on:` | `on:2025-12-25` | Photos taken on a specific date |
| `place:` | `place:karachi` | Photos taken at a location |
| `tag:` | `tag:receipt` | Photos with a specific AI/user tag |
| `face:` | `face:"Mom"` | Photos containing a named face group (available when AI is enabled) |
| `type:` | `type:video` | Filter by media type |
| `in:` | `in:photos` | Limit results to the Files or Photos area |
| `ext:` | `ext:pdf` | Filter by file extension |
| `size:` | `size:>10MB` | Filter by file size |

Operators can be combined with free text:

```
receipts after:2025 place:lahore
```

> The operator list above is the planned syntax and may change during development.

---

## 🗂️ Sidecar Metadata

Every photo has a companion JSON file stored alongside it. The full original filename is kept in the sidecar name to avoid collisions (e.g. `IMG_0001.jpg` and `IMG_0001.png`).

```
photos/
├── IMG_0001.jpg
├── IMG_0001.jpg.json
├── IMG_0002.png
└── IMG_0002.png.json
```

Example `IMG_0001.jpg.json` (draft schema):

```json
{
  "schemaVersion": 1,
  "file": {
    "name": "IMG_0001.jpg",
    "size": 2483921,
    "hash": "sha256:…",
    "mimeType": "image/jpeg"
  },
  "description": "Lunch with the team",
  "takenAt": "2025-11-14T13:42:10+05:00",
  "location": {
    "lat": 24.8607,
    "lon": 67.0011,
    "place": "Karachi, Sindh, Pakistan"
  },
  "camera": {
    "make": "Google",
    "model": "Pixel 8"
  },
  "ai": {
    "processed": true,
    "model": "<model-name>@<version>",
    "processedAt": "2025-11-15T02:10:00Z",
    "tags": [
      { "label": "receipt", "confidence": 0.94 },
      { "label": "document", "confidence": 0.81 }
    ],
    "faces": [
      {
        "groupId": "face_7f3a",
        "groupName": "Mom",
        "box": { "x": 0.21, "y": 0.18, "w": 0.12, "h": 0.16 }
      }
    ]
  },
  "userTags": ["tax-2025"]
}
```

The sidecar files are the **source of truth**. Any search index built by the application is a cache that can be rebuilt from them at any time.

---

## 🔒 Privacy Principles

- **Local-first:** all storage and AI processing happen on your own hardware
- **AI is opt-in:** the NAS works fully without it
- **No telemetry, no cloud dependencies**
- **Portable data:** metadata lives next to your files in an open, readable format

---

## 🚀 Getting Started

> Installation instructions will be added once the first release is ready. To run the current development version from source, see the Development section below.

Planned deployment options:
- [ ] Docker Compose (primary), with images for x86-64 and ARM64 (e.g. Raspberry Pi); optional AI component enabled with a Compose profile
- [ ] Native install on Linux (single binary + systemd service)
- [ ] Native install on Windows and macOS (under consideration)

---

## 🛠️ Development

The project is in its first stage: the NAS core, a REST API on this computer only, with no web interface or login yet. To run it from source:

**You need** Git and Go 1.27 or newer. The exact Go version the project pins (go1.27.1) is downloaded automatically on the first build. Docker is optional.

### Run it on your computer (Linux, macOS, or Windows)

```sh
git clone https://github.com/KhizirFarrukh/local-ai-nas.git
cd local-ai-nas
go run ./cmd/local-ai-nas serve --storage-root "$PWD/dev/data"
```

The same command works in PowerShell. The server creates the folder, its database, and its log file (`dev/data/.local-ai-nas/`), and listens on `http://127.0.0.1:8080`. Check it from a second terminal:

```sh
curl http://127.0.0.1:8080/api/v1/system/health
# {"status":"ok","version":"dev","checks":[{"name":"database","status":"ok"}]}
```

The API documentation is at `http://127.0.0.1:8080/api/docs/` (it works offline). Stop the server with Ctrl+C. The `dev/` folder is ignored by Git.

To use a config file instead of flags, copy [`deploy/config.example.toml`](deploy/config.example.toml) to `dev/config.toml`, set `storage.root` to an absolute path, and run `go run ./cmd/local-ai-nas serve --config dev/config.toml`. Every setting can also come from an environment variable or a flag (`storage.root` = `LOCALAINAS_STORAGE_ROOT` = `--storage-root`); flags win over environment variables, which win over the file. `go run ./cmd/local-ai-nas serve -h` lists all flags.

Other commands: `migrate up` and `migrate status` (database migrations; `serve` also migrates on start) and `version`.

### The web interface

The server includes the web interface (stage 2, being built) when the interface is built first. For that you need Node.js 24 LTS and pnpm 12.6.0 (`npm install -g pnpm@12.6.0`):

```sh
cd web
pnpm install
pnpm build
cd ..
go run ./cmd/local-ai-nas serve --storage-root "$PWD/dev/data"
```

Then open `http://127.0.0.1:8080/` in a browser on this computer. Without the build, the server shows a short notice there, and the API works as before.

To work on the interface, run the server as above, and in a second terminal run `cd web` and `pnpm dev`. The development server at `http://localhost:5173/` reloads on every change and passes `/api` requests on to the server.

### Use the API

With the server running, you manage the Files area over HTTP. For example, from a second terminal on Linux or macOS:

```sh
API=http://127.0.0.1:8080/api/v1
curl -s -X POST "$API/files/folders" -H 'Content-Type: application/json' -d '{"path":"/docs"}'
curl -s -T README.md "$API/files/content?path=/docs/README.md"
curl -s "$API/files/items?path=/docs"
```

The guide [docs/api/usage.md](docs/api/usage.md) goes through every operation with curl, for Linux and macOS and for Windows PowerShell:
- creating folders, uploading, and listing;
- downloading, including ranges;
- renaming, moving, copying, and deleting;
- resumable (tus) uploads, including a resume after a cut connection;
- errors.

In PowerShell, use `curl.exe`, and send JSON bodies with the guide's small helper. To run all of it at once, with every answer checked, use `scripts/demo.sh` or, in PowerShell, `powershell -ExecutionPolicy Bypass -File scripts\demo.ps1`. The rules every endpoint follows are in [docs/api/conventions.md](docs/api/conventions.md), and the running server serves the full reference at `http://127.0.0.1:8080/api/docs/`.

### Run it in Docker

```sh
docker compose -f deploy/compose.dev.yaml up --build
```

The container runs your working copy with `go run` and keeps its data in a Docker volume. It is reachable at `http://127.0.0.1:8080` from this computer only. Inside the container the server listens on all of the container's interfaces (`LOCALAINAS_SERVER_ALLOW_CONTAINER_BIND`, set in the compose file); on its own the server accepts only loopback addresses for `server.bind` until login and HTTPS exist (stage 3). After a code change: `docker compose -f deploy/compose.dev.yaml restart`.

### Tests and checks

| What | Command |
|---|---|
| Tests | `go test ./...` |
| Coverage (minimum 80%) | `scripts/coverage.sh` |
| Lint and format | `scripts/install-golangci-lint.sh` once, then `./bin/golangci-lint run ./...` and `./bin/golangci-lint fmt --diff` |
| Dependency licenses | `scripts/check-licenses.sh` |
| Performance baseline | `scripts/perf-baseline.sh` (results in [docs/perf/](docs/perf/)) |
| Web interface: format, lint, type checks | in `web/`: `pnpm format:check`, `pnpm lint`, `pnpm check` |
| Web interface licenses | `node scripts/check-web-licenses.mjs` |

On Windows, run the `scripts/*.sh` files from Git Bash. More in [docs/testing.md](docs/testing.md). CI runs all of these except the performance baseline for every push, on Linux and Windows (the web interface checks on Linux).

---

## 🗺️ Roadmap

- [x] Stage 1: Basic NAS (storage service and API; Files and Photos areas)
- [ ] Stage 2: Web interface for the NAS
- [ ] Stage 3: Security (login, HTTPS, hardening)
- [ ] Stage 4: Media management (Photos area: timeline, albums, viewer)
- [ ] Stage 5: Media metadata (EXIF, sidecar JSON, offline place names)
- [ ] Stage 6: Search across Files and Photos (fuzzy, synonyms, operators)
- [ ] Stage 7: Multiple users and sharing
- [ ] Final stage: optional local AI (auto-classification, face grouping)

---

## 🤝 Contributing

Contributions, ideas, and feedback are welcome! Feel free to open an issue to discuss a feature or report a bug.

---

## 📄 License

local-ai-nas is free software, licensed under the **GNU Affero General Public License v3.0 or later** (`AGPL-3.0-or-later`). See [LICENSE](LICENSE).

If you run a modified version on a server that other people use over a network, the AGPL requires you to offer them its source code. Third-party components keep their own licenses. The policy is in [docs/licensing.md](docs/licensing.md).
