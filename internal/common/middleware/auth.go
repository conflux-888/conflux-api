package middleware

import (
	"strings"

	"github.com/conflux-888/conflux-api/internal/common/jwt"
	"github.com/conflux-888/conflux-api/internal/common/response"
	"github.com/gin-gonic/gin"
)

const (
	keyUserID = "userID"
	keyEmail  = "email"
)

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(parts[1], secret)
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(keyUserID, claims.Subject)
		c.Set(keyEmail, claims.Email)
		c.Next()
	}
}

func UserIDFromContext(c *gin.Context) string {
	return c.GetString(keyUserID)
}

func EmailFromContext(c *gin.Context) string {
	return c.GetString(keyEmail)
}
