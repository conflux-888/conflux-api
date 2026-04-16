# Notification System — Frontend Integration Spec

## Business Overview

Conflux API adds an **In-App Notification System** using a **polling pattern** (not push), because the iOS app currently runs on Simulator only where FCM/APNs is unreliable.

### Use Cases

| Type | Trigger | When |
|------|---------|------|
| `critical_nearby` | GDELT sync finds a new critical event within the user's radius | Every 15 minutes |
| `daily_briefing` | Daily LLM summary finishes generating | Every 30 minutes (scheduler) |

**Backend logic:**
- `critical_nearby` — filters by user's `last_location` + `radius_km` + `min_severity`
- `daily_briefing` — broadcast to every user with `notifications_enabled = true`
- Deduplication — no duplicate notifications for the same (user, event) pair
- Auto-cleanup — notifications older than 30 days are deleted automatically (TTL index)

---

## Data Models

### `Notification`
```json
{
  "id": "670a2f...",
  "user_id": "665f1a...",
  "type": "critical_nearby" | "daily_briefing",
  "title": "Critical threat 3.2km from you",
  "body": "Military force in Chatuchak, Bangkok",
  "event_id": "69d1...",
  "summary_date": "2026-04-17",
  "distance_km": 3.2,
  "read_at": "2026-04-17T08:30:00Z",
  "created_at": "2026-04-17T08:00:00Z"
}
```

| Field | Populated when | Description |
|-------|---------------|-------------|
| `event_id` | type = `critical_nearby` | Used to navigate to event detail |
| `summary_date` | type = `daily_briefing` | Used to navigate to the daily summary |
| `distance_km` | type = `critical_nearby` | Distance of the event from the user's location |
| `read_at` | After marking as read | `null` if unread |

### `UserPreferences`
```json
{
  "id": "670a...",
  "user_id": "665f1a...",
  "notifications_enabled": true,
  "min_severity": "critical",
  "radius_km": 50,
  "last_location": {
    "type": "Point",
    "coordinates": [100.5018, 13.7563]
  },
  "last_location_at": "2026-04-17T08:00:00Z",
  "created_at": "2026-04-17T00:00:00Z",
  "updated_at": "2026-04-17T08:00:00Z"
}
```

**Defaults for new users:**
- `notifications_enabled: true`
- `min_severity: "critical"`
- `radius_km: 50`
- `last_location: null` (until the iOS client updates it)

---

## API Endpoints

Base URL: `http://host:port/api/v1`
All endpoints require `Authorization: Bearer <token>`

### Preferences

#### `GET /preferences`
Returns the user's preferences (auto-creates defaults if not yet set).

**Response 200:**
```json
{
  "data": {
    "notifications_enabled": true,
    "min_severity": "critical",
    "radius_km": 50,
    "last_location": null,
    "last_location_at": null
  }
}
```

#### `PUT /preferences`
Updates preferences (partial — send only the fields you want to change).

**Request body (all fields optional):**
```json
{
  "notifications_enabled": false,
  "min_severity": "high",
  "radius_km": 100
}
```

| Field | Type | Validation |
|-------|------|------------|
| `notifications_enabled` | bool | - |
| `min_severity` | string | `critical`, `high`, `medium`, `low` |
| `radius_km` | float | 1-500 |

**Response 200:** Updated preferences object.

#### `PUT /preferences/location`
Updates the user's current location (called periodically by iOS).

**Request body:**
```json
{
  "latitude": 13.7563,
  "longitude": 100.5018
}
```

**Response 200:**
```json
{ "data": { "message": "location updated" } }
```

---

### Notifications

#### `GET /notifications/me`
Lists the user's notifications (paginated, sorted by `created_at desc`).

**Query params:**
| Param | Default | Description |
|-------|---------|-------------|
| `unread_only` | `false` | Only show unread notifications |
| `page` | 1 | |
| `limit` | 20 | Max 50 |

**Response 200:**
```json
{
  "data": [
    {
      "id": "...",
      "type": "critical_nearby",
      "title": "Critical threat 3.2km from you",
      "body": "Military force in Chatuchak, Bangkok",
      "event_id": "...",
      "distance_km": 3.2,
      "read_at": null,
      "created_at": "2026-04-17T08:00:00Z"
    }
  ],
  "pagination": { "page": 1, "limit": 20, "total": 3 }
}
```

#### `GET /notifications/me/unread-count`
Badge count — use this to display a number on the bell icon 🔔.

**Response 200:**
```json
{ "data": { "unread_count": 5 } }
```

#### `POST /notifications/:id/read`
Marks a single notification as read.

**Response 200:**
```json
{ "data": { "message": "marked as read" } }
```

**404** if the notification does not exist or does not belong to the current user.

#### `POST /notifications/read-all`
Marks all of the user's notifications as read.

**Response 200:**
```json
{ "data": { "modified_count": 12 } }
```

