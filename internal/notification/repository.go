package notification

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrNotFound = errors.New("notification not found")

type Repository struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	col := db.Collection("notifications")

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "read_at", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "event_id", Value: 1}},
		},
		{
			// TTL: auto-delete notifications after 30 days
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(30 * 24 * 3600),
		},
	}
	col.Indexes().CreateMany(context.Background(), indexes)

	return &Repository{col: col}
}

func (r *Repository) BulkCreate(ctx context.Context, notifs []Notification) error {
	if len(notifs) == 0 {
		return nil
	}
	docs := make([]any, len(notifs))
	now := time.Now()
	for i := range notifs {
		notifs[i].CreatedAt = now
		docs[i] = notifs[i]
	}
	_, err := r.col.InsertMany(ctx, docs)
	return err
}

func (r *Repository) Create(ctx context.Context, n *Notification) error {
	n.CreatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, n)
	return err
}

func (r *Repository) FindByUser(ctx context.Context, userID bson.ObjectID, unreadOnly bool, since *time.Time, page, limit int) ([]Notification, int64, error) {
	filter := bson.M{"user_id": userID}
	if unreadOnly {
		filter["read_at"] = bson.M{"$exists": false}
	}
	if since != nil {
		filter["created_at"] = bson.M{"$gt": *since}
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * limit)
	cursor, err := r.col.Find(ctx, filter,
		options.Find().
			SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(skip).
			SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	notifs := []Notification{}
	if err := cursor.All(ctx, &notifs); err != nil {
		return nil, 0, err
	}
	return notifs, total, nil
}

func (r *Repository) CountUnread(ctx context.Context, userID bson.ObjectID) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{
		"user_id": userID,
		"read_at": bson.M{"$exists": false},
	})
}

func (r *Repository) MarkRead(ctx context.Context, id, userID bson.ObjectID) error {
	result, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id, "user_id": userID},
		bson.M{"$set": bson.M{"read_at": time.Now()}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) MarkAllRead(ctx context.Context, userID bson.ObjectID) (int64, error) {
	result, err := r.col.UpdateMany(ctx,
		bson.M{"user_id": userID, "read_at": bson.M{"$exists": false}},
		bson.M{"$set": bson.M{"read_at": time.Now()}},
	)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}

// ExistsForUserAndEvent checks if a notification was already sent for a user/event pair (dedup)
func (r *Repository) ExistsForUserAndEvent(ctx context.Context, userID, eventID bson.ObjectID) (bool, error) {
	count, err := r.col.CountDocuments(ctx,
		bson.M{"user_id": userID, "event_id": eventID},
		options.Count().SetLimit(1),
	)
	return count > 0, err
}
