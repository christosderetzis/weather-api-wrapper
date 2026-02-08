package weather

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWeatherClient(t *testing.T) {
	client := NewWeatherClient("test-key", "http://api.example.com")

	assert.Equal(t, "test-key", client.ApiKey)
	assert.Equal(t, "http://api.example.com", client.BaseUrl)
}

func TestWeatherClient_FetchWeatherFromApi(t *testing.T) {
	tests := []struct {
		name             string
		serverStatus     int
		serverResponse   string
		city             string
		expectError      bool
		expectedLocation string
	}{
		{
			name:             "Success",
			serverStatus:     http.StatusOK,
			serverResponse:   `{"location":{"name":"London","country":"UK"},"current":{"temp_c":15,"condition":{"text":"Partly cloudy"}}}`,
			city:             "London",
			expectError:      false,
			expectedLocation: "London",
		},
		{
			name:           "Bad request",
			serverStatus:   http.StatusBadRequest,
			serverResponse: "bad request",
			city:           "London",
			expectError:    true,
		},
		{
			name:           "Internal server error",
			serverStatus:   http.StatusInternalServerError,
			serverResponse: "internal server error",
			city:           "London",
			expectError:    true,
		},
		{
			name:           "Invalid JSON response",
			serverStatus:   http.StatusOK,
			serverResponse: "not valid json",
			city:           "London",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := NewWeatherClient("test-key", server.URL)

			result, err := client.FetchWeatherFromApi(tt.city)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedLocation, result.Location.Name)
			}
		})
	}
}

func TestWeatherClient_FetchWeatherFromApi_HttpError(t *testing.T) {
	client := NewWeatherClient("test-key", "http://invalid.invalid.invalid:99999")

	_, err := client.FetchWeatherFromApi("London")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrorFailedToFetchWeather)
}

func TestWeatherClient_FetchWeatherFromApi_UrlEncoding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "New York", r.URL.Query().Get("q"))
		json.NewEncoder(w).Encode(sampleWeatherResponse())
	}))
	defer server.Close()

	client := NewWeatherClient("test-key", server.URL)

	_, err := client.FetchWeatherFromApi("New York")

	assert.NoError(t, err)
}

func sampleWeatherResponse() *WeatherResponse {
	return &WeatherResponse{
		Location: Location{
			Name:    "London",
			Country: "UK",
		},
		Current: Current{
			TempC: 15.0,
			Condition: Condition{
				Text: "Partly cloudy",
			},
		},
	}
}
