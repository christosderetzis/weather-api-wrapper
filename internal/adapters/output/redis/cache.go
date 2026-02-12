package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"weather-api-wrapper/internal/domain/weather"
)

// Cache implements the WeatherCache port using Redis
type Cache struct {
	client *redis.Client
}

// NewCache creates a new Redis cache adapter
func NewCache(host string, port string) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
	})

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Cache{
		client: client,
	}, nil
}

// Get retrieves weather data from Redis cache
// Returns nil and no error if the key doesn't exist (cache miss)
func (c *Cache) Get(ctx context.Context, location string) (*weather.Weather, error) {
	data, err := c.client.Get(ctx, location).Result()
	if err != nil {
		// redis.Nil indicates the key doesn't exist (cache miss)
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var weatherData weather.Weather
	if err := json.Unmarshal([]byte(data), &weatherData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return &weatherData, nil
}

// Set stores weather data in Redis cache with the given TTL
func (c *Cache) Set(ctx context.Context, location string, data *weather.Weather, ttl time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal weather data: %w", err)
	}

	if err := c.client.Set(ctx, location, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Close closes the Redis client connection
func (c *Cache) Close() error {
	return c.client.Close()
}
