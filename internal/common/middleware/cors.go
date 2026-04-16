package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var devAllowedOrigins = map[string]bool{
	"http://localhost:5173":  true,
	"http://127.0.0.1:5173":  true,
	"http://localhost:8080":  true,
}

// CORS returns a middleware. When allowLocalhost is false, the middleware is a no-op
// (production same-origin serving needs no CORS). When true, allows the admin UI dev
// server to call the API with credentials.
func CORS(allowLocalhost bool) gin.HandlerFunc {
	if !allowLocalhost {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if devAllowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
			c.Header("Access-Control-Max-Age", "600")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
