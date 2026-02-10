package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var (
	ErrorFailedToFetchWeather   = errors.New("failed to fetch weather data")
	ErrorApiReturnedNonOKStatus = errors.New("API returned non-OK status")
	ErrorParseWeatherData       = errors.New("failed to parse weather data")
	ErrorSerializarionData      = errors.New("failed to serialize weather data")
)

type WeatherClientInterface interface {
	FetchWeatherFromApi(ctx context.Context, location string) (*WeatherResponse, error)
}

type WeatherClientImpl struct {
	ApiKey  string
	BaseUrl string
}

func NewWeatherClient(apiKey string, baseUrl string) *WeatherClientImpl {
	return &WeatherClientImpl{
		ApiKey:  apiKey,
		BaseUrl: baseUrl,
	}
}

func (wc *WeatherClientImpl) FetchWeatherFromApi(ctx context.Context, location string) (*WeatherResponse, error) {
	reqURL := fmt.Sprintf("%s?key=%s&q=%s&aqi=no", wc.BaseUrl, wc.ApiKey, url.QueryEscape(location))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrorFailedToFetchWeather, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("%w: %v", ErrorFailedToFetchWeather, ctx.Err())
		}
		return nil, ErrorFailedToFetchWeather
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrorParseWeatherData, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d, response: %s", ErrorApiReturnedNonOKStatus, resp.StatusCode, string(body))
	}

	var weather WeatherResponse
	if err := json.Unmarshal(body, &weather); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrorSerializarionData, err)
	}

	return &weather, nil
}
