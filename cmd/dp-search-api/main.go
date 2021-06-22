package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	dphttp "github.com/ONSdigital/dp-net/http"
	esauth "github.com/ONSdigital/dp-elasticsearch/v2/awsauth"
	"github.com/ONSdigital/dp-search-api/api"
	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	"github.com/ONSdigital/dp-search-api/query"
	"github.com/ONSdigital/dp-search-api/transformer"
	"github.com/ONSdigital/log.go/log"
)

const serviceName = "dp-search-api"

var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
)

func main() {
	log.Namespace = serviceName

	cfg, err := config.Get()
	if err != nil {
		log.Event(nil, "error retrieving config", log.Error(err), log.FATAL)
		os.Exit(1)
	}

	// sensitive fields are omitted from config.String().
	log.Event(nil, "config on startup", log.Data{"config": cfg}, log.INFO)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	apiErrors := make(chan error, 1)

	log.Event(nil, "initialising query builder", log.INFO)
	queryBuilder, err := query.NewQueryBuilder()
	if err != nil {
		log.Event(nil, "error initialising query builder", log.Error(err), log.FATAL)
		os.Exit(1)
	}

	var esSigner *esauth.Signer
	if cfg.SignElasticsearchRequests {
		esSigner, err = esauth.NewAwsSigner("", "", cfg.AwsRegion, cfg.AwsService)
		if err != nil {
			log.Event(ctx, "failed to create aws v4 signer", log.ERROR, log.Error(err))
			os.Exit(1)
		}
	}

	elasticSearchClient := elasticsearch.New(cfg.ElasticSearchAPIURL, dphttp.NewClient(), cfg.SignElasticsearchRequests, esSigner, cfg.AwsRegion, cfg.AwsService)
	transformer := transformer.New()

	if err := api.CreateAndInitialise(cfg.BindAddr, queryBuilder, elasticSearchClient, transformer, apiErrors); err != nil {
		log.Event(nil, "error initialising API", log.Error(err), log.FATAL)
		os.Exit(1)
	}

	gracefulShutdown := func() {
		log.Event(nil, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": cfg.GracefulShutdownTimeout}, log.INFO)
		ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)

		// stop any incoming requests before closing any outbound connections
		if err := api.Close(ctx); err != nil {
			log.Event(ctx, "error closing API", log.Error(err), log.ERROR)
		}

		log.Event(ctx, "shutdown complete", log.INFO)
		cancel()
	}

	// blocks until a fatal error occurs
	select {
	case err := <-apiErrors:
		log.Event(nil, "search api error received", log.Error(err), log.FATAL)
	case <-signals:
		log.Event(nil, "os signal received", log.INFO)
		gracefulShutdown()
	}

	os.Exit(1)
}
