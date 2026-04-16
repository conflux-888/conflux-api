package report

import (
	"context"
	"time"

	"github.com/conflux-888/conflux-api/internal/common/response"
	"github.com/conflux-888/conflux-api/internal/event"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	eventRepo *event.Repository
}

func NewService(eventRepo *event.Repository) *Service {
	return &Service{eventRepo: eventRepo}
}

func (s *Service) SubmitReport(ctx context.Context, userID string, req CreateReportRequest) (*event.Event, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		log.Warn().Str("user_id", userID).Msg("[report.SubmitReport] invalid user id")
		return nil, err
	}

	e := &event.Event{
		Source:       event.SourceUserReport,
		EventType:    req.EventType,
		Severity:     req.Severity,
		Title:        req.Title,
		Description:  req.Description,
		Country:      req.Country,
		LocationName: req.LocationName,
		Location: event.GeoJSONPoint{
			Type:        "Point",
			Coordinates: [2]float64{req.Longitude, req.Latitude},
		},
		EventDate:  time.Now(),
		ReportedBy: &uid,
		IsDeleted:  false,
	}

	if err := s.eventRepo.Create(ctx, e); err != nil {
		log.Error().Err(err).Msg("[report.SubmitReport] failed to create event")
		return nil, err
	}

	log.Info().Str("event_id", e.ID.Hex()).Str("user_id", userID).Msg("[report.SubmitReport] report created")
	return e, nil
}

func (s *Service) GetMyReports(ctx context.Context, userID string, page, limit int) ([]event.Event, *response.Pagination, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		log.Warn().Str("user_id", userID).Msg("[report.GetMyReports] invalid user id")
		return nil, nil, err
	}

	events, total, err := s.eventRepo.FindByReportedBy(ctx, uid, page, limit)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("[report.GetMyReports] failed to query reports")
		return nil, nil, err
	}

	log.Info().Str("user_id", userID).Int("count", len(events)).Int64("total", total).Msg("[report.GetMyReports] reports listed")

	return events, &response.Pagination{Page: page, Limit: limit, Total: total}, nil
}

func (s *Service) DeleteMyReport(ctx context.Context, userID, eventID string) error {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		log.Warn().Str("user_id", userID).Msg("[report.DeleteMyReport] invalid user id")
		return event.ErrNotFound
	}
	eid, err := bson.ObjectIDFromHex(eventID)
	if err != nil {
		log.Warn().Str("event_id", eventID).Msg("[report.DeleteMyReport] invalid event id")
		return event.ErrNotFound
	}

	if err := s.eventRepo.SoftDeleteByID(ctx, eid, uid); err != nil {
		log.Error().Err(err).Str("user_id", userID).Str("event_id", eventID).Msg("[report.DeleteMyReport] failed to delete report")
		return err
	}

	log.Info().Str("user_id", userID).Str("event_id", eventID).Msg("[report.DeleteMyReport] report deleted")
	return nil
}
