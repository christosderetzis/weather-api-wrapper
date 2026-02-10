package weather

import (
	"context"
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
		name              string
		serverStatus      int
		serverResponse    string
		city              string
		expectError       bool
		expectedLocation  string
		expectedTempC     float64
		expectedCondition string
	}{
		{
			name:              "Success",
			serverStatus:      http.StatusOK,
			serverResponse:    `{"location":{"name":"London","country":"UK"},"current":{"temp_c":15,"condition":{"text":"Partly cloudy"}}}`,
			city:              "London",
			expectError:       false,
			expectedLocation:  "London",
			expectedTempC:     15.0,
			expectedCondition: "Partly cloudy",
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

				assert.Equal(t, "London", r.URL.Query().Get("q"))
				assert.Equal(t, "no", r.URL.Query().Get("aqi"))
				assert.Equal(t, http.MethodGet, r.Method)
			}))
			defer server.Close()

			client := NewWeatherClient("test-key", server.URL)

			result, err := client.FetchWeatherFromApi(context.Background(), tt.city)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedLocation, result.Location.Name)
				assert.Equal(t, tt.expectedTempC, result.Current.TempC)
				assert.Equal(t, tt.expectedCondition, result.Current.Condition.Text)
			}
		})
	}
}

func TestWeatherClient_FetchWeatherFromApi_HttpError(t *testing.T) {
	client := NewWeatherClient("test-key", "http://invalid.invalid.invalid:99999")

	_, err := client.FetchWeatherFromApi(context.Background(), "London")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrorFailedToFetchWeather)
}

func TestWeatherClient_FetchWeatherFromApi_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		<-r.Context().Done()
	}))
	defer server.Close()

	client := NewWeatherClient("test-key", server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.FetchWeatherFromApi(ctx, "London")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrorFailedToFetchWeather)
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
