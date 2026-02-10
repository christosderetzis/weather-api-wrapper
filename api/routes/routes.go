package routes

import (
	"net/http"
	"weather-api-wrapper/api/handler"
	"weather-api-wrapper/api/middleware/logging"
	"weather-api-wrapper/api/middleware/rate_limiter"
)

func SetupRoutes(handler *handler.WeatherHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/weather", handler.GetWeatherHandler)

	rateLimiter := rate_limiter.NewRateLimiter(30)
	withRateLimit := rateLimiter.Middleware(mux)

	return logging.LoggingMiddleware(withRateLimit)
}
