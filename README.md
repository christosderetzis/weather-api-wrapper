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

## Project Structure

```
.
├── api/
│   ├── dto/          # Data transfer objects
│   ├── handler/      # HTTP handlers
│   ├── middleware/   # HTTP middleware (rate limiting, logging)
│   └── routes/       # Route definitions
├── cmd/server/       # Application entrypoint
├── docker/           # Docker compose files
├── internal/
│   ├── cache/        # Redis client
│   └── config/       # Configuration
└── weather/          # Weather domain (client, service, repository)
```
