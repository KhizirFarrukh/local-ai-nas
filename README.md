# local-ai-nas

> A self-hosted NAS with Google Photos–style photo management and optional, fully local AI.

**local-ai-nas** is a network-attached storage platform that anyone can deploy on their own hardware. Beyond storing and serving files, it organizes your photos, understands what's in them, and lets you search your library with natural, forgiving queries, all without sending a single byte to the cloud.

> ⚠️ **Status: early development.** This project is in its initial stage. Features described below are the planned scope and are not yet implemented.

---

## ✨ Features

### 📁 NAS Core
- Store, browse, upload, and download files over your local network
- Self-hosted on your own machine: your data never leaves your system
- Designed to be simple to deploy on any system

### 🖼️ Photo Management
- Timeline and album views, similar to Google Photos
- Automatic extraction of EXIF data (date/time, camera, GPS location)
- Each photo gets its own **sidecar JSON metadata file** stored right next to it
- Your metadata stays portable: it lives with your files, not locked inside a database

### 🤖 Optional Local AI (opt-in)
When enabled, an AI model runs **locally** alongside the NAS software. Nothing is sent to external services.

- **Auto-classification:** photos are analyzed and tagged by content (e.g. `receipt`, `document`, `food`, `landscape`, `pet`, `screenshot`). Upload a random receipt and it becomes findable by searching `receipts`, with no manual tagging needed.
- **Face detection & grouping:** detects human faces and clusters similar faces together, so you can browse all photos of the same person. A single photo can contain multiple faces, and each face is grouped independently.
- **Results are persisted:** all classifications and face groups are written into the photo's JSON metadata file. The AI runs once per photo at processing time, **not** at search time.

### 🔍 Smart Search
Search runs over stored metadata, so it is fast and doesn't require the AI model to be running.

Searchable fields:
- File name
- Description
- Place (derived from GPS data)
- Date and time
- AI-generated tags and face groups

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
| `face:` | `face:"Mom"` | Photos containing a named face group |
| `type:` | `type:video` | Filter by media type |

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
- [ ] Docker / Docker Compose
- [ ] Native install (Linux, Windows, macOS)

---

## 🗺️ Roadmap

- [ ] Core file storage and web UI
- [ ] Photo upload, EXIF extraction, sidecar JSON generation
- [ ] Timeline and album views
- [ ] Metadata search index
- [ ] Fuzzy search with synonyms and typo tolerance
- [ ] Search operators (`before:`, `after:`, `place:`, …)
- [ ] Reverse geocoding for GPS → place names (offline)
- [ ] Optional local AI: image classification
- [ ] Optional local AI: face detection and grouping
- [ ] Face group naming and merging in the UI
- [ ] User accounts and permissions

---

## 🤝 Contributing

Contributions, ideas, and feedback are welcome! Feel free to open an issue to discuss a feature or report a bug.

---

## 📄 License

License to be decided.
