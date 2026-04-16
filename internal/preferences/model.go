package preferences

import (
	"time"

	"github.com/conflux-888/conflux-api/internal/event"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// UserPreferences represents a user's notification settings and last known location
type UserPreferences struct {
	ID                   bson.ObjectID       `bson:"_id,omitempty" json:"id"`
	UserID               bson.ObjectID       `bson:"user_id" json:"user_id"`
	NotificationsEnabled bool                `bson:"notifications_enabled" json:"notifications_enabled"`
	MinSeverity          string              `bson:"min_severity" json:"min_severity"`
	RadiusKM             float64             `bson:"radius_km" json:"radius_km"`
	LastLocation         *event.GeoJSONPoint `bson:"last_location,omitempty" json:"last_location,omitempty"`
	LastLocationAt       *time.Time          `bson:"last_location_at,omitempty" json:"last_location_at,omitempty"`
	CreatedAt            time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time           `bson:"updated_at" json:"updated_at"`
}

// Default preferences for new users
func Default(userID bson.ObjectID) *UserPreferences {
	now := time.Now()
	return &UserPreferences{
		UserID:               userID,
		NotificationsEnabled: true,
		MinSeverity:          event.SeverityCritical,
		RadiusKM:             50,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// Request DTOs
type UpdateRequest struct {
	NotificationsEnabled *bool    `json:"notifications_enabled"`
	MinSeverity          string   `json:"min_severity" binding:"omitempty,oneof=critical high medium low"`
	RadiusKM             *float64 `json:"radius_km" binding:"omitempty,min=1,max=500"`
}

type UpdateLocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}
