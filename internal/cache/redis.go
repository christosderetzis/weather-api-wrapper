package cache

import (
	"context"

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
