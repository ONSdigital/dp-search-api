//go:build !aws
// +build !aws

package main

import (
	"context"
	"log"
	"os"

	"github.com/ONSdigital/dp-search-api/config"
)

var Name = "development"

func getConfig(ctx context.Context) cliConfig {
	cfg, err := config.Get()
	if err != nil {
		log.Fatal(ctx, "error retrieving config", err)
		os.Exit(1)
	}

	return cliConfig{
		aws: AWSConfig{
			filename:              cfg.AWS.Filename,
			profile:               cfg.AWS.Profile,
			region:                cfg.AWS.Region,
			service:               "es",
			tlsInsecureSkipVerify: true,
		},
		esURL:            "http://localhost:11200",
		signRequests:     false,
		zebedeeURL:       "http://localhost:8082",
		datasetURL:       "http://localhost:22000",
		ServiceAuthToken: "",
		TestSubset:       false,
		IgnoreZebedee:    false,
	}
}
