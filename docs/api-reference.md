# API Reference

Base URL: `http://localhost:8080/api/v1`

## Response Format

All responses follow a consistent format.

**Success (single):**
```json
{ "data": { ... } }
```

**Success (list):**
```json
{
  "data": [ ... ],
  "pagination": { "page": 1, "limit": 50, "total": 1234 }
}
```

**Error:**
```json
{
  "error": { "code": "VALIDATION_ERROR", "message": "email is required" }
}
```

**Error codes:**

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `VALIDATION_ERROR` | Invalid request body/params |
| 401 | `UNAUTHORIZED` | Missing/invalid/expired token |
| 404 | `NOT_FOUND` | Resource not found |
| 409 | `CONFLICT` | Duplicate (e.g., email taken) |
| 500 | `INTERNAL_ERROR` | Server error |

---

## Health Check

### GET /health (outside /api/v1)

No auth required. URL: `http://localhost:8080/health`

**Response:**
```json
{ "status": "ok" }
```

### Swagger UI (outside /api/v1)

URL: `http://localhost:8080/swagger/index.html`

---

## Authentication

### POST /api/v1/auth/register

Create a new user account.

**Request body:**
```json
{
  "email": "user@example.com",
  "password": "minimum8chars",
  "display_name": "John Doe"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| email | string | yes | valid email, unique |
| password | string | yes | min 8 characters |
| display_name | string | yes | |

**Response (201):**
```json
{
  "data": {
    "id": "665f1a2b3c4d5e6f7a8b9c0d",
    "email": "user@example.com",
    "display_name": "John Doe",
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  }
}
```

**Errors:** 409 (email taken), 400 (validation)

---

### POST /api/v1/auth/login

Authenticate and receive a JWT access token.

**Request body:**
```json
{
  "email": "user@example.com",
  "password": "minimum8chars"
}
```

**Response (200):**
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 3600
  }
}
```

Token is valid for **1 hour**. No refresh token — login again when expired.

**Errors:** 401 (invalid email or password)

---

## User Profile

All endpoints require `Authorization: Bearer <token>` header.

### GET /api/v1/users/me

**Response (200):**
```json
{
  "data": {
    "id": "665f1a2b3c4d5e6f7a8b9c0d",
    "email": "user@example.com",
    "display_name": "John Doe",
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  }
}
```

### PUT /api/v1/users/me

**Request body:**
```json
{
  "display_name": "Jane Doe"
}
```

**Response (200):** Same format as GET /api/v1/users/me with updated fields.

---

## Events

All endpoints require `Authorization: Bearer <token>` header.

### GET /api/v1/events

List events with filtering, sorting, and pagination. Returns both GDELT and user report events.

**Query parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| severity | string | | Comma-separated: `critical,high,medium,low` |
| event_type | string | | Exact match on event type |
| country | string | | Country code (FIPS 2-char for GDELT, free text for user reports) |
| source | string | | `gdelt` or `user_report` |
| date_from | string | | `YYYY-MM-DD`, event_date >= |
| date_to | string | | `YYYY-MM-DD`, event_date <= |
| bbox | string | | Bounding box: `min_lng,min_lat,max_lng,max_lat` |
| page | int | 1 | Page number |
| limit | int | 50 | Results per page (max 200) |
| sort | string | date_desc | `date_desc`, `date_asc`, `severity` |

**Example requests:**
```
GET /api/v1/events?severity=critical,high&limit=100
GET /api/v1/events?bbox=100.3,13.5,100.8,14.0&source=gdelt
GET /api/v1/events?country=IR&date_from=2025-01-01&sort=date_desc
```

**Response (200):**
```json
{
  "data": [
    {
      "id": "69d11b89997825febc88f036",
      "source": "gdelt",
      "external_id": "1297630653",
      "event_type": "Violent clash",
      "sub_event_type": "CAMEO 182",
      "event_root_code": "18",
      "severity": "critical",
      "title": "Violent clash in Mashhad, (IR30), Iran",
      "description": "https://www.independent.co.uk/...",
      "country": "IR",
      "location_name": "Mashhad, (IR30), Iran",
      "location": {
        "type": "Point",
        "coordinates": [59.6062, 36.297]
      },
      "num_sources": 1,
      "num_articles": 3,
      "actors": ["MASHHAD"],
      "event_date": "2025-04-04T00:00:00Z",
      "is_deleted": false,
      "created_at": "2026-04-04T14:09:13.321Z",
      "updated_at": "2026-04-04T14:17:12.383Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 84
  }
}
```

---

### GET /api/v1/events/:id

Get a single event by ID.

**Response (200):** Single event object in `data` field.

**Errors:** 404 (event not found)

---

### GET /api/v1/events/nearby

Find events near a geographic point.

**Query parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| lat | float | | **Required.** Latitude |
| lng | float | | **Required.** Longitude |
| radius_km | float | 50 | Search radius in km (max 500) |
| severity | string | | Filter by severity level |
| limit | int | 20 | Max results (max 100) |

**Example:**
```
GET /api/v1/events/nearby?lat=36.3&lng=59.6&radius_km=10&severity=critical
```

**Response (200):** Array of events in `data` field (no pagination, limited by `limit` param).

**Errors:** 400 (missing lat/lng)

---

## User Reports

All endpoints require `Authorization: Bearer <token>` header.

### POST /api/v1/reports

Submit a user report. Creates an event with `source: "user_report"` and auto-clusters with nearby reports.

**Request body:**
```json
{
  "event_type": "armed_conflict",
  "severity": "high",
  "title": "Gunfire near border checkpoint",
  "description": "Heard multiple rounds of gunfire from the south...",
  "latitude": 13.7563,
  "longitude": 100.5018,
  "location_name": "Bangkok",
  "country": "TH"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| event_type | string | yes | One of: `armed_conflict`, `use_of_force`, `explosion`, `terrorism`, `civil_unrest`, `other` |
| severity | string | yes | One of: `critical`, `high`, `medium`, `low` |
| title | string | yes | Max 200 characters |
| description | string | no | Max 2000 characters |
| latitude | float | yes | |
| longitude | float | yes | |
| location_name | string | no | |
| country | string | yes | |

**Response (201):** Created event object in `data` field.

**Clustering behavior:**
After creating the event, the system checks for an existing report cluster within 5km radius, same event_type, and within the last 24 hours. If found, the report is added to the cluster. Otherwise, a new cluster is created.

---

### GET /api/v1/reports/me

List the authenticated user's own reports.

**Query parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| page | int | 1 | Page number |
| limit | int | 20 | Results per page (max 100) |

**Response (200):** Paginated list of events.

---

### DELETE /api/v1/reports/:id

Soft-delete a report. Only the report owner can delete it.

**Response (200):**
```json
{ "data": { "message": "report deleted" } }
```

**Errors:** 404 (not found or not owner)

---

## Admin — Sync

All endpoints require `Authorization: Bearer <token>` header.

### GET /api/v1/admin/sync/status

Get the current GDELT sync state.

**Response (200):**
```json
{
  "data": {
    "id": "gdelt",
    "last_sync_timestamp": "20260404141500",
    "last_sync_at": "2026-04-04T14:15:02Z",
    "status": "success",
    "events_synced": 42,
    "error_message": ""
  }
}
```

### POST /api/v1/admin/sync/trigger

Manually trigger a GDELT sync cycle. Runs synchronously — returns after sync completes.

**Response (200):** Same format as GET /api/v1/admin/sync/status with updated values.
