package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"flexgin/internal/config"
)

// Global returns the default stack in the order it must run:
// RequestID -> Logger -> Recovery -> CORS -> Timeout -> ErrorHandler.
// Logger sits outside Recovery so panics are logged with their final status.
// ErrorHandler is innermost so it sees errors from handlers.
func Global(cfg config.Config, log *slog.Logger) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		RequestID(),
		RequestLogger(log),
		Recovery(log),
		CORS(cfg.CORS),
		Timeout(cfg.HTTP.RequestTimeout),
		ErrorHandler(log),
	}
}
