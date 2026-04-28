package devicetoken

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	g := r.Group("/users/me/device-tokens")
	g.Use(authMiddleware)
	{
		g.POST("", h.HandleRegister)
		g.DELETE("/:token", h.HandleUnregister)
	}
}
