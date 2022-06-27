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
		esURL:            "https://vpc-sandbox-site-xczo3vpa3wd7hzu3mvjjxczq4u.eu-west-2.es.amazonaws.com",
		signRequests:     true,
		datasetURL:       "http://localhost:10400",
		ServiceAuthToken: "aevaep0chohj0phi7ephaew5chi0ohnaephaipheibahcheipoughoy4sie5tooH",
	}
}
