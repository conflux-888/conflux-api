package notification

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/conflux-888/conflux-api/internal/common/middleware"
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

// HandleListNotifications godoc
// @Summary      List my notifications
// @Tags         notifications
// @Produce      json
// @Param        unread_only  query  bool  false  "Filter unread only"   default(false)
// @Param        page         query  int   false  "Page number"          default(1)
// @Param        limit        query  int   false  "Items per page"       default(20)
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /notifications/me [get]
func (h *Handler) HandleListNotifications(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	unreadOnly, _ := strconv.ParseBool(c.DefaultQuery("unread_only", "false"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	notifs, pagination, err := h.svc.GetMyNotifications(c.Request.Context(), userID, unreadOnly, page, limit)
	if err != nil {
		log.Error().Err(err).Msg("[notification.HandleListNotifications] error")
		response.InternalError(c)
		return
	}
	response.List(c, notifs, *pagination)
}

// HandleUnreadCount godoc
// @Summary      Get unread notification count
// @Tags         notifications
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /notifications/me/unread-count [get]
func (h *Handler) HandleUnreadCount(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	count, err := h.svc.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("[notification.HandleUnreadCount] error")
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"unread_count": count})
}

// HandleMarkRead godoc
// @Summary      Mark notification as read
// @Tags         notifications
// @Produce      json
// @Param        id   path  string  true  "Notification ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /notifications/{id}/read [post]
func (h *Handler) HandleMarkRead(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	id := c.Param("id")

	if err := h.svc.MarkRead(c.Request.Context(), userID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "notification not found")
			return
		}
		log.Error().Err(err).Msg("[notification.HandleMarkRead] error")
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "marked as read"})
}

// HandleMarkAllRead godoc
// @Summary      Mark all notifications as read
// @Tags         notifications
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /notifications/read-all [post]
func (h *Handler) HandleMarkAllRead(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	count, err := h.svc.MarkAllRead(c.Request.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("[notification.HandleMarkAllRead] error")
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"modified_count": count})
}
