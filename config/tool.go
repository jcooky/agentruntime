package config

type ToolConfig struct {
	OpenWeatherApiKey string `json:"openWeatherApiKey" jsonschema:"description=OpenWeather Map API key"`
	SerpApiKey        string `json:"serpApiKey" jsonschema:"description=SerpAPI API key"`
}
