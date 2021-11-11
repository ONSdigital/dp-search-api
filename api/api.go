package api

//go:generate moq -out mocks.go -pkg api . ElasticSearcher QueryBuilder ResponseTransformer

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

var (
	httpServer *server.Server
	update     = auth.Permissions{Update: true}
)

//SearchAPI provides an API around elasticseach
type SearchAPI struct {
	Router        *mux.Router
	QueryBuilder  QueryBuilder
	ElasticSearch ElasticSearcher
	Transformer   ResponseTransformer
	permissions   AuthHandler
}

// AuthHandler provides authorisation checks on requests
type AuthHandler interface {
	Require(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc
}

// ElasticSearcher provides client methods for the elasticsearch package
type ElasticSearcher interface {
	Search(ctx context.Context, index string, docType string, request []byte) ([]byte, error)
	MultiSearch(ctx context.Context, index string, docType string, request []byte) ([]byte, error)
	GetStatus(ctx context.Context) ([]byte, error)
	CreateNewEmptyIndex(ctx context.Context, indexName string) (bool, error)
}

// QueryBuilder provides methods for the search package
type QueryBuilder interface {
	BuildSearchQuery(ctx context.Context, q, contentTypes, sort string, limit, offset int) ([]byte, error)
}

// ResponseTransformer provides methods for the transform package
type ResponseTransformer interface {
	TransformSearchResponse(ctx context.Context, responseData []byte, query string, highlight bool) ([]byte, error)
}

// NewSearchAPI returns a new Search API struct after registering the routes
func NewSearchAPI(router *mux.Router, elasticSearch ElasticSearcher, queryBuilder QueryBuilder, transformer ResponseTransformer, permissions AuthHandler) *SearchAPI {

	api := &SearchAPI{
		Router:        router,
		QueryBuilder:  queryBuilder,
		ElasticSearch: elasticSearch,
		Transformer:   transformer,
		permissions:   permissions,
	}

	router.HandleFunc("/search", SearchHandlerFunc(queryBuilder, api.ElasticSearch, api.Transformer)).Methods("GET")
	router.HandleFunc("/timeseries/{cdid}", TimeseriesLookupHandlerFunc(api.ElasticSearch)).Methods("GET")
	router.HandleFunc("/data", DataLookupHandlerFunc(api.ElasticSearch)).Methods("GET")
	createSearchIndexHandler := permissions.Require(update, CreateSearchIndexHandlerFunc(api.ElasticSearch))
	router.HandleFunc("/search", createSearchIndexHandler).Methods("POST")

	return api
}

// Close represents the graceful shutting down of the http server
func Close(ctx context.Context) error {
	if err := httpServer.Shutdown(ctx); err != nil {
		return err
	}
	log.Info(ctx, "graceful shutdown of http server complete")
	return nil
}
