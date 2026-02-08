package main

import (
	"fmt"
	"log"
	"net/http"
	"weather-api-wrapper/api/handler"
	"weather-api-wrapper/api/routes"
	"weather-api-wrapper/internal/cache"
	"weather-api-wrapper/internal/config"
	"weather-api-wrapper/weather"
)

func main() {
	redisClient, err := cache.InitRedisClient("localhost:6379")
	if err != nil {
		fmt.Printf("Failed to connect to Redis: %v\n", err)
		return
	}

	configuration := config.LoadConfig()

	// dependency injection
	weatherClient := weather.NewWeatherClient(configuration.ApiKey, configuration.BaseApiUrl)
	weatherRepository := weather.NewWeatherRepository(redisClient)
	weatherService := weather.NewWeatherService(weatherClient, weatherRepository)
	weatherHandler := handler.NewWeatherHandler(weatherService)

	router := routes.SetupRoutes(weatherHandler)

	port := ":8080"
	fmt.Printf("Server starting on port: %s\n", port)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
