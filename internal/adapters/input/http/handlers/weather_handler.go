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
		h.writeErrorJSON(w, "city query parameter is required", http.StatusBadRequest)
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
		h.writeErrorJSON(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleError maps domain errors to appropriate HTTP status codes
func (h *WeatherHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, weather.ErrInvalidLocation):
		h.writeErrorJSON(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, weather.ErrWeatherNotFound):
		h.writeErrorJSON(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, weather.ErrWeatherUnavailable):
		h.writeErrorJSON(w, "weather service is currently unavailable", http.StatusServiceUnavailable)
	case errors.Is(err, weather.ErrCacheUnavailable):
		// Cache errors shouldn't reach here, but if they do, treat as server error
		h.writeErrorJSON(w, "internal server error", http.StatusInternalServerError)
	default:
		h.writeErrorJSON(w, "internal server error", http.StatusInternalServerError)
	}
}

// writeErrorJSON writes a JSON error response
func (h *WeatherHandler) writeErrorJSON(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := dto.ErrorResponse{
		Message: message,
	}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		// Fallback to plain text if JSON encoding fails
		http.Error(w, message, statusCode)
	}
}
