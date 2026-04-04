package user

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, h *Handler, authMiddleware gin.HandlerFunc) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", h.HandleRegister)
		auth.POST("/login", h.HandleLogin)
	}

	users := r.Group("/users")
	users.Use(authMiddleware)
	{
		users.GET("/me", h.HandleGetMe)
		users.PUT("/me", h.HandleUpdateMe)
	}
}
