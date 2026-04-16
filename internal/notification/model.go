package notification

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	TypeCriticalNearby = "critical_nearby"
	TypeDailyBriefing  = "daily_briefing"
)

// Notification represents an in-app notification for a user
type Notification struct {
	ID          bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID      bson.ObjectID  `bson:"user_id" json:"user_id"`
	Type        string         `bson:"type" json:"type"`
	Title       string         `bson:"title" json:"title"`
	Body        string         `bson:"body" json:"body"`
	EventID     *bson.ObjectID `bson:"event_id,omitempty" json:"event_id,omitempty"`
	SummaryDate string         `bson:"summary_date,omitempty" json:"summary_date,omitempty"`
	DistanceKM  float64        `bson:"distance_km,omitempty" json:"distance_km,omitempty"`
	ReadAt      *time.Time     `bson:"read_at,omitempty" json:"read_at,omitempty"`
	CreatedAt   time.Time      `bson:"created_at" json:"created_at"`
}
