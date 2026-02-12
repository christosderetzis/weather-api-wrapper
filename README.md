# Weather API Wrapper

A Go REST API that wraps an external weather API with Redis caching. Inspiration: [Weather API](https://roadmap.sh/projects/weather-api-wrapper-service)

![Architecture Diagram](diagram.jpg)

## Prerequisites

- Go 1.21+
- Docker (for Redis)

## Configuration

Set the following environment variables (or create a `.env` file):

| Variable | Description | Default |
|----------|-------------|---------|
| `WEATHER_API_KEY` | API key for weather provider | `test_api_key` |
| `WEATHER_API_BASE_URL` | Base URL for weather API | `https://base-url.com` |

## Running

1. Start Redis:
```bash
docker compose -f docker/docker-compose.yml up -d
```

2. Run the server:
```bash
go run cmd/server/main.go
```

The server starts on port `8080`.

## API

### Get Weather

```
GET /weather?city={city}
```

**Response:**
```json
{
  "location": "London",
  "temperature_c": 15.5,
  "condition_text": "Partly cloudy"
}
```

**Rate Limiting:**
- Maximum 30 requests per minute per IP address
- Returns `429 Too Many Requests` when limit is exceeded

## Features

### Structured Logging
All HTTP requests are logged with structured format:
```
[2026-02-10 15:04:05] GET /weather?city=London 200
[2026-02-10 15:04:12] GET /weather?city=Paris 200
[2026-02-10 15:05:01] GET /weather?city=Tokyo 429
```

Format: `[timestamp] METHOD path status_code`

## Testing

```bash
go test ./...
```

## Architecture

This project implements **Hexagonal Architecture** (Ports and Adapters pattern), providing clear separation between business logic and infrastructure concerns.

### Key Principles

- **Domain Layer**: Pure business logic with zero infrastructure dependencies
- **Ports**: Interfaces defining boundaries (input ports for use cases, output ports for dependencies)
- **Adapters**: Infrastructure implementations (HTTP handlers, external APIs, databases)
- **Dependency Inversion**: All dependencies point inward toward the domain

### Dependency Flow

```
Domain (internal/domain/*)          ← No external dependencies
  ↑
Ports (internal/ports/*)            ← Defines interfaces, depends on domain
  ↑
Application (internal/application/*) ← Implements use cases, depends on ports
  ↑
Adapters (internal/adapters/*)      ← Implements ports, depends on everything above
  ↑
Main (cmd/server/main.go)           ← Wires everything together
```

## Project Structure

```
weather-api-wrapper/
├── cmd/
│   └── server/
│       └── main.go                    # Application composition root
│
├── internal/
│   ├── domain/                        # CORE - Pure business logic
│   │   └── weather/
│   │       ├── weather.go             # Rich domain entities with behavior
│   │       ├── errors.go              # Domain-specific errors
│   │       └── validation.go          # Business validation rules
│   │
│   ├── ports/                         # PORTS - Interfaces
│   │   ├── input/
│   │   │   └── weather_service.go     # GetWeatherUseCase interface
│   │   └── output/
│   │       ├── weather_provider.go    # External weather API port
│   │       └── weather_cache.go       # Cache port
│   │
│   ├── application/                   # Use case implementations
│   │   └── weather/
│   │       ├── service.go             # Implements GetWeatherUseCase
│   │       └── service_test.go        # Unit tests with mocked ports
│   │
│   └── adapters/                      # ADAPTERS - Infrastructure
│       ├── input/                     # Primary adapters (drivers)
│       │   └── http/
│       │       ├── handlers/          # HTTP handlers
│       │       ├── dto/               # HTTP-specific DTOs
│       │       ├── middleware/        # Logging, rate limiting
│       │       └── routes/            # Route configuration
│       │
│       └── output/                    # Secondary adapters (driven)
│           ├── weatherapi/            # WeatherAPI.com client
│           ├── redis/                 # Redis cache implementation
│           └── config/                # Configuration loader
│
└── docker/
    └── docker-compose.yml             # Redis container
```

### Architecture Benefits

- ✅ **Testability**: Domain and application layers tested without infrastructure
- ✅ **Flexibility**: Easy to swap implementations (different cache, API, HTTP framework)
- ✅ **Maintainability**: Clear separation of concerns, SOLID principles
- ✅ **Independence**: Business logic isolated from frameworks and external services
