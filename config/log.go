package config

type LogConfig struct {
	LogLevel   string `json:"logLevel"`
	LogHandler string `json:"logHandler"`
}

func NewLogConfig() *LogConfig {
	return &LogConfig{
		LogLevel:   "debug",
		LogHandler: "default",
	}
}
