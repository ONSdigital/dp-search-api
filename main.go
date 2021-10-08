package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/service"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
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
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Error(ctx, "application unexpectedly failed", err)
		os.Exit(1)
	}

}

func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	//Create the service, providing an error channel for fatal errors
	svcErrors := make(chan error, 1)
	svcList := service.NewServiceList(&service.Init{})

	//Read config
	cfg, err := config.Get()
	if err != nil {
		log.Fatal(ctx, "error retrieving configuration", err)
		return err
	}

	log.Info(ctx, "config on startup", log.Data{"config": cfg, "build_time": BuildTime, "git-commit": GitCommit})

	//Run the service
	svc := service.New(cfg, svcList)
	if err := svc.Run(ctx, BuildTime, GitCommit, Version, svcErrors); err != nil {
		return errors.Wrap(err, "running service failed")
	}
	// Blocks until a fatal error occurs
	select {
	case err := <-svcErrors:
		log.Fatal(ctx, "search api error received", err)
	case <-signals:
		log.Info(ctx, "os signal received")
	}
	return svc.Close(ctx)
}

//	log.Info(ctx, "initialising query builder")
//	queryBuilder, err := query.NewQueryBuilder()
//	if err != nil {
//		log.Fatal(ctx, "error initialising query builder", err)
//		os.Exit(1)
//	}
//
//	var esSigner *esauth.Signer
//	if cfg.SignElasticsearchRequests {
//		esSigner, err = esauth.NewAwsSigner("", "", cfg.AwsRegion, cfg.AwsService)
//		if err != nil {
//			log.Error(ctx, "failed to create aws v4 signer", err)
//			os.Exit(1)
//		}
//	}
//
//	elasticHTTPClient := dphttp.NewClient()
//	elasticSearchClient := elasticsearch.New(cfg.ElasticSearchAPIURL, dphttp.NewClient(), cfg.SignElasticsearchRequests, esSigner, cfg.AwsRegion, cfg.AwsService)
//	transformer := transformer.New()
//
//
//	// Get HealthCheck
//
//	hc, err := registerCheckers(ctx, cfg, elasticHTTPClient, esSigner, svcList)
//	if err != nil {
//		log.Fatal(ctx, "could not register healthcheck", err)
//		os.Exit(1)
//	}
//
//	if err := api.CreateAndInitialise(cfg, queryBuilder, elasticSearchClient, transformer, hc, svcErrors); err != nil {
//		log.Fatal(ctx, "error initialising API", err)
//		os.Exit(1)
//	}
//
//	gracefulShutdown := func() {
//		log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": cfg.GracefulShutdownTimeout})
//		ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)
//
//		// stop any incoming requests before closing any outbound connections
//		if err := api.Close(ctx); err != nil {
//			log.Error(ctx, "error closing API", err)
//		}
//
//		hc.Stop()
//
//		log.Info(ctx, "shutdown complete")
//		cancel()
//	}
//
//
//
//	os.Exit(1)
//}
//
//func registerCheckers(ctx context.Context,
//	cfg *config.Config,
//	elasticHTTPClient dphttp.Clienter,
//	esSigner *esauth.Signer,
//	svcList *service.ExternalServiceList) (service.HealthChecker, error) {
//
//	hasErrors := false
//
//	hc, err := svcList.GetHealthCheck(cfg, BuildTime, GitCommit, Version)
//	if err != nil {
//		return nil, err
//	}
//
//	elasticClient := elastic.NewClientWithHTTPClientAndAwsSigner(cfg.ElasticSearchAPIURL, esSigner, cfg.SignElasticsearchRequests, elasticHTTPClient)
//	if err = hc.AddCheck("Elasticsearch", elasticClient.Checker); err != nil {
//		log.Error(ctx, "error creating elasticsearch health check", err)
//		hasErrors = true
//	}
//
//	if hasErrors {
//		return nil, err
//
//	}
//	return hc, nil
//}
