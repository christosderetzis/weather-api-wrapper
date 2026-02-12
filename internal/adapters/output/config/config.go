package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for the application
type Config struct {
	WeatherAPIKey     string
	WeatherAPIBaseURL string
	RedisHost         string
	RedisPort         string
}

// Load loads configuration from environment variables with fallback defaults
// It attempts to load from .env file if present
func Load() *Config {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	return &Config{
		WeatherAPIKey:     getEnv("WEATHER_API_KEY", "test_api_key"),
		WeatherAPIBaseURL: getEnv("WEATHER_API_BASE_URL", "https://base-url.com"),
		RedisHost:         getEnv("REDIS_HOST", "localhost"),
		RedisPort:         getEnv("REDIS_PORT", "6379"),
	}
}

// getEnv retrieves an environment variable or returns a fallback value
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
