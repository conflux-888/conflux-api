package summary

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrNotFound = errors.New("summary not found")

type Repository struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	col := db.Collection("daily_summaries")

	col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "summary_date", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	return &Repository{col: col}
}

func (r *Repository) Upsert(ctx context.Context, s *DailySummary) error {
	now := time.Now()
	s.UpdatedAt = now

	_, err := r.col.UpdateOne(
		ctx,
		bson.M{"summary_date": s.SummaryDate},
		bson.M{
			"$set":         s,
			"$setOnInsert": bson.M{"created_at": now},
		},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (r *Repository) FindByDate(ctx context.Context, date string) (*DailySummary, error) {
	var s DailySummary
	err := r.col.FindOne(ctx, bson.M{"summary_date": date}).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	return &s, err
}

func (r *Repository) FindByDateRange(ctx context.Context, from, to string, page, limit int) ([]DailySummary, int64, error) {
	filter := bson.M{
		"summary_date": bson.M{"$gte": from, "$lte": to},
		"status":       "completed",
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * limit)
	cursor, err := r.col.Find(ctx, filter,
		options.Find().
			SetSort(bson.D{{Key: "summary_date", Value: -1}}).
			SetSkip(skip).
			SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var summaries []DailySummary
	if err := cursor.All(ctx, &summaries); err != nil {
		return nil, 0, err
	}
	return summaries, total, nil
}

func (r *Repository) FindLatest(ctx context.Context, limit int) ([]DailySummary, error) {
	cursor, err := r.col.Find(ctx,
		bson.M{"status": "completed"},
		options.Find().
			SetSort(bson.D{{Key: "summary_date", Value: -1}}).
			SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var summaries []DailySummary
	if err := cursor.All(ctx, &summaries); err != nil {
		return nil, err
	}
	return summaries, nil
}
