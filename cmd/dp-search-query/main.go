package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/dp-search-query/api"
	"github.com/ONSdigital/dp-search-query/config"
	"github.com/ONSdigital/dp-search-query/elasticsearch"
	"github.com/ONSdigital/log.go/log"
)

func main() {
	log.Namespace = "dp-search-query"

	cfg, err := config.Get()
	if err != nil {
		log.Event(nil, "error retreiving config", log.Error(err), log.FATAL)
		os.Exit(1)
	}

	// sensitive fields are omitted from config.String().
	log.Event(nil, "config on startup", log.Data{"config": cfg})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	apiErrors := make(chan error, 1)

	elasticSearchClient := elasticsearch.New(cfg.ElasticSearchAPIURL, rchttp.NewClient())
	if err := api.CreateAndInitialise(cfg.BindAddr, elasticSearchClient, apiErrors); err != nil {
		log.Event(nil, "error initialising API", log.Error(err), nil)
		os.Exit(1)
	}

	gracefulShutdown := func() {
		log.Event(nil, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": cfg.GracefulShutdownTimeout})
		ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)

		// stop any incoming requests before closing any outbound connections
		if err := api.Close(ctx); err != nil {
			log.Event(ctx, "error closing API", log.Error(err))
		}

		log.Event(ctx, "shutdown complete")
		cancel()
	}

	// blocks until a fatal error occurs
	select {
	case err := <-apiErrors:
		log.Event(nil, "search-query api error received", log.Error(err), log.FATAL)
	case <-signals:
		log.Event(nil, "os signal received")
		gracefulShutdown()
	}

	os.Exit(1)
}
