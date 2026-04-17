package event

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	SourceGDELT      = "gdelt"
	SourceUserReport = "user_report"

	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
)

type GeoJSONPoint struct {
	Type        string     `bson:"type" json:"type"`
	Coordinates [2]float64 `bson:"coordinates" json:"coordinates"` // [lng, lat]
}

type Event struct {
	ID            bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	Source        string         `bson:"source" json:"source"`
	ExternalID    string         `bson:"external_id,omitempty" json:"external_id,omitempty"`
	EventType     string         `bson:"event_type" json:"event_type"`
	SubEventType  string         `bson:"sub_event_type" json:"sub_event_type"`
	EventRootCode string         `bson:"event_root_code" json:"event_root_code"`
	Severity      string         `bson:"severity" json:"severity"`
	Title         string         `bson:"title" json:"title"`
	Description   string         `bson:"description" json:"description"`
	Country       string         `bson:"country" json:"country"`
	LocationName  string         `bson:"location_name" json:"location_name"`
	Location      GeoJSONPoint   `bson:"location" json:"location"`
	NumSources    int            `bson:"num_sources" json:"num_sources"`
	NumArticles   int            `bson:"num_articles" json:"num_articles"`
	Actors        []string       `bson:"actors" json:"actors"`
	EventDate     time.Time      `bson:"event_date" json:"event_date"`
	ReportedBy    *bson.ObjectID `bson:"reported_by,omitempty" json:"reported_by,omitempty"`
	IsDeleted     bool           `bson:"is_deleted" json:"is_deleted"`
	CreatedAt     time.Time      `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time      `bson:"updated_at" json:"updated_at"`
}

type SeedAdminEventRequest struct {
	Title         string  `json:"title" binding:"required,min=3,max=200"`
	Latitude      float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude     float64 `json:"longitude" binding:"required,min=-180,max=180"`
	Severity      string  `json:"severity" binding:"required,oneof=critical high medium low"`
	Country       string  `json:"country" binding:"required,len=2"`
	LocationName  string  `json:"location_name"`
	EventType     string  `json:"event_type"`
	EventRootCode string  `json:"event_root_code" binding:"omitempty,oneof=14 15 16 17 18 19 20"`
	Description   string  `json:"description"`
	NumArticles   int     `json:"num_articles"`
}

type EventFilter struct {
	Severity  []string
	EventType string
	Country   string
	Source    string
	DateFrom  *time.Time
	DateTo    *time.Time
	BBox      *[4]float64 // [min_lng, min_lat, max_lng, max_lat]
	Page      int
	Limit     int
	Sort      string // date_desc, date_asc, severity
}
