package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"roadmap/internal/domain"
	pkgauth "roadmap/internal/pkg/auth"
)

const (
	ContextUserIDKey = "auth_user_id"
	ContextRoleKey   = "auth_role"
)

// Auth проверяет Bearer access-токен и кладёт user_id/role в context.
func Auth(tokens *pkgauth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": domain.ErrUnauthorized.Error()})
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		claims, err := tokens.ParseAccess(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": domain.ErrInvalidToken.Error()})
			return
		}
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextRoleKey, claims.Role)
		c.Next()
	}
}

// RequireAdmin пускает только role=admin.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(ContextRoleKey)
		if role != domain.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": domain.ErrForbidden.Error()})
			return
		}
		c.Next()
	}
}

// UserID достаёт user_id из gin context.
func UserID(c *gin.Context) (string, bool) {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return "", false
	}
	id, ok := v.(string)
	return id, ok && id != ""
}
