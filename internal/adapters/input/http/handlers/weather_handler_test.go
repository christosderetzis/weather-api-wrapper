package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"weather-api-wrapper/internal/adapters/input/http/dto"
	"weather-api-wrapper/internal/domain/weather"
)

// MockGetWeatherUseCase mocks the GetWeatherUseCase input port
type MockGetWeatherUseCase struct {
	mock.Mock
}

func (m *MockGetWeatherUseCase) GetWeather(ctx context.Context, location string) (*weather.Weather, error) {
	args := m.Called(ctx, location)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*weather.Weather), args.Error(1)
}

func createSampleDomainWeather() *weather.Weather {
	return &weather.Weather{
		Location: weather.Location{
			Name:    "Athens",
			Country: "Greece",
		},
		Current: weather.CurrentWeather{
			Temperature: weather.Temperature{
				Celsius:    24.5,
				Fahrenheit: 76.1,
			},
			Condition: weather.Condition{
				Text: "Sunny",
				Code: 1000,
			},
		},
		UpdatedAt: time.Now(),
	}
}

func TestGetWeatherHandler_MissingCity(t *testing.T) {
	// Arrange
	useCase := new(MockGetWeatherUseCase)
	handler := NewWeatherHandler(useCase)

	req := httptest.NewRequest(http.MethodGet, "/weather", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.GetWeatherHandler(rec, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "city query parameter is required")

	// Use case should not be called
	useCase.AssertNotCalled(t, "GetWeather", mock.Anything, mock.Anything)
}

func TestGetWeatherHandler_InvalidLocation(t *testing.T) {
	// Arrange
	useCase := new(MockGetWeatherUseCase)
	handler := NewWeatherHandler(useCase)
	ctx := context.Background()

	req := httptest.NewRequest(http.MethodGet, "/weather?city=  ", nil)
	rec := httptest.NewRecorder()

	useCase.
		On("GetWeather", ctx, "  ").
		Return(nil, weather.ErrInvalidLocation).
		Once()

	// Act
	handler.GetWeatherHandler(rec, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid location")

	useCase.AssertExpectations(t)
}

func TestGetWeatherHandler_WeatherNotFound(t *testing.T) {
	// Arrange
	useCase := new(MockGetWeatherUseCase)
	handler := NewWeatherHandler(useCase)
	ctx := context.Background()

	req := httptest.NewRequest(http.MethodGet, "/weather?city=NonExistent", nil)
	rec := httptest.NewRecorder()

	useCase.
		On("GetWeather", ctx, "NonExistent").
		Return(nil, weather.ErrWeatherNotFound).
		Once()

	// Act
	handler.GetWeatherHandler(rec, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "weather data not found")

	useCase.AssertExpectations(t)
}

func TestGetWeatherHandler_WeatherUnavailable(t *testing.T) {
	// Arrange
	useCase := new(MockGetWeatherUseCase)
	handler := NewWeatherHandler(useCase)
	ctx := context.Background()

	req := httptest.NewRequest(http.MethodGet, "/weather?city=Athens", nil)
	rec := httptest.NewRecorder()

	useCase.
		On("GetWeather", ctx, "Athens").
		Return(nil, weather.ErrWeatherUnavailable).
		Once()

	// Act
	handler.GetWeatherHandler(rec, req)

	// Assert
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), "weather service is currently unavailable")

	useCase.AssertExpectations(t)
}

func TestGetWeatherHandler_UnknownError(t *testing.T) {
	// Arrange
	useCase := new(MockGetWeatherUseCase)
	handler := NewWeatherHandler(useCase)
	ctx := context.Background()

	req := httptest.NewRequest(http.MethodGet, "/weather?city=Athens", nil)
	rec := httptest.NewRecorder()

	useCase.
		On("GetWeather", ctx, "Athens").
		Return(nil, errors.New("unknown error")).
		Once()

	// Act
	handler.GetWeatherHandler(rec, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "internal server error")

	useCase.AssertExpectations(t)
}

func TestGetWeatherHandler_Success(t *testing.T) {
	// Arrange
	useCase := new(MockGetWeatherUseCase)
	handler := NewWeatherHandler(useCase)
	ctx := context.Background()

	req := httptest.NewRequest(http.MethodGet, "/weather?city=Athens", nil)
	rec := httptest.NewRecorder()

	weatherData := createSampleDomainWeather()

	useCase.
		On("GetWeather", ctx, "Athens").
		Return(weatherData, nil).
		Once()

	// Act
	handler.GetWeatherHandler(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	// Verify response body
	var response dto.WeatherResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)

	require.NoError(t, err)
	assert.Equal(t, "Athens", response.Location)
	assert.Equal(t, 24.5, response.Temperature)
	assert.Equal(t, "Sunny", response.Condition)

	useCase.AssertExpectations(t)
}

func TestGetWeatherHandler_SpecialCharactersInCity(t *testing.T) {
	// Arrange
	useCase := new(MockGetWeatherUseCase)
	handler := NewWeatherHandler(useCase)
	ctx := context.Background()

	req := httptest.NewRequest(http.MethodGet, "/weather?city=São Paulo", nil)
	rec := httptest.NewRecorder()

	weatherData := &weather.Weather{
		Location: weather.Location{
			Name:    "São Paulo",
			Country: "Brazil",
		},
		Current: weather.CurrentWeather{
			Temperature: weather.Temperature{
				Celsius: 28.0,
			},
			Condition: weather.Condition{
				Text: "Partly cloudy",
			},
		},
		UpdatedAt: time.Now(),
	}

	useCase.
		On("GetWeather", ctx, "São Paulo").
		Return(weatherData, nil).
		Once()

	// Act
	handler.GetWeatherHandler(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.WeatherResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)

	require.NoError(t, err)
	assert.Equal(t, "São Paulo", response.Location)

	useCase.AssertExpectations(t)
}
