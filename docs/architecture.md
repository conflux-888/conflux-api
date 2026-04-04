# Architecture Overview

## System Purpose

Conflux API is a **Global Threat Report** backend that aggregates conflict event data from GDELT (machine-coded from global news every 15 minutes) and user-submitted reports, then serves them through a REST API for map-based visualization.

## High-Level Architecture

```
┌─────────────┐       ┌──────────────┐       ┌──────────┐
│  Conflux UI │──────▸│  Conflux API │──────▸│  MongoDB  │
│  (Frontend) │◂──────│   (Go/Gin)   │◂──────│  Atlas    │
└─────────────┘       └──────┬───────┘       └──────────┘
                             │                     ▲
                             │  Background sync    │
                             │  (every 15 min)     │
                             ▼                     │
                      ┌──────────────┐             │
                      │  GDELT 2.0   │─────────────┘
                      │  (CSV files) │
                      └──────────────┘
```

### Data Flows

**Read flow**: UI → API → MongoDB → response
**Sync flow (background)**: GDELT CSV → parse → filter conflicts → BulkUpsert → MongoDB
**Report flow**: User → API → MongoDB (events + report_clusters)

## Tech Stack

| Layer | Technology | Version |
|-------|------------|---------|
| Language | Go | 1.25+ |
| Web Framework | Gin | v1.12 |
| Database | MongoDB | Atlas (driver v2) |
| External Data | GDELT 2.0 | CSV export files |
| Auth | Custom JWT | golang-jwt/jwt v5 |
| Password Hash | bcrypt | golang.org/x/crypto |
| Logging | zerolog | JSON structured logs |
| Config | godotenv | .env file |

## Domain-Based Architecture

The codebase is organized by **domain** (not by layer). Each domain encapsulates its own model, repository, service, handler, and routes.

```
internal/
├── config/          # App configuration
├── infrastructure/  # Database connection, HTTP server setup
├── common/          # Shared utilities (response, middleware, jwt, logger)
├── user/            # User management domain
├── event/           # Event data domain (shared model + read API)
├── report/          # User report domain (write + clustering)
└── sync/            # GDELT sync domain (background process)
```

### Domain Dependencies

```
user     → (standalone, uses common/jwt for token generation)
event    → (standalone, shared repository used by report + sync)
report   → event (writes to events collection, reads event repository)
sync     → event (writes to events collection via bulk upsert)
```

No circular dependencies. Cross-domain interaction happens through shared repositories passed via dependency injection in `main.go`.

## Dependency Injection

All wiring happens in `main.go` — no DI framework, plain constructor injection:

```
config → logger → database
  → userRepo → userService → userHandler
  → eventRepo → eventService → eventHandler
  → eventRepo + reportClusterRepo → reportService → reportHandler
  → syncClient + eventRepo + syncStateRepo → syncService → syncHandler
  → router (all routes registered)
  → go syncService.Start(ctx) (background goroutine)
  → http.Server.ListenAndServe()
```

## Background Sync

The GDELT sync runs as a goroutine launched from `main.go`:
- Executes immediately on startup
- Then repeats every `SYNC_INTERVAL_MINUTES` (default 15)
- Stops gracefully when the context is cancelled (SIGINT/SIGTERM)
- Downloads GDELT CSV, parses, filters conflict events (CAMEO root code 18-20), bulk upserts

## Graceful Shutdown

Uses `signal.NotifyContext` to handle SIGINT/SIGTERM:
1. Context cancelled → sync goroutine stops
2. HTTP server shutdown with 5s timeout
3. In-flight requests complete before exit

## Data Model

All events (both GDELT and user reports) are stored in a single `events` MongoDB collection, differentiated by the `source` field (`"gdelt"` or `"user_report"`). This enables unified querying across both data sources with a single API endpoint.

User reports additionally create entries in the `report_clusters` collection for spatial grouping (5km radius, 24hr window, same event type).
