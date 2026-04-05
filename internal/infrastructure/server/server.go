package server

import (
	"github.com/gin-gonic/gin"
)

func NewRouter() (*gin.Engine, *gin.RouterGroup) {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")

	return r, v1
}
