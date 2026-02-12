package weatherapi

import (
	"time"

	"weather-api-wrapper/internal/domain/weather"
)

// MapAPIResponseToDomain converts the external API response model to domain model
// This keeps the domain layer clean from infrastructure concerns (JSON tags, API structure)
func MapAPIResponseToDomain(apiResponse *APIWeatherResponse) *weather.Weather {
	return &weather.Weather{
		Location: weather.Location{
			Name:      apiResponse.Location.Name,
			Region:    apiResponse.Location.Region,
			Country:   apiResponse.Location.Country,
			Latitude:  apiResponse.Location.Lat,
			Longitude: apiResponse.Location.Lon,
			Timezone:  apiResponse.Location.TzID,
			LocalTime: time.Unix(apiResponse.Location.LocaltimeEpoch, 0),
		},
		Current: weather.CurrentWeather{
			LastUpdated: time.Unix(apiResponse.Current.LastUpdatedEpoch, 0),
			Temperature: weather.Temperature{
				Celsius:    apiResponse.Current.TempC,
				Fahrenheit: apiResponse.Current.TempF,
				FeelsLike: weather.FeelsLike{
					Celsius:    apiResponse.Current.FeelslikeC,
					Fahrenheit: apiResponse.Current.FeelslikeF,
				},
				Windchill: weather.TemperatureValue{
					Celsius:    apiResponse.Current.WindchillC,
					Fahrenheit: apiResponse.Current.WindchillF,
				},
				HeatIndex: weather.TemperatureValue{
					Celsius:    apiResponse.Current.HeatindexC,
					Fahrenheit: apiResponse.Current.HeatindexF,
				},
				Dewpoint: weather.TemperatureValue{
					Celsius:    apiResponse.Current.DewpointC,
					Fahrenheit: apiResponse.Current.DewpointF,
				},
			},
			Condition: weather.Condition{
				Text: apiResponse.Current.Condition.Text,
				Code: apiResponse.Current.Condition.Code,
				Icon: apiResponse.Current.Condition.Icon,
			},
			Wind: weather.Wind{
				SpeedKph:  apiResponse.Current.WindKph,
				SpeedMph:  apiResponse.Current.WindMph,
				Direction: apiResponse.Current.WindDir,
				Degree:    apiResponse.Current.WindDegree,
				GustKph:   apiResponse.Current.GustKph,
				GustMph:   apiResponse.Current.GustMph,
			},
			Pressure: weather.Pressure{
				Millibars: apiResponse.Current.PressureMb,
				Inches:    apiResponse.Current.PressureIn,
			},
			Precipitation: weather.Precipitation{
				Millimeters: apiResponse.Current.PrecipMm,
				Inches:      apiResponse.Current.PrecipIn,
			},
			Humidity:   apiResponse.Current.Humidity,
			CloudCover: apiResponse.Current.Cloud,
			Visibility: weather.Distance{
				Kilometers: apiResponse.Current.VisKm,
				Miles:      apiResponse.Current.VisMiles,
			},
			UVIndex: apiResponse.Current.UV,
			IsDay:   apiResponse.Current.IsDay == 1,
			Radiation: weather.Radiation{
				ShortWave: apiResponse.Current.ShortRad,
				Diffuse:   apiResponse.Current.DiffRad,
				DNI:       apiResponse.Current.DNI,
				GTI:       apiResponse.Current.GTI,
			},
		},
		UpdatedAt: time.Now(),
	}
}
