package middleware

import (
	"log/slog"

	"flex/internal/apperror"
	"flex/internal/httpx"

	"github.com/gin-gonic/gin"
)

// ErrorHandler renders the first error attached via c.Error(err) as the
// standard envelope. Handlers attach the error and return; nothing else.
// Internal errors are logged with their cause; clients only see the safe message.
func ErrorHandler(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		err := c.Errors[0].Err
		ae := apperror.From(err)

		if ae.Kind == apperror.KindInternal {
			log.ErrorContext(
				c.Request.Context(), "request failed",
				slog.String("request_id", c.GetString(httpx.RequestIDKey)),
				slog.String("code", ae.Code),
				slog.Any("error", err),
			)
		}
		httpx.WriteError(c, ae)
	}
}
