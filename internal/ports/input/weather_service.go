package input

import (
	"context"

	"weather-api-wrapper/internal/domain/weather"
)

// GetWeatherUseCase defines the business capability to retrieve weather information
// This is the primary/driving port that external actors (like HTTP handlers) use
// to interact with the application's core business logic
type GetWeatherUseCase interface {
	// GetWeather retrieves weather information for a given location
	// It returns domain weather data or a domain error
	GetWeather(ctx context.Context, location string) (*weather.Weather, error)
}
