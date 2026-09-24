# local-ai-nas

> A self-hosted NAS with Google Photos–style photo management and optional, fully local AI.

**local-ai-nas** is a network-attached storage platform that anyone can deploy on their own hardware. It keeps two separate areas: **Files**, a general-purpose file store, and **Photos**, a Google Photos–style library. It organizes your photos, understands what's in them, and lets you search both areas with natural, forgiving queries, all without sending a single byte to the cloud.

> ⚠️ **Status: early development.** This project is in its initial stage. Features described below are the planned scope and are not yet implemented.

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

> Installation instructions will be added once the first release is ready.

Planned deployment options:
- [ ] Docker Compose (primary), with images for x86-64 and ARM64 (e.g. Raspberry Pi); optional AI component enabled with a Compose profile
- [ ] Native install on Linux (single binary + systemd service)
- [ ] Native install on Windows and macOS (under consideration)

---

## 🗺️ Roadmap

- [ ] Stage 1: Basic NAS (storage service and API; Files and Photos areas)
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
