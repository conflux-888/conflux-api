package devicetoken

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	PlatformIOS     = "ios"
	EnvSandbox      = "sandbox"
	EnvProduction   = "production"
)

// DeviceToken represents a push notification token for a user's device.
type DeviceToken struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     bson.ObjectID `bson:"user_id" json:"user_id"`
	Token      string        `bson:"token" json:"token"`
	Platform   string        `bson:"platform" json:"platform"`
	Env        string        `bson:"env" json:"env"`
	BundleID   string        `bson:"bundle_id" json:"bundle_id"`
	LastSeenAt time.Time     `bson:"last_seen_at" json:"last_seen_at"`
	CreatedAt  time.Time     `bson:"created_at" json:"created_at"`
}

type RegisterRequest struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform" binding:"required,oneof=ios"`
	Env      string `json:"env" binding:"required,oneof=sandbox production"`
	BundleID string `json:"bundle_id" binding:"required"`
}
