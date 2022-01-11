package api

//go:generate moq -out mocks.go -pkg api . ElasticSearcher QueryBuilder ResponseTransformer AuthHandler

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-authorisation/auth"
	dpelastic "github.com/ONSdigital/dp-elasticsearch/v3/elasticsearch"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var (
	update = auth.Permissions{Update: true}
)

// SearchAPI provides an API around elasticseach
type SearchAPI struct {
	Router             *mux.Router
	QueryBuilder       QueryBuilder
	dpESClient         *dpelastic.Client
	deprecatedESClient ElasticSearcher
	Transformer        ResponseTransformer
	permissions        AuthHandler
}

// AuthHandler provides authorisation checks on requests
type AuthHandler interface {
	Require(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc
}

// ElasticSearcher provides client methods for the elasticsearch package - now deprecated, due to be replaced
// with the methods in dp-elasticsearch
type ElasticSearcher interface {
	Search(ctx context.Context, index string, docType string, request []byte) ([]byte, error)
	MultiSearch(ctx context.Context, index string, docType string, request []byte) ([]byte, error)
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
func NewSearchAPI(router *mux.Router, dpESClient *dpelastic.Client, deprecatedESClient ElasticSearcher, queryBuilder QueryBuilder, transformer ResponseTransformer, permissions AuthHandler) (*SearchAPI, error) {
	errData := SetupData()
	if errData != nil {
		return nil, errors.Wrap(errData, "Failed to setup data templates")
	}

	errTimeseries := SetupTimeseries()
	if errTimeseries != nil {
		return nil, errors.Wrap(errTimeseries, "Failed to setup timeseries templates")
	}

	api := &SearchAPI{
		Router:             router,
		QueryBuilder:       queryBuilder,
		dpESClient:         dpESClient,
		deprecatedESClient: deprecatedESClient,
		Transformer:        transformer,
		permissions:        permissions,
	}

	router.HandleFunc("/search", SearchHandlerFunc(queryBuilder, api.deprecatedESClient, api.Transformer)).Methods("GET")
	router.HandleFunc("/timeseries/{cdid}", TimeseriesLookupHandlerFunc(api.deprecatedESClient)).Methods("GET")
	router.HandleFunc("/data", DataLookupHandlerFunc(api.deprecatedESClient)).Methods("GET")

	createSearchIndexHandler := permissions.Require(update, CreateSearchIndexHandlerFunc(api.dpESClient))
	router.HandleFunc("/search", createSearchIndexHandler).Methods("POST")

	return api, nil
}
