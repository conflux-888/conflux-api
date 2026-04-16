package preferences

import (
	"context"
	"errors"
	"time"

	"github.com/conflux-888/conflux-api/internal/event"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	col := db.Collection("user_preferences")

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "last_location", Value: "2dsphere"}},
		},
	}
	col.Indexes().CreateMany(context.Background(), indexes)

	return &Repository{col: col}
}

// Get returns user's preferences or default if not yet created
func (r *Repository) Get(ctx context.Context, userID bson.ObjectID) (*UserPreferences, error) {
	var prefs UserPreferences
	err := r.col.FindOne(ctx, bson.M{"user_id": userID}).Decode(&prefs)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Default(userID), nil
	}
	return &prefs, err
}

func (r *Repository) Upsert(ctx context.Context, prefs *UserPreferences) error {
	now := time.Now()
	prefs.UpdatedAt = now

	set := bson.M{
		"notifications_enabled": prefs.NotificationsEnabled,
		"min_severity":          prefs.MinSeverity,
		"radius_km":             prefs.RadiusKM,
		"updated_at":            now,
	}
	if prefs.LastLocation != nil {
		set["last_location"] = prefs.LastLocation
		set["last_location_at"] = now
	}

	_, err := r.col.UpdateOne(ctx,
		bson.M{"user_id": prefs.UserID},
		bson.M{
			"$set": set,
			"$setOnInsert": bson.M{
				"user_id":    prefs.UserID,
				"created_at": now,
			},
		},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (r *Repository) UpdateLocation(ctx context.Context, userID bson.ObjectID, lng, lat float64) error {
	now := time.Now()
	loc := event.GeoJSONPoint{Type: "Point", Coordinates: [2]float64{lng, lat}}

	_, err := r.col.UpdateOne(ctx,
		bson.M{"user_id": userID},
		bson.M{
			"$set": bson.M{
				"last_location":    loc,
				"last_location_at": now,
				"updated_at":       now,
			},
			"$setOnInsert": bson.M{
				"user_id":               userID,
				"notifications_enabled": true,
				"min_severity":          event.SeverityCritical,
				"radius_km":             50.0,
				"created_at":            now,
			},
		},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

// FindNearbyEnabled returns users within maxDistKm of the point with notifications enabled
func (r *Repository) FindNearbyEnabled(ctx context.Context, lng, lat, maxDistKm float64) ([]UserPreferences, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$geoNear", Value: bson.M{
			"near":          event.GeoJSONPoint{Type: "Point", Coordinates: [2]float64{lng, lat}},
			"distanceField": "distance_m",
			"maxDistance":   maxDistKm * 1000,
			"query":         bson.M{"notifications_enabled": true},
			"spherical":     true,
		}}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	prefs := []UserPreferences{}
	if err := cursor.All(ctx, &prefs); err != nil {
		return nil, err
	}
	return prefs, nil
}

// FindAllEnabled returns all users with notifications enabled (for daily briefing broadcasts)
func (r *Repository) FindAllEnabled(ctx context.Context) ([]UserPreferences, error) {
	cursor, err := r.col.Find(ctx, bson.M{"notifications_enabled": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	prefs := []UserPreferences{}
	if err := cursor.All(ctx, &prefs); err != nil {
		return nil, err
	}
	return prefs, nil
}
