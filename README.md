# Weather API Wrapper

A Go REST API that wraps an external weather API with Redis caching. Inspiration: [Weather API](https://roadmap.sh/projects/weather-api-wrapper-service)

![Architecture Diagram](diagram.jpg)

## Prerequisites

- Go 1.21+
- Docker and Docker Compose

## Configuration

Copy `.env.example` to `.env` and set your values:

```bash
cp .env.example .env
```

| Variable | Description | Default |
|----------|-------------|---------|
| `WEATHER_API_KEY` | API key for weather provider | `test_api_key` |
| `WEATHER_API_BASE_URL` | Base URL for weather API | `https://api.weatherapi.com/v1/current.json` |
| `REDIS_HOST` | Redis hostname | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |

## Running

### Option 1: Full Stack with Docker Compose (Recommended)

Run the entire stack (API, Redis, Prometheus, Grafana) with one command:

```bash
docker compose -f docker/docker-compose.yml up -d
```

This starts:
- **Weather API**: http://localhost:8080
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Redis**: localhost:6379

To view logs:
```bash
docker compose -f docker/docker-compose.yml logs -f weather-api
```

To stop:
```bash
docker compose -f docker/docker-compose.yml down
```

### Option 2: Local Development

1. Start Redis:
```bash
docker compose -f docker/docker-compose.yml up -d redis
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

### Metrics Endpoint

```
GET /metrics
```

Prometheus-formatted metrics including:
- HTTP request counts and latency
- Cache hit/miss rates
- External API call performance
- Rate limit violations

## Features

### Structured Logging
All HTTP requests are logged with structured format:
```
[2026-02-10 15:04:05] GET /weather?city=London 200
[2026-02-10 15:04:12] GET /weather?city=Paris 200
[2026-02-10 15:05:01] GET /weather?city=Tokyo 429
```

Format: `[timestamp] METHOD path status_code`

### Prometheus Metrics
Comprehensive metrics exposed at `/metrics`:

- **HTTP Metrics**
  - `weather_api_http_requests_total` - Total HTTP requests by method, path, and status
  - `weather_api_http_request_duration_seconds` - Request latency histogram

- **Cache Metrics**
  - `weather_api_cache_hits_total` - Total cache hits
  - `weather_api_cache_misses_total` - Total cache misses
  - `weather_api_cache_errors_total` - Total cache errors

- **External API Metrics**
  - `weather_api_external_api_calls_total` - External API calls by provider and status
  - `weather_api_external_api_call_duration_seconds` - External API call latency

- **Rate Limiting Metrics**
  - `weather_api_rate_limit_exceeded_total` - Total rate limit violations

### Grafana Dashboard
Pre-configured dashboard with panels for:
- Request rate and latency (p50, p95, p99)
- Cache hit rate and operations
- Error rates (4xx, 5xx)
- External API performance
- Rate limit events

Access at http://localhost:3000 (admin/admin) when running via Docker Compose.

## Testing

Run all tests:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out
```

Run linters (requires golangci-lint):
```bash
golangci-lint run --timeout=5m
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
├── docker/
│   ├── docker-compose.yml             # Full stack (API, Redis, Prometheus, Grafana)
│   ├── grafana/
│   │   ├── dashboards/
│   │   │   └── weather-api-dashboard.json # Pre-configured dashboard
│   │   └── provisioning/              # Auto-provisioning configs
│   │       ├── datasources/
│   │       └── dashboards/
│   └── prometheus/
│       └── prometheus.yml             # Prometheus configuration
│
├── Dockerfile                         # Multi-stage production build
└── .dockerignore                      # Docker build exclusions
```

### Architecture Benefits

- ✅ **Testability**: Domain and application layers tested without infrastructure
- ✅ **Flexibility**: Easy to swap implementations (different cache, API, HTTP framework)
- ✅ **Maintainability**: Clear separation of concerns, SOLID principles
- ✅ **Independence**: Business logic isolated from frameworks and external services
