package routes

import (
	"net/http"

	"weather-api-wrapper/internal/adapters/input/http/handlers"
	"weather-api-wrapper/internal/adapters/input/http/middleware/logging"
	"weather-api-wrapper/internal/adapters/input/http/middleware/rate_limiter"
)

// SetupRoutes configures the HTTP routes with middleware chain
// Middleware order: Logging (outer) -> Rate Limiter -> Handler (inner)
func SetupRoutes(handler *handlers.WeatherHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/weather", handler.GetWeatherHandler)

	// Apply rate limiting (30 requests per minute)
	rateLimiter := rate_limiter.NewRateLimiter(30)
	withRateLimit := rateLimiter.Middleware(mux)

	// Apply logging (outermost middleware for visibility)
	return logging.LoggingMiddleware(withRateLimit)
}
