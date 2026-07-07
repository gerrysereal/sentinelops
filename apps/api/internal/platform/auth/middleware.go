package auth

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sentinelops/sentinelops/apps/api/internal/config"
)

type Middleware struct {
	cfg config.Config
}

type Principal struct {
	Name        string
	Role        string
	Permissions map[string]bool
}

func NewMiddleware(cfg config.Config) *Middleware {
	return &Middleware{cfg: cfg}
}

func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := m.principalFromRequest(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("principal", principal.Name)
		c.Set("role", principal.Role)
		c.Set("permissions", principal.Permissions)
		ctx := context.WithValue(c.Request.Context(), "actor", principal.Name)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func (m *Middleware) RequirePermission(required ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, ok := c.Get("permissions")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permission context missing"})
			return
		}
		permissionMap, ok := permissions.(map[string]bool)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid permission context"})
			return
		}
		for _, permission := range required {
			if permissionMap[permission] || permissionMap["*"] {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permission denied"})
	}
}

func (m *Middleware) principalFromRequest(c *gin.Context) (Principal, bool) {
	if !m.cfg.AuthEnabled {
		role := strings.TrimSpace(c.GetHeader("X-User-Role"))
		if role == "" {
			role = "platform-admin"
		}
		return Principal{Name: "local-admin", Role: role, Permissions: permissionsForRole(role)}, true
	}

	authorization := strings.TrimSpace(c.GetHeader("Authorization"))
	if !strings.HasPrefix(authorization, "Bearer ") {
		return Principal{}, false
	}
	token := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
	if m.cfg.BootstrapToken != "" && subtle.ConstantTimeCompare([]byte(token), []byte(m.cfg.BootstrapToken)) == 1 {
		return Principal{Name: "bootstrap-admin", Role: "platform-admin", Permissions: permissionsForRole("platform-admin")}, true
	}
	return Principal{}, false
}

func permissionsForRole(role string) map[string]bool {
	if role == "platform-admin" || role == "admin" {
		return map[string]bool{"*": true}
	}
	baseRead := map[string]bool{
		"platform:read":    true,
		"settings:read":    true,
		"integration:read": true,
		"application:read": true,
	}
	switch role {
	case "platform-engineer":
		baseRead["settings:write"] = true
		baseRead["integration:write"] = true
		baseRead["integration:delete"] = true
		baseRead["integration:operate"] = true
		baseRead["application:write"] = true
		baseRead["pipeline:operate"] = true
		baseRead["deployment:operate"] = true
	case "security-engineer":
		baseRead["security:operate"] = true
		baseRead["integration:operate"] = true
	case "developer":
		baseRead["pipeline:operate"] = true
	}
	return baseRead
}
