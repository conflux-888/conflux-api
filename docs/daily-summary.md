# Daily LLM Summary

## Overview

Generates daily news-style briefings of conflict events using Google Gemini 2.5 Flash. Output reads like a professional Reuters/BBC article — continuous prose, no bullet points.

## How It Works

```
Background Scheduler (every 30 min)
  → Check: does today's summary need (re)generation?
  → Query events from MongoDB for target date
  → Pre-process: group by region, sort by severity
  → Send to Gemini API with journalist system prompt
  → Parse structured JSON response (title + content)
  → Store in daily_summaries collection
```

## Generation Schedule

| Date | Behavior |
|------|----------|
| Today | Generate on first check, re-generate every 6 hours (for live updates) |
| Yesterday | Generate if missing or failed |
| Past 7 days | Backfill if missing or failed |
| Older | Keep as-is, never re-generate |

The scheduler checks every 30 minutes but only triggers generation when needed:
- Missing summary → generate
- Failed summary → retry
- Today's summary older than 6 hours → re-generate
- Completed past summary → skip

## Architecture

```
internal/common/gemini/        ← Generic Gemini client (reusable)
  client.go                    # Generate(), GenerateJSON()

internal/summary/              ← Summary business domain
  model.go                     # DailySummary, SeverityBreakdown, country-to-region mapping
  repository.go                # daily_summaries MongoDB CRUD
  prompt.go                    # Prompt building, event pre-processing, LLM call
  service.go                   # Orchestrate: query events → generate → store
  scheduler.go                 # Background job (ticker loop)
  handler.go                   # HTTP handlers
  routes.go                    # Route registration
```

Key separation: `gemini.Client` is a generic LLM client that can be reused by other features. The summary-specific prompt engineering and event pre-processing stay in the `summary` domain.

## Gemini Integration

| Setting | Value |
|---------|-------|
| SDK | `github.com/google/generative-ai-go/genai` |
| Auth | API key via `GEMINI_API_KEY` env var |
| Model | `gemini-2.5-flash` |
| Temperature | 0.2 (factual, low creativity) |
| Output | Structured JSON via `ResponseMIMEType` + `ResponseSchema` |

### Structured Output Schema

```json
{
  "title": "string (news headline)",
  "content": "string (full article, 400-800 words)"
}
```

Enforced by Gemini's `ResponseSchema` — guarantees valid JSON matching our struct. No parsing failures.

### System Prompt

```
You are a senior conflict correspondent writing for an international
news wire service like Reuters or BBC World Service. You produce daily
conflict briefings that are factual, authoritative, and written in
clear professional English.

Rules:
- Write in third person, past tense
- Lead with the most significant developments
- Group related events by region, flowing naturally between paragraphs
- Include specific details: locations, actor names, number of incidents
- End with a brief statistical context paragraph
- Do NOT use bullet points, headers, or lists — write continuous prose
- Do NOT speculate or editorialize — report only what the data shows
- Length: 400-800 words depending on event volume
```

### Input Pre-Processing

Events are grouped by geographic region and sorted by severity before sending to the LLM:

```
Date: April 5, 2025
Total: 142 events | Critical: 5, High: 30, Medium: 65, Low: 42

Region: Eastern Europe (45 events)
- [CRITICAL] Military force in Kharkiv Oblast, Ukraine — Actors: RUSSIA, UKRAINE — 15 sources
- [HIGH] Violent clash in Donetsk, Ukraine — Actors: RUSSIA — 8 sources
...

Region: Middle East (32 events)
- [CRITICAL] Use of force in Gaza, Palestine — Actors: ISRAEL — 12 sources
...
```

Country-to-region mapping is hardcoded for ~50 conflict-relevant countries in `model.go`. Unknown countries map to "Other".

Events are capped at 20 per region in the prompt to manage token usage.

## Timezone Strategy

Summaries use **UTC days** — a summary for "2025-04-05" covers events where `event_date` is between `2025-04-05T00:00:00Z` and `2025-04-05T23:59:59Z`.

This aligns with GDELT events which are parsed as midnight UTC. See [architecture.md](architecture.md) for the full timezone rationale.

`summary_date` is stored as a string `"YYYY-MM-DD"` (not `time.Time`) to avoid timezone interpretation.

## MongoDB Schema

Collection: `daily_summaries`

| Field | Type | Description |
|-------|------|-------------|
| `_id` | ObjectID | Auto-generated |
| `summary_date` | string | `"YYYY-MM-DD"` (unique) |
| `status` | string | `completed`, `failed`, `no_events` |
| `event_count` | int | Number of events summarized |
| `title` | string | LLM-generated headline |
| `content` | string | LLM-generated news article (continuous prose) |
| `severity_breakdown` | object | `{ critical, high, medium, low }` — computed from data |
| `model` | string | LLM model used (e.g., `gemini-2.5-flash`) |
| `prompt_tokens` | int | Input token count (for cost tracking) |
| `completion_tokens` | int | Output token count |
| `generation_number` | int | How many times this date was (re)generated |
| `generated_at` | datetime | When LLM was called |
| `created_at` | datetime | First creation |
| `updated_at` | datetime | Last update |
| `error_message` | string | Error details if status is `failed` |

Index: `summary_date` (unique)

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/summaries` | List summaries (paginated, optional date range) |
| GET | `/api/v1/summaries/latest` | Most recent completed summary |
| GET | `/api/v1/summaries/:date` | Summary for specific date (YYYY-MM-DD) |
| POST | `/api/v1/admin/summaries/trigger` | Manually generate `{ "date": "2025-04-05" }` |

All require auth.

### GET /api/v1/summaries

| Param | Default | Description |
|-------|---------|-------------|
| from | 7 days ago | Start date (YYYY-MM-DD) |
| to | today | End date (YYYY-MM-DD) |
| page | 1 | Page number |
| limit | 7 | Items per page (max 30) |

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Gemini API fails | Store status `failed` + error_message, retry next cycle |
| No events for date | Store status `no_events` with canned message, don't retry |
| JSON parse error | Treat as failed, retry |
| No `GEMINI_API_KEY` | Feature disabled, scheduler doesn't start, log warning |

## Cost

- ~1000 events/day × ~60 tokens/event = ~60K input tokens per call
- 4 re-generations/day × 60K = ~240K tokens/day
- $0.15/1M input tokens → **~$0.04/day**

## Configuration

| Env Var | Default | Description |
|---------|---------|-------------|
| `GEMINI_API_KEY` | (none) | Google AI API key. Feature disabled if empty. |
| `SUMMARY_CHECK_INTERVAL_MIN` | 30 | How often scheduler checks (minutes) |
| `SUMMARY_BACKFILL_DAYS` | 7 | How many past days to backfill on startup |

## Key Files

| File | Purpose |
|------|---------|
| `internal/common/gemini/client.go` | Generic Gemini client (reusable) |
| `internal/summary/model.go` | DailySummary struct + country-to-region mapping |
| `internal/summary/repository.go` | MongoDB CRUD for daily_summaries |
| `internal/summary/prompt.go` | Prompt building + event pre-processing |
| `internal/summary/service.go` | Business logic orchestration |
| `internal/summary/scheduler.go` | Background job |
| `internal/summary/handler.go` | HTTP handlers |
| `internal/summary/routes.go` | Route registration |
