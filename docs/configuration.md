# Configuration

## Environment Variables

Configuration is loaded from `.env` file (via godotenv) and environment variables. Env vars override `.env` file values.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8080` | HTTP server port |
| `MONGODB_URI` | No | `mongodb://localhost:27017` | MongoDB connection string (supports `mongodb+srv://`) |
| `MONGODB_DATABASE` | No | `conflux` | MongoDB database name |
| `JWT_SECRET` | **Yes** | `change-me-in-production` | Secret key for signing JWT tokens. **Must change in production.** |
| `SYNC_INTERVAL_MINUTES` | No | `15` | How often to sync from GDELT (minutes) |
| `GEMINI_API_KEY` | No | (empty) | Google AI API key for LLM summaries. Feature disabled if empty. |
| `SUMMARY_CHECK_INTERVAL_MIN` | No | `30` | How often summary scheduler checks (minutes) |
| `SUMMARY_BACKFILL_DAYS` | No | `7` | How many past days to backfill on startup |
| `LOG_LEVEL` | No | `info` | Log level: `debug`, `info`, `warn`, `error`, `fatal` |

## .env File

Copy `.env.example` and fill in values:

```bash
cp .env.example .env
```

Example `.env`:
```
PORT=8080
MONGODB_URI=mongodb+srv://user:password@cluster.mongodb.net/?appName=myapp
MONGODB_DATABASE=conflux
JWT_SECRET=a-very-long-random-secret-string
SYNC_INTERVAL_MINUTES=15
LOG_LEVEL=info
```

## Config Loading

Configuration is loaded once at startup in `main.go`:

```go
cfg := config.Load()
```

The `Config` struct is then passed to constructors via dependency injection — it is never accessed as a global.

Implementation: `internal/config/config.go`

## Security Notes

- `.env` is in `.gitignore` — never committed to git
- `JWT_SECRET` default is `"change-me-in-production"` — **must** be changed
- MongoDB credentials are part of `MONGODB_URI` — keep this secret
- No GDELT credentials needed (public API)

## Hot Reload (Development)

Uses [air](https://github.com/air-verse/air) for hot reload:

```bash
# Install air
go install github.com/air-verse/air@latest

# Run with hot reload
make dev
```

Configuration: `.air.toml`
- Watches `.go`, `.toml`, `.yaml` files
- Rebuilds to `./tmp/main`
- Excludes `tmp/`, `vendor/`, `.git/`
