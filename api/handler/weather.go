package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"weather-api-wrapper/api/dto"
	"weather-api-wrapper/weather"
)

type WeatherServiceInterface interface {
	GetWeather(ctx context.Context, location string) (*weather.WeatherResponse, error)
}

type WeatherHandler struct {
	Service WeatherServiceInterface
}

func NewWeatherHandler(service WeatherServiceInterface) *WeatherHandler {
	return &WeatherHandler{
		Service: service,
	}
}

func (wh *WeatherHandler) GetWeatherHandler(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, "city query parameter is required", http.StatusBadRequest)
		return
	}

	weatherData, err := wh.Service.GetWeather(r.Context(), city)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	weatherResponseBody := dto.WeatherResponse{
		Location:    weatherData.Location.Name,
		Temperature: weatherData.Current.TempC,
		Condition:   weatherData.Current.Condition.Text,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(weatherResponseBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
