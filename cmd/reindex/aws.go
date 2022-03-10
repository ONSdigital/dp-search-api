//go:build aws
// +build aws

package main

import "context"

var Name = "aws"

func getConfig(ctx context.Context) cliConfig {
	return cliConfig{
		zebedeeURL:   "http://localhost:10050",
		esURL:        "ES_URL_Replaceme",
		signRequests: true,
	}
}
