package event

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrNotFound = errors.New("event not found")

type Repository struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	col := db.Collection("events")

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "external_id", Value: 1}, {Key: "source", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys: bson.D{{Key: "location", Value: "2dsphere"}},
		},
		{
			Keys: bson.D{{Key: "severity", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "event_date", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "country", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "reported_by", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
	}

	col.Indexes().CreateMany(context.Background(), indexes)

	return &Repository{col: col}
}

func (r *Repository) UpsertByExternalID(ctx context.Context, externalID, source string, e *Event) error {
	now := time.Now()
	filter := bson.M{"external_id": externalID, "source": source}
	update := bson.M{
		"$set": bson.M{
			"event_type":     e.EventType,
			"sub_event_type":  e.SubEventType,
			"event_root_code": e.EventRootCode,
			"severity":        e.Severity,
			"title":          e.Title,
			"description":    e.Description,
			"country":        e.Country,
			"location_name":  e.LocationName,
			"location":       e.Location,
			"num_sources":    e.NumSources,
			"num_articles":   e.NumArticles,
			"actors":         e.Actors,
			"event_date":     e.EventDate,
			"is_deleted":     false,
			"updated_at":     now,
		},
		"$setOnInsert": bson.M{
			"external_id": externalID,
			"source":      source,
			"created_at":  now,
		},
	}

	_, err := r.col.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
	return err
}

func (r *Repository) BulkUpsert(ctx context.Context, events []Event, source string) (int64, error) {
	if len(events) == 0 {
		return 0, nil
	}

	now := time.Now()
	models := make([]mongo.WriteModel, 0, len(events))

	for _, e := range events {
		filter := bson.M{"external_id": e.ExternalID, "source": source}
		update := bson.M{
			"$set": bson.M{
				"event_type":     e.EventType,
				"sub_event_type":  e.SubEventType,
				"event_root_code": e.EventRootCode,
				"severity":        e.Severity,
				"title":          e.Title,
				"description":    e.Description,
				"country":        e.Country,
				"location_name":  e.LocationName,
				"location":       e.Location,
				"num_sources":    e.NumSources,
				"num_articles":   e.NumArticles,
				"actors":         e.Actors,
				"event_date":     e.EventDate,
				"is_deleted":     false,
				"updated_at":     now,
			},
			"$setOnInsert": bson.M{
				"external_id": e.ExternalID,
				"source":      source,
				"created_at":  now,
			},
		}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}

	result, err := r.col.BulkWrite(ctx, models)
	if err != nil {
		return 0, err
	}
	return result.UpsertedCount + result.ModifiedCount, nil
}

func (r *Repository) SoftDeleteByExternalIDs(ctx context.Context, externalIDs []string, source string) (int64, error) {
	if len(externalIDs) == 0 {
		return 0, nil
	}

	filter := bson.M{
		"external_id": bson.M{"$in": externalIDs},
		"source":      source,
	}
	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"updated_at": time.Now(),
		},
	}

	result, err := r.col.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}

func (r *Repository) FindByID(ctx context.Context, id bson.ObjectID) (*Event, error) {
	var e Event
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&e)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	return &e, err
}
