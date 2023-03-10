package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ONSdigital/dp-elasticsearch/v3/client"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	"github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/dp-search-api/query"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
)

const defaultContentTypes string = "article," +
	"article_download," +
	"bulletin," +
	"compendium_landing_page," +
	"compendium_chapter," +
	"compendium_data," +
	"dataset," +
	"dataset_landing_page," +
	"product_page," +
	"reference_tables," +
	"release," +
	"static_adhoc," +
	"static_article," +
	"static_foi," +
	"static_landing_page," +
	"static_methodology," +
	"static_methodology_download," +
	"static_page," +
	"static_qmi," +
	"timeseries," +
	"timeseries_dataset"

var serverErrorMessage = "internal server error"

func paramGet(params url.Values, key, defaultValue string) string {
	value := params.Get(key)
	if len(value) < 1 {
		value = defaultValue
	}
	return value
}

func paramGetBool(params url.Values, key string, defaultValue bool) bool {
	value := params.Get(key)
	if len(value) < 1 {
		return defaultValue
	}
	return value == "true"
}

// CreateSearchRequest reads the parameters from the request and generates the corresponding SearchRequest
// If any validation fails, the http.Error is already handled, and nil is returned: in this case the caller may return straight away
func CreateSearchRequest(w http.ResponseWriter, req *http.Request, validator QueryParamValidator) (string, *query.SearchRequest) {
	ctx := req.Context()
	params := req.URL.Query()

	q := params.Get("q")
	sanitisedQuery := sanitiseDoubleQuotes(q)

	sortParam := paramGet(params, "sort", "relevance")
	sort, err := validator.Validate(ctx, "sort", sortParam)
	if err != nil {
		log.Warn(ctx, err.Error(), log.Data{"param": "sort", "value": sortParam})
		http.Error(w, "Invalid sort parameter", http.StatusBadRequest)
		return "", nil
	}

	highlight := paramGetBool(params, "highlight", true)

	topicsParam := paramGet(params, "topics", "")
	topics := sanitiseURLParams(topicsParam)
	log.Info(ctx, "topic extracted and sanitised from the request url params", log.Data{
		"param": "topics",
		"value": topics,
	})

	limitParam := paramGet(params, "limit", "10")
	limit, err := validator.Validate(ctx, "limit", limitParam)
	if err != nil {
		log.Warn(ctx, err.Error(), log.Data{"param": "limit", "value": limitParam})
		http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
		return "", nil
	}

	offsetParam := paramGet(params, "offset", "0")
	offset, err := validator.Validate(ctx, "offset", offsetParam)
	if err != nil {
		log.Warn(ctx, err.Error(), log.Data{"param": "offset", "value": offsetParam})
		http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
		return "", nil
	}

	typesParam := paramGet(params, "content_type", defaultContentTypes)

	return q, &query.SearchRequest{
		Term:      sanitisedQuery,
		From:      offset.(int),
		Size:      limit.(int),
		Types:     strings.Split(typesParam, ","),
		Topic:     topics,
		SortBy:    sort.(string),
		Highlight: highlight,
		Now:       time.Now().UTC().Format(time.RFC3339),
	}
}

