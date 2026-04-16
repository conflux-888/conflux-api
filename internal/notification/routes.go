package notification

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	notifs := r.Group("/notifications")
	notifs.Use(authMiddleware)
	{
		notifs.GET("/me", h.HandleListNotifications)
		notifs.GET("/me/unread-count", h.HandleUnreadCount)
		notifs.POST("/read-all", h.HandleMarkAllRead)
		notifs.POST("/:id/read", h.HandleMarkRead)
	}
}
