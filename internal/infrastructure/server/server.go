package server

import (
	"github.com/conflux-888/conflux-api/internal/common/middleware"
	"github.com/gin-gonic/gin"
)

type RouterOptions struct {
	CORSAllowLocalhost bool
}

func NewRouter(opts RouterOptions) (*gin.Engine, *gin.RouterGroup) {
	r := gin.Default()
	r.Use(middleware.CORS(opts.CORSAllowLocalhost))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")

	return r, v1
}
