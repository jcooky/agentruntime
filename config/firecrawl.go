package config

import (
	"os"

	"github.com/pkg/errors"
)

type FireCrawlConfig struct {
	APIKey string `json:"api_key"`
	APIUrl string `json:"api_url"`
}

func (c *FireCrawlConfig) Validate() error {
	if c.APIKey == "" {
		return errors.New("api_key is required")
	}
	return nil
}

func NewFireCrawlConfig() *FireCrawlConfig {
	config := &FireCrawlConfig{
		APIKey: os.Getenv("FIRECRAWL_API_KEY"),
		APIUrl: os.Getenv("FIRECRAWL_API_URL"),
	}

	if config.APIUrl == "" {
		config.APIUrl = "https://api.firecrawl.dev"
	}

	return config
}
