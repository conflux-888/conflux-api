package summary

import (
	"context"
	"time"

	"github.com/conflux-888/conflux-api/internal/common/gemini"
	"github.com/conflux-888/conflux-api/internal/common/response"
	"github.com/conflux-888/conflux-api/internal/event"
	"github.com/rs/zerolog/log"
)

type Service struct {
	repo      *Repository
	eventRepo *event.Repository
	gemini    *gemini.Client
}

func NewService(repo *Repository, eventRepo *event.Repository, geminiClient *gemini.Client) *Service {
	return &Service{repo: repo, eventRepo: eventRepo, gemini: geminiClient}
}

func (s *Service) GenerateSummaryForDate(ctx context.Context, date string) error {
	startTime := time.Now()
	log.Info().Str("date", date).Msg("[summary.GenerateSummaryForDate] starting generation")

	// Parse date range (UTC day)
	dateStart, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Error().Err(err).Str("date", date).Msg("[summary.GenerateSummaryForDate] invalid date format")
		return err
	}
	dateEnd := dateStart.Add(24*time.Hour - time.Second)

	log.Debug().Str("date", date).Time("from", dateStart).Time("to", dateEnd).Msg("[summary.GenerateSummaryForDate] querying events")

	// Query events for this date
	events, total, err := s.eventRepo.Find(ctx, event.EventFilter{
		DateFrom: &dateStart,
		DateTo:   &dateEnd,
		Page:     1,
		Limit:    10000,
		Sort:     "date_desc",
	})
	if err != nil {
		log.Error().Err(err).Str("date", date).Msg("[summary.GenerateSummaryForDate] failed to query events")
		return err
	}

	log.Info().Str("date", date).Int64("event_count", total).Msg("[summary.GenerateSummaryForDate] events queried")

	// Compute severity breakdown from data
	breakdown := computeSeverityBreakdown(events)
	log.Debug().
		Str("date", date).
		Int("critical", breakdown.Critical).
		Int("high", breakdown.High).
		Int("medium", breakdown.Medium).
		Int("low", breakdown.Low).
		Msg("[summary.GenerateSummaryForDate] severity breakdown")

	// Get existing summary for generation_number tracking
	existing, _ := s.repo.FindByDate(ctx, date)
	genNum := 1
	if existing != nil {
		genNum = existing.GenerationNumber + 1
		log.Debug().Str("date", date).Int("prev_generation", existing.GenerationNumber).Msg("[summary.GenerateSummaryForDate] previous summary found")
	}

	// No events
	if total == 0 {
		log.Info().Str("date", date).Msg("[summary.GenerateSummaryForDate] no events for date, storing no_events")
		return s.repo.Upsert(ctx, &DailySummary{
			SummaryDate:       date,
			Status:            "no_events",
			EventCount:        0,
			Title:             "No conflict events recorded — " + date,
			Content:           "No significant conflict events were recorded for this date.",
			SeverityBreakdown: breakdown,
			Model:             modelName,
			GenerationNumber:  genNum,
			GeneratedAt:       time.Now(),
		})
	}

	// Call Gemini
	log.Info().Str("date", date).Int64("event_count", total).Msg("[summary.GenerateSummaryForDate] calling Gemini")
	result, err := generateSummary(ctx, s.gemini, date, events)
	if err != nil {
		log.Error().Err(err).Str("date", date).Msg("[summary.GenerateSummaryForDate] Gemini API failed")
		return s.repo.Upsert(ctx, &DailySummary{
			SummaryDate:      date,
			Status:           "failed",
			EventCount:       int(total),
			Model:            modelName,
			GenerationNumber: genNum,
			GeneratedAt:      time.Now(),
			ErrorMessage:     err.Error(),
		})
	}

	// Store completed summary
	summary := &DailySummary{
		SummaryDate:       date,
		Status:            "completed",
		EventCount:        int(total),
		IncidentCount:     result.IncidentCount,
		Title:             result.Output.Title,
		Content:           result.Output.Content,
		TopEvents:         result.Output.TopEvents,
		SeverityBreakdown: breakdown,
		Model:             modelName,
		PromptTokens:      result.PromptTokens,
		CompletionTokens:  result.CompletionTokens,
		GenerationNumber:  genNum,
		GeneratedAt:       time.Now(),
	}

	if err := s.repo.Upsert(ctx, summary); err != nil {
		log.Error().Err(err).Str("date", date).Msg("[summary.GenerateSummaryForDate] failed to store summary")
		return err
	}

	log.Info().
		Str("date", date).
		Int("event_count", int(total)).
		Int("prompt_tokens", result.PromptTokens).
		Int("completion_tokens", result.CompletionTokens).
		Int("incidents", result.IncidentCount).
		Int("generation", genNum).
		Dur("duration", time.Since(startTime)).
		Msg("[summary.GenerateSummaryForDate] summary generated")

	return nil
}

func (s *Service) GetSummaryByDate(ctx context.Context, date string) (*DailySummary, error) {
	return s.repo.FindByDate(ctx, date)
}

func (s *Service) ListSummaries(ctx context.Context, from, to string, page, limit int) ([]DailySummary, *response.Pagination, error) {
	summaries, total, err := s.repo.FindByDateRange(ctx, from, to, page, limit)
	if err != nil {
		return nil, nil, err
	}
	return summaries, &response.Pagination{Page: page, Limit: limit, Total: total}, nil
}

func (s *Service) GetLatestSummary(ctx context.Context) (*DailySummary, error) {
	summaries, err := s.repo.FindLatest(ctx, 1)
	if err != nil {
		return nil, err
	}
	if len(summaries) == 0 {
		return nil, ErrNotFound
	}
	return &summaries[0], nil
}

func computeSeverityBreakdown(events []event.Event) SeverityBreakdown {
	var b SeverityBreakdown
	for _, e := range events {
		switch e.Severity {
		case "critical":
			b.Critical++
		case "high":
			b.High++
		case "medium":
			b.Medium++
		case "low":
			b.Low++
		}
	}
	return b
}
