package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Custom latency buckets optimized for web/API services (5ms to 10s)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{"method", "path", "status"},
	)

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed.",
		},
		[]string{"method", "path", "status"},
	)

	httpInFlightRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_in_flight_requests",
			Help: "Current number of HTTP requests being processed simultaneously.",
		},
	)
)

func init() {
	// Register metrics with the default Prometheus registry safely
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpInFlightRequests)
}

// PrometheusMiddleware returns a hardened Gin middleware for production telemetry.
// ignoredPaths allows you to skip metrics tracking for system routes (e.g., /metrics, /healthz).
func PrometheusMiddleware(ignoredPaths ...string) gin.HandlerFunc {
	ignoreMap := make(map[string]bool, len(ignoredPaths))
	for _, p := range ignoredPaths {
		ignoreMap[p] = true
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip high-frequency internal routes to prevent metric bloat
		if ignoreMap[path] {
			c.Next()
			return
		}

		// Track concurrent load (Saturation)
		httpInFlightRequests.Inc()
		defer httpInFlightRequests.Dec()

		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		// CRITICAL: Use c.FullPath() to get the registered route pattern (e.g. /api/v1/users/:id)
		// rather than the raw URL path, protecting against high cardinality memory leaks.
		routePath := c.FullPath()
		if routePath == "" {
			routePath = "unmatched"
		}

		httpRequestDuration.WithLabelValues(method, routePath, status).Observe(duration)
		httpRequestsTotal.WithLabelValues(method, routePath, status).Inc()
	}
}
