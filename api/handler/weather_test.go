package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"weather-api-wrapper/api/dto"
	"weather-api-wrapper/weather"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockWeatherService struct {
	mock.Mock
}

func (m *MockWeatherService) GetWeather(
	ctx context.Context,
	location string,
) (*weather.WeatherResponse, error) {
	args := m.Called(ctx, location)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*weather.WeatherResponse), args.Error(1)
}

func TestGetWeatherHandler_MissingCity(t *testing.T) {
	// given
	service := new(MockWeatherService)
	handler := NewWeatherHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/weather", nil)
	rec := httptest.NewRecorder()

	// when
	handler.GetWeatherHandler(rec, req)

	// then
	assert.Equal(t, rec.Code, http.StatusBadRequest)
	assert.Contains(t, rec.Body.String(), "city query parameter is required")

	// and service should not be called
	service.AssertNotCalled(t, "GetWeather", mock.Anything, mock.Anything)
}

func TestGetWeatherHandler_ServiceError(t *testing.T) {
	// given
	service := new(MockWeatherService)
	handler := NewWeatherHandler(service)
	ctx := context.Background()

	req := httptest.NewRequest(http.MethodGet, "/weather?city=Athens", nil)
	rec := httptest.NewRecorder()

	service.
		On("GetWeather", ctx, "Athens").
		Return(nil, errors.New("service failed")).
		Once()

	// when
	handler.GetWeatherHandler(rec, req)

	// then
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "service failed")

	// and service should be called once
	service.AssertExpectations(t)
}

func TestGetWeatherHandler_Success(t *testing.T) {
	// given
	service := new(MockWeatherService)
	handler := NewWeatherHandler(service)
	ctx := context.Background()

	req := httptest.NewRequest(http.MethodGet, "/weather?city=Athens", nil)
	rec := httptest.NewRecorder()

	weatherResponse := &weather.WeatherResponse{
		Location: weather.Location{
			Name: "Athens",
		},
		Current: weather.Current{
			TempC: 24.5,
			Condition: weather.Condition{
				Text: "Sunny",
			},
		},
	}

	service.
		On("GetWeather", ctx, "Athens").
		Return(weatherResponse, nil).
		Once()

	// when
	handler.GetWeatherHandler(rec, req)

	// then
	assert.Equal(t, rec.Code, http.StatusOK)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	// and response body should be correct
	var body dto.WeatherResponse
	err := json.Unmarshal(rec.Body.Bytes(), &body)

	assert.NoError(t, err)
	assert.Equal(t, "Athens", body.Location)
	assert.Equal(t, 24.5, body.Temperature)
	assert.Equal(t, "Sunny", body.Condition)

	service.AssertExpectations(t)
}
