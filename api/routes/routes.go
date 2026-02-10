package routes

import (
	"net/http"
	"weather-api-wrapper/api/handler"
	"weather-api-wrapper/api/middleware"
)

func SetupRoutes(handler *handler.WeatherHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/weather", handler.GetWeatherHandler)

	rateLimiter := middleware.NewRateLimiter(30)
	withRateLimit := rateLimiter.Middleware(mux)

	return middleware.LoggingMiddleware(withRateLimit)
}
