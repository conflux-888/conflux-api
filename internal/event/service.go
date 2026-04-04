package event

import (
	"context"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
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

func (s *Service) GetNearbyEvents(ctx context.Context, lng, lat, radiusKM float64, severity string, limit int) ([]Event, error) {
	events, err := s.repo.FindNearby(ctx, lng, lat, radiusKM, severity, limit)
	if err != nil {
		log.Error().Err(err).Float64("lng", lng).Float64("lat", lat).Float64("radius_km", radiusKM).Msg("[event.GetNearbyEvents] failed to query nearby events")
		return nil, err
	}
	log.Info().Int("count", len(events)).Float64("lng", lng).Float64("lat", lat).Float64("radius_km", radiusKM).Msg("[event.GetNearbyEvents] nearby events found")
	return events, nil
}