// SearchHandlerFunc returns a http handler function handling search api requests.
func SearchHandlerFunc(validator QueryParamValidator, queryBuilder QueryBuilder, elasticSearchClient DpElasticSearcher, transformer ResponseTransformer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		params := req.URL.Query()

		q, reqParams := CreateSearchRequest(w, req, validator)
		if reqParams == nil {
			return // error already handled
		}

		var (
			resDataChan        = make(chan []byte)
			resCountChan       = make(chan []byte)
			responseSearchData []byte
			responseCountData  []byte
			count              int
			err                error
		)

		go func() {
			processCountQuery(ctx, elasticSearchClient, queryBuilder, reqParams.Term, resCountChan)
		}()

		go func() {
			processSearchQuery(ctx, elasticSearchClient, queryBuilder, reqParams, resDataChan)
		}()

		for i := 0; i < 2; i++ {
			select {
			case responseSearchData = <-resDataChan:
			case responseCountData = <-resCountChan:
			}
		}

		if !paramGetBool(params, "raw", false) {
			if responseSearchData == nil {
				log.Error(ctx, "call to elastic multisearch api failed", errors.New("nil response data"))
				http.Error(w, "call to elastic multisearch api failed", http.StatusInternalServerError)
				return
			}

			responseSearchData, err = transformer.TransformSearchResponse(ctx, responseSearchData, q, reqParams.Highlight)
			if err != nil {
				log.Error(ctx, "transformation of response data failed", err)
				http.Error(w, "failed to transform search result", http.StatusInternalServerError)
				return
			}

			if responseCountData == nil {
				log.Error(ctx, "call to elasticsearch count api failed due to", errors.New("nil response data"))
				http.Error(w, "call to elasticsearch count api failed due to", http.StatusInternalServerError)
				return
			}
			count, err = transformer.TransformCountResponse(ctx, responseCountData)
			if err != nil {
				log.Error(ctx, "transformation of response count data failed", err)
				http.Error(w, "failed to transform count result", http.StatusInternalServerError)
				return
			}
			//TODO: This needs to be refactored as it involves multiple marshal and unmarshal code. So basically the
			// transformSearchResponse function can return an interface that would satisfy both legacy search response and
			// new search response instead of bytes. So here we just have to add the count instead of unmarshalling the bytes
			// and adding the count and marshalling it again. Will be done in a separate pr very soon.
			var esSearchResponse models.SearchResponse
			if SearchRespErr := json.Unmarshal(responseSearchData, &esSearchResponse); SearchRespErr != nil {
				log.Error(ctx, "failed to unmarshal the essearchResponse data due to", SearchRespErr)
				http.Error(w, "failed to unmarshal the essearchResponse data due to", http.StatusInternalServerError)
				return
			}
			esSearchResponse.DistinctItemsCount = count
			var responseDataErr error
			responseSearchData, responseDataErr = json.Marshal(esSearchResponse)
			if responseDataErr != nil {
				log.Error(ctx, "failed to marshal the elasticsearch response data due to", responseDataErr)
				http.Error(w, "failed to transform search result", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		if _, err := w.Write(responseSearchData); err != nil {
			log.Error(ctx, "writing response failed", err)
			http.Error(w, "Failed to write http response", http.StatusInternalServerError)
			return
		}
	}
}

// LegacySearchHandlerFunc returns a http handler function handling search api requests.
// TODO: This wil be deleted once the switch over is done to ES 7.10
func LegacySearchHandlerFunc(validator QueryParamValidator, queryBuilder QueryBuilder, elasticSearchClient ElasticSearcher, transformer ResponseTransformer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		params := req.URL.Query()

		q, reqParams := CreateSearchRequest(w, req, validator)
		if reqParams == nil {
			return
		}

		formattedQuery, err := queryBuilder.BuildSearchQuery(ctx, reqParams, false)
		if err != nil {
			log.Error(ctx, "creation of search query failed", err, log.Data{"q": q, "sort": reqParams.SortBy, "limit": reqParams.Size, "offset": reqParams.From})
			http.Error(w, "Failed to create search query", http.StatusInternalServerError)
			return
		}

		responseData, err := elasticSearchClient.MultiSearch(ctx, "ons", "", formattedQuery)
		if err != nil {
			log.Error(ctx, "elasticsearch query failed", err)
			http.Error(w, "Failed to run search query", http.StatusInternalServerError)
			return
		}

		if !json.Valid(responseData) {
			log.Error(ctx, "elastic search returned invalid JSON for search query", errors.New("elastic search returned invalid JSON for search query"))
			http.Error(w, "Failed to process search query", http.StatusInternalServerError)
			return
		}

		if !paramGetBool(params, "raw", false) {
			responseData, err = transformer.TransformSearchResponse(ctx, responseData, q, reqParams.Highlight)
			if err != nil {
				log.Error(ctx, "transformation of response data failed", err)
				http.Error(w, "Failed to transform search result", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		_, err = w.Write(responseData)
		if err != nil {
			log.Error(ctx, "writing response failed", err)
			http.Error(w, "Failed to write http response", http.StatusInternalServerError)
			return
		}
	}
}

func (a SearchAPI) CreateSearchIndexHandlerFunc(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	indexName := createIndexName("ons")
	fmt.Printf("Index created: %s\n", indexName)
	indexCreated := true

	err := a.dpESClient.CreateIndex(ctx, indexName, elasticsearch.GetSearchIndexSettings())
	if err != nil {
		log.Error(ctx, "error creating index", err, log.Data{"index_name": indexName})
		indexCreated = false
	}

	if !indexCreated {
		if err != nil {
			log.Error(ctx, "creating index failed with this error", err)
		}
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	createIndexResponse := models.CreateIndexResponse{IndexName: indexName}
	jsonResponse, _ := json.Marshal(createIndexResponse)

	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Error(ctx, "writing response failed", err)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
}

func createIndexName(s string) string {
	now := time.Now()
	return fmt.Sprintf("%s%d", s, now.UnixMicro())
}

func sanitiseURLParams(str string) []string {
	if str == "" {
		return nil
	}
	return strings.Split(strings.ReplaceAll(str, " ", ""), ",")
}

func sanitiseDoubleQuotes(str string) string {
	b := strconv.Quote(str)
	return b[1 : len(b)-1]
}

// func processSearchQuery(ctx context.Context, elasticSearchClient DpElasticSearcher, queryBuilder QueryBuilder, sanitisedQuery, typesParam, sort string, topics []string, limit, offset int, responseDataChan chan []byte) {
func processSearchQuery(ctx context.Context, elasticSearchClient DpElasticSearcher, queryBuilder QueryBuilder, reqParams *query.SearchRequest, responseDataChan chan []byte) {
	formattedQuery, err := queryBuilder.BuildSearchQuery(ctx, reqParams, true)
	if err != nil {
		log.Error(ctx, "creation of search query failed", err, log.Data{"q": reqParams.Term, "sort": reqParams.SortBy, "limit": reqParams.Size, "offset": reqParams.From})
		responseDataChan <- nil
		return
	}

	var searches []client.Search

	if marshalErr := json.Unmarshal(formattedQuery, &searches); marshalErr != nil {
		log.Error(ctx, "creation of search query failed", marshalErr, log.Data{"q": reqParams.From, "sort": reqParams.From, "limit": reqParams.Size, "offset": reqParams.From})
		responseDataChan <- nil
		return
	}

	enableTotalHitsCount := true
	responseData, err := elasticSearchClient.MultiSearch(ctx, searches, &client.QueryParams{
		EnableTotalHitsCounter: &enableTotalHitsCount,
	})
	if err != nil {
		log.Error(ctx, "elasticsearch query failed", err)
		responseDataChan <- nil
		return
	}

	if !json.Valid(responseData) {
		log.Error(ctx, "elasticsearch returned invalid JSON for search query", errors.New("elasticsearch returned invalid JSON for search query"))
		responseDataChan <- nil
		return
	}
	responseDataChan <- responseData
}

func processCountQuery(ctx context.Context, elasticSearchClient DpElasticSearcher, queryBuilder QueryBuilder, sanitisedQuery string, resCountChan chan []byte) {
	countQBytes, err := queryBuilder.BuildCountQuery(ctx, sanitisedQuery)
	if err != nil {
		log.Error(ctx, "creation of count query failed", err, log.Data{"q": sanitisedQuery})
		resCountChan <- nil
		return
	}

	countRes, err := elasticSearchClient.Count(ctx, client.Count{
		Query: countQBytes,
	})
	if err != nil {
		log.Error(ctx, "call to elasticsearch count api failed", err, log.Data{"q": sanitisedQuery})
		resCountChan <- nil
		return
	}
	resCountChan <- countRes
}
