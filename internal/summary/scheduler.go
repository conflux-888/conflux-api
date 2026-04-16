package summary

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
)

const regenThreshold = 6 * time.Hour

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
	log.Info().Dur("interval", s.interval).Int("backfill_days", s.backfillDays).Msg("[summary.Scheduler] starting summary scheduler")

	s.runCycle(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.runCycle(ctx)
		case <-ctx.Done():
			log.Info().Msg("[summary.Scheduler] scheduler stopped")
			return
		}
	}
}

func (s *Scheduler) runCycle(ctx context.Context) {
	now := time.Now().UTC()
	todayStr := now.Format("2006-01-02")

	// 1. Current day: generate if missing / failed / stale (> 6 hours)
	s.generateIfNeeded(ctx, todayStr, true)

	// 2. Yesterday: generate if missing / failed
	yesterdayStr := now.AddDate(0, 0, -1).Format("2006-01-02")
	s.generateIfNeeded(ctx, yesterdayStr, false)

	// 3. Backfill: check past N days for missing summaries
	for i := 2; i <= s.backfillDays; i++ {
		dateStr := now.AddDate(0, 0, -i).Format("2006-01-02")
		s.generateIfNeeded(ctx, dateStr, false)
	}
}

func (s *Scheduler) generateIfNeeded(ctx context.Context, date string, allowRegen bool) {
	existing, err := s.svc.GetSummaryByDate(ctx, date)

	// Not found → generate
	if errors.Is(err, ErrNotFound) {
		log.Info().Str("date", date).Msg("[summary.Scheduler] no summary found, generating")
		if err := s.svc.GenerateSummaryForDate(ctx, date); err != nil {
			log.Error().Err(err).Str("date", date).Msg("[summary.Scheduler] generation failed")
		}
		return
	}
	if err != nil {
		log.Error().Err(err).Str("date", date).Msg("[summary.Scheduler] failed to check summary")
		return
	}

	// Failed → retry
	if existing.Status == "failed" {
		log.Info().Str("date", date).Msg("[summary.Scheduler] retrying failed summary")
		if err := s.svc.GenerateSummaryForDate(ctx, date); err != nil {
			log.Error().Err(err).Str("date", date).Msg("[summary.Scheduler] retry failed")
		}
		return
	}

	// Current day: re-generate if stale
	if allowRegen && existing.Status == "completed" && time.Since(existing.GeneratedAt) > regenThreshold {
		log.Info().Str("date", date).Int("generation", existing.GenerationNumber).Msg("[summary.Scheduler] re-generating stale summary")
		if err := s.svc.GenerateSummaryForDate(ctx, date); err != nil {
			log.Error().Err(err).Str("date", date).Msg("[summary.Scheduler] re-generation failed")
		}
	}
}
