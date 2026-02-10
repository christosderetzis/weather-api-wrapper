package weather

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockWeatherClient struct {
	mock.Mock
}

func (m *MockWeatherClient) FetchWeatherFromApi(ctx context.Context, location string) (*WeatherResponse, error) {
	args := m.Called(ctx, location)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*WeatherResponse), args.Error(1)
}

type MockWeatherRepository struct {
	mock.Mock
}

func (m *MockWeatherRepository) GetWeatherFromCache(
	ctx context.Context,
	location string,
) (*WeatherResponse, error) {
	args := m.Called(ctx, location)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*WeatherResponse), args.Error(1)
}

func (m *MockWeatherRepository) SetWeatherInCache(
	ctx context.Context,
	location string,
	weather *WeatherResponse,
	ttl time.Duration,
) error {
	args := m.Called(ctx, location, weather, ttl)
	return args.Error(0)
}

func TestGetWeather_CacheHit(t *testing.T) {
	ctx := context.Background()
	location := "Athens"

	expected := &WeatherResponse{
		Location: Location{
			Name: "Athens",
		},
		Current: Current{
			TempC: 25.0,
			Condition: Condition{
				Text: "Sunny",
			},
		},
	}

	client := new(MockWeatherClient)
	repo := new(MockWeatherRepository)

	repo.
		On("GetWeatherFromCache", ctx, location).
		Return(expected, nil)

	service := NewWeatherService(client, repo)

	result, err := service.GetWeather(ctx, location)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	client.AssertNotCalled(t, "FetchWeatherFromApi", mock.Anything)
	repo.AssertExpectations(t)
}

func TestGetWeather_CacheMiss_Success(t *testing.T) {
	ctx := context.Background()
	location := "Athens"

	expected := &WeatherResponse{
		Location: Location{
			Name: "Athens",
		},
		Current: Current{
			TempC: 25.0,
			Condition: Condition{
				Text: "Sunny",
			},
		},
	}

	client := new(MockWeatherClient)
	repo := new(MockWeatherRepository)

	repo.
		On("GetWeatherFromCache", ctx, location).
		Return(nil, errors.New("cache miss"))

	client.
		On("FetchWeatherFromApi", ctx, location).
		Return(expected, nil)

	repo.
		On(
			"SetWeatherInCache",
			ctx,
			location,
			expected,
			time.Hour*12,
		).
		Return(nil)

	service := NewWeatherService(client, repo)

	result, err := service.GetWeather(ctx, location)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	client.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestGetWeather_ApiError(t *testing.T) {
	ctx := context.Background()
	location := "Athens"

	client := new(MockWeatherClient)
	repo := new(MockWeatherRepository)

	repo.
		On("GetWeatherFromCache", ctx, location).
		Return(nil, errors.New("cache miss"))

	client.
		On("FetchWeatherFromApi", ctx, location).
		Return(nil, errors.New("api down"))

	service := NewWeatherService(client, repo)

	result, err := service.GetWeather(ctx, location)

	assert.Error(t, err)
	assert.Nil(t, result)

	repo.AssertNotCalled(
		t,
		"SetWeatherInCache",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	)
}

func TestGetWeather_CacheSetError(t *testing.T) {
	ctx := context.Background()
	location := "Athens"

	expected := &WeatherResponse{
		Location: Location{
			Name: "Athens",
		},
		Current: Current{
			TempC: 25.0,
			Condition: Condition{
				Text: "Sunny",
			},
		},
	}

	client := new(MockWeatherClient)
	repo := new(MockWeatherRepository)

	repo.
		On("GetWeatherFromCache", ctx, location).
		Return(nil, errors.New("cache miss"))

	client.
		On("FetchWeatherFromApi", ctx, location).
		Return(expected, nil)

	repo.
		On(
			"SetWeatherInCache",
			ctx,
			location,
			expected,
			time.Hour*12,
		).
		Return(errors.New("cache set failed"))

	service := NewWeatherService(client, repo)

	result, err := service.GetWeather(ctx, location)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
