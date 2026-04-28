package devicetoken

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrNotFound = errors.New("device token not found")

type Repository struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	col := db.Collection("device_tokens")

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
	}
	col.Indexes().CreateMany(context.Background(), indexes)

	return &Repository{col: col}
}

func (r *Repository) Upsert(ctx context.Context, t *DeviceToken) error {
	now := time.Now()
	t.LastSeenAt = now
	_, err := r.col.UpdateOne(ctx,
		bson.M{"token": t.Token},
		bson.M{
			"$set": bson.M{
				"user_id":      t.UserID,
				"platform":     t.Platform,
				"env":          t.Env,
				"bundle_id":    t.BundleID,
				"last_seen_at": now,
			},
			"$setOnInsert": bson.M{
				"token":      t.Token,
				"created_at": now,
			},
		},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (r *Repository) DeleteByToken(ctx context.Context, token string) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"token": token})
	return err
}

func (r *Repository) DeleteByUserAndToken(ctx context.Context, userID bson.ObjectID, token string) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"user_id": userID, "token": token})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) FindByUser(ctx context.Context, userID bson.ObjectID) ([]DeviceToken, error) {
	cursor, err := r.col.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	tokens := []DeviceToken{}
	if err := cursor.All(ctx, &tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *Repository) FindByUsers(ctx context.Context, userIDs []bson.ObjectID) ([]DeviceToken, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	cursor, err := r.col.Find(ctx, bson.M{"user_id": bson.M{"$in": userIDs}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	tokens := []DeviceToken{}
	if err := cursor.All(ctx, &tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}
