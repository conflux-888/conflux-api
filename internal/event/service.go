package event

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Notifier is the minimal interface the event service needs to dispatch notifications.
// The notification package's Service already satisfies it — kept here to avoid import cycle.
type Notifier interface {
	NotifyNearbyCritical(ctx context.Context, events []Event)
	DeleteNotificationsForEvent(ctx context.Context, eventID bson.ObjectID) int64
	DeleteNotificationsForEvents(ctx context.Context, eventIDs []bson.ObjectID) int64
}

type Service struct {
	repo     *Repository
	notifier Notifier
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SetNotifier(n Notifier) {
	s.notifier = n
}

func (s *Service) ListEvents(ctx context.Context, filter EventFilter) ([]Event, int64, error) {
	events, total, err := s.repo.Find(ctx, filter)
	if err != nil {
		log.Error().Err(err).Msg("[event.ListEvents] failed to query events")
		return nil, 0, err
	}
	log.Info().Int("count", len(events)).Int64("total", total).Int("page", filter.Page).Msg("[event.ListEvents] events listed")
	return events, total, nil
}

func (s *Service) GetEvent(ctx context.Context, id string) (*Event, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		log.Warn().Str("id", id).Msg("[event.GetEvent] invalid event id")
		return nil, ErrNotFound
	}
	e, err := s.repo.FindByID(ctx, oid)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("[event.GetEvent] failed to find event")
		return nil, err
	}
	return e, nil
}

// SeedAdminEvent inserts a synthetic event and triggers NotifyNearbyCritical on the same
// code path used by production sync. Used by the admin UI to test notification delivery.
func (s *Service) SeedAdminEvent(ctx context.Context, req SeedAdminEventRequest) (*Event, error) {
	eventType := req.EventType
	if eventType == "" {
		eventType = "Admin Test Event"
	}
	rootCode := req.EventRootCode
	if rootCode == "" {
		rootCode = "19"
	}
	numArticles := req.NumArticles
	if numArticles <= 0 {
		numArticles = 10
	}

	e := &Event{
		Source:        SourceGDELT,
		ExternalID:    "ADMIN_TEST_" + uuid.NewString(),
		EventType:     eventType,
		EventRootCode: rootCode,
		Severity:      req.Severity,
		Title:         req.Title,
		Description:   req.Description,
		Country:       req.Country,
		LocationName:  req.LocationName,
		Location: GeoJSONPoint{
			Type:        "Point",
			Coordinates: [2]float64{req.Longitude, req.Latitude},
		},
		NumSources:  1,
		NumArticles: numArticles,
		EventDate:   time.Now().UTC(),
		IsDeleted:   false,
	}

	if err := s.repo.Create(ctx, e); err != nil {
		log.Error().Err(err).Msg("[event.SeedAdminEvent] failed to insert event")
		return nil, err
	}

	log.Info().Str("event_id", e.ID.Hex()).Str("external_id", e.ExternalID).Str("severity", e.Severity).
		Msg("[event.SeedAdminEvent] event seeded")

	// Dispatch notification asynchronously — same pattern as sync.runSync
	if s.notifier != nil {
		go s.notifier.NotifyNearbyCritical(context.Background(), []Event{*e})
	}

	return e, nil
}

// ListSeededEvents returns admin-seeded events (most recent first).
func (s *Service) ListSeededEvents(ctx context.Context, page, limit int) ([]Event, int64, error) {
	events, total, err := s.repo.FindSeeded(ctx, page, limit)
	if err != nil {
		log.Error().Err(err).Msg("[event.ListSeededEvents] failed")
		return nil, 0, err
	}
	return events, total, nil
}

// DeleteSeededEvent hard-deletes a seeded event and its related notifications.
// Refuses to delete events that are not admin-seeded (external_id must have ADMIN_TEST_ prefix).
// Returns the number of notifications that were cleaned up.
func (s *Service) DeleteSeededEvent(ctx context.Context, id string) (int64, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		log.Warn().Str("id", id).Msg("[event.DeleteSeededEvent] invalid id")
		return 0, ErrNotFound
	}

	e, err := s.repo.FindByID(ctx, oid)
	if err != nil {
		return 0, err
	}
	if len(e.ExternalID) < 11 || e.ExternalID[:11] != "ADMIN_TEST_" {
		log.Warn().Str("id", id).Str("external_id", e.ExternalID).Msg("[event.DeleteSeededEvent] refused: not a seeded event")
		return 0, ErrNotFound
	}

	if err := s.repo.HardDeleteByID(ctx, oid); err != nil {
		log.Error().Err(err).Str("id", id).Msg("[event.DeleteSeededEvent] failed to delete event")
		return 0, err
	}

	var notifsDeleted int64
	if s.notifier != nil {
		notifsDeleted = s.notifier.DeleteNotificationsForEvent(ctx, oid)
	}

	log.Info().Str("id", id).Str("external_id", e.ExternalID).Int64("notifications_deleted", notifsDeleted).
		Msg("[event.DeleteSeededEvent] deleted")
	return notifsDeleted, nil
}

// DeleteAllSeededEvents removes every admin-seeded event and their notifications.
// Returns (events deleted, notifications deleted).
func (s *Service) DeleteAllSeededEvents(ctx context.Context) (int64, int64, error) {
	ids, err := s.repo.FindSeededIDs(ctx)
	if err != nil {
		log.Error().Err(err).Msg("[event.DeleteAllSeededEvents] failed to list seeded IDs")
		return 0, 0, err
	}
	if len(ids) == 0 {
		return 0, 0, nil
	}

	eventsDeleted, err := s.repo.HardDeleteSeeded(ctx)
	if err != nil {
		log.Error().Err(err).Msg("[event.DeleteAllSeededEvents] failed to delete events")
		return 0, 0, err
	}

	var notifsDeleted int64
	if s.notifier != nil {
		notifsDeleted = s.notifier.DeleteNotificationsForEvents(ctx, ids)
	}

	log.Info().Int64("events_deleted", eventsDeleted).Int64("notifications_deleted", notifsDeleted).
		Msg("[event.DeleteAllSeededEvents] done")
	return eventsDeleted, notifsDeleted, nil
}

func (s *Service) GetNearbyEvents(ctx context.Context, lng, lat, radiusKM float64, severity string, limit int) ([]Event, error) {
	events, err := s.repo.FindNearby(ctx, lng, lat, radiusKM, severity, limit)
	if err != nil {
		log.Error().Err(err).Float64("lng", lng).Float64("lat", lat).Float64("radius_km", radiusKM).Msg("[event.GetNearbyEvents] failed to query nearby events")
		return nil, err
	}
	log.Info().Int("count", len(events)).Float64("lng", lng).Float64("lat", lat).Float64("radius_km", radiusKM).Msg("[event.GetNearbyEvents] nearby events found")
	return events, nil
}
