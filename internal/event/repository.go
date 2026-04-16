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
			"event_type":      e.EventType,
			"sub_event_type":  e.SubEventType,
			"event_root_code": e.EventRootCode,
			"severity":        e.Severity,
			"title":           e.Title,
			"description":     e.Description,
			"country":         e.Country,
			"location_name":   e.LocationName,
			"location":        e.Location,
			"num_sources":     e.NumSources,
			"num_articles":    e.NumArticles,
			"actors":          e.Actors,
			"event_date":      e.EventDate,
			"is_deleted":      false,
			"updated_at":      now,
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

type BulkUpsertResult struct {
	InsertedEvents []Event // events that were newly inserted (with _id set)
	ModifiedCount  int64
}

func (r *Repository) BulkUpsert(ctx context.Context, events []Event, source string) (*BulkUpsertResult, error) {
	if len(events) == 0 {
		return &BulkUpsertResult{}, nil
	}

	now := time.Now()
	models := make([]mongo.WriteModel, 0, len(events))

	for _, e := range events {
		filter := bson.M{"external_id": e.ExternalID, "source": source}
		update := bson.M{
			"$set": bson.M{
				"event_type":      e.EventType,
				"sub_event_type":  e.SubEventType,
				"event_root_code": e.EventRootCode,
				"severity":        e.Severity,
				"title":           e.Title,
				"description":     e.Description,
				"country":         e.Country,
				"location_name":   e.LocationName,
				"location":        e.Location,
				"num_sources":     e.NumSources,
				"num_articles":    e.NumArticles,
				"actors":          e.Actors,
				"event_date":      e.EventDate,
				"is_deleted":      false,
				"updated_at":      now,
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
		return nil, err
	}

	// Map upserted indices back to events + set their IDs
	inserted := make([]Event, 0, len(result.UpsertedIDs))
	for idx, id := range result.UpsertedIDs {
		e := events[idx]
		if oid, ok := id.(bson.ObjectID); ok {
			e.ID = oid
		}
		inserted = append(inserted, e)
	}

	return &BulkUpsertResult{
		InsertedEvents: inserted,
		ModifiedCount:  result.ModifiedCount,
	}, nil
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

func (r *Repository) Create(ctx context.Context, e *Event) error {
	now := time.Now()
	e.CreatedAt = now
	e.UpdatedAt = now
	_, err := r.col.InsertOne(ctx, e)
	return err
}

func (r *Repository) FindByReportedBy(ctx context.Context, userID bson.ObjectID, page, limit int) ([]Event, int64, error) {
	filter := bson.M{"reported_by": userID, "is_deleted": false}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * limit)
	cursor, err := r.col.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetSkip(skip).SetLimit(int64(limit)))
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	events := []Event{}
	if err := cursor.All(ctx, &events); err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (r *Repository) SoftDeleteByID(ctx context.Context, id, userID bson.ObjectID) error {
	filter := bson.M{"_id": id, "reported_by": userID, "is_deleted": false}
	update := bson.M{"$set": bson.M{"is_deleted": true, "updated_at": time.Now()}}

	result, err := r.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Find(ctx context.Context, f EventFilter) ([]Event, int64, error) {
	filter := bson.M{"is_deleted": false}

	if len(f.Severity) > 0 {
		filter["severity"] = bson.M{"$in": f.Severity}
	}
	if f.EventType != "" {
		filter["event_type"] = f.EventType
	}
	if f.Country != "" {
		filter["country"] = f.Country
	}
	if f.Source != "" {
		filter["source"] = f.Source
	}
	if f.DateFrom != nil || f.DateTo != nil {
		dateFilter := bson.M{}
		if f.DateFrom != nil {
			dateFilter["$gte"] = *f.DateFrom
		}
		if f.DateTo != nil {
			dateFilter["$lte"] = *f.DateTo
		}
		filter["event_date"] = dateFilter
	}
	if f.BBox != nil {
		filter["location"] = bson.M{
			"$geoWithin": bson.M{
				"$box": bson.A{
					bson.A{f.BBox[0], f.BBox[1]}, // [min_lng, min_lat]
					bson.A{f.BBox[2], f.BBox[3]}, // [max_lng, max_lat]
				},
			},
		}
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	var sortKey bson.D
	switch f.Sort {
	case "date_asc":
		sortKey = bson.D{{Key: "event_date", Value: 1}}
	case "severity":
		sortKey = bson.D{{Key: "severity", Value: 1}, {Key: "event_date", Value: -1}}
	default:
		sortKey = bson.D{{Key: "event_date", Value: -1}}
	}

	skip := int64((f.Page - 1) * f.Limit)
	cursor, err := r.col.Find(ctx, filter, options.Find().SetSort(sortKey).SetSkip(skip).SetLimit(int64(f.Limit)))
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	events := []Event{}
	if err := cursor.All(ctx, &events); err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (r *Repository) FindNearby(ctx context.Context, lng, lat, radiusKM float64, severity string, limit int) ([]Event, error) {
	query := bson.M{"is_deleted": false}
	if severity != "" {
		query["severity"] = severity
	}

	pipeline := mongo.Pipeline{
		{{Key: "$geoNear", Value: bson.M{
			"near":          GeoJSONPoint{Type: "Point", Coordinates: [2]float64{lng, lat}},
			"distanceField": "distance",
			"maxDistance":   radiusKM * 1000, // km to meters
			"query":         query,
			"spherical":     true,
		}}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	events := []Event{}
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}
