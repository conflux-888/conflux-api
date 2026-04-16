# Project Structure

## Directory Layout

```
conflux-api/
├── main.go                              # Entry point — wires everything together
├── go.mod                               # Go module (github.com/conflux-888/conflux-api)
├── go.sum
├── Makefile                             # dev (air), build, run commands
├── .air.toml                            # Hot reload config
├── .env                                 # Local environment vars (gitignored)
├── .env.example                         # Template for env vars
├── .gitignore
├── requirement.md                       # Original system requirements
│
├── docs/                                # Documentation
│   ├── architecture.md
│   ├── api-reference.md
│   ├── gdelt-sync.md
│   ├── user-reports.md
│   ├── authentication.md
│   ├── database-schema.md
│   ├── configuration.md
│   └── project-structure.md             # (this file)
│
└── internal/                            # Private application code
    ├── config/
    │   └── config.go                    # Config struct + Load() from env
    │
    ├── infrastructure/
    │   ├── database/
    │   │   └── mongo.go                 # MongoDB Connect()
    │   └── server/
    │       └── server.go                # Gin engine + /health endpoint
    │
    ├── common/
    │   ├── response/
    │   │   └── response.go              # Success(), List(), Error() JSON helpers
    │   ├── middleware/
    │   │   └── auth.go                  # JWT auth middleware
    │   ├── jwt/
    │   │   └── jwt.go                   # GenerateToken(), ParseToken()
    │   ├── logger/
    │   │   └── logger.go                # zerolog Init()
    │   └── gemini/
    │       └── client.go                # Generic Gemini API client (reusable)
    │
    ├── user/                            # USER DOMAIN
    │   ├── model.go                     # User, RegisterRequest, LoginRequest, etc.
    │   ├── repository.go                # CRUD on users collection
    │   ├── service.go                   # Register, Login, GetProfile, UpdateProfile
    │   ├── handler.go                   # HTTP handlers
    │   └── routes.go                    # /api/v1/auth/*, /api/v1/users/*
    │
    ├── event/                           # EVENT DOMAIN (shared)
    │   ├── model.go                     # Event, GeoJSONPoint, EventFilter, constants
    │   ├── repository.go                # CRUD + Find + FindNearby + BulkUpsert
    │   ├── service.go                   # ListEvents, GetEvent, GetNearbyEvents
    │   ├── handler.go                   # HTTP handlers
    │   └── routes.go                    # /api/v1/events/*
    │
    ├── report/                          # REPORT DOMAIN
    │   ├── model.go                     # CreateReportRequest, ReportCluster
    │   ├── repository.go                # Cluster CRUD with $geoNear
    │   ├── service.go                   # SubmitReport + clustering, GetMyReports, Delete
    │   ├── handler.go                   # HTTP handlers
    │   └── routes.go                    # /api/v1/reports/*
    │
    └── sync/                            # SYNC DOMAIN
        ├── model.go                     # SyncState, GDELTEvent
        ├── severity.go                  # GoldsteinScale → severity classification
        ├── state_repository.go          # sync_state collection CRUD
        ├── client.go                    # GDELT HTTP client, CSV download + parse
        ├── service.go                   # Sync orchestrator + ticker loop
        ├── handler.go                   # Admin HTTP handlers
        └── routes.go                    # /api/v1/admin/sync/*
    │
    └── summary/                         # SUMMARY DOMAIN
        ├── model.go                     # DailySummary, SeverityBreakdown, region mapping
        ├── repository.go                # daily_summaries collection CRUD
        ├── prompt.go                    # Prompt building + event pre-processing
        ├── service.go                   # Orchestrate: events → Gemini → store
        ├── scheduler.go                 # Background job (every 30 min)
        ├── handler.go                   # HTTP handlers
        └── routes.go                    # /api/v1/summaries/*, /api/v1/admin/summaries/*
```

## Domain Pattern

Each domain follows the same 5-file pattern:

| File | Purpose | Depends on |
|------|---------|------------|
| `model.go` | Data structures, DTOs, constants | Nothing |
| `repository.go` | MongoDB operations | model |
| `service.go` | Business logic | repository, other services |
| `handler.go` | HTTP request/response handling | service, common/response, common/middleware |
| `routes.go` | Route registration | handler |

### Adding a New Domain

1. Create `internal/<domain>/` directory
2. Create the 5 files following the pattern above
3. In `main.go`:
   - Create repository: `repo := domain.NewRepository(db)`
   - Create service: `svc := domain.NewService(repo, ...)`
   - Create handler: `handler := domain.NewHandler(svc)`
   - Register routes: `domain.RegisterRoutes(router, handler, authMW)`

## Key Patterns

### Constructor Injection

No DI framework. All dependencies passed through constructors:

```go
repo := user.NewRepository(db)
svc := user.NewService(repo, cfg.JWTSecret)
handler := user.NewHandler(svc)
```

### Error Handling

Domain-specific sentinel errors defined at package level:

```go
var ErrNotFound = errors.New("user not found")
var ErrEmailTaken = errors.New("email already taken")
```

Handlers map these to HTTP status codes:

```go
if errors.Is(err, user.ErrEmailTaken) {
    response.Conflict(c, "email already taken")
    return
}
```

### Logging Convention

All log messages follow `[module.method]` format:

```go
log.Info().Str("user_id", id).Msg("[user.Register] user registered")
log.Error().Err(err).Msg("[sync.runSync] failed to fetch events")
log.Warn().Str("email", email).Msg("[user.Login] invalid password")
```

Log levels:
- `info` — successful operations, progress updates
- `warn` — expected failures (invalid input, not found)
- `error` — unexpected failures (DB errors, external API errors)
- `fatal` — startup failures (DB connection, server start)

### Response Helpers

All handlers use `internal/common/response` for consistent JSON output:

```go
response.Success(c, http.StatusCreated, data)   // { "data": ... }
response.List(c, data, pagination)              // { "data": [...], "pagination": {...} }
response.ValidationError(c, "email is required") // 400
response.Unauthorized(c, "invalid token")        // 401
response.NotFound(c, "not found")               // 404
response.Conflict(c, "already exists")          // 409
response.InternalError(c)                       // 500
```

### Route Registration

Each domain exports a `RegisterRoutes` function:

```go
func RegisterRoutes(r *gin.Engine, h *Handler, authMiddleware gin.HandlerFunc) {
    group := r.Group("/path")
    group.Use(authMiddleware)  // apply to all routes in group
    {
        group.GET("", h.HandleList)
        group.POST("", h.HandleCreate)
    }
}
```

### Soft Deletes

Events use `is_deleted` flag instead of hard deletes. All queries include `is_deleted: false` filter.

### GeoJSON

All geographic data stored as GeoJSON Point with 2dsphere index:

```go
type GeoJSONPoint struct {
    Type        string     `bson:"type"`
    Coordinates [2]float64 `bson:"coordinates"` // [lng, lat]
}
```

## Go Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/gin-gonic/gin` | HTTP framework |
| `go.mongodb.org/mongo-driver/v2` | MongoDB driver |
| `github.com/golang-jwt/jwt/v5` | JWT token handling |
| `golang.org/x/crypto` | bcrypt password hashing |
| `github.com/rs/zerolog` | Structured JSON logging |
| `github.com/joho/godotenv` | .env file loading |
