package dto

type WeatherResponse struct {
	Location    string  `json:"location"`
	Temperature float64 `json:"temperature_c"`
	Condition   string  `json:"condition_text"`
}
