package weather

import (
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
)

type WeatherClient struct {
	ApiKey  string
	BaseUrl string
}

func NewWeatherClient(apiKey string, baseUrl string) *WeatherClient {
	return &WeatherClient{
		ApiKey:  apiKey,
		BaseUrl: baseUrl,
	}
}

func (wc *WeatherClient) FetchWeatherFromApi(location string) (*WeatherResponse, error) {
	// Implementation goes here
	reqURL := fmt.Sprintf("%s?key=%s&q=%s&aqi=no", wc.BaseUrl, wc.ApiKey, url.QueryEscape(location))

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, ErrorFailedToFetchWeather
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d, response: %s", ErrorApiReturnedNonOKStatus, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrorParseWeatherData, err)
	}

	var weather WeatherResponse
	if err := json.Unmarshal(body, &weather); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrorParseWeatherData, err)
	}

	return &weather, nil
}
