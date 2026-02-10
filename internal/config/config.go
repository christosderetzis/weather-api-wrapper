package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ApiKey     string
	BaseApiUrl string
	RedisHost  string
	RedisPort  string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		ApiKey:     getEnv("WEATHER_API_KEY", "test_api_key"),
		BaseApiUrl: getEnv("WEATHER_API_BASE_URL", "https://base-url.com"),
		RedisHost:  getEnv("REDIS_HOST", "localhost"),
		RedisPort:  getEnv("REDIS_PORT", "6379"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
