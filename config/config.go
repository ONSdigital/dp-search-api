package config

import (
	"encoding/json"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config is the search API handler config
type Config struct {
	AWS                        AWS
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	ElasticSearchAPIURL        string        `envconfig:"ELASTIC_SEARCH_URL"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	OTServiceName              string        `envconfig:"OTEL_SERVICE_NAME"`
	OTExporterOTLPEndpoint     string        `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	ZebedeeURL                 string        `envconfig:"ZEBEDEE_URL"`
}

type AWS struct {
	Filename              string `envconfig:"AWS_FILENAME"`
	Profile               string `envconfig:"AWS_PROFILE"`
	Region                string `envconfig:"AWS_REGION"`
	Service               string `envconfig:"AWS_SERVICE"`
	Signer                bool   `envconfig:"AWS_SIGNER"`
	TLSInsecureSkipVerify bool   `envconfig:"AWS_TLS_INSECURE_SKIP_VERIFY"`
}

var cfg *Config

// Get configures the application and returns the Config
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                   ":23900",
		ElasticSearchAPIURL:        "http://localhost:11200",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		OTExporterOTLPEndpoint:     "localhost:4317",
		OTServiceName:              "dp-search-api",
		ZebedeeURL:                 "http://localhost:8082",
	}

	cfg.AWS = AWS{
		Filename:              "",
		Profile:               "",
		Region:                "eu-west-2",
		Service:               "es",
		Signer:                false,
		TLSInsecureSkipVerify: false,
	}

	return cfg, envconfig.Process("", cfg)
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Config) String() string {
	data, _ := json.Marshal(config)
	return string(data)
}
