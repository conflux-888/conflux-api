package notification

import (
	"context"
	"errors"
	"os"

	"github.com/conflux-888/conflux-api/internal/devicetoken"
	"github.com/rs/zerolog/log"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// APNSConfig holds Apple Push credentials. KeyPath should point to the .p8 file.
type APNSConfig struct {
	KeyPath    string
	KeyID      string
	TeamID     string
	BundleID   string
	Production bool
}

func (c APNSConfig) Enabled() bool {
	return c.KeyPath != "" && c.KeyID != "" && c.TeamID != "" && c.BundleID != ""
}

type APNSPusher struct {
	cfg    APNSConfig
	client *apns2.Client
	repo   *devicetoken.Repository
}

func NewAPNSPusher(cfg APNSConfig, repo *devicetoken.Repository) (*APNSPusher, error) {
	if !cfg.Enabled() {
		return nil, errors.New("apns config incomplete")
	}
	keyBytes, err := os.ReadFile(cfg.KeyPath)
	if err != nil {
		return nil, err
	}
	authKey, err := token.AuthKeyFromBytes(keyBytes)
	if err != nil {
		return nil, err
	}
	tk := &token.Token{
		AuthKey: authKey,
		KeyID:   cfg.KeyID,
		TeamID:  cfg.TeamID,
	}
	client := apns2.NewTokenClient(tk)
	if cfg.Production {
		client.Production()
	} else {
		client.Development()
	}
	return &APNSPusher{cfg: cfg, client: client, repo: repo}, nil
}

// Push sends a notification to all device tokens for the given users.
// Stale tokens (BadDeviceToken / Unregistered) are pruned automatically.
func (p *APNSPusher) Push(ctx context.Context, userIDs []bson.ObjectID, title, body string, data map[string]interface{}, badge *int) {
	if p == nil || p.client == nil {
		return
	}
	if len(userIDs) == 0 {
		return
	}
	tokens, err := p.repo.FindByUsers(ctx, userIDs)
	if err != nil {
		log.Error().Err(err).Msg("[notification.APNSPusher.Push] failed to load tokens")
		return
	}
	if len(tokens) == 0 {
		return
	}

	pl := payload.NewPayload().AlertTitle(title).AlertBody(body).Sound("default")
	if badge != nil {
		pl = pl.Badge(*badge)
	}
	for k, v := range data {
		pl = pl.Custom(k, v)
	}

	sent, dropped := 0, 0
	for _, t := range tokens {
		// APNs sandbox vs production environments must match the build that registered the token.
		// We trust the env stored on the token; if it disagrees with our pusher's environment we skip it.
		if p.cfg.Production && t.Env != devicetoken.EnvProduction {
			continue
		}
		if !p.cfg.Production && t.Env != devicetoken.EnvSandbox {
			continue
		}

		notif := &apns2.Notification{
			DeviceToken: t.Token,
			Topic:       p.cfg.BundleID,
			Payload:     pl,
		}
		res, err := p.client.PushWithContext(ctx, notif)
		if err != nil {
			log.Warn().Err(err).Str("token", t.Token).Msg("[notification.APNSPusher.Push] send error")
			continue
		}
		if res.Sent() {
			sent++
			continue
		}
		if res.StatusCode == 410 || res.Reason == "BadDeviceToken" || res.Reason == "Unregistered" {
			if err := p.repo.DeleteByToken(ctx, t.Token); err != nil {
				log.Warn().Err(err).Str("token", t.Token).Msg("[notification.APNSPusher.Push] failed to delete stale token")
			} else {
				dropped++
			}
			continue
		}
		log.Warn().Int("status", res.StatusCode).Str("reason", res.Reason).Str("token", t.Token).Msg("[notification.APNSPusher.Push] non-success response")
	}

	log.Info().Int("sent", sent).Int("dropped", dropped).Int("total", len(tokens)).Msg("[notification.APNSPusher.Push] done")
}
