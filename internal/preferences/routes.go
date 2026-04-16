package preferences

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	prefs := r.Group("/preferences")
	prefs.Use(authMiddleware)
	{
		prefs.GET("", h.HandleGet)
		prefs.PUT("", h.HandleUpdate)
		prefs.PUT("/location", h.HandleUpdateLocation)
	}
}
