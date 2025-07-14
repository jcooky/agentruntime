package tool_test

import (
	"fmt"
	"os"

	"github.com/habiliai/agentruntime/tool"
	"github.com/mitchellh/mapstructure"
)

func (s *TestSuite) TestGetWeather() {
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

	// 3. 출력
	fmt.Printf("🌡️ 최고 기온: %.2f°C\n", weatherSummary.Temperature.Max)
	fmt.Printf("🌡️ 최저 기온: %.2f°C\n", weatherSummary.Temperature.Min)
	fmt.Printf("🌡️ 오후 기온(12:00): %.2f°C\n", weatherSummary.Temperature.Afternoon)
	fmt.Printf("🌡️ 아침 기온(06:00): %.2f°C\n", weatherSummary.Temperature.Morning)
	fmt.Printf("🌡️ 저녁 기온(18:00): %.2f°C\n", weatherSummary.Temperature.Evening)
	fmt.Printf("🌡️ 밤 기온(00:00): %.2f°C\n", weatherSummary.Temperature.Night)
	fmt.Printf("💧 오후 습도: %.2f\n", weatherSummary.Humidity.Afternoon)
	fmt.Printf("🌬️ 최대 풍속: %.2fm/s (방향: %.2f°)\n", weatherSummary.Wind.Max.Speed, weatherSummary.Wind.Max.Direction)
}
