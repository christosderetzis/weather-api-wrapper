package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "weather_api_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "weather_api_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// Cache metrics
	CacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "weather_api_cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	CacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "weather_api_cache_misses_total",
			Help: "Total number of cache misses",
		},
	)

	CacheErrorsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "weather_api_cache_errors_total",
			Help: "Total number of cache errors",
		},
	)

	// External API metrics
	ExternalAPICallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "weather_api_external_api_calls_total",
			Help: "Total number of external API calls",
		},
		[]string{"provider", "status"},
	)

	ExternalAPICallDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "weather_api_external_api_call_duration_seconds",
			Help:    "External API call latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider"},
	)

	// Rate limiter metrics
	RateLimitExceeded = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "weather_api_rate_limit_exceeded_total",
			Help: "Total number of requests that exceeded rate limit",
		},
	)
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// MetricsMiddleware records HTTP request metrics
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default status
		}

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrapped.statusCode)
		path := r.URL.Path

		httpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, path, status).Observe(duration)
	})
}
