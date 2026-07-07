package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sentinelops/sentinelops/apps/api/internal/config"
)

type Middleware struct {
	cfg config.Config
}

func NewMiddleware(cfg config.Config) *Middleware {
	return &Middleware{cfg: cfg}
}

func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.cfg.AuthEnabled {
			c.Set("principal", "local-admin")
			c.Set("role", "admin")
			c.Next()
			return
		}

		authorization := c.GetHeader("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		// Production integration point:
		// Validate the JWT against Keycloak JWKS and check audience/issuer claims.
		// The middleware boundary is intentionally kept here so handlers remain auth-agnostic.
		c.Set("principal", "keycloak-user")
		c.Set("role", "platform-viewer")
		c.Next()
	}
}
