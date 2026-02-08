package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ApiKey     string
	BaseApiUrl string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		ApiKey:     getEnv("WEATHER_API_KEY", "test_api_key"),
		BaseApiUrl: getEnv("WEATHER_API_BASE_URL", "https://base-url.com"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
