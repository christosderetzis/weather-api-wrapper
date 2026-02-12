package output

import (
	"context"

	"weather-api-wrapper/internal/domain/weather"
)

// WeatherProvider abstracts external weather data sources
// This is a secondary/driven port that defines how the application
// retrieves weather data from external providers (like WeatherAPI.com)
type WeatherProvider interface {
	// FetchWeather retrieves weather information from an external source
	// It returns domain weather data or an error
	FetchWeather(ctx context.Context, location string) (*weather.Weather, error)
}
