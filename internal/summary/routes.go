package summary

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, authMiddleware, adminAuthMiddleware gin.HandlerFunc) {
	summaries := r.Group("/summaries")
	summaries.Use(authMiddleware)
	{
		summaries.GET("", h.HandleListSummaries)
		summaries.GET("/latest", h.HandleGetLatestSummary)
		summaries.GET("/:date", h.HandleGetSummary)
	}

	admin := r.Group("/admin/summaries")
	admin.Use(adminAuthMiddleware)
	{
		admin.POST("/trigger", h.HandleTriggerSummary)
	}
}
