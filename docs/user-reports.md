# User Report System

## Overview

Users can submit their own threat event reports through the API. Reports are stored alongside GDELT events in the same `events` collection (with `source: "user_report"`) and are automatically clustered with nearby reports to establish credibility.

## Why User Reports Matter

GDELT data is machine-coded from news articles — it has coverage gaps and false positives. User reports add **ground-truth data** from people on the ground who witness events firsthand.

| | GDELT Events | User Reports |
|---|---|---|
| Source | News articles (machine-coded) | Eyewitness (human-submitted) |
| Delay | ~15 minutes from article publication | Immediate |
| Description | Source URL only | Free-text description |
| Accuracy | Noisy (false positives) | Unverified (could be false) |
| Credibility | Single source per event | Increases with cluster report_count |

## Event Types

User report event types are designed to align with GDELT's conflict categories (root codes 18-20):

| Event Type | Description | Comparable GDELT Code |
|------------|-------------|----------------------|
| `armed_conflict` | Armed clash, gunfire, warfare | 18 (Violent clash), 20 (Military force) |
| `use_of_force` | Police/military force, blockade, raid | 19 (Use of force) |
| `explosion` | Bombing, IED, explosion | — |
| `terrorism` | Terrorist attack | — |
| `civil_unrest` | Violent protest, riot | — |
| `other` | Other threat events | — |

Validated via Gin binding tag: `binding:"required,oneof=armed_conflict use_of_force explosion terrorism civil_unrest other"`

## Submission Flow

```
1. User sends POST /reports with event details + coordinates
2. Handler validates request (event_type, severity, title, lat/lng, country)
3. Service creates Event in events collection:
   - source = "user_report"
   - reported_by = authenticated user's ID
   - event_date = now
4. Service runs clustering logic:
   a. Search report_clusters for nearby cluster
      (same event_type, within 5km, within last 24hr)
   b. If found → add report to cluster
   c. If not found → create new cluster
5. Return created event to user
```

## Clustering System

### Purpose

Clustering answers: "How many people reported the same thing in the same area?"

A single user report has low credibility. But when 15 people independently report "explosion" at the same location within 24 hours, it's very likely real. The `report_count` field enables the UI to:
- Display larger points for high-count clusters
- Prioritize verified reports over single reports
- Show "X people reported this"

### Clustering Criteria

| Criteria | Value | Reason |
|----------|-------|--------|
| Max distance | 5 km | Same neighborhood/area |
| Time window | 24 hours | Same ongoing incident |
| Event type | Must match | Different types = different events |

### Cluster Operations

**New report comes in:**

```
Query report_clusters with $geoNear:
  - center within 5000 meters of report coordinates
  - event_type matches
  - last_reported_at >= 24 hours ago
  - Limit 1 (nearest)

If found → AddToCluster:
  - Push event._id into report_ids array
  - report_count++
  - Recalculate center (weighted average):
    new_lng = (old_lng * count + report_lng) / (count + 1)
    new_lat = (old_lat * count + report_lat) / (count + 1)
  - Update severity to max(existing, new report's severity)
  - Update last_reported_at = now

If not found → CreateCluster:
  - center = report coordinates
  - report_ids = [event._id]
  - report_count = 1
  - severity = report's severity
  - first_reported_at = now
  - last_reported_at = now
```

### Severity Escalation

When a new report is added to a cluster, the cluster's severity is updated to the **higher** of the existing and new severity:

```
Rank: low (0) < medium (1) < high (2) < critical (3)

Cluster severity = "medium"
New report severity = "critical"
→ Cluster severity becomes "critical"
```

Implementation: `report.HigherSeverity(a, b string) string`

### Cluster Data Model

MongoDB collection: `report_clusters`

| Field | Type | Description |
|-------|------|-------------|
| `_id` | ObjectID | Auto-generated |
| `event_type` | string | Shared event type |
| `severity` | string | Highest severity in cluster |
| `center` | GeoJSON Point | Weighted geographic center |
| `report_ids` | []ObjectID | References to events in `events` collection |
| `report_count` | int | Number of reports |
| `first_reported_at` | datetime | When first report was submitted |
| `last_reported_at` | datetime | When most recent report was added |
| `created_at` | datetime | Cluster creation time |
| `updated_at` | datetime | Last update time |

Index: `center` (2dsphere) for `$geoNear` queries.

### Cluster vs Client-Side Clustering

| | Server-side (report_clusters) | Client-side (map library) |
|---|---|---|
| Purpose | Credibility weighting (report_count) | Visual grouping (reduce map clutter) |
| When computed | On write (pre-computed) | On render (real-time) |
| Data | report_count, severity escalation | Just coordinates |
| Performance | O(1) read | Recomputed every zoom/pan |

Both serve different purposes and can coexist.

## Ownership & Deletion

- Only the report creator can delete their own report
- Delete is **soft delete** (`is_deleted: true`)
- Ownership checked via `reported_by` field matching authenticated user's ID
- Attempting to delete another user's report returns 404 (not 403, to avoid leaking existence)

## Key Files

| File | Purpose |
|------|---------|
| `internal/report/model.go` | CreateReportRequest, ReportCluster, HigherSeverity |
| `internal/report/repository.go` | Cluster CRUD with $geoNear |
| `internal/report/service.go` | SubmitReport + clustering, GetMyReports, DeleteMyReport |
| `internal/report/handler.go` | HTTP handlers |
| `internal/report/routes.go` | Route registration |
