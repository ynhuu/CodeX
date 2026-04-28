package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	httpRequestActive = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "http_request_active",
		Help: "Number of active requests",
	})
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, httpRequestActive)
}

func NewPrometheus() gin.HandlerFunc {
	return func(c *gin.Context) {
		endpoint := c.Request.URL.Path
		if strings.Contains(endpoint, "metrics") || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		httpRequestActive.Inc()
		defer httpRequestActive.Dec()

		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		method := c.Request.Method
		httpRequestsTotal.WithLabelValues(method, endpoint, http.StatusText(status)).Inc()
		httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
	}
}
