package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"weather-api-wrapper/api/handler"
	"weather-api-wrapper/api/routes"
	"weather-api-wrapper/internal/cache"
	"weather-api-wrapper/internal/config"
	"weather-api-wrapper/weather"
)

func main() {
	configuration := config.LoadConfig()

	// initialize Redis client
	redisClient, err := cache.InitRedisClient(fmt.Sprintf("%s:%s", configuration.RedisHost, configuration.RedisPort))
	if err != nil {
		fmt.Printf("Failed to connect to Redis: %v\n", err)
		return
	}

	// dependency injection
	weatherClient := weather.NewWeatherClient(configuration.ApiKey, configuration.BaseApiUrl)
	weatherRepository := weather.NewWeatherRepository(redisClient)
	weatherService := weather.NewWeatherService(weatherClient, weatherRepository)
	weatherHandler := handler.NewWeatherHandler(weatherService)

	router := routes.SetupRoutes(weatherHandler)

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
	if err := redisClient.Close(); err != nil {
		log.Printf("Redis client close error: %v", err)
	} else {
		log.Println("Redis connection closed")
	}

	log.Println("Shutdown complete")
}
