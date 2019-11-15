package api

//go:generate moq -out mocks.go -pkg api . ElasticSearcher

import (
	"context"

	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var httpServer *server.Server

//SearchQueryAPI provides an API around elasticseach
type SearchQueryAPI struct {
	Router        *mux.Router
	ElasticSearch ElasticSearcher
}

// ElasticSearcher provides client methods for the elasticsearch package
type ElasticSearcher interface {
	Search(ctx context.Context, index string, docType string, request []byte) ([]byte, error)
	MultiSearch(ctx context.Context, index string, docType string, request []byte) ([]byte, error)
}

// CreateAndInitialise initiates a new Search Query API
func CreateAndInitialise(bindAddr string, elasticSearchClient ElasticSearcher, errorChan chan error) error {

	if elasticSearchClient == nil {
		return errors.New("CreateAndInitialise called without a valid elasticsearch client")
	}

	router := mux.NewRouter()

	errSearch := SetupSearch()
	if errSearch != nil {
		return errors.Wrap(errSearch, "Failed to setup search templates")
	}

	errData := SetupData()
	if errData != nil {
		return errors.Wrap(errData, "Failed to setup data templates")
	}

	errTimeseries := SetupTimeseries()
	if errTimeseries != nil {
		return errors.Wrap(errTimeseries, "Failed to setup timeseries templates")
	}

	api := NewSearchQueryAPI(router, elasticSearchClient)

	httpServer = server.New(bindAddr, api.Router)

	// Disable this here to allow service to manage graceful shutdown of the entire app.
	httpServer.HandleOSSignals = false

	go func() {
		log.Debug("Starting search-query api...", nil)
		if err := httpServer.ListenAndServe(); err != nil {
			log.ErrorC("search-query api http server returned error", err, nil)
			errorChan <- err
		}
	}()

	return nil
}

// NewSearchQueryAPI returns a new Search Query API struct after registering the routes
func NewSearchQueryAPI(router *mux.Router, elasticSearch ElasticSearcher) *SearchQueryAPI {

	api := &SearchQueryAPI{
		Router:        router,
		ElasticSearch: elasticSearch,
	}

	router.HandleFunc("/search", SearchHandlerFunc(api.ElasticSearch)).Methods("GET")
	router.HandleFunc("/timeseries/{cdid}", TimeseriesLookupHandlerFunc(api.ElasticSearch)).Methods("GET")
	router.HandleFunc("/data", DataLookupHandlerFunc(api.ElasticSearch)).Methods("GET")
	router.HandleFunc("/healthcheck", HealthCheckHandlerCreator()).Methods("GET")
	return api
}

// Close represents the graceful shutting down of the http server
func Close(ctx context.Context) error {
	if err := httpServer.Shutdown(ctx); err != nil {
		return err
	}
	log.Info("graceful shutdown of http server complete", nil)
	return nil
}
