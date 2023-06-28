//go:build aws
// +build aws

package main

import (
	"context"
	"time"
)

var Name = "aws"

func getConfig(ctx context.Context) cliConfig {
	return cliConfig{
		aws: AWSConfig{
			region:                "eu-west-2",
			service:               "es",
			tlsInsecureSkipVerify: false,
		},
		zebedeeURL:       "http://localhost:10050",
		esURL:            "ES_URL_Replaceme",
		signRequests:     true,
		datasetURL:       "http://localhost:10400",
		ServiceAuthToken: "SAuthToken_Replaceme",
		PaginationLimit:  DefaultPaginationLimit,
		TestSubset:       false,
		IgnoreZebedee:    false,
		MaxRetries:       2,
		Timeout:          30 * time.Second,
	}
}
