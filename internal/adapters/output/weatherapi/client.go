package weatherapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"weather-api-wrapper/internal/domain/weather"
)

// Infrastructure-specific errors
var (
	ErrFailedToFetchWeather   = errors.New("failed to fetch weather data")
	ErrAPIReturnedNonOKStatus = errors.New("API returned non-OK status")
	ErrParseWeatherData       = errors.New("failed to parse weather data")
	ErrSerializationData      = errors.New("failed to serialize weather data")
)

// Client implements the WeatherProvider port for WeatherAPI.com
type Client struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewClient creates a new WeatherAPI client adapter
func NewClient(apiKey string, baseURL string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  http.DefaultClient,
	}
}

// FetchWeather implements the WeatherProvider port
// It fetches weather data from the external WeatherAPI.com service
// and converts the response to domain models
func (c *Client) FetchWeather(ctx context.Context, location string) (*weather.Weather, error) {
	// Build the API request URL
	reqURL := fmt.Sprintf("%s?key=%s&q=%s&aqi=no", c.baseURL, c.apiKey, url.QueryEscape(location))

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToFetchWeather, err)
	}

	// Execute the request
	resp, err := c.client.Do(req)
	if err != nil {
		// Check if the error is due to context cancellation
		if ctx.Err() != nil {
			return nil, fmt.Errorf("%w: %v", ErrFailedToFetchWeather, ctx.Err())
		}
		return nil, fmt.Errorf("%w: %v", ErrFailedToFetchWeather, err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseWeatherData, err)
	}

	// Check for non-OK status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d, response: %s", ErrAPIReturnedNonOKStatus, resp.StatusCode, string(body))
	}

	// Unmarshal into API-specific model
	var apiResponse APIWeatherResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSerializationData, err)
	}

	// Map API model to domain model
	domainWeather := MapAPIResponseToDomain(&apiResponse)

	return domainWeather, nil
}
