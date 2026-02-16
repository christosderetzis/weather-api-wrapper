package weather

import (
	"context"
	"fmt"
	"log"
	"time"

	"weather-api-wrapper/internal/adapters/input/http/middleware/metrics"
	"weather-api-wrapper/internal/domain/weather"
	"weather-api-wrapper/internal/ports/output"
)

const defaultCacheTTL = 12 * time.Hour

// Service implements the GetWeatherUseCase use case
// It orchestrates weather data retrieval using a cache-aside pattern
type Service struct {
	weatherProvider output.WeatherProvider
	cache           output.WeatherCache
}

// NewService creates a new weather application service
func NewService(provider output.WeatherProvider, cache output.WeatherCache) *Service {
	return &Service{
		weatherProvider: provider,
		cache:           cache,
	}
}

// GetWeather retrieves weather information for a given location
// It implements the cache-aside pattern:
// 1. Check cache first
// 2. On cache miss, fetch from weather provider
// 3. Update cache with fresh data
// 4. Return weather data
func (s *Service) GetWeather(ctx context.Context, location string) (*weather.Weather, error) {
	// Domain validation
	if err := weather.ValidateLocation(location); err != nil {
		return nil, err
	}

	// Try to get from cache first
	cachedWeather, err := s.cache.Get(ctx, location)
	if err == nil && cachedWeather != nil {
		log.Printf("Cache hit for location: %s", location)
		metrics.CacheHitsTotal.Inc()
		return cachedWeather, nil
	}

	if err != nil {
		log.Printf("Cache error for location %s: %v", location, err)
		metrics.CacheErrorsTotal.Inc()
	}

	// Cache miss - fetch from weather provider
	log.Printf("Cache miss for location: %s", location)
	metrics.CacheMissesTotal.Inc()
	weatherData, err := s.weatherProvider.FetchWeather(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", weather.ErrWeatherUnavailable, err)
	}

	// Update the timestamp
	weatherData.UpdatedAt = time.Now()

	// Store in cache (non-blocking - don't fail the request if caching fails)
	if err := s.cache.Set(ctx, location, weatherData, defaultCacheTTL); err != nil {
		log.Printf("Warning: failed to cache weather data for %s: %v", location, err)
		// Continue - caching failure shouldn't break the request
	}

	return weatherData, nil
}
