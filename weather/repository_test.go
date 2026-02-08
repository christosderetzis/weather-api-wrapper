package weather

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	t.Cleanup(func() {
		client.Close()
		mr.Close()
	})

	return mr, client
}

func TestWeatherRepository_SetWeatherInCache(t *testing.T) {
	mr, client := setupTestRedis(t)
	repo := NewWeatherRepository(client)
	ctx := context.Background()

	weather := sampleWeatherResponse()
	location := "London"
	ttl := time.Hour

	err := repo.SetWeatherInCache(ctx, location, weather, ttl)

	require.NoError(t, err)

	// Verify data was stored in Redis
	data, err := mr.Get(location)
	require.NoError(t, err)

	var stored WeatherResponse
	err = json.Unmarshal([]byte(data), &stored)
	require.NoError(t, err)

	assert.Equal(t, weather.Location.Name, stored.Location.Name)
	assert.Equal(t, weather.Current.TempC, stored.Current.TempC)
}

func TestWeatherRepository_GetWeatherFromCache(t *testing.T) {
	mr, client := setupTestRedis(t)
	repo := NewWeatherRepository(client)
	ctx := context.Background()

	weather := sampleWeatherResponse()
	location := "London"

	// Pre-populate cache
	data, err := json.Marshal(weather)
	require.NoError(t, err)
	mr.Set(location, string(data))

	result, err := repo.GetWeatherFromCache(ctx, location)

	require.NoError(t, err)
	assert.Equal(t, weather.Location.Name, result.Location.Name)
	assert.Equal(t, weather.Location.Country, result.Location.Country)
	assert.Equal(t, weather.Current.TempC, result.Current.TempC)
}

func TestWeatherRepository_GetWeatherFromCache_NotFound(t *testing.T) {
	_, client := setupTestRedis(t)
	repo := NewWeatherRepository(client)
	ctx := context.Background()

	result, err := repo.GetWeatherFromCache(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, redis.Nil)
}

func TestWeatherRepository_GetWeatherFromCache_InvalidJSON(t *testing.T) {
	mr, client := setupTestRedis(t)
	repo := NewWeatherRepository(client)
	ctx := context.Background()

	mr.Set("London", "not valid json")

	result, err := repo.GetWeatherFromCache(ctx, "London")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestWeatherRepository_SetAndGet(t *testing.T) {
	_, client := setupTestRedis(t)
	repo := NewWeatherRepository(client)
	ctx := context.Background()

	weather := sampleWeatherResponse()
	location := "Paris"

	err := repo.SetWeatherInCache(ctx, location, weather, time.Hour)
	require.NoError(t, err)

	result, err := repo.GetWeatherFromCache(ctx, location)

	require.NoError(t, err)
	assert.Equal(t, weather.Location.Name, result.Location.Name)
	assert.Equal(t, weather.Current.TempC, result.Current.TempC)
	assert.Equal(t, weather.Current.Condition.Text, result.Current.Condition.Text)
}
