package server

import (
	"github.com/gin-gonic/gin"
)

type DomainRoutes interface {
	RegisterRoutes(r *gin.Engine, authMiddleware gin.HandlerFunc)
}

func NewRouter(domains ...DomainRoutes) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
