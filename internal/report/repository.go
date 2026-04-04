package report

import (
	"context"
	"errors"
	"time"

	"github.com/conflux-888/conflux-api/internal/event"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrClusterNotFound = errors.New("cluster not found")

type Repository struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	col := db.Collection("report_clusters")

	col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{{Key: "center", Value: "2dsphere"}},
	})

	return &Repository{col: col}
}

func (r *Repository) FindNearbyCluster(ctx context.Context, eventType string, lng, lat float64, maxDistMeters float64, since time.Time) (*ReportCluster, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$geoNear", Value: bson.M{
			"near":          event.GeoJSONPoint{Type: "Point", Coordinates: [2]float64{lng, lat}},
			"distanceField": "dist",
			"maxDistance":   maxDistMeters,
			"query": bson.M{
				"event_type":       eventType,
				"last_reported_at": bson.M{"$gte": since},
			},
			"spherical": true,
		}}},
		{{Key: "$limit", Value: 1}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return nil, ErrClusterNotFound
	}

	var cluster ReportCluster
	if err := cursor.Decode(&cluster); err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *Repository) CreateCluster(ctx context.Context, cluster *ReportCluster) error {
	now := time.Now()
	cluster.CreatedAt = now
	cluster.UpdatedAt = now
	_, err := r.col.InsertOne(ctx, cluster)
	return err
}

func (r *Repository) AddToCluster(ctx context.Context, clusterID, eventID bson.ObjectID, severity string, lng, lat float64) error {
	// Get current cluster to calculate weighted center
	var cluster ReportCluster
	err := r.col.FindOne(ctx, bson.M{"_id": clusterID}).Decode(&cluster)
	if err != nil {
		return err
	}

	// Weighted average of coordinates
	count := float64(cluster.ReportCount)
	newLng := (cluster.Center.Coordinates[0]*count + lng) / (count + 1)
	newLat := (cluster.Center.Coordinates[1]*count + lat) / (count + 1)

	now := time.Now()
	update := bson.M{
		"$push": bson.M{"report_ids": eventID},
		"$inc":  bson.M{"report_count": 1},
		"$set": bson.M{
			"center": event.GeoJSONPoint{
				Type:        "Point",
				Coordinates: [2]float64{newLng, newLat},
			},
			"severity":         HigherSeverity(cluster.Severity, severity),
			"last_reported_at": now,
			"updated_at":       now,
		},
	}

	_, err = r.col.UpdateByID(ctx, clusterID, update)
	return err
}
