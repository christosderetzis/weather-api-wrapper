package weatherapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"weather-api-wrapper/internal/adapters/input/http/middleware/metrics"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-key", "http://api.example.com")

	assert.NotNil(t, client)
	assert.Equal(t, "test-key", client.apiKey)
	assert.Equal(t, "http://api.example.com", client.baseURL)
}

func TestClient_FetchWeather(t *testing.T) {
	tests := []struct {
		name              string
		serverStatus      int
		serverResponse    string
		location          string
		expectError       bool
		expectedLocation  string
		expectedTempC     float64
		expectedCondition string
		expectedMetricLbl string // expected label for ExternalAPICallsTotal
	}{
		{
			name:         "Success",
			serverStatus: http.StatusOK,
			serverResponse: `{
				"location": {
					"name": "London",
					"region": "City of London, Greater London",
					"country": "United Kingdom",
					"lat": 51.52,
					"lon": -0.11,
					"tz_id": "Europe/London",
					"localtime_epoch": 1234567890,
					"localtime": "2009-02-13 23:31"
				},
				"current": {
					"last_updated_epoch": 1234567890,
					"last_updated": "2009-02-13 23:30",
					"temp_c": 15.0,
					"temp_f": 59.0,
					"is_day": 0,
					"condition": {
						"text": "Partly cloudy",
						"icon": "//cdn.weatherapi.com/weather/64x64/night/116.png",
						"code": 1003
					},
					"wind_mph": 8.1,
					"wind_kph": 13.0,
					"wind_degree": 230,
					"wind_dir": "SW",
					"pressure_mb": 1012.0,
					"pressure_in": 29.88,
					"precip_mm": 0.0,
					"precip_in": 0.0,
					"humidity": 82,
					"cloud": 75,
					"feelslike_c": 13.5,
					"feelslike_f": 56.4,
					"windchill_c": 13.5,
					"windchill_f": 56.4,
					"heatindex_c": 15.0,
					"heatindex_f": 59.0,
					"dewpoint_c": 12.0,
					"dewpoint_f": 53.6,
					"vis_km": 10.0,
					"vis_miles": 6.0,
					"uv": 1.0,
					"gust_mph": 14.1,
					"gust_kph": 22.7,
					"short_rad": 0.0,
					"diff_rad": 0.0,
					"dni": 0.0,
					"gti": 0.0
				}
			}`,
			location:          "London",
			expectError:       false,
			expectedLocation:  "London",
			expectedTempC:     15.0,
			expectedCondition: "Partly cloudy",
			expectedMetricLbl: "200",
		},
		{
			name:              "Bad request",
			serverStatus:      http.StatusBadRequest,
			serverResponse:    `{"error": {"message": "Invalid location"}}`,
			location:          "InvalidLocation",
			expectError:       true,
			expectedMetricLbl: "400",
		},
		{
			name:              "Internal server error",
			serverStatus:      http.StatusInternalServerError,
			serverResponse:    "internal server error",
			location:          "London",
			expectError:       true,
			expectedMetricLbl: "500",
		},
		{
			name:              "Invalid JSON response",
			serverStatus:      http.StatusOK,
			serverResponse:    "not valid json",
			location:          "London",
			expectError:       true,
			expectedMetricLbl: "200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				_, _ = w.Write([]byte(tt.serverResponse))

				// Verify query parameters
				assert.Equal(t, tt.location, r.URL.Query().Get("q"))
				assert.Equal(t, "no", r.URL.Query().Get("aqi"))
			}))
			defer server.Close()

			// Create client with mock server URL
			client := NewClient("test-key", server.URL)

			// Capture initial metric value
			initialAPICalls := testutil.ToFloat64(metrics.ExternalAPICallsTotal.WithLabelValues("weatherapi", tt.expectedMetricLbl))

			// Execute
			ctx := context.Background()
			result, err := client.FetchWeather(ctx, tt.location)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedLocation, result.Location.Name)
				assert.Equal(t, tt.expectedTempC, result.Current.Temperature.Celsius)
				assert.Equal(t, tt.expectedCondition, result.Current.Condition.Text)
				assert.Equal(t, "United Kingdom", result.Location.Country)
				assert.Equal(t, 51.52, result.Location.Latitude)
				assert.Equal(t, 13.0, result.Current.Wind.SpeedKph)
				assert.Equal(t, 82, result.Current.Humidity)
				assert.False(t, result.Current.IsDay) // is_day: 0 -> false
			}

			// Verify API call metric was incremented
			assert.Equal(t, initialAPICalls+1, testutil.ToFloat64(metrics.ExternalAPICallsTotal.WithLabelValues("weatherapi", tt.expectedMetricLbl)))
		})
	}
}

func TestClient_FetchWeather_ContextCancellation(t *testing.T) {
	// Create a server that never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block forever
		select {}
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL)

	// Capture initial metric value
	initialAPIErrors := testutil.ToFloat64(metrics.ExternalAPICallsTotal.WithLabelValues("weatherapi", "error"))

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := client.FetchWeather(ctx, "London")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrFailedToFetchWeather)

	// Verify error metric was incremented
	assert.Equal(t, initialAPIErrors+1, testutil.ToFloat64(metrics.ExternalAPICallsTotal.WithLabelValues("weatherapi", "error")))
}

func TestClient_FetchWeather_URLEncoding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that the location is properly URL-encoded
		location := r.URL.Query().Get("q")
		assert.Equal(t, "New York", location) // httptest automatically decodes

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"location": {"name": "New York"},
			"current": {"temp_c": 20.0, "condition": {"text": "Clear"}}
		}`))
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL)

	// Capture initial metric value
	initialAPICalls := testutil.ToFloat64(metrics.ExternalAPICallsTotal.WithLabelValues("weatherapi", "200"))

	ctx := context.Background()
	result, err := client.FetchWeather(ctx, "New York")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "New York", result.Location.Name)

	// Verify API call metric was incremented
	assert.Equal(t, initialAPICalls+1, testutil.ToFloat64(metrics.ExternalAPICallsTotal.WithLabelValues("weatherapi", "200")))
}
