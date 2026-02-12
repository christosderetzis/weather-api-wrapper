package redis

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"weather-api-wrapper/internal/domain/weather"
)

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *Cache) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	// Create Redis client connected to miniredis
	cache := &Cache{
		client: goredis.NewClient(&goredis.Options{
			Addr: mr.Addr(),
		}),
	}

	t.Cleanup(func() {
		cache.Close()
		mr.Close()
	})

	return mr, cache
}

func createSampleWeather() *weather.Weather {
	return &weather.Weather{
		Location: weather.Location{
			Name:    "London",
			Country: "United Kingdom",
		},
		Current: weather.CurrentWeather{
			Temperature: weather.Temperature{
				Celsius:    15.0,
				Fahrenheit: 59.0,
			},
			Condition: weather.Condition{
				Text: "Partly cloudy",
				Code: 1003,
			},
		},
		UpdatedAt: time.Now(),
	}
}

func TestCache_Set(t *testing.T) {
	mr, cache := setupTestRedis(t)
	ctx := context.Background()

	weatherData := createSampleWeather()
	location := "London"
	ttl := time.Hour

	err := cache.Set(ctx, location, weatherData, ttl)

	require.NoError(t, err)

	// Verify data was stored in Redis
	data, err := mr.Get(location)
	require.NoError(t, err)

	var stored weather.Weather
	err = json.Unmarshal([]byte(data), &stored)
	require.NoError(t, err)

	assert.Equal(t, weatherData.Location.Name, stored.Location.Name)
	assert.Equal(t, weatherData.Current.Temperature.Celsius, stored.Current.Temperature.Celsius)
}

func TestCache_Get(t *testing.T) {
	mr, cache := setupTestRedis(t)
	ctx := context.Background()

	weatherData := createSampleWeather()
	location := "London"

	// Pre-populate cache
	data, err := json.Marshal(weatherData)
	require.NoError(t, err)
	mr.Set(location, string(data))

	result, err := cache.Get(ctx, location)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, weatherData.Location.Name, result.Location.Name)
	assert.Equal(t, weatherData.Location.Country, result.Location.Country)
	assert.Equal(t, weatherData.Current.Temperature.Celsius, result.Current.Temperature.Celsius)
}

func TestCache_Get_NotFound(t *testing.T) {
	_, cache := setupTestRedis(t)
	ctx := context.Background()

	result, err := cache.Get(ctx, "nonexistent")

	// Should return nil, nil for cache miss (not an error)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestCache_Get_InvalidJSON(t *testing.T) {
	mr, cache := setupTestRedis(t)
	ctx := context.Background()

	mr.Set("London", "not valid json")

	result, err := cache.Get(ctx, "London")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCache_SetAndGet(t *testing.T) {
	_, cache := setupTestRedis(t)
	ctx := context.Background()

	weatherData := createSampleWeather()
	location := "Paris"

	err := cache.Set(ctx, location, weatherData, time.Hour)
	require.NoError(t, err)

	result, err := cache.Get(ctx, location)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, weatherData.Location.Name, result.Location.Name)
	assert.Equal(t, weatherData.Current.Temperature.Celsius, result.Current.Temperature.Celsius)
	assert.Equal(t, weatherData.Current.Condition.Text, result.Current.Condition.Text)
}

func TestCache_TTL(t *testing.T) {
	mr, cache := setupTestRedis(t)
	ctx := context.Background()

	weatherData := createSampleWeather()
	location := "London"
	ttl := 2 * time.Second

	err := cache.Set(ctx, location, weatherData, ttl)
	require.NoError(t, err)

	// Verify data exists immediately
	result, err := cache.Get(ctx, location)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Fast-forward time in miniredis
	mr.FastForward(3 * time.Second)

	// Data should be expired now
	result, err = cache.Get(ctx, location)
	require.NoError(t, err)
	assert.Nil(t, result) // Should be nil (expired)
}
