package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout sets a deadline on the request context. Handlers and DB calls
// must pass c.Request.Context() down for it to take effect.
func Timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
