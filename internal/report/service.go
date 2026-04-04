package report

import (
	"context"
	"errors"
	"time"

	"github.com/conflux-888/conflux-api/internal/common/response"
	"github.com/conflux-888/conflux-api/internal/event"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	clusterRadiusMeters = 5000           // 5 km
	clusterTimeWindow   = 24 * time.Hour // 24 hours
)

var ErrNotOwner = errors.New("not the owner of this report")

type Service struct {
	eventRepo   *event.Repository
	clusterRepo *Repository
}

func NewService(eventRepo *event.Repository, clusterRepo *Repository) *Service {
	return &Service{eventRepo: eventRepo, clusterRepo: clusterRepo}
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

	// Clustering
	s.handleClustering(ctx, e)

	return e, nil
}

func (s *Service) handleClustering(ctx context.Context, e *event.Event) {
	since := time.Now().Add(-clusterTimeWindow)
	lng := e.Location.Coordinates[0]
	lat := e.Location.Coordinates[1]

	cluster, err := s.clusterRepo.FindNearbyCluster(ctx, e.EventType, lng, lat, clusterRadiusMeters, since)
	if errors.Is(err, ErrClusterNotFound) {
		// Create new cluster
		newCluster := &ReportCluster{
			EventType:       e.EventType,
			Severity:        e.Severity,
			Center:          e.Location,
			ReportIDs:       []bson.ObjectID{e.ID},
			ReportCount:     1,
			FirstReportedAt: time.Now(),
			LastReportedAt:  time.Now(),
		}
		if err := s.clusterRepo.CreateCluster(ctx, newCluster); err != nil {
			log.Error().Err(err).Msg("[report.handleClustering] failed to create cluster")
		}
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("[report.handleClustering] failed to find nearby cluster")
		return
	}

	// Add to existing cluster
	if err := s.clusterRepo.AddToCluster(ctx, cluster.ID, e.ID, e.Severity, lng, lat); err != nil {
		log.Error().Err(err).Msg("[report.handleClustering] failed to add to cluster")
	} else {
		log.Info().Str("cluster_id", cluster.ID.Hex()).Int("report_count", cluster.ReportCount+1).Msg("[report.handleClustering] added to cluster")
	}
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

	pagination := &response.Pagination{
		Page:  page,
		Limit: limit,
		Total: total,
	}

	return events, pagination, nil
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
