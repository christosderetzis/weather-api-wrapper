package dto

import "weather-api-wrapper/internal/domain/weather"

// WeatherResponse is the HTTP response DTO for weather endpoints
// It provides a simplified view of weather data for API clients
type WeatherResponse struct {
	Location    string  `json:"location"`
	Temperature float64 `json:"temperature_c"`
	Condition   string  `json:"condition_text"`
}

// FromDomain maps domain weather data to HTTP response DTO
// This keeps the HTTP layer decoupled from domain structure
func FromDomain(w *weather.Weather) WeatherResponse {
	return WeatherResponse{
		Location:    w.Location.Name,
		Temperature: w.Current.Temperature.Celsius,
		Condition:   w.Current.Condition.Text,
	}
}