---

## iOS Implementation Flow

### 1. First Launch / Login

```
User logs in successfully
  → GET /preferences  (returns defaults)
  → Request Location permission (iOS CoreLocation)
    → If granted → PUT /preferences/location
    → If denied → nearby notifications won't work (daily_briefing still works)
  → Request Notification permission (UNUserNotificationCenter)
    → Required for local notification banners
```

### 2. Polling Loop (while app is active)

```swift
Timer.scheduledTimer(withTimeInterval: 60, repeats: true) { _ in
    // 1. Fetch unread count → update badge
    getUnreadCount() → update tab bar badge

    // 2. Fetch new notifications since last poll
    getNotifications(unreadOnly: true, page: 1, limit: 20)
    → for each new notification:
        - Add to inbox UI
        - Show in-app banner (if app foreground)
        - Trigger UNUserNotificationCenter (for OS banner)
}
```

**Recommended:**
- Store `lastSeenNotificationID` in UserDefaults
- When fetching new notifications, compare each `id` with `lastSeenNotificationID` → only show banner for newer ones

### 3. Location Update Strategy

**Don't update every second** — iOS significant location change or a timer is enough.

```swift
// Option A: Significant location change (battery friendly)
locationManager.startMonitoringSignificantLocationChanges()
// → triggers when user moves > 500m

// Option B: Periodic update (every 15-30 min while app active)
Timer.scheduledTimer(withTimeInterval: 1800, ...) {
    PUT /preferences/location
}
```

Recommendation: Option A + update on app foreground.

### 4. User Taps a Notification

```
User taps notification in inbox
  → POST /notifications/:id/read (mark as read, update badge)
  → Navigate based on type:
    - critical_nearby → Event detail page (use event_id)
    - daily_briefing  → Summary page (use summary_date)
```

### 5. Settings Screen

```
User opens "Notification Settings"
  → GET /preferences (show current values)
  → User toggles/changes:
    - Notifications on/off → PUT /preferences {notifications_enabled}
    - Severity filter      → PUT /preferences {min_severity}
    - Radius slider        → PUT /preferences {radius_km}
```

---

## Recommended UI Components

### 1. Bell Icon with Badge
- Placement: navigation bar or tab bar
- Badge: `unread_count` (poll every 60s)
- Tap → open Inbox

### 2. Inbox Screen
- Paginated list of notifications
- Unread: bold + dot indicator
- Pull to refresh
- Swipe to mark as read
- Empty state: "No notifications yet"
- Toolbar: "Mark all as read" button

### 3. In-App Banner (foreground)
- Top banner slides down when a new notification arrives
- Auto-dismiss after 4-5 seconds
- Tap → navigate to event/summary

### 4. OS-Level Banner (background/foreground)
```swift
let content = UNMutableNotificationContent()
content.title = notif.title
content.body = notif.body
content.sound = .default

let trigger = UNTimeIntervalNotificationTrigger(timeInterval: 1, repeats: false)
let request = UNNotificationRequest(identifier: notif.id, content: content, trigger: trigger)
UNUserNotificationCenter.current().add(request)
```

### 5. Settings Screen
- Toggle: Enable notifications
- Picker: Minimum severity
- Slider: Radius (1-500 km) — suggested labels: 10, 25, 50, 100, 250, 500

---

## Edge Cases & Caveats

| Case | Behavior |
|------|----------|
| User has no `last_location` | No `critical_nearby` notifications, but `daily_briefing` still works |
| User `notifications_enabled = false` | No notifications created at all |
| App fully closed | No notifications received (polling stops) |
| Same event triggers again (same GlobalEventID) | No duplicate notification (dedup by event_id) |
| Notification older than 30 days | Automatically deleted by the backend |
| MVP has no push notification | User must keep the app open to receive notifications |

---

## Example End-to-End Flow

```
1. User A opens the app and logs in
   → GET /preferences → { enabled, radius 50, severity critical }
   → Grants location permission
   → PUT /preferences/location { lat: 13.75, lng: 100.50 } (Bangkok)

2. GDELT sync at 08:00 UTC
   → Finds critical event: "Military force in Bangkok" at 13.78, 100.53
   → Matches User A (distance 3.2km < 50km, severity critical matches)
   → Creates a notification for User A

3. iOS polling at 08:00:30
   → GET /notifications/me/unread-count → 1
   → Badge shows "1"
   → GET /notifications/me?unread_only=true → [new notification]
   → Shows banner "Critical threat 3.2km from you — Military force in Bangkok"
   → Triggers local notification

4. User A taps the banner
   → Navigates to event detail (uses event_id)
   → POST /notifications/:id/read → badge becomes "0"

5. At 09:00 UTC the daily summary completes
   → Broadcasts notification to all users with notifications_enabled
   → User A receives "Your daily conflict briefing is ready"
   → Badge increments to "1" on the next poll
```
