package sync

import "time"

// SyncState represents the sync_state collection document
type SyncState struct {
	ID                string    `bson:"_id" json:"id"`
	LastSyncTimestamp string    `bson:"last_sync_timestamp" json:"last_sync_timestamp"` // GDELT DATEADDED (YYYYMMDDHHmmSS)
	LastSyncAt        time.Time `bson:"last_sync_at" json:"last_sync_at"`
	Status            string    `bson:"status" json:"status"`
	EventsSynced      int       `bson:"events_synced" json:"events_synced"`
	ErrorMessage      string    `bson:"error_message,omitempty" json:"error_message,omitempty"`
}

// GDELTEvent represents a parsed row from the GDELT export CSV
type GDELTEvent struct {
	GlobalEventID       string
	Day                 string  // YYYYMMDD
	Actor1Name          string
	Actor2Name          string
	IsRootEvent         string
	EventCode           string  // Full CAMEO code
	EventBaseCode       string
	EventRootCode       string  // 14-20 = conflict
	QuadClass           string  // 4 = Material Conflict
	GoldsteinScale      float64
	NumMentions         int
	NumSources          int
	NumArticles         int
	AvgTone             float64
	ActionGeoType       string
	ActionGeoFullName   string
	ActionGeoCountryCode string // FIPS 2-char
	ActionGeoLat        float64
	ActionGeoLong       float64
	DateAdded           string  // YYYYMMDDHHmmSS
	SourceURL           string
}
