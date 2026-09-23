# ADR-0013: Offline reverse geocoding with GeoNames

| Field | Value |
|---|---|
| Number | ADR-0013 |
| Status | **Accepted** |
| Date proposed | 2026-09-24 (session S003) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none |
| Superseded by | none |

## Context

S05.4 converts GPS coordinates into place names (city, region, country) **offline**: FR-062, FR-063, I6. The `place:` operator (S06.3) matches any place level.

## Options considered

| Option | Pros | Cons |
|---|---|---|
| **GeoNames cities dataset + in-memory nearest-city lookup** (chosen) | Offline; small; attribution license (CC BY 4.0); includes admin and country tables and alternate names | Nearest-populated-place, not boundary-accurate |
| Natural Earth / OSM boundary polygons | Boundary-accurate country and region | Larger; polygon lookup code; weaker at city level |
| Online geocoder (Nominatim, etc.) | Accurate | Network calls at runtime, which violates I6 |

## Decision

- **Dataset:** the **GeoNames** cities dataset, license **Creative Commons Attribution 4.0** (verified in `download.geonames.org/export/dump/readme.txt`). **Attribution is required** in the documentation and on an **About page** in the web UI.
- **Lookup:** an in-memory nearest-city lookup over a **k-d tree**.
- **Region and country names:** `admin1CodesASCII.txt` (0.2 MB) and `countryInfo.txt`.
- The dataset is **bundled at build/package time** (Docker image; native install package). **It is never downloaded at runtime** (I6). The dataset version (GeoNames export date) is recorded in each sidecar's location section.

### Implementation details chosen by agent
| Detail | Choice | Trade-off recorded |
|---|---|---|
| Granularity | **`cities500`**: "all cities with a population > 500 or seats of adm div down to PPLA4 (ca 185.000)", 13.9 MB zipped (verified) | Denser rural coverage (relevant for the user's region) for +2.8 MB compared with `cities1000` (11.1 MB). `cities5000` (5.7 MB) is too coarse for village-level photos. In-memory cost is estimated at tens of MB, to be measured in S05.4, and the lookup structure can be loaded lazily only while geocoding jobs run |
| k-d tree | **Small custom implementation** (2-D on unit-sphere coordinates, or lat/lon with a haversine check), about 150 lines, no dependency | Trivial, well understood, avoids a low-activity dependency |
| Alternate names | Use the GeoNames `alternatenames` column from the cities file for `place:` matching (e.g. Bombay → Mumbai). The large `alternateNames` dump is not bundled | Keeps size small (FR-063) |
| Oceans and remote points | Return country-only or "unknown" beyond a distance threshold (e.g. 50 km), and never invent a city | Avoids misleading places |

## Consequences

- **Easier:** fully offline place names; small footprint; simple code.
- **Harder:** accuracy near borders (documented limitation; Natural Earth polygons remain a later option); keeping the dataset fresh (updated with releases).
- **Required:** attribution text (docs + About page, S05.4 and S02 or later GUI); fixture coordinates tests (S05.4); dataset version in sidecars.

## Approval record

> Decision made by the user in plan change request #3 (`code-agent-docs/prompts/P003-technology-stack.json`, `technology_stack.reverse_geocoding`). Granularity and name tables are "agent decides" items
> (2026-09-24, session S003). License and file sizes verified in S003 log E005.

## Links

- **Related requirements:** FR-062, FR-063, NFR-001
- **Related ADRs:** ADR-0009 (About page), ADR-0014 (`place:` search)
- **Related stages:** S05.4, S06.3
- **Plan version:** 0.3.0
