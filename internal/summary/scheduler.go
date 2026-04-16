package summary

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	regenThreshold = 6 * time.Hour
	rateLimitDelay = 15 * time.Second // delay between Gemini calls to avoid rate limit
)

type Scheduler struct {
	svc          *Service
	interval     time.Duration
	backfillDays int
}

func NewScheduler(svc *Service, checkIntervalMin, backfillDays int) *Scheduler {
	return &Scheduler{
		svc:          svc,
		interval:     time.Duration(checkIntervalMin) * time.Minute,
		backfillDays: backfillDays,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	log.Info().
		Dur("interval", s.interval).
		Int("backfill_days", s.backfillDays).
		Dur("rate_limit_delay", rateLimitDelay).
		Msg("[summary.Scheduler] starting summary scheduler")

	s.runCycle(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Info().Msg("[summary.Scheduler] tick — starting new cycle")
			s.runCycle(ctx)
		case <-ctx.Done():
			log.Info().Msg("[summary.Scheduler] scheduler stopped (context cancelled)")
			return
		}
	}
}

func (s *Scheduler) runCycle(ctx context.Context) {
	cycleStart := time.Now()
	now := time.Now().UTC()
	todayStr := now.Format("2006-01-02")
	generated := 0
	skipped := 0

	log.Info().Str("today", todayStr).Int("backfill_days", s.backfillDays).Msg("[summary.Scheduler] cycle started")

	// 1. Current day
	if s.generateIfNeeded(ctx, todayStr, true) {
		generated++
		log.Info().Dur("delay", rateLimitDelay).Msg("[summary.Scheduler] rate limit delay before next call")
		time.Sleep(rateLimitDelay)
	} else {
		skipped++
	}

	// 2. Yesterday
	yesterdayStr := now.AddDate(0, 0, -1).Format("2006-01-02")
	if s.generateIfNeeded(ctx, yesterdayStr, false) {
		generated++
		log.Info().Dur("delay", rateLimitDelay).Msg("[summary.Scheduler] rate limit delay before next call")
		time.Sleep(rateLimitDelay)
	} else {
		skipped++
	}

	// 3. Backfill
	for i := 2; i <= s.backfillDays; i++ {
		dateStr := now.AddDate(0, 0, -i).Format("2006-01-02")
		if s.generateIfNeeded(ctx, dateStr, false) {
			generated++
			log.Info().Dur("delay", rateLimitDelay).Msg("[summary.Scheduler] rate limit delay before next call")
			time.Sleep(rateLimitDelay)
		} else {
			skipped++
		}
	}

	log.Info().
		Int("generated", generated).
		Int("skipped", skipped).
		Dur("duration", time.Since(cycleStart)).
		Msg("[summary.Scheduler] cycle completed")
}

// generateIfNeeded returns true if it called Gemini (for rate limiting)
func (s *Scheduler) generateIfNeeded(ctx context.Context, date string, allowRegen bool) bool {
	existing, err := s.svc.GetSummaryByDate(ctx, date)

	// Not found → generate
	if errors.Is(err, ErrNotFound) {
		log.Info().Str("date", date).Msg("[summary.Scheduler] no summary found, generating")
		if err := s.svc.GenerateSummaryForDate(ctx, date); err != nil {
			log.Error().Err(err).Str("date", date).Msg("[summary.Scheduler] generation failed")
		} else {
			log.Info().Str("date", date).Msg("[summary.Scheduler] generation succeeded")
		}
		return true
	}
	if err != nil {
		log.Error().Err(err).Str("date", date).Msg("[summary.Scheduler] failed to check summary")
		return false
	}

	// Already completed (past day) → skip
	if existing.Status == "completed" && !allowRegen {
		log.Debug().Str("date", date).Str("status", existing.Status).Msg("[summary.Scheduler] already completed, skipping")
		return false
	}

	// No events → skip
	if existing.Status == "no_events" {
		log.Debug().Str("date", date).Msg("[summary.Scheduler] no events for date, skipping")
		return false
	}

	// Failed → retry
	if existing.Status == "failed" {
		log.Info().Str("date", date).Str("error", existing.ErrorMessage).Msg("[summary.Scheduler] retrying failed summary")
		if err := s.svc.GenerateSummaryForDate(ctx, date); err != nil {
			log.Error().Err(err).Str("date", date).Msg("[summary.Scheduler] retry failed")
		} else {
			log.Info().Str("date", date).Msg("[summary.Scheduler] retry succeeded")
		}
		return true
	}

	// Current day: re-generate if stale
	if allowRegen && existing.Status == "completed" {
		age := time.Since(existing.GeneratedAt)
		if age > regenThreshold {
			log.Info().
				Str("date", date).
				Int("generation", existing.GenerationNumber).
				Dur("age", age).
				Msg("[summary.Scheduler] re-generating stale summary")
			if err := s.svc.GenerateSummaryForDate(ctx, date); err != nil {
				log.Error().Err(err).Str("date", date).Msg("[summary.Scheduler] re-generation failed")
			} else {
				log.Info().Str("date", date).Int("generation", existing.GenerationNumber+1).Msg("[summary.Scheduler] re-generation succeeded")
			}
			return true
		}
		log.Debug().Str("date", date).Dur("age", age).Dur("threshold", regenThreshold).Msg("[summary.Scheduler] today's summary still fresh, skipping")
	}

	return false
}
