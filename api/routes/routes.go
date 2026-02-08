package routes

import (
	"net/http"
	"weather-api-wrapper/api/handler"
)

func SetupRoutes(handler *handler.WeatherHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/weather", handler.GetWeatherHandler)

	return mux
}
