package middleware

import (
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	"flexgin/internal/apperror"
	"flexgin/internal/domain/auth"
)

// Authenticate requires "Authorization: Bearer <jwt>" with an active session.
func Authenticate(svc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme, token, ok := strings.Cut(c.GetHeader("Authorization"), " ")
		token = strings.TrimSpace(token)
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			abort(c, auth.ErrMissingToken)
			return
		}
		claims, err := svc.Authenticate(c.Request.Context(), token)
		if err != nil {
			abort(c, err)
			return
		}
		auth.SetClaims(c, claims)
		c.Next()
	}
}

// RequireRole must run after Authenticate.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := auth.ClaimsFrom(c)
		if !ok {
			abort(c, auth.ErrMissingToken)
			return
		}
		if !slices.Contains(roles, claims.Role) {
			abort(c, apperror.Forbidden("insufficient permissions"))
			return
		}
		c.Next()
	}
}

func abort(c *gin.Context, err error) {
	_ = c.Error(err)
	c.Abort()
}
