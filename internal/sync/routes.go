package sync

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, h *Handler, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/sync")
	admin.Use(authMiddleware)
	{
		admin.GET("/status", h.HandleGetStatus)
		admin.POST("/trigger", h.HandleTriggerSync)
	}
}
