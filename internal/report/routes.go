package report

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	reports := r.Group("/reports")
	reports.Use(authMiddleware)
	{
		reports.POST("", h.HandleCreateReport)
		reports.GET("/me", h.HandleGetMyReports)
		reports.DELETE("/:id", h.HandleDeleteReport)
	}
}
