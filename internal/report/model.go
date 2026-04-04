package report

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/conflux-888/conflux-api/internal/event"
)

type CreateReportRequest struct {
	EventType    string  `json:"event_type" binding:"required,oneof=armed_conflict use_of_force explosion terrorism civil_unrest other"`
	Severity     string  `json:"severity" binding:"required,oneof=critical high medium low"`
	Title        string  `json:"title" binding:"required,max=200"`
	Description  string  `json:"description" binding:"max=2000"`
	Latitude     float64 `json:"latitude" binding:"required"`
	Longitude    float64 `json:"longitude" binding:"required"`
	LocationName string  `json:"location_name"`
	Country      string  `json:"country" binding:"required"`
}

type ReportCluster struct {
	ID              bson.ObjectID      `bson:"_id,omitempty" json:"id"`
	EventType       string             `bson:"event_type" json:"event_type"`
	Severity        string             `bson:"severity" json:"severity"`
	Center          event.GeoJSONPoint `bson:"center" json:"center"`
	ReportIDs       []bson.ObjectID    `bson:"report_ids" json:"report_ids"`
	ReportCount     int                `bson:"report_count" json:"report_count"`
	FirstReportedAt time.Time          `bson:"first_reported_at" json:"first_reported_at"`
	LastReportedAt  time.Time          `bson:"last_reported_at" json:"last_reported_at"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

var severityRank = map[string]int{
	"low":      0,
	"medium":   1,
	"high":     2,
	"critical": 3,
}

func HigherSeverity(a, b string) string {
	if severityRank[a] >= severityRank[b] {
		return a
	}
	return b
}
