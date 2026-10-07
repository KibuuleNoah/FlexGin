package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"FlexGin/internal/article"
	"FlexGin/internal/httpx"
	"FlexGin/internal/middleware"
	"FlexGin/internal/store"
)

// RegisterRoutes builds the router: global middleware, probes, then every feature group.
func (s *Server) RegisterRoutes(q *store.Queries) http.Handler {
	if s.cfg.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}
	httpx.UseJSONFieldNames()

	r := gin.New()
	r.Use(middleware.Global(s.cfg, s.log)...)

	s.registerProbeRoutes(r)
	s.registerAPIRoutes(r, q)

	return r
}

func (s *Server) registerProbeRoutes(r *gin.Engine) {
	r.GET("/healthz", s.liveness)
	r.GET("/readyz", s.readiness)
}

// registerAPIRoutes is the single place feature routes are mounted.
// Adding a feature = one line here.
func (s *Server) registerAPIRoutes(r *gin.Engine, q *store.Queries) {
	v1 := r.Group("/api/v1")

	article.RegisterArticleRoutes(v1, q)
	// user.RegisterRoutes(v1, s.users)
	// item.RegisterRoutes(v1, s.items)
	// cart.RegisterRoutes(v1, s.carts)
}

// liveness: process is up. No dependency checks.
func (s *Server) liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// readiness: dependencies are reachable. Used by orchestrators to gate traffic.
func (s *Server) readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := s.db.Health(ctx); err != nil {
		s.log.ErrorContext(ctx, "readiness check failed", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
