package middleware

import (
	"time"

	"flexgin/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS(cfg config.CORS) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", requestIDHeader},
		ExposeHeaders:    []string{requestIDHeader},
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           12 * time.Hour,
	})
}
