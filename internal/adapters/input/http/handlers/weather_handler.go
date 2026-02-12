package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"weather-api-wrapper/internal/adapters/input/http/dto"
	"weather-api-wrapper/internal/domain/weather"
	"weather-api-wrapper/internal/ports/input"
)

// WeatherHandler handles HTTP requests for weather data
type WeatherHandler struct {
	weatherUseCase input.GetWeatherUseCase
}

// NewWeatherHandler creates a new weather HTTP handler
func NewWeatherHandler(useCase input.GetWeatherUseCase) *WeatherHandler {
	return &WeatherHandler{
		weatherUseCase: useCase,
	}
}

// GetWeatherHandler handles GET /weather requests
func (h *WeatherHandler) GetWeatherHandler(w http.ResponseWriter, r *http.Request) {
	// Validate required query parameter
	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, "city query parameter is required", http.StatusBadRequest)
		return
	}

	// Call use case
	weatherData, err := h.weatherUseCase.GetWeather(r.Context(), city)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Convert domain model to DTO
	response := dto.FromDomain(weatherData)

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleError maps domain errors to appropriate HTTP status codes
func (h *WeatherHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, weather.ErrInvalidLocation):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, weather.ErrWeatherNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, weather.ErrWeatherUnavailable):
		http.Error(w, "weather service is currently unavailable", http.StatusServiceUnavailable)
	case errors.Is(err, weather.ErrCacheUnavailable):
		// Cache errors shouldn't reach here, but if they do, treat as server error
		http.Error(w, "internal server error", http.StatusInternalServerError)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
