package weather

import (
	"context"
	"fmt"
	"log"
	"time"
)

type WeatherService struct {
	Client     WeatherClientInterface
	Repository WeatherRepositoryInterface
}

func NewWeatherService(weatherClient WeatherClientInterface, repository WeatherRepositoryInterface) *WeatherService {
	return &WeatherService{
		Client:     weatherClient,
		Repository: repository,
	}
}

func (ws *WeatherService) GetWeather(ctx context.Context, location string) (*WeatherResponse, error) {
	cachedWeather, err := ws.Repository.GetWeatherFromCache(ctx, location)
	if err == nil && cachedWeather != nil {
		log.Println("Cache hit for location:", location)
		return cachedWeather, nil
	}

	log.Println("Cache miss for location:", location)
	weather, err := ws.Client.FetchWeatherFromApi(ctx, location)
	if err != nil {
		return nil, err
	}

	if err := ws.Repository.SetWeatherInCache(ctx, location, weather, time.Hour*12); err != nil {
		fmt.Printf("Warning: failed to cache weather data: %v\n", err)
	}

	return weather, nil
}
