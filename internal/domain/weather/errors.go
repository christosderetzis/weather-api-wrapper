package weather

import "errors"

// Domain-specific errors
var (
	// ErrInvalidLocation indicates that the provided location is invalid
	ErrInvalidLocation = errors.New("invalid location: location cannot be empty")

	// ErrWeatherNotFound indicates that weather data was not found for the requested location
	ErrWeatherNotFound = errors.New("weather data not found")

	// ErrWeatherUnavailable indicates that the weather service is unavailable
	ErrWeatherUnavailable = errors.New("weather service unavailable")

	// ErrCacheUnavailable indicates that the cache service is unavailable
	ErrCacheUnavailable = errors.New("cache service unavailable")
)
