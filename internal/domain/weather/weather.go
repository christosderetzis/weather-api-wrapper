package weather

import "time"

// Weather is the core domain entity representing weather information for a location
type Weather struct {
	Location  Location
	Current   CurrentWeather
	UpdatedAt time.Time
}

// Location represents geographic information
type Location struct {
	Name      string
	Region    string
	Country   string
	Latitude  float64
	Longitude float64
	Timezone  string
	LocalTime time.Time
}

// CurrentWeather represents current meteorological conditions
type CurrentWeather struct {
	LastUpdated    time.Time
	Temperature    Temperature
	Condition      Condition
	Wind           Wind
	Pressure       Pressure
	Precipitation  Precipitation
	Humidity       int
	CloudCover     int
	Visibility     Distance
	UVIndex        float64
	IsDay          bool
	Radiation      Radiation
}

// Temperature holds temperature measurements in different units
type Temperature struct {
	Celsius    float64
	Fahrenheit float64
	FeelsLike  FeelsLike
	Windchill  TemperatureValue
	HeatIndex  TemperatureValue
	Dewpoint   TemperatureValue
}

// FeelsLike represents the perceived temperature
type FeelsLike struct {
	Celsius    float64
	Fahrenheit float64
}

// TemperatureValue represents a temperature in both Celsius and Fahrenheit
type TemperatureValue struct {
	Celsius    float64
	Fahrenheit float64
}

// Condition describes the weather phenomenon
type Condition struct {
	Text string
	Code int
	Icon string
}

// Wind holds wind measurements
type Wind struct {
	SpeedKph  float64
	SpeedMph  float64
	Direction string
	Degree    int
	GustKph   float64
	GustMph   float64
}

// Pressure holds atmospheric pressure measurements
type Pressure struct {
	Millibars float64
	Inches    float64
}

// Precipitation holds precipitation measurements
type Precipitation struct {
	Millimeters float64
	Inches      float64
}

// Distance holds distance measurements
type Distance struct {
	Kilometers float64
	Miles      float64
}

// Radiation holds solar radiation measurements
type Radiation struct {
	ShortWave float64 // Short-wave radiation
	Diffuse   float64 // Diffuse radiation
	DNI       float64 // Direct Normal Irradiance
	GTI       float64 // Global Tilted Irradiance
}

// Business Methods - Rich Domain Behavior

// IsFreezing returns true if the temperature is at or below freezing (0°C)
func (t Temperature) IsFreezing() bool {
	return t.Celsius <= 0
}

// IsHot returns true if the temperature is considered hot (>=30°C)
func (t Temperature) IsHot() bool {
	return t.Celsius >= 30
}

// IsCold returns true if the temperature is considered cold (<=10°C)
func (t Temperature) IsCold() bool {
	return t.Celsius <= 10
}

// IsExtreme returns true if temperature is extremely hot (>=40°C) or cold (<=−20°C)
func (t Temperature) IsExtreme() bool {
	return t.Celsius >= 40 || t.Celsius <= -20
}

// GetComfortLevel returns a human-readable comfort level based on temperature
func (t Temperature) GetComfortLevel() string {
	switch {
	case t.Celsius < -20:
		return "Extreme Cold"
	case t.Celsius < 0:
		return "Freezing"
	case t.Celsius < 10:
		return "Cold"
	case t.Celsius < 20:
		return "Cool"
	case t.Celsius < 25:
		return "Comfortable"
	case t.Celsius < 30:
		return "Warm"
	case t.Celsius < 40:
		return "Hot"
	default:
		return "Extreme Heat"
	}
}

// IsStrongWind returns true if wind speed is considered strong (>= 50 kph)
func (w Wind) IsStrongWind() bool {
	return w.SpeedKph >= 50
}

// IsGale returns true if wind speed indicates gale conditions (>= 62 kph)
func (w Wind) IsGale() bool {
	return w.SpeedKph >= 62
}

// GetBeaufortScale returns the Beaufort scale number (0-12) for the wind speed
func (w Wind) GetBeaufortScale() int {
	kph := w.SpeedKph
	switch {
	case kph < 1:
		return 0 // Calm
	case kph < 6:
		return 1 // Light air
	case kph < 12:
		return 2 // Light breeze
	case kph < 20:
		return 3 // Gentle breeze
	case kph < 29:
		return 4 // Moderate breeze
	case kph < 39:
		return 5 // Fresh breeze
	case kph < 50:
		return 6 // Strong breeze
	case kph < 62:
		return 7 // Near gale
	case kph < 75:
		return 8 // Gale
	case kph < 89:
		return 9 // Strong gale
	case kph < 103:
		return 10 // Storm
	case kph < 118:
		return 11 // Violent storm
	default:
		return 12 // Hurricane
	}
}

// IsRaining returns true if there is measurable precipitation
func (p Precipitation) IsRaining() bool {
	return p.Millimeters > 0
}

// IsHeavyRain returns true if precipitation is heavy (>= 10mm)
func (p Precipitation) IsHeavyRain() bool {
	return p.Millimeters >= 10
}

// GetRainfallIntensity returns a description of rainfall intensity
func (p Precipitation) GetRainfallIntensity() string {
	mm := p.Millimeters
	switch {
	case mm == 0:
		return "No Rain"
	case mm < 2.5:
		return "Light Rain"
	case mm < 10:
		return "Moderate Rain"
	case mm < 50:
		return "Heavy Rain"
	default:
		return "Violent Rain"
	}
}

// IsHighHumidity returns true if humidity is high (>= 70%)
func (cw CurrentWeather) IsHighHumidity() bool {
	return cw.Humidity >= 70
}

// IsLowHumidity returns true if humidity is low (<= 30%)
func (cw CurrentWeather) IsLowHumidity() bool {
	return cw.Humidity <= 30
}

// IsCloudy returns true if cloud cover is significant (>= 50%)
func (cw CurrentWeather) IsCloudy() bool {
	return cw.CloudCover >= 50
}

// IsClear returns true if cloud cover is minimal (<= 20%)
func (cw CurrentWeather) IsClear() bool {
	return cw.CloudCover <= 20
}

// GetUVRisk returns a human-readable UV risk level
func (cw CurrentWeather) GetUVRisk() string {
	uv := cw.UVIndex
	switch {
	case uv < 3:
		return "Low"
	case uv < 6:
		return "Moderate"
	case uv < 8:
		return "High"
	case uv < 11:
		return "Very High"
	default:
		return "Extreme"
	}
}

// IsPoorVisibility returns true if visibility is poor (<= 2km)
func (cw CurrentWeather) IsPoorVisibility() bool {
	return cw.Visibility.Kilometers <= 2
}

// IsGoodVisibility returns true if visibility is good (>= 10km)
func (cw CurrentWeather) IsGoodVisibility() bool {
	return cw.Visibility.Kilometers >= 10
}

// GetFullDescription returns a comprehensive weather description
func (w Weather) GetFullDescription() string {
	desc := w.Current.Condition.Text
	desc += ", " + w.Current.Temperature.GetComfortLevel()

	if w.Current.Precipitation.IsRaining() {
		desc += ", " + w.Current.Precipitation.GetRainfallIntensity()
	}

	if w.Current.Wind.IsStrongWind() {
		desc += ", Strong Winds"
	}

	return desc
}

// IsExtreme returns true if weather conditions are extreme
func (w Weather) IsExtreme() bool {
	return w.Current.Temperature.IsExtreme() ||
		w.Current.Wind.IsGale() ||
		w.Current.Precipitation.IsHeavyRain() ||
		w.Current.IsPoorVisibility()
}
