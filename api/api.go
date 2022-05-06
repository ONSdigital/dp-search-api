package api

//go:generate moq -out mocks.go -pkg api . ElasticSearcher DpElasticSearcher QueryParamValidator QueryBuilder ReleaseQueryBuilder ResponseTransformer AuthHandler

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/dp-elasticsearch/v3/client"
	"github.com/ONSdigital/dp-search-api/query"
	"github.com/gorilla/mux"
)

var (
	update = auth.Permissions{Update: true}
)

// SearchAPI provides an API around elasticseach
type SearchAPI struct {
	Router             *mux.Router
	QueryBuilder       QueryBuilder
	dpESClient         DpElasticSearcher
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

// DpElasticSearcher provides an interface for the dp-elasticsearch functionality
type DpElasticSearcher interface {
	CreateIndex(ctx context.Context, indexName string, indexSettings []byte) error
	MultiSearch(ctx context.Context, searches []client.Search) ([]byte, error)
}

// QueryParamValidator provides an interface to validate api query parameters (used for /search/releases)
type QueryParamValidator interface {
	Validate(ctx context.Context, name, value string) (interface{}, error)
}

// QueryBuilder provides methods for the search package
type QueryBuilder interface {
	BuildSearchQuery(ctx context.Context, q, contentTypes, sort string, topics []string, limit, offset int) ([]byte, error)
}

// ReleaseQueryBuilder provides an interface to build a search query for the Release content type
type ReleaseQueryBuilder interface {
	BuildSearchQuery(ctx context.Context, request query.ReleaseSearchRequest) ([]byte, error)
}

// ResponseTransformer provides methods for the transform package
type ResponseTransformer interface {
	TransformSearchResponse(ctx context.Context, responseData []byte, query string, highlight bool) ([]byte, error)
}

type ReleaseResponseTransformer interface {
	TransformSearchResponse(ctx context.Context, responseData []byte, req query.ReleaseSearchRequest, highlight bool) ([]byte, error)
}

// NewSearchAPI returns a new Search API struct after registering the routes
func NewSearchAPI(router *mux.Router, dpESClient DpElasticSearcher, deprecatedESClient ElasticSearcher, queryBuilder QueryBuilder, transformer ResponseTransformer, permissions AuthHandler, elasticVersion710 bool) (*SearchAPI, error) {
	api := &SearchAPI{
		Router:             router,
		QueryBuilder:       queryBuilder,
		dpESClient:         dpESClient,
		deprecatedESClient: deprecatedESClient,
		Transformer:        transformer,
		permissions:        permissions,
	}

	if elasticVersion710 {
		router.HandleFunc("/search", SearchHandlerFunc(queryBuilder, api.dpESClient, api.Transformer)).Methods("GET")
	} else {
		router.HandleFunc("/search", LegacySearchHandlerFunc(queryBuilder, api.deprecatedESClient, api.Transformer)).Methods("GET")
	}
	createSearchIndexHandler := permissions.Require(update, api.CreateSearchIndexHandlerFunc)
	router.HandleFunc("/search", createSearchIndexHandler).Methods("POST")
	return api, nil
}

func (a *SearchAPI) AddSearchReleaseAPI(validator QueryParamValidator, builder ReleaseQueryBuilder, searcher ElasticSearcher, transformer ReleaseResponseTransformer) *SearchAPI {
	a.Router.HandleFunc("/search/releases", SearchReleasesHandlerFunc(validator, builder, searcher, transformer)).Methods("GET")

	return a
}
