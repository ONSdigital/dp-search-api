//go:build aws
// +build aws

package main

import "context"

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
		TestSubset:       false,
		IgnoreZebedee:    false,
	}
}
