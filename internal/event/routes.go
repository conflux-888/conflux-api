package event

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	events := r.Group("/events")
	events.Use(authMiddleware)
	{
		events.GET("", h.HandleListEvents)
		events.GET("/nearby", h.HandleGetNearby)
		events.GET("/:id", h.HandleGetEvent)
	}
}
