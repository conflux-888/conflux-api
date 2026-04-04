package sync

import (
	"context"
	"time"

	"github.com/conflux-888/conflux-api/internal/event"
	"github.com/rs/zerolog/log"
)

// CAMEO event root code descriptions
var cameoDescriptions = map[string]string{
	"14": "Military posture",
	"15": "Conventional attack",
	"16": "Unconventional mass violence",
	"17": "Riotous forces",
	"18": "Violent clash",
	"19": "Use of force",
	"20": "Military force",
}

type Service struct {
	client    *Client
	eventRepo *event.Repository
	stateRepo *StateRepository
	interval  time.Duration
}

func NewService(client *Client, eventRepo *event.Repository, stateRepo *StateRepository, intervalMinutes int) *Service {
	return &Service{
		client:    client,
		eventRepo: eventRepo,
		stateRepo: stateRepo,
		interval:  time.Duration(intervalMinutes) * time.Minute,
	}
}

func (s *Service) Start(ctx context.Context) {
	log.Info().Dur("interval", s.interval).Msg("[sync.Start] starting GDELT sync")

	s.runSync(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.runSync(ctx)
		case <-ctx.Done():
			log.Info().Msg("[sync.Start] sync stopped")
			return
		}
	}
}

func (s *Service) runSync(ctx context.Context) {
	log.Info().Msg("[sync.runSync] sync cycle started")

	events, maxDateAdded, err := s.client.FetchLatestEvents(ctx)
	if err != nil {
		log.Error().Err(err).Msg("[sync.runSync] failed to fetch GDELT events")
		s.updateState(ctx, "", 0, "failed", err.Error())
		return
	}

	if len(events) == 0 {
		log.Info().Msg("[sync.runSync] no conflict events in latest batch")
		s.updateState(ctx, maxDateAdded, 0, "success", "")
		return
	}

	// Map GDELT events to domain events
	batch := make([]event.Event, 0, len(events))
	for _, g := range events {
		e := mapGDELTEvent(g)
		batch = append(batch, e)
	}

	// Bulk upsert
	upserted, err := s.eventRepo.BulkUpsert(ctx, batch, event.SourceGDELT)
	if err != nil {
		log.Error().Err(err).Msg("[sync.runSync] bulk upsert failed")
		s.updateState(ctx, "", 0, "failed", err.Error())
		return
	}

	s.updateState(ctx, maxDateAdded, int(upserted), "success", "")

	log.Info().
		Int("batch_size", len(batch)).
		Int64("upserted", upserted).
		Str("max_date_added", maxDateAdded).
		Msg("[sync.runSync] sync cycle completed")
}

func (s *Service) updateState(ctx context.Context, timestamp string, synced int, status, errMsg string) {
	state := &SyncState{
		LastSyncAt:   time.Now(),
		Status:       status,
		EventsSynced: synced,
		ErrorMessage: errMsg,
	}
	if timestamp != "" {
		state.LastSyncTimestamp = timestamp
	}

	if err := s.stateRepo.Upsert(ctx, state); err != nil {
		log.Error().Err(err).Msg("[sync.runSync] failed to update sync state")
	}
}

func (s *Service) GetStatus(ctx context.Context) (*SyncState, error) {
	return s.stateRepo.Get(ctx)
}

func (s *Service) TriggerSync(ctx context.Context) {
	s.runSync(ctx)
}

func mapGDELTEvent(g GDELTEvent) event.Event {
	eventDate, _ := time.Parse("20060102", g.Day)

	desc := cameoDescriptions[g.EventRootCode]
	if desc == "" {
		desc = "Event code " + g.EventCode
	}

	title := desc
	if g.ActionGeoFullName != "" {
		title += " in " + g.ActionGeoFullName
	}

	var actors []string
	if g.Actor1Name != "" {
		actors = append(actors, g.Actor1Name)
	}
	if g.Actor2Name != "" {
		actors = append(actors, g.Actor2Name)
	}

	return event.Event{
		Source:       event.SourceGDELT,
		ExternalID:   g.GlobalEventID,
		EventType:     desc,
		SubEventType:  "CAMEO " + g.EventCode,
		EventRootCode: g.EventRootCode,
		Severity:      ClassifySeverity(g.GoldsteinScale),
		Title:        title,
		Description:  g.SourceURL,
		Country:      g.ActionGeoCountryCode,
		LocationName: g.ActionGeoFullName,
		Location: event.GeoJSONPoint{
			Type:        "Point",
			Coordinates: [2]float64{g.ActionGeoLong, g.ActionGeoLat},
		},
		NumSources:  g.NumSources,
		NumArticles: g.NumArticles,
		Actors:      actors,
		EventDate:  eventDate,
		IsDeleted:  false,
	}
}
