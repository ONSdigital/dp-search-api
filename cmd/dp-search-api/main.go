package main

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-search-api/service"
	"os"
	"os/signal"
	"syscall"

	dphttp "github.com/ONSdigital/dp-net/http"
	esauth "github.com/ONSdigital/dp-elasticsearch/v2/awsauth"
	elastic "github.com/ONSdigital/dp-elasticsearch/v2/elasticsearch"
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
			log.Event(nil, "failed to create aws v4 signer", log.ERROR, log.Error(err))
			os.Exit(1)
		}
	}

	elasticHTTPClient := dphttp.NewClient()
	elasticSearchClient := elasticsearch.New(cfg.ElasticSearchAPIURL, dphttp.NewClient(), cfg.SignElasticsearchRequests, esSigner, cfg.AwsRegion, cfg.AwsService)
	transformer := transformer.New()
	svcList := service.NewServiceList(&service.Init{})

	// Get HealthCheck
	ctx := context.Background()
	hc, err := svcList.GetHealthCheck(cfg, BuildTime, GitCommit, Version)
	if err != nil {
		log.Event(nil, "could not instantiate healthcheck", log.FATAL, log.Error(err))
		os.Exit(1)
	}
	if err := registerCheckers(ctx, cfg, elasticHTTPClient, esSigner); err != nil {
		log.Event(nil, "failed to register checks", log.ERROR, log.Error(err))
		os.Exit(1)
	}

	if err := api.CreateAndInitialise(cfg, queryBuilder, elasticSearchClient, transformer, hc, apiErrors); err != nil {
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

		hc.Stop()

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

func registerCheckers(ctx context.Context,
	cfg *config.Config,
	elasticHTTPClient dphttp.Clienter,
	esSigner *esauth.Signer) *healthcheck.HealthCheck {

	hasErrors := false

	versionInfo, err := healthcheck.NewVersionInfo(BuildTime, GitCommit, Version)
	if err != nil {
		log.Event(ctx, "error creating version info", log.FATAL, log.Error(err))
		hasErrors = true
	}

	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)

	elasticClient := elastic.NewClientWithHTTPClientAndAwsSigner(cfg.ElasticSearchAPIURL, esSigner, cfg.SignElasticsearchRequests, elasticHTTPClient)
	if err = hc.AddCheck("Elasticsearch", elasticClient.Checker); err != nil {
		log.Event(ctx, "error creating elasticsearch health check", log.ERROR, log.Error(err))
		hasErrors = true
	}

	if hasErrors {
		log.Event(nil, "failed to successfully register API checks - exiting", log.ERROR, log.Error(err))
		os.Exit(1)
	}
	return &hc
}
