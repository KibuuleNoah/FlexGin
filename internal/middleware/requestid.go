package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"

	"FlexGin/internal/httpx" // replace "yourmodule" with the module path from go.mod
)

const requestIDHeader = "X-Request-ID"

// RequestID reuses a sane inbound X-Request-ID or generates one,
// stores it in the gin context, and echoes it on the response.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(requestIDHeader)
		if id == "" || len(id) > 128 {
			id = newID()
		}
		c.Set(httpx.RequestIDKey, id)
		c.Header(requestIDHeader, id)
		c.Next()
	}
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
