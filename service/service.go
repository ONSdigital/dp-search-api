package service

import (
	"context"

	"github.com/ONSdigital/dp-search-api/api"
	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var httpServer *server.Server

// CreateAndInitialise initiates a new Search API
func CreateAndInitialise(cfg *config.Config, queryBuilder api.QueryBuilder, elasticSearchClient api.ElasticSearcher, transformer api.ResponseTransformer, hc HealthChecker, errorChan chan error, svcList *ExternalServiceList) error {

	if elasticSearchClient == nil {
		return errors.New("CreateAndInitialise called without a valid elasticsearch client")
	}

	if queryBuilder == nil {
		return errors.New("CreateAndInitialise called without a valid query builder")
	}
	router := mux.NewRouter()

	errData := api.SetupData()
	if errData != nil {
		return errors.Wrap(errData, "Failed to setup data templates")
	}

	errTimeseries := api.SetupTimeseries()
	if errTimeseries != nil {
		return errors.Wrap(errTimeseries, "Failed to setup timeseries templates")
	}

	ctx := context.Background()
	router.StrictSlash(true).Path("/health").HandlerFunc(hc.Handler)
	hc.Start(ctx)

	permissions := svcList.GetAuthorisationHandlers(cfg)

	api := api.NewSearchAPI(router, elasticSearchClient, queryBuilder, transformer, permissions)

	httpServer = server.New(cfg.BindAddr, api.Router)

	// Disable this here to allow service to manage graceful shutdown of the entire app.
	httpServer.HandleOSSignals = false

	go func() {
		ctx := context.Background()
		log.Info(ctx, "search api starting")
		if err := httpServer.ListenAndServe(); err != nil {
			log.Error(ctx, "search api http server returned error", err)
			errorChan <- err
		}
	}()

	return nil
}
