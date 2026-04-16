package preferences

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrInvalidUser = errors.New("invalid user id")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context, userID string) (*UserPreferences, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		log.Warn().Str("user_id", userID).Msg("[preferences.Get] invalid user id")
		return nil, ErrInvalidUser
	}
	return s.repo.Get(ctx, uid)
}

func (s *Service) Update(ctx context.Context, userID string, req UpdateRequest) (*UserPreferences, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		log.Warn().Str("user_id", userID).Msg("[preferences.Update] invalid user id")
		return nil, ErrInvalidUser
	}

	prefs, err := s.repo.Get(ctx, uid)
	if err != nil {
		return nil, err
	}

	if req.NotificationsEnabled != nil {
		prefs.NotificationsEnabled = *req.NotificationsEnabled
	}
	if req.MinSeverity != "" {
		prefs.MinSeverity = req.MinSeverity
	}
	if req.RadiusKM != nil {
		prefs.RadiusKM = *req.RadiusKM
	}

	if err := s.repo.Upsert(ctx, prefs); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("[preferences.Update] upsert failed")
		return nil, err
	}
	log.Info().Str("user_id", userID).Msg("[preferences.Update] preferences updated")
	return prefs, nil
}

func (s *Service) UpdateLocation(ctx context.Context, userID string, lat, lng float64) error {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		log.Warn().Str("user_id", userID).Msg("[preferences.UpdateLocation] invalid user id")
		return ErrInvalidUser
	}
	if err := s.repo.UpdateLocation(ctx, uid, lng, lat); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("[preferences.UpdateLocation] failed")
		return err
	}
	log.Info().Str("user_id", userID).Float64("lat", lat).Float64("lng", lng).Msg("[preferences.UpdateLocation] location updated")
	return nil
}
