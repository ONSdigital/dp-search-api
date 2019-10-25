package config

import (
	"encoding/json"

	"github.com/kelseyhightower/envconfig"
)

// Config is the search query handler config
type Config struct {
	BindAddr            string `envconfig:"BIND_ADDR"`
	ElasticSearchAPIURL string `envconfig:"ELASTIC_SEARCH_URL"`
}

var cfg *Config

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:            ":23900",
		ElasticSearchAPIURL: "http://localhost:9200",
	}

	return cfg, envconfig.Process("", cfg)
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Config) String() string {
	json, _ := json.Marshal(config)
	return string(json)
}
