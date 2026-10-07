package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"flexgin/internal/httpx"
)

// RequestLogger logs one structured line per request after it completes.
func RequestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		path := c.FullPath() // route template, avoids high-cardinality IDs in logs
		if path == "" {
			path = "unmatched"
		}

		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}

		log.LogAttrs(
			c.Request.Context(), level, "http request",
			slog.String("request_id", c.GetString(httpx.RequestIDKey)),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", time.Since(start)),
			slog.Int("bytes", c.Writer.Size()),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}
