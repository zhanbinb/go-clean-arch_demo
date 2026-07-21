package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics records request count and latency per route+method.
func Metrics(reg prometheus.Registerer) gin.HandlerFunc {
	requestCount := promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed, labelled by method, path and status.",
		},
		[]string{"method", "path", "status"},
	)
	requestLatency := promauto.With(reg).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request latencies in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		status := strconv.Itoa(c.Writer.Status())
		requestCount.WithLabelValues(c.Request.Method, path, status).Inc()
		requestLatency.WithLabelValues(c.Request.Method, path).Observe(time.Since(start).Seconds())
	}
}
