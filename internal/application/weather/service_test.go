package weather

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"weather-api-wrapper/internal/domain/weather"
)

// Mock implementations for testing

type MockWeatherProvider struct {
	mock.Mock
}

func (m *MockWeatherProvider) FetchWeather(ctx context.Context, location string) (*weather.Weather, error) {
	args := m.Called(ctx, location)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*weather.Weather), args.Error(1)
}

type MockWeatherCache struct {
	mock.Mock
}

func (m *MockWeatherCache) Get(ctx context.Context, location string) (*weather.Weather, error) {
	args := m.Called(ctx, location)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*weather.Weather), args.Error(1)
}

func (m *MockWeatherCache) Set(ctx context.Context, location string, data *weather.Weather, ttl time.Duration) error {
	args := m.Called(ctx, location, data, ttl)
	return args.Error(0)
}

// Helper function to create sample weather data
func createSampleWeather(locationName string, tempC float64) *weather.Weather {
	return &weather.Weather{
		Location: weather.Location{
			Name:    locationName,
			Country: "Greece",
		},
		Current: weather.CurrentWeather{
			Temperature: weather.Temperature{
				Celsius:    tempC,
				Fahrenheit: tempC*9/5 + 32,
			},
			Condition: weather.Condition{
				Text: "Sunny",
				Code: 1000,
			},
		},
		UpdatedAt: time.Now(),
	}
}

// Tests

func TestGetWeather_CacheHit(t *testing.T) {
	// Arrange
	ctx := context.Background()
	location := "Athens"
	expected := createSampleWeather("Athens", 25.0)

	provider := new(MockWeatherProvider)
	cache := new(MockWeatherCache)

	cache.On("Get", ctx, location).Return(expected, nil)

	service := NewService(provider, cache)

	// Act
	result, err := service.GetWeather(ctx, location)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expected.Location.Name, result.Location.Name)
	assert.Equal(t, expected.Current.Temperature.Celsius, result.Current.Temperature.Celsius)

	// Verify provider was never called (cache hit)
	provider.AssertNotCalled(t, "FetchWeather", mock.Anything, mock.Anything)
	cache.AssertExpectations(t)
}

func TestGetWeather_CacheMiss_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	location := "Athens"
	expected := createSampleWeather("Athens", 25.0)

	provider := new(MockWeatherProvider)
	cache := new(MockWeatherCache)

	cache.On("Get", ctx, location).Return(nil, errors.New("cache miss"))
	provider.On("FetchWeather", ctx, location).Return(expected, nil)
	cache.On("Set", ctx, location, mock.AnythingOfType("*weather.Weather"), defaultCacheTTL).Return(nil)

	service := NewService(provider, cache)

	// Act
	result, err := service.GetWeather(ctx, location)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expected.Location.Name, result.Location.Name)
	assert.Equal(t, expected.Current.Temperature.Celsius, result.Current.Temperature.Celsius)

	provider.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestGetWeather_ProviderError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	location := "Athens"
	providerError := errors.New("api down")

	provider := new(MockWeatherProvider)
	cache := new(MockWeatherCache)

	cache.On("Get", ctx, location).Return(nil, errors.New("cache miss"))
	provider.On("FetchWeather", ctx, location).Return(nil, providerError)

	service := NewService(provider, cache)

	// Act
	result, err := service.GetWeather(ctx, location)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, weather.ErrWeatherUnavailable)

	// Cache Set should not be called when provider fails
	cache.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGetWeather_CacheSetError_StillReturnsData(t *testing.T) {
	// Arrange
	ctx := context.Background()
	location := "Athens"
	expected := createSampleWeather("Athens", 25.0)
	cacheError := errors.New("cache set failed")

	provider := new(MockWeatherProvider)
	cache := new(MockWeatherCache)

	cache.On("Get", ctx, location).Return(nil, errors.New("cache miss"))
	provider.On("FetchWeather", ctx, location).Return(expected, nil)
	cache.On("Set", ctx, location, mock.AnythingOfType("*weather.Weather"), defaultCacheTTL).Return(cacheError)

	service := NewService(provider, cache)

	// Act
	result, err := service.GetWeather(ctx, location)

	// Assert
	// Cache failure should not fail the request - data should still be returned
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expected.Location.Name, result.Location.Name)

	provider.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestGetWeather_InvalidLocation_Empty(t *testing.T) {
	// Arrange
	ctx := context.Background()
	location := ""

	provider := new(MockWeatherProvider)
	cache := new(MockWeatherCache)

	service := NewService(provider, cache)

	// Act
	result, err := service.GetWeather(ctx, location)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, weather.ErrInvalidLocation)

	// Neither cache nor provider should be called for invalid input
	cache.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
	provider.AssertNotCalled(t, "FetchWeather", mock.Anything, mock.Anything)
}

func TestGetWeather_InvalidLocation_Whitespace(t *testing.T) {
	// Arrange
	ctx := context.Background()
	location := "   "

	provider := new(MockWeatherProvider)
	cache := new(MockWeatherCache)

	service := NewService(provider, cache)

	// Act
	result, err := service.GetWeather(ctx, location)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, weather.ErrInvalidLocation)

	// Neither cache nor provider should be called for invalid input
	cache.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
	provider.AssertNotCalled(t, "FetchWeather", mock.Anything, mock.Anything)
}

func TestGetWeather_UpdatedAtTimestamp(t *testing.T) {
	// Arrange
	ctx := context.Background()
	location := "Athens"
	weatherData := createSampleWeather("Athens", 25.0)
	// Set an old timestamp
	weatherData.UpdatedAt = time.Now().Add(-24 * time.Hour)

	provider := new(MockWeatherProvider)
	cache := new(MockWeatherCache)

	cache.On("Get", ctx, location).Return(nil, errors.New("cache miss"))
	provider.On("FetchWeather", ctx, location).Return(weatherData, nil)
	cache.On("Set", ctx, location, mock.AnythingOfType("*weather.Weather"), defaultCacheTTL).Return(nil)

	service := NewService(provider, cache)

	// Act
	beforeCall := time.Now()
	result, err := service.GetWeather(ctx, location)
	afterCall := time.Now()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Verify that UpdatedAt was set to current time
	assert.True(t, result.UpdatedAt.After(beforeCall) || result.UpdatedAt.Equal(beforeCall))
	assert.True(t, result.UpdatedAt.Before(afterCall) || result.UpdatedAt.Equal(afterCall))
}

func TestGetWeather_TTL(t *testing.T) {
	// Arrange
	ctx := context.Background()
	location := "Athens"
	expected := createSampleWeather("Athens", 25.0)

	provider := new(MockWeatherProvider)
	cache := new(MockWeatherCache)

	cache.On("Get", ctx, location).Return(nil, errors.New("cache miss"))
	provider.On("FetchWeather", ctx, location).Return(expected, nil)

	// Verify that TTL is exactly 12 hours
	cache.On("Set", ctx, location, mock.AnythingOfType("*weather.Weather"), 12*time.Hour).Return(nil)

	service := NewService(provider, cache)

	// Act
	_, err := service.GetWeather(ctx, location)

	// Assert
	assert.NoError(t, err)
	cache.AssertExpectations(t)
}
