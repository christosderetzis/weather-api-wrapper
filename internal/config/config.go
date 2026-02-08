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
	_ = godotenv.Load() // Load .env file if it exists

	return &Config{
		ApiKey:     os.Getenv("WEATHER_API_KEY"),
		BaseApiUrl: os.Getenv("WEATHER_API_BASE_URL"),
	}
}
