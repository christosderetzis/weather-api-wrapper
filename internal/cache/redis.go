package cache

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type Cache[T any] interface {
	Set(ctx context.Context, key string, value *T) error
	Get(ctx context.Context, key string) (*T, error)
}

var redisClient *redis.Client

func InitRedisClient(addr string) (*redis.Client, error) {
	redisClient = redis.NewClient(&redis.Options{
		Addr: addr,
	})
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return redisClient, nil
}

func Set[T any](ctx context.Context, c *redis.Client, key string, value *T) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, data, 0).Err()
}

func Get[T any](ctx context.Context, c redis.Client, key string) (*T, error) {
	data, err := c.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
