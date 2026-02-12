package output

import (
	"context"
	"time"

	"weather-api-wrapper/internal/domain/weather"
)

// WeatherCache abstracts caching mechanisms for weather data
// This is a secondary/driven port that defines how the application
// stores and retrieves weather data from cache (like Redis)
type WeatherCache interface {
	// Get retrieves weather data from cache for a given location
	// Returns nil and no error if the key doesn't exist (cache miss)
	Get(ctx context.Context, location string) (*weather.Weather, error)

	// Set stores weather data in cache with a time-to-live duration
	// The cache implementation should handle serialization
	Set(ctx context.Context, location string, data *weather.Weather, ttl time.Duration) error
}
