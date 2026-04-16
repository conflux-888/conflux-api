# Database Schema

Database: MongoDB (Atlas recommended)
Database name: configured via `MONGODB_DATABASE` env (default: `conflux`)

## Collections

### users

Stores registered user accounts.

| Field | Type | Description |
|-------|------|-------------|
| `_id` | ObjectID | Auto-generated |
| `email` | string | Unique, indexed |
| `password_hash` | string | bcrypt hash (cost 12) |
| `display_name` | string | Display name |
| `created_at` | datetime | Account creation time |
| `updated_at` | datetime | Last update time |

**Indexes:**
| Fields | Type | Properties |
|--------|------|------------|
| `email` | single | unique |

---

### events

Stores all threat events from both GDELT sync and user reports in a single collection.

| Field | Type | Description |
|-------|------|-------------|
| `_id` | ObjectID | Auto-generated |
| `source` | string | `"gdelt"` or `"user_report"` |
| `external_id` | string | GDELT GlobalEventID (null for user reports) |
| `event_type` | string | Event type (GDELT: "Violent clash" etc., User: "armed_conflict" etc.) |
| `sub_event_type` | string | GDELT: "CAMEO 182", User: empty |
| `event_root_code` | string | CAMEO root code ("18", "19", "20") — GDELT only |
| `severity` | string | `"critical"`, `"high"`, `"medium"`, `"low"` |
| `title` | string | Short summary |
| `description` | string | GDELT: source URL / User: free-text description |
| `country` | string | Country code or name |
| `location_name` | string | Human-readable location |
| `location` | GeoJSON Point | `{ type: "Point", coordinates: [lng, lat] }` |
| `num_sources` | int | Number of distinct news sources (GDELT) |
| `num_articles` | int | Total articles mentioning event (GDELT) |
| `actors` | [string] | Involved parties |
| `event_date` | datetime | When the event occurred |
| `reported_by` | ObjectID | User who submitted (user reports only, null for GDELT) |
| `is_deleted` | bool | Soft delete flag (default: false) |
| `created_at` | datetime | When inserted into Conflux |
| `updated_at` | datetime | Last update time |

**Indexes:**
| Fields | Type | Properties |
|--------|------|------------|
| `external_id` + `source` | compound | unique, sparse |
| `location` | 2dsphere | geospatial queries |
| `severity` | single | filter |
| `event_date` | single (desc) | sort |
| `country` | single | filter |
| `reported_by` | single | sparse, user report queries |

**Notes:**
- The compound unique sparse index on `(external_id, source)` prevents duplicate GDELT events while allowing user reports (which have no external_id) to coexist.
- The 2dsphere index on `location` enables `$geoWithin` (bbox queries) and `$geoNear` (nearby queries).
- Soft deletes: queries always include `is_deleted: false` filter.

---

### report_clusters

Spatial clusters of user reports that are geographically close, share the same event type, and were reported within a 24-hour window.

| Field | Type | Description |
|-------|------|-------------|
| `_id` | ObjectID | Auto-generated |
| `event_type` | string | Shared event type across all reports in cluster |
| `severity` | string | Highest severity among all reports |
| `center` | GeoJSON Point | Weighted geographic center of all reports |
| `report_ids` | [ObjectID] | References to events in `events` collection |
| `report_count` | int | Number of reports in cluster |
| `first_reported_at` | datetime | When first report was submitted |
| `last_reported_at` | datetime | When most recent report was added |
| `created_at` | datetime | Cluster creation time |
| `updated_at` | datetime | Last update time |

**Indexes:**
| Fields | Type | Properties |
|--------|------|------------|
| `center` | 2dsphere | $geoNear queries for finding nearby clusters |

**Notes:**
- `center` is recalculated as a weighted average when new reports are added.
- `severity` escalates upward only (if a "critical" report joins a "medium" cluster, cluster becomes "critical").
- Clusters with `last_reported_at` older than 24 hours are effectively "closed" — new reports in the same area create a new cluster.

---

### sync_state

Tracks the state of the GDELT background sync process. Contains a single document.

| Field | Type | Description |
|-------|------|-------------|
| `_id` | string | Always `"gdelt"` (fixed key) |
| `last_sync_timestamp` | string | GDELT DATEADDED of most recent synced event (YYYYMMDDHHmmSS) |
| `last_sync_at` | datetime | When sync last completed |
| `status` | string | `"success"` or `"failed"` |
| `events_synced` | int | Number of events upserted in last sync cycle |
| `error_message` | string | Error details if status is "failed" |

**No custom indexes** — single document accessed by `_id`.

---

### daily_summaries

LLM-generated daily conflict briefings.

| Field | Type | Description |
|-------|------|-------------|
| `_id` | ObjectID | Auto-generated |
| `summary_date` | string | `"YYYY-MM-DD"` (UTC day) |
| `status` | string | `"completed"`, `"failed"`, `"no_events"` |
| `event_count` | int | Number of events summarized |
| `title` | string | LLM-generated news headline |
| `content` | string | LLM-generated article (continuous prose) |
| `severity_breakdown` | object | `{ critical, high, medium, low }` — computed from data |
| `model` | string | LLM model used |
| `prompt_tokens` | int | Input token count |
| `completion_tokens` | int | Output token count |
| `generation_number` | int | Re-generation counter |
| `generated_at` | datetime | When LLM was called |
| `created_at` | datetime | First creation |
| `updated_at` | datetime | Last update |
| `error_message` | string | Error details if failed |

**Indexes:**
| Fields | Type | Properties |
|--------|------|------------|
| `summary_date` | single | unique |

---

## GeoJSON Format

All geographic data uses MongoDB's GeoJSON Point format:

```json
{
  "type": "Point",
  "coordinates": [longitude, latitude]
}
```

**Important:** Coordinates are `[longitude, latitude]` (not `[lat, lng]`). This follows the GeoJSON specification and is required for MongoDB 2dsphere indexes.

## Document Relationships

```
users._id ←──── events.reported_by (user reports only)
events._id ←─── report_clusters.report_ids[] (many-to-one)
```

No foreign key enforcement — MongoDB doesn't have it. Relationships maintained at application level.
