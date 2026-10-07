package middleware

import (
	"fmt"
	"log/slog"
	"runtime/debug"

	"flex/internal/apperror"
	"flex/internal/httpx"

	"github.com/gin-gonic/gin"
)

// Recovery turns panics into a 500 envelope and logs the stack.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.ErrorContext(
					c.Request.Context(), "panic recovered",
					slog.String("request_id", c.GetString(httpx.RequestIDKey)),
					slog.Any("panic", r),
					slog.String("stack", string(debug.Stack())),
				)
				httpx.WriteError(c, apperror.Internal(fmt.Errorf("panic: %v", r)))
			}
		}()
		c.Next()
	}
}
