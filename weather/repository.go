package weather

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type WeatherRepositoryInterface interface {
	GetWeatherFromCache(ctx context.Context, location string) (*WeatherResponse, error)
	SetWeatherInCache(ctx context.Context, location string, weather *WeatherResponse, ttl time.Duration) error
}

type WeatherRepository struct {
	Cache *redis.Client
}

func NewWeatherRepository(cache *redis.Client) *WeatherRepository {
	return &WeatherRepository{
		Cache: cache,
	}
}

func (wr *WeatherRepository) GetWeatherFromCache(ctx context.Context, location string) (*WeatherResponse, error) {
	data, err := wr.Cache.Get(ctx, location).Result()
	if err != nil {
		return nil, err
	}

	var weather WeatherResponse
	if err := json.Unmarshal([]byte(data), &weather); err != nil {
		return nil, err
	}

	return &weather, nil
}

func (wr *WeatherRepository) SetWeatherInCache(ctx context.Context, location string, weather *WeatherResponse, ttl time.Duration) error {
	data, err := json.Marshal(weather)
	if err != nil {
		return err
	}

	return wr.Cache.Set(ctx, location, data, ttl).Err()
}
