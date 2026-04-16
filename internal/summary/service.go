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
	log.Info().Str("date", date).Msg("[summary.GenerateSummaryForDate] starting generation")

	// Parse date range (UTC day)
	dateStart, err := time.Parse("2006-01-02", date)
	if err != nil {
		return err
	}
	dateEnd := dateStart.Add(24*time.Hour - time.Second)

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

	// Compute severity breakdown from data
	breakdown := computeSeverityBreakdown(events)

	// Get existing summary for generation_number tracking
	existing, _ := s.repo.FindByDate(ctx, date)
	genNum := 1
	if existing != nil {
		genNum = existing.GenerationNumber + 1
	}

	// No events
	if total == 0 {
		log.Info().Str("date", date).Msg("[summary.GenerateSummaryForDate] no events for date")
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
	output, promptTokens, completionTokens, err := generateSummary(ctx, s.gemini, date, events)
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
		Title:             output.Title,
		Content:           output.Content,
		SeverityBreakdown: breakdown,
		Model:             modelName,
		PromptTokens:      promptTokens,
		CompletionTokens:  completionTokens,
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
		Int("prompt_tokens", promptTokens).
		Int("generation", genNum).
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
