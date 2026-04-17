package middleware

import (
	"strings"

	"github.com/conflux-888/conflux-api/internal/common/jwt"
	"github.com/conflux-888/conflux-api/internal/common/response"
	"github.com/gin-gonic/gin"
)

const (
	keyAdminUser = "adminUser"
	adminClaim   = "admin"
)

// AdminAuth accepts only JWTs marked with typ="admin" (issued by adminauth.Service).
// Rejects user JWTs with 401 even if the signature is valid.
func AdminAuth(secret string) gin.HandlerFunc {
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
		if claims.Type != adminClaim {
			response.Unauthorized(c, "admin scope required")
			c.Abort()
			return
		}

		c.Set(keyAdminUser, claims.Subject)
		c.Next()
	}
}

func AdminUserFromContext(c *gin.Context) string {
	return c.GetString(keyAdminUser)
}
