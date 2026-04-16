package summary

import (
	"errors"
	"net/http"
	"strconv"
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

// HandleListSummaries godoc
// @Summary      List daily summaries
// @Description  Returns paginated list of daily conflict summaries
// @Tags         summaries
// @Produce      json
// @Param        from   query  string  false  "Start date (YYYY-MM-DD)"
// @Param        to     query  string  false  "End date (YYYY-MM-DD)"
// @Param        page   query  int     false  "Page number"     default(1)
// @Param        limit  query  int     false  "Items per page"  default(7)
// @Success      200    {object}  map[string]interface{}
// @Failure      401    {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /summaries [get]
func (h *Handler) HandleListSummaries(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "7"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 30 {
		limit = 7
	}

	// Default: last 7 days
	to := c.DefaultQuery("to", time.Now().UTC().Format("2006-01-02"))
	from := c.DefaultQuery("from", time.Now().UTC().AddDate(0, 0, -7).Format("2006-01-02"))

	summaries, pagination, err := h.svc.ListSummaries(c.Request.Context(), from, to, page, limit)
	if err != nil {
		log.Error().Err(err).Msg("[summary.HandleListSummaries] unexpected error")
		response.InternalError(c)
		return
	}

	response.List(c, summaries, *pagination)
}

// HandleGetLatestSummary godoc
// @Summary      Get latest summary
// @Description  Returns the most recent completed daily summary
// @Tags         summaries
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /summaries/latest [get]
func (h *Handler) HandleGetLatestSummary(c *gin.Context) {
	s, err := h.svc.GetLatestSummary(c.Request.Context())
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "no summaries available")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("[summary.HandleGetLatestSummary] unexpected error")
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, s)
}

// HandleGetSummary godoc
// @Summary      Get summary by date
// @Description  Returns the daily summary for a specific date
// @Tags         summaries
// @Produce      json
// @Param        date  path  string  true  "Date (YYYY-MM-DD)"
// @Success      200   {object}  map[string]interface{}
// @Failure      401   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /summaries/{date} [get]
func (h *Handler) HandleGetSummary(c *gin.Context) {
	date := c.Param("date")

	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		response.ValidationError(c, "invalid date format, use YYYY-MM-DD")
		return
	}

	s, err := h.svc.GetSummaryByDate(c.Request.Context(), date)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "summary not found for this date")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("[summary.HandleGetSummary] unexpected error")
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, s)
}

// HandleTriggerSummary godoc
// @Summary      Trigger summary generation
// @Description  Manually generate a summary for a specific date
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request  body      TriggerRequest  true  "Date to generate"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /admin/summaries/trigger [post]
func (h *Handler) HandleTriggerSummary(c *gin.Context) {
	var req TriggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		response.ValidationError(c, "invalid date format, use YYYY-MM-DD")
		return
	}

	if err := h.svc.GenerateSummaryForDate(c.Request.Context(), req.Date); err != nil {
		log.Error().Err(err).Msg("[summary.HandleTriggerSummary] generation failed")
		response.InternalError(c)
		return
	}

	s, err := h.svc.GetSummaryByDate(c.Request.Context(), req.Date)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, s)
}
