package report

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/conflux-888/conflux-api/internal/common/middleware"
	"github.com/conflux-888/conflux-api/internal/common/response"
	"github.com/conflux-888/conflux-api/internal/event"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// HandleCreateReport godoc
// @Summary      Submit a user report
// @Description  Creates a threat event report and auto-clusters with nearby reports
// @Tags         reports
// @Accept       json
// @Produce      json
// @Param        request  body      CreateReportRequest  true  "Report data"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /reports [post]
func (h *Handler) HandleCreateReport(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)

	var req CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	e, err := h.svc.SubmitReport(c.Request.Context(), userID, req)
	if err != nil {
		log.Error().Err(err).Msg("[report.HandleCreateReport] unexpected error")
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusCreated, e)
}

// HandleGetMyReports godoc
// @Summary      List my reports
// @Description  Returns paginated list of the authenticated user's reports
// @Tags         reports
// @Produce      json
// @Param        page   query  int  false  "Page number"     default(1)
// @Param        limit  query  int  false  "Items per page"  default(20)
// @Success      200    {object}  map[string]interface{}
// @Failure      401    {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /reports/me [get]
func (h *Handler) HandleGetMyReports(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	events, pagination, err := h.svc.GetMyReports(c.Request.Context(), userID, page, limit)
	if err != nil {
		log.Error().Err(err).Msg("[report.HandleGetMyReports] unexpected error")
		response.InternalError(c)
		return
	}

	response.List(c, events, *pagination)
}

// HandleDeleteReport godoc
// @Summary      Delete my report
// @Description  Soft-deletes a report (only the owner can delete)
// @Tags         reports
// @Produce      json
// @Param        id   path  string  true  "Report/Event ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /reports/{id} [delete]
func (h *Handler) HandleDeleteReport(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	eventID := c.Param("id")

	err := h.svc.DeleteMyReport(c.Request.Context(), userID, eventID)
	if errors.Is(err, event.ErrNotFound) {
		response.NotFound(c, "report not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("[report.HandleDeleteReport] unexpected error")
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "report deleted"})
}
