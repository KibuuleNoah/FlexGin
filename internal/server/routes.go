package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"flexgin/internal/domain/article"
	"flexgin/internal/domain/auth"
	"flexgin/internal/httpx"
	"flexgin/internal/middleware"
	"flexgin/internal/store"

	_ "flexgin/docs"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"
)

func init() {
	if _, err := swag.ReadDoc(); err != nil {
		panic("swagger: " + err.Error())
	}
}

// RegisterRoutes builds the router: global middleware, probes, then every feature group
func (s *Server) RegisterRoutes(q *store.Queries) http.Handler {
	if s.cfg.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}
	httpx.UseJSONFieldNames()

	r := gin.New()
	r.Use(middleware.PrometheusMiddleware("/metrics", "/healthz", "/readyz"))
	r.Use(gin.Recovery())

	r.Use(middleware.Global(s.cfg, s.log)...)

	if !s.cfg.IsProd() {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	s.registerProbeRoutes(r)
	s.registerAPIV1Routes(r, q)

	return r
}

func (s *Server) registerProbeRoutes(r *gin.Engine) {
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/healthz", s.liveness)
	r.GET("/readyz", s.readiness)
}

// registerAPIRoutes is were all routes are mounted
func (s *Server) registerAPIV1Routes(r *gin.Engine, q *store.Queries) {
	v1 := r.Group("/api/v1")
	authn := middleware.Authenticate(s.authSvc)

	auth.RegisterAuthRoutes(v1, s.authSvc, authn)
	article.RegisterArticleRoutes(v1, q)
	// user.RegisterRoutes(v1, s.users)
	// item.RegisterRoutes(v1, s.items)
	// cart.RegisterRoutes(v1, s.carts)
}

func (s *Server) liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

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
