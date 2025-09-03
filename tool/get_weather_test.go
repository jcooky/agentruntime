package tool_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/habiliai/agentruntime/tool"
	"github.com/mitchellh/mapstructure"
)

func (s *TestSuite) TestGetWeather() {
	if testing.Short() {
		s.T().Skip("Skipping test in short mode")
	}

	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		s.T().Skip("OPENWEATHER_API_KEY is not set")
	}

	getWeatherTool := s.toolManager.GetTool("get_weather")
	s.Require().NotNil(getWeatherTool)

	res, err := getWeatherTool.RunRaw(s, map[string]any{
		"location": "Seoul",
		"date":     "2023-10-01",
	})
	s.Require().NoError(err)

	var weatherSummary tool.GetWeatherResponse
	s.Require().NoError(mapstructure.Decode(res, &weatherSummary))

	s.T().Logf("contents: %v", weatherSummary)

	// 3. Output results
	fmt.Printf("🌡️ Max Temperature: %.2f°C\n", weatherSummary.Temperature.Max)
	fmt.Printf("🌡️ Min Temperature: %.2f°C\n", weatherSummary.Temperature.Min)
	fmt.Printf("🌡️ Afternoon Temperature (12:00): %.2f°C\n", weatherSummary.Temperature.Afternoon)
	fmt.Printf("🌡️ Morning Temperature (06:00): %.2f°C\n", weatherSummary.Temperature.Morning)
	fmt.Printf("🌡️ Evening Temperature (18:00): %.2f°C\n", weatherSummary.Temperature.Evening)
	fmt.Printf("🌡️ Night Temperature (00:00): %.2f°C\n", weatherSummary.Temperature.Night)
	fmt.Printf("💧 Afternoon Humidity: %.2f\n", weatherSummary.Humidity.Afternoon)
	fmt.Printf("🌬️ Max Wind Speed: %.2fm/s (Direction: %.2f°)\n", weatherSummary.Wind.Max.Speed, weatherSummary.Wind.Max.Direction)
}
