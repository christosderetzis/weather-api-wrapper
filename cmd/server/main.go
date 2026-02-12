package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"weather-api-wrapper/internal/adapters/input/http/handlers"
	"weather-api-wrapper/internal/adapters/input/http/routes"
	"weather-api-wrapper/internal/adapters/output/config"
	"weather-api-wrapper/internal/adapters/output/redis"
	"weather-api-wrapper/internal/adapters/output/weatherapi"
	weatherapp "weather-api-wrapper/internal/application/weather"
)

func main() {
	// 1. Load configuration (configuration adapter)
	cfg := config.Load()
	log.Println("Configuration loaded successfully")

	// 2. Initialize output adapters (secondary/driven)

	// Initialize Redis cache adapter
	redisCache, err := redis.NewCache(cfg.RedisHost, cfg.RedisPort)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Redis cache connected successfully")

	// Initialize Weather API client adapter
	weatherAPIClient := weatherapi.NewClient(cfg.WeatherAPIKey, cfg.WeatherAPIBaseURL)
	log.Println("Weather API client initialized")

	// 3. Initialize application service (core business logic)
	weatherService := weatherapp.NewService(weatherAPIClient, redisCache)
	log.Println("Weather application service initialized")

	// 4. Initialize input adapter (primary/driving)
	weatherHandler := handlers.NewWeatherHandler(weatherService)
	log.Println("HTTP handler initialized")

	// 5. Setup routes with middleware chain
	router := routes.SetupRoutes(weatherHandler)
	log.Println("Routes configured with middleware")

	port := ":8080"
	server := &http.Server{
		Addr:    port,
		Handler: router,
	}

	// Channel to receive server errors
	serverErr := make(chan error, 1)

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Channel to listen for shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErr:
		log.Fatalf("Server failed to start: %v", err)
	case sig := <-shutdown:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
	}

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown of HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
		// Force close if graceful shutdown fails
		if err := server.Close(); err != nil {
			log.Printf("HTTP server force close error: %v", err)
		}
	} else {
		log.Println("HTTP server stopped gracefully")
	}

	// Close Redis connection
	if err := redisCache.Close(); err != nil {
		log.Printf("Redis cache close error: %v", err)
	} else {
		log.Println("Redis cache closed")
	}

	log.Println("Shutdown complete")
}
