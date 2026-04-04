# GDELT Sync System

## Overview

The sync module downloads conflict event data from the GDELT 2.0 project every 15 minutes. GDELT monitors news worldwide and extracts structured event records in near-realtime. No API key or authentication required.

## Data Source

| Item | Detail |
|------|--------|
| Source | GDELT 2.0 Event Database |
| URL | `http://data.gdeltproject.org/gdeltv2/lastupdate.txt` |
| Format | Tab-separated CSV in ZIP archive |
| Update frequency | Every 15 minutes |
| Auth | None required |
| Rate limit | None for file downloads |
| Events per batch | ~500-1000 per 15-minute window |

## Sync Flow

```
Every 15 minutes (+ immediately on startup):

1. GET lastupdate.txt
   → Parse to find .export.CSV.zip URL

2. Download ZIP
   → Unzip in memory

3. Parse tab-separated CSV (61 fields per row)
   → Create []GDELTEvent structs

4. Filter:
   → QuadClass == "4" (Material Conflict only)
   → EventRootCode >= "18" (violent clash, use of force, military force)

5. Map to domain Event structs
   → Classify severity from GoldsteinScale
   → Generate title from CAMEO description + location
   → Set source = "gdelt"

6. BulkUpsert to MongoDB events collection
   → Upsert by external_id + source (idempotent)

7. Update sync_state document
   → last_sync_timestamp = max DateAdded seen
   → status = "success" / "failed"
```

## GDELT Event Filtering

### QuadClass Filter

GDELT classifies events into 4 quad classes:
1. Verbal Cooperation
2. Material Cooperation
3. Verbal Conflict
4. **Material Conflict** (we only keep this)

### EventRootCode Filter (CAMEO codes)

| Root Code | Description | Included? |
|-----------|-------------|-----------|
| 01-13 | Cooperation, diplomacy, aid | No |
| 14 | Military posture | No (too broad, noisy) |
| 15 | Conventional attack | No (threats, not actual) |
| 16 | Unconventional mass violence | No (threats) |
| 17 | Riotous forces | No (noisy — includes graffiti, protests) |
| **18** | **Violent clash** | **Yes** — physical assault, armed clash, killing |
| **19** | **Use of force** | **Yes** — blockade, occupation, use of force |
| **20** | **Military force** | **Yes** — warfare, military operations |

Result: from ~600 total events per 15-minute batch, ~40-80 pass both filters.

## Severity Classification

Based on the GDELT GoldsteinScale (-10.0 to +10.0), which measures theoretical impact on country stability:

| GoldsteinScale | Severity | Meaning |
|----------------|----------|---------|
| <= -7.0 | `critical` | Mass violence, warfare, WMD |
| -7.0 to -5.0 | `high` | Conventional attacks, armed clashes |
| -5.0 to -2.0 | `medium` | Military posture, blockades |
| > -2.0 | `low` | Threats, minor incidents |

Implementation: `internal/sync/severity.go` → `ClassifySeverity(goldsteinScale float64) string`

## GDELT CSV Field Mapping

The export CSV has 61 tab-separated fields. We parse these (0-indexed):

| Index | GDELT Field | Maps to Event field |
|-------|-------------|-------------------|
| 0 | GlobalEventID | `external_id` |
| 1 | SQLDATE (YYYYMMDD) | `event_date` |
| 6 | Actor1Name | `actors[0]` |
| 16 | Actor2Name | `actors[1]` |
| 25 | IsRootEvent | (used for filtering) |
| 26 | EventCode | `sub_event_type` ("CAMEO " + code) |
| 28 | EventRootCode | `event_root_code` |
| 29 | QuadClass | (used for filtering, must be "4") |
| 30 | GoldsteinScale | `severity` (via ClassifySeverity) |
| 31 | NumMentions | (parsed, not stored currently) |
| 32 | NumSources | `num_sources` |
| 33 | NumArticles | `num_articles` |
| 34 | AvgTone | (parsed, not stored currently) |
| 51 | ActionGeo_Type | (parsed) |
| 52 | ActionGeo_FullName | `location_name` |
| 53 | ActionGeo_CountryCode | `country` (FIPS 2-char) |
| 56 | ActionGeo_Lat | `location.coordinates[1]` |
| 57 | ActionGeo_Long | `location.coordinates[0]` |
| 59 | DATEADDED | (used for sync timestamp tracking) |
| 60 | SOURCEURL | `description` (news article link) |

## Event Title Generation

Title is constructed from CAMEO root code description + location:

```
cameoDescriptions[EventRootCode] + " in " + ActionGeoFullName
```

Examples:
- "Violent clash in Mashhad, (IR30), Iran"
- "Use of force in Pahalgam, Jammu and Kashmir, India"
- "Military force in Kyiv, Kyyivs'ka, Ukraine"

## Sync State

Stored in MongoDB `sync_state` collection with fixed `_id: "gdelt"`:

| Field | Type | Description |
|-------|------|-------------|
| `_id` | string | Always `"gdelt"` |
| `last_sync_timestamp` | string | GDELT DATEADDED (YYYYMMDDHHmmSS) of most recent event |
| `last_sync_at` | datetime | When sync last completed |
| `status` | string | `"success"` or `"failed"` |
| `events_synced` | int | Count from last sync cycle |
| `error_message` | string | Error details if failed |

## Error Handling

- **Download failure**: Sync cycle marked as `"failed"`, retries on next tick
- **Parse errors**: Individual events with invalid fields are skipped (logged as warn), cycle continues
- **MongoDB errors**: BulkUpsert failure marks cycle as `"failed"`, retries on next tick
- **No events**: Not an error — logged as info, state updated normally

## Admin Endpoints

- `GET /admin/sync/status` — view current sync state
- `POST /admin/sync/trigger` — manually trigger a sync cycle (runs synchronously)

## Key Files

| File | Purpose |
|------|---------|
| `internal/sync/client.go` | GDELT HTTP client, CSV download + parse |
| `internal/sync/model.go` | SyncState + GDELTEvent structs |
| `internal/sync/service.go` | Sync orchestrator, ticker loop, event mapping |
| `internal/sync/severity.go` | GoldsteinScale → severity classification |
| `internal/sync/state_repository.go` | sync_state MongoDB CRUD |
| `internal/sync/handler.go` | Admin HTTP handlers |
| `internal/sync/routes.go` | Route registration |

## GDELT Limitations

- **Machine-coded**: Events are extracted by NLP from news articles — false positives occur (e.g., graffiti classified as "riot")
- **No description**: GDELT provides source URL but no human-written summary
- **No fatalities count**: Unlike ACLED, GDELT has no fatalities field
- **Country codes**: Uses FIPS 2-character codes (e.g., "IR" for Iran), not ISO 3166
- **Duplicates possible**: Same real-world event may appear multiple times from different source articles
