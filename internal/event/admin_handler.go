package event

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/conflux-888/conflux-api/internal/common/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// HandleSeedEvent godoc
// @Summary      Seed a synthetic event (admin only)
// @Description  Inserts an admin-controlled event and triggers the same notification pipeline as GDELT sync. Used for testing notification delivery.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request  body      SeedAdminEventRequest  true  "Seed parameters"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /admin/events/seed [post]
func (h *Handler) HandleSeedEvent(c *gin.Context) {
	var req SeedAdminEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	e, err := h.svc.SeedAdminEvent(c.Request.Context(), req)
	if err != nil {
		log.Error().Err(err).Msg("[event.HandleSeedEvent] failed")
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"event":                 e,
		"notification_dispatch": "queued",
	})
}

// HandleListSeededEvents godoc
// @Summary      List admin-seeded events
// @Description  Returns events created via admin seed (external_id prefix ADMIN_TEST_), newest first.
// @Tags         admin
// @Produce      json
// @Param        page   query  int  false  "Page number"      default(1)
// @Param        limit  query  int  false  "Items per page"   default(20)
// @Success      200    {object}  map[string]interface{}
// @Failure      401    {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /admin/events/seeded [get]
func (h *Handler) HandleListSeededEvents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	events, total, err := h.svc.ListSeededEvents(c.Request.Context(), page, limit)
	if err != nil {
		log.Error().Err(err).Msg("[event.HandleListSeededEvents] failed")
		response.InternalError(c)
		return
	}

	response.List(c, events, response.Pagination{Page: page, Limit: limit, Total: total})
}

// HandleDeleteAllSeededEvents godoc
// @Summary      Delete ALL admin-seeded events
// @Description  Hard-deletes every seeded event (external_id prefix ADMIN_TEST_) and their notifications.
// @Tags         admin
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /admin/events/seeded [delete]
func (h *Handler) HandleDeleteAllSeededEvents(c *gin.Context) {
	eventsDeleted, notifsDeleted, err := h.svc.DeleteAllSeededEvents(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("[event.HandleDeleteAllSeededEvents] failed")
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusOK, gin.H{
		"events_deleted":        eventsDeleted,
		"notifications_deleted": notifsDeleted,
	})
}

// HandleDeleteSeededEvent godoc
// @Summary      Delete an admin-seeded event
// @Description  Hard-deletes the event and its related notifications. Only works on seeded events (external_id prefix ADMIN_TEST_).
// @Tags         admin
// @Produce      json
// @Param        id   path  string  true  "Event ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /admin/events/{id} [delete]
func (h *Handler) HandleDeleteSeededEvent(c *gin.Context) {
	id := c.Param("id")
	notifsDeleted, err := h.svc.DeleteSeededEvent(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "seeded event not found")
			return
		}
		log.Error().Err(err).Msg("[event.HandleDeleteSeededEvent] failed")
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusOK, gin.H{
		"message":               "deleted",
		"notifications_deleted": notifsDeleted,
	})
}
