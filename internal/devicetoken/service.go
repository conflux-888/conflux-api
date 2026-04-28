package devicetoken

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Repository() *Repository {
	return s.repo
}

func (s *Service) Register(ctx context.Context, userID string, req RegisterRequest) error {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return ErrNotFound
	}
	return s.repo.Upsert(ctx, &DeviceToken{
		UserID:   uid,
		Token:    req.Token,
		Platform: req.Platform,
		Env:      req.Env,
		BundleID: req.BundleID,
	})
}

func (s *Service) Unregister(ctx context.Context, userID, token string) error {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return ErrNotFound
	}
	return s.repo.DeleteByUserAndToken(ctx, uid, token)
}
