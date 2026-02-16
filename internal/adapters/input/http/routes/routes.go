package routes

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"weather-api-wrapper/internal/adapters/input/http/handlers"
	"weather-api-wrapper/internal/adapters/input/http/middleware/logging"
	"weather-api-wrapper/internal/adapters/input/http/middleware/metrics"
	"weather-api-wrapper/internal/adapters/input/http/middleware/rate_limiter"
)

// SetupRoutes configures the HTTP routes with middleware chain
// Middleware order: Logging (outer) -> Metrics -> Rate Limiter -> Handler (inner)
func SetupRoutes(handler *handlers.WeatherHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/weather", handler.GetWeatherHandler)

	// Expose Prometheus metrics endpoint (no rate limiting for metrics)
	mux.Handle("/metrics", promhttp.Handler())

	// Apply rate limiting (30 requests per minute)
	rateLimiter := rate_limiter.NewRateLimiter(30)
	withRateLimit := rateLimiter.Middleware(mux)

	// Apply metrics middleware
	withMetrics := metrics.MetricsMiddleware(withRateLimit)

	// Apply logging (outermost middleware for visibility)
	return logging.LoggingMiddleware(withMetrics)
}
