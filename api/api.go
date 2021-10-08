package api

//go:generate moq -out mocks.go -pkg api . ElasticSearcher QueryBuilder ResponseTransformer

import (
	"context"
	"github.com/ONSdigital/dp-search-api/config"

	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var httpServer *server.Server

//SearchAPI provides an API around elasticseach
type SearchAPI struct {
	Router        *mux.Router
	QueryBuilder  QueryBuilder
	ElasticSearch ElasticSearcher
	Transformer   ResponseTransformer
}

// ElasticSearcher provides client methods for the elasticsearch package
type ElasticSearcher interface {
	Search(ctx context.Context, index string, docType string, request []byte) ([]byte, error)
	MultiSearch(ctx context.Context, index string, docType string, request []byte) ([]byte, error)
	GetStatus(ctx context.Context) ([]byte, error)
}

// QueryBuilder provides methods for the search package
type QueryBuilder interface {
	BuildSearchQuery(ctx context.Context, q, contentTypes, sort string, limit, offset int) ([]byte, error)
}

// ResponseTransformer provides methods for the transform package
type ResponseTransformer interface {
	TransformSearchResponse(ctx context.Context, responseData []byte, query string, highlight bool) ([]byte, error)
}

// CreateAndInitialise initiates a new Search API
func CreateAndInitialise(cfg *config.Configuration, queryBuilder QueryBuilder, elasticSearchClient ElasticSearcher, transformer ResponseTransformer, errorChan chan error) error {

	if elasticSearchClient == nil {
		return errors.New("CreateAndInitialise called without a valid elasticsearch client")
	}

	if queryBuilder == nil {
		return errors.New("CreateAndInitialise called without a valid query builder")
	}
	router := mux.NewRouter()

	errData := SetupData()
	if errData != nil {
		return errors.Wrap(errData, "Failed to setup data templates")
	}

	errTimeseries := SetupTimeseries()
	if errTimeseries != nil {
		return errors.Wrap(errTimeseries, "Failed to setup timeseries templates")
	}

	//TODO... delete...
	//ctx := context.Background()
	//router.StrictSlash(true).Path("/health").HandlerFunc(hc.Handler)
	//hc.Start(ctx)

	api := NewSearchAPI(router, elasticSearchClient, queryBuilder, transformer)

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

// NewSearchAPI returns a new Search API struct after registering the routes
func NewSearchAPI(router *mux.Router, elasticSearch ElasticSearcher, queryBuilder QueryBuilder, transformer ResponseTransformer) *SearchAPI {

	api := &SearchAPI{
		Router:        router,
		QueryBuilder:  queryBuilder,
		ElasticSearch: elasticSearch,
		Transformer:   transformer,
	}

	router.HandleFunc("/search", SearchHandlerFunc(queryBuilder, api.ElasticSearch, api.Transformer)).Methods("GET")
	router.HandleFunc("/timeseries/{cdid}", TimeseriesLookupHandlerFunc(api.ElasticSearch)).Methods("GET")
	router.HandleFunc("/data", DataLookupHandlerFunc(api.ElasticSearch)).Methods("GET")
	return api
}

// Close represents the graceful shutting down of the http server
//TODO: delete...
//func Close(ctx context.Context) error {
//	if err := httpServer.Shutdown(ctx); err != nil {
//		return err
//	}
//	log.Info(ctx, "graceful shutdown of http server complete")
//	return nil
//}
