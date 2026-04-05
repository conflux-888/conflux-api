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

// HandleListEvents godoc
// @Summary      List events
// @Description  Returns paginated, filterable list of conflict events (GDELT + user reports)
// @Tags         events
// @Produce      json
// @Param        severity    query  string  false  "Comma-separated: critical,high,medium,low"
// @Param        event_type  query  string  false  "Filter by event type"
// @Param        country     query  string  false  "Filter by country code"
// @Param        source      query  string  false  "gdelt or user_report"
// @Param        date_from   query  string  false  "Start date (YYYY-MM-DD)"
// @Param        date_to     query  string  false  "End date (YYYY-MM-DD)"
// @Param        bbox        query  string  false  "Bounding box: min_lng,min_lat,max_lng,max_lat"
// @Param        page        query  int     false  "Page number"     default(1)
// @Param        limit       query  int     false  "Items per page"  default(50)
// @Param        sort        query  string  false  "date_desc, date_asc, severity"  default(date_desc)
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /events [get]
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

// HandleGetEvent godoc
// @Summary      Get event by ID
// @Tags         events
// @Produce      json
// @Param        id   path  string  true  "Event ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /events/{id} [get]
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

// HandleGetNearby godoc
// @Summary      Find nearby events
// @Description  Returns events within a radius of a geographic point
// @Tags         events
// @Produce      json
// @Param        lat        query  number  true   "Latitude"
// @Param        lng        query  number  true   "Longitude"
// @Param        radius_km  query  number  false  "Radius in km (max 500)"  default(50)
// @Param        severity   query  string  false  "Filter by severity"
// @Param        limit      query  int     false  "Max results (max 100)"   default(20)
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /events/nearby [get]
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
