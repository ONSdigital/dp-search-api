package api

//go:generate moq -out mocks.go -pkg api . ElasticSearcher DpElasticSearcher QueryParamValidator QueryBuilder ReleaseQueryBuilder ResponseTransformer AuthHandler

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/dp-elasticsearch/v3/client"
	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/nlp"
	"github.com/ONSdigital/dp-search-api/query"
	"github.com/gorilla/mux"
)

var (
	update = auth.Permissions{Update: true}
)

// SearchAPI provides an API around elasticseach
type SearchAPI struct {
	Router      *mux.Router
	dpESClient  DpElasticSearcher
	permissions AuthHandler
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
	MultiSearch(ctx context.Context, searches []client.Search, params *client.QueryParams) ([]byte, error)
	Count(ctx context.Context, count client.Count) ([]byte, error)
}

// QueryParamValidator provides an interface to validate api query parameters (used for /search/releases)
type QueryParamValidator interface {
	Validate(ctx context.Context, name, value string) (interface{}, error)
}

// QueryBuilder provides methods for the search package
type QueryBuilder interface {
	AddNlpCategorySearch(nlpCriteria *query.NlpCriteria, category string, subCategory string, categoryWeighting float32) *query.NlpCriteria
	AddNlpSubdivisionSearch(nlpCriteria *query.NlpCriteria, subdivisionWords string) *query.NlpCriteria
	BuildSearchQuery(ctx context.Context, req *query.SearchRequest, esVersion710 bool) ([]byte, error)
	BuildCountQuery(ctx context.Context, req *query.CountRequest) ([]byte, error)
}

// ReleaseQueryBuilder provides an interface to build a search query for the Release content type
type ReleaseQueryBuilder interface {
	BuildSearchQuery(ctx context.Context, request interface{}) ([]client.Search, error)
}

// ResponseTransformer provides methods for the transform package
type ResponseTransformer interface {
	TransformSearchResponse(ctx context.Context, responseData []byte, query string, highlight bool) ([]byte, error)
	TransformCountResponse(ctx context.Context, responseData []byte) (int, error)
}

type ReleaseResponseTransformer interface {
	TransformSearchResponse(ctx context.Context, responseData []byte, req query.ReleaseSearchRequest, highlight bool) ([]byte, error)
}

// NewSearchAPI returns a new Search API struct after registering the routes
func NewSearchAPI(router *mux.Router, dpESClient DpElasticSearcher, permissions AuthHandler) *SearchAPI {
	return &SearchAPI{
		Router:      router,
		dpESClient:  dpESClient,
		permissions: permissions,
	}
}

// RegisterGetSearch registers the handler for GET /search endpoint
// with the provided validator and query builder
// as well as the API's elasticsearch client and response transformer
func (a *SearchAPI) RegisterGetSearch(validator QueryParamValidator, builder QueryBuilder, settingsNLP config.NLP, cli *nlp.Client, transformer ResponseTransformer) *SearchAPI {
	a.Router.HandleFunc(
		"/search",
		SearchHandlerFunc(
			validator,
			builder,
			settingsNLP,
			cli,
			a.dpESClient,
			transformer,
		),
	).Methods(http.MethodGet)
	return a
}

// RegisterGetSearch registers the handler for GET /nlp/search endpoint
// with the provided validator and query builder
// as well as the API's elasticsearch client and response transformer
func (a *SearchAPI) RegisterGetNLPSearch(nlpClient *nlp.Client) *SearchAPI {
	a.Router.HandleFunc(
		"/nlp/search",
		NLPSearchHandlerFunc(nlpClient),
	).Methods(http.MethodGet)
	return a
}

// RegisterPostSearch registers the handler for POST /search endpoint
// enforcing required update permissions
func (a *SearchAPI) RegisterPostSearch() *SearchAPI {
	a.Router.HandleFunc(
		"/search",
		a.permissions.Require(
			update,
			a.CreateSearchIndexHandlerFunc,
		),
	).Methods(http.MethodPost)
	return a
}

// RegisterGetSearchRelease registers the handler for GET /search/releases endpoint
// with the provided validator, query builder, searcher and validator
func (a *SearchAPI) RegisterGetSearchReleases(validator QueryParamValidator, builder ReleaseQueryBuilder, transformer ReleaseResponseTransformer) *SearchAPI {
	a.Router.HandleFunc(
		"/search/releases",
		SearchReleasesHandlerFunc(
			validator,
			builder,
			a.dpESClient,
			transformer,
		),
	).Methods(http.MethodGet)
	return a
}
