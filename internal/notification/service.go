package notification

import (
	"context"
	"fmt"
	"math"

	"github.com/conflux-888/conflux-api/internal/common/response"
	"github.com/conflux-888/conflux-api/internal/event"
	"github.com/conflux-888/conflux-api/internal/preferences"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var severityRank = map[string]int{
	"low":      0,
	"medium":   1,
	"high":     2,
	"critical": 3,
}

type Service struct {
	repo      *Repository
	prefsRepo *preferences.Repository
}

func NewService(repo *Repository, prefsRepo *preferences.Repository) *Service {
	return &Service{repo: repo, prefsRepo: prefsRepo}
}

// NotifyNearbyCritical is called by sync after new events are inserted.
// Creates notifications for users within their configured radius of new critical events.
func (s *Service) NotifyNearbyCritical(ctx context.Context, events []event.Event) {
	log.Info().Int("events", len(events)).Msg("[notification.NotifyNearbyCritical] processing new events")

	created := 0
	for _, e := range events {
		if e.Severity != event.SeverityCritical {
			continue
		}
		if len(e.Location.Coordinates) != 2 {
			continue
		}
		lng := e.Location.Coordinates[0]
		lat := e.Location.Coordinates[1]

		// Find any user within 500km (covers max radius preference)
		candidates, err := s.prefsRepo.FindNearbyEnabled(ctx, lng, lat, 500)
		if err != nil {
			log.Error().Err(err).Str("event_id", e.ID.Hex()).Msg("[notification.NotifyNearbyCritical] failed to find candidates")
			continue
		}

		notifs := []Notification{}
		for _, pref := range candidates {
			if pref.LastLocation == nil {
				continue
			}
			if severityRank[e.Severity] < severityRank[pref.MinSeverity] {
				continue
			}

			userLng := pref.LastLocation.Coordinates[0]
			userLat := pref.LastLocation.Coordinates[1]
			distKm := haversineKM(lat, lng, userLat, userLng)
			if distKm > pref.RadiusKM {
				continue
			}

			exists, err := s.repo.ExistsForUserAndEvent(ctx, pref.UserID, e.ID)
			if err != nil || exists {
				continue
			}

			eventIDCopy := e.ID
			notifs = append(notifs, Notification{
				UserID:     pref.UserID,
				Type:       TypeCriticalNearby,
				Title:      fmt.Sprintf("Critical threat %.1fkm from you", distKm),
				Body:       e.Title,
				EventID:    &eventIDCopy,
				DistanceKM: distKm,
			})
		}

		if len(notifs) > 0 {
			if err := s.repo.BulkCreate(ctx, notifs); err != nil {
				log.Error().Err(err).Msg("[notification.NotifyNearbyCritical] failed to bulk create")
				continue
			}
			created += len(notifs)
		}
	}

	log.Info().Int("created", created).Msg("[notification.NotifyNearbyCritical] done")
}

// NotifyDailyBriefing broadcasts a notification to all users when a daily summary is completed.
func (s *Service) NotifyDailyBriefing(ctx context.Context, summaryDate, title string) {
	log.Info().Str("date", summaryDate).Msg("[notification.NotifyDailyBriefing] creating notifications")

	users, err := s.prefsRepo.FindAllEnabled(ctx)
	if err != nil {
		log.Error().Err(err).Msg("[notification.NotifyDailyBriefing] failed to find users")
		return
	}

	notifs := make([]Notification, 0, len(users))
	for _, u := range users {
		notifs = append(notifs, Notification{
			UserID:      u.UserID,
			Type:        TypeDailyBriefing,
			Title:       "Your daily conflict briefing is ready",
			Body:        title,
			SummaryDate: summaryDate,
		})
	}

	if err := s.repo.BulkCreate(ctx, notifs); err != nil {
		log.Error().Err(err).Msg("[notification.NotifyDailyBriefing] failed to bulk create")
		return
	}
	log.Info().Int("created", len(notifs)).Msg("[notification.NotifyDailyBriefing] done")
}

// User-facing methods

func (s *Service) GetMyNotifications(ctx context.Context, userID string, unreadOnly bool, page, limit int) ([]Notification, *response.Pagination, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, nil, ErrNotFound
	}
	notifs, total, err := s.repo.FindByUser(ctx, uid, unreadOnly, nil, page, limit)
	if err != nil {
		return nil, nil, err
	}
	return notifs, &response.Pagination{Page: page, Limit: limit, Total: total}, nil
}

func (s *Service) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return 0, ErrNotFound
	}
	return s.repo.CountUnread(ctx, uid)
}

func (s *Service) MarkRead(ctx context.Context, userID, notifID string) error {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return ErrNotFound
	}
	nid, err := bson.ObjectIDFromHex(notifID)
	if err != nil {
		return ErrNotFound
	}
	return s.repo.MarkRead(ctx, nid, uid)
}

func (s *Service) MarkAllRead(ctx context.Context, userID string) (int64, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return 0, ErrNotFound
	}
	return s.repo.MarkAllRead(ctx, uid)
}

// haversineKM returns the great-circle distance between two points in kilometers
func haversineKM(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	toRad := func(d float64) float64 { return d * math.Pi / 180 }

	dLat := toRad(lat2 - lat1)
	dLng := toRad(lng2 - lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
