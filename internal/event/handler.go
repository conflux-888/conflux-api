package event

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/conflux-888/conflux-api/internal/common/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) HandleListEvents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}

	filter := EventFilter{
		EventType: c.Query("event_type"),
		Country:   c.Query("country"),
		Source:    c.Query("source"),
		Sort:      c.DefaultQuery("sort", "date_desc"),
		Page:      page,
		Limit:     limit,
	}

	if s := c.Query("severity"); s != "" {
		filter.Severity = strings.Split(s, ",")
	}

	if df := c.Query("date_from"); df != "" {
		if t, err := time.Parse("2006-01-02", df); err == nil {
			filter.DateFrom = &t
		}
	}
	if dt := c.Query("date_to"); dt != "" {
		if t, err := time.Parse("2006-01-02", dt); err == nil {
			filter.DateTo = &t
		}
	}

	if bbox := c.Query("bbox"); bbox != "" {
		parts := strings.Split(bbox, ",")
		if len(parts) == 4 {
			var box [4]float64
			valid := true
			for i, p := range parts {
				v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
				if err != nil {
					valid = false
					break
				}
				box[i] = v
			}
			if valid {
				filter.BBox = &box
			}
		}
	}

	events, total, err := h.svc.ListEvents(c.Request.Context(), filter)
	if err != nil {
		log.Error().Err(err).Msg("[event.HandleListEvents] unexpected error")
		response.InternalError(c)
		return
	}

	response.List(c, events, response.Pagination{
		Page:  page,
		Limit: limit,
		Total: total,
	})
}

func (h *Handler) HandleGetEvent(c *gin.Context) {
	id := c.Param("id")

	e, err := h.svc.GetEvent(c.Request.Context(), id)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "event not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("[event.HandleGetEvent] unexpected error")
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, e)
}

func (h *Handler) HandleGetNearby(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	if latStr == "" || lngStr == "" {
		response.ValidationError(c, "lat and lng are required")
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		response.ValidationError(c, "invalid lat")
		return
	}
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		response.ValidationError(c, "invalid lng")
		return
	}

	radiusKM, _ := strconv.ParseFloat(c.DefaultQuery("radius_km", "50"), 64)
	if radiusKM <= 0 || radiusKM > 500 {
		radiusKM = 50
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	severity := c.Query("severity")

	events, err := h.svc.GetNearbyEvents(c.Request.Context(), lng, lat, radiusKM, severity, limit)
	if err != nil {
		log.Error().Err(err).Msg("[event.HandleGetNearby] unexpected error")
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, events)
}
