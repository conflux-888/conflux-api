package sync

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts sync admin endpoints. Caller must pass the admin-scoped middleware.
func RegisterRoutes(r *gin.RouterGroup, h *Handler, adminAuthMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/sync")
	admin.Use(adminAuthMiddleware)
	{
		admin.GET("/status", h.HandleGetStatus)
		admin.POST("/trigger", h.HandleTriggerSync)
	}
}
