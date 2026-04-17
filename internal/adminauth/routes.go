package adminauth

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the admin login endpoint. Login is NOT wrapped with authMiddleware;
// it's the way to obtain an admin token.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	g := r.Group("/admin/auth")
	{
		g.POST("/login", h.HandleLogin)
	}
}
