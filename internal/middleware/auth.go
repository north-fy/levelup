package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/north-fy/levelup/internal/domain"
	"github.com/north-fy/levelup/internal/pkg/jwt"
)

const userIDKey = "user_id"

// UserID returns the authenticated user id stored in the context.
func UserID(c *gin.Context) uint {
	if v, ok := c.Get(userIDKey); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// Authenticate validates a Bearer access token and stores the user id.
func Authenticate(manager *jwt.Manager, blacklist func(c *gin.Context, tokenID string) (bool, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := manager.ParseAccess(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": domain.ErrInvalidToken.Error()})
			return
		}

		if blacklist != nil {
			revoked, err := blacklist(c, claims.ID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
				return
			}
			if revoked {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": domain.ErrTokenRevoked.Error()})
				return
			}
		}

		c.Set(userIDKey, claims.UserID)
		c.Next()
	}
}
