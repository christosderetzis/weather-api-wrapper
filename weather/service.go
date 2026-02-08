package weather

import (
	"context"
	"fmt"
	"log"
	"weather-api-wrapper/internal/cache"

	"github.com/redis/go-redis/v9"
)

type WeatherService struct {
	WeatherClient *WeatherClient
	Cache         *redis.Client
}

func NewWeatherService(weatherClient *WeatherClient, cache *redis.Client) *WeatherService {
	return &WeatherService{
		WeatherClient: weatherClient,
		Cache:         cache,
	}
}

func (ws *WeatherService) GetWeather(ctx context.Context, location string) (*WeatherResponse, error) {
	cachedWeather, err := cache.Get[WeatherResponse](ctx, *ws.Cache, location)
	if err == nil && cachedWeather != nil {
		log.Println("Cache hit for location:", location)
		return cachedWeather, nil
	}

	log.Println("Cache miss for location:", location)
	weather, err := ws.WeatherClient.FetchWeatherFromApi(location)
	if err != nil {
		return nil, err
	}

	if err := cache.Set(ctx, ws.Cache, location, weather); err != nil {
		fmt.Printf("Warning: failed to cache weather data: %v\n", err)
	}

	return weather, nil
}
