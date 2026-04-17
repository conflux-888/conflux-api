package event

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, authMiddleware, adminAuthMiddleware gin.HandlerFunc) {
	events := r.Group("/events")
	events.Use(authMiddleware)
	{
		events.GET("", h.HandleListEvents)
		events.GET("/nearby", h.HandleGetNearby)
		events.GET("/:id", h.HandleGetEvent)
	}

	admin := r.Group("/admin/events")
	admin.Use(adminAuthMiddleware)
	{
		admin.POST("/seed", h.HandleSeedEvent)
		admin.GET("/seeded", h.HandleListSeededEvents)
		admin.DELETE("/seeded", h.HandleDeleteAllSeededEvents)
		admin.DELETE("/:id", h.HandleDeleteSeededEvent)
	}
}
