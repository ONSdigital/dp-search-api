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

// SearchHandlerFunc returns a http handler function handling search api requests.
func SearchHandlerFunc(queryBuilder QueryBuilder, elasticSearchClient DpElasticSearcher, transformer ResponseTransformer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		params := req.URL.Query()

		q := params.Get("q")
		sanitisedQuery := sanitiseDoubleQuotes(q)
		sort := paramGet(params, "sort", "relevance")

		highlight := paramGetBool(params, "highlight", true)
		topics := paramGet(params, "topics", "")
		topicSlice := sanitiseURLParams(topics)
		log.Info(ctx, "topic extracted and sanitised from the request url params", log.Data{
			"param": "topics",
			"value": topicSlice,
		})
		limitParam := paramGet(params, "limit", "10")
		limit, err := strconv.Atoi(limitParam)
		if err != nil {
			log.Warn(ctx, "numeric search parameter provided with non numeric characters", log.Data{
				"param": "limit",
				"value": limitParam,
			})
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
		if limit < 0 {
			log.Warn(ctx, "numeric search parameter provided with negative value", log.Data{
				"param": "limit",
				"value": limitParam,
			})
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}

		offsetParam := paramGet(params, "offset", "0")
		offset, err := strconv.Atoi(offsetParam)
		if err != nil {
			log.Warn(ctx, "numeric search parameter provided with non numeric characters", log.Data{
				"param": "from",
				"value": offsetParam,
			})
			http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
			return
		}
		if offset < 0 {
			log.Warn(ctx, "numeric search parameter provided with negative value", log.Data{
				"param": "from",
				"value": offsetParam,
			})
			http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
			return
		}

		typesParam := paramGet(params, "content_type", defaultContentTypes)

		var resDataChan = make(chan []byte)
		var resCountChan = make(chan []byte)
		var responseSearchData []byte
		go func() {
			defer close(resCountChan)
			processCountQuery(ctx, elasticSearchClient, queryBuilder, sanitisedQuery, resCountChan)
		}()
		go func() {
			defer close(resDataChan)
			processSearchQuery(ctx, elasticSearchClient, queryBuilder, sanitisedQuery, typesParam, sort, topicSlice, limit, offset, resDataChan)
		}()

		for responseData := range resDataChan {
			if responseData == nil {
				log.Error(ctx, "call to elastic multisearch api failed", errors.New("nil response data"))
				http.Error(w, "call to elastic multisearch api failed", http.StatusInternalServerError)
				return
			}
			if !paramGetBool(params, "raw", false) {
				responseData, err = transformer.TransformSearchResponse(ctx, responseData, q, highlight)
				if err != nil {
					log.Error(ctx, "transformation of response data failed", err)
					http.Error(w, "failed to transform search result", http.StatusInternalServerError)
					return
				}
			}
			responseSearchData = responseData
		}

		for responseCountData := range resCountChan {
			if responseCountData == nil {
				log.Error(ctx, "call to elastic count api failed due to", errors.New("nil response data"))
				http.Error(w, "call to elastic count api failed due to", http.StatusInternalServerError)
				return
			}
			if !paramGetBool(params, "raw", false) {
				count, CountAPIErr := transformer.TransformCountResponse(ctx, responseCountData)
				if CountAPIErr != nil {
					log.Error(ctx, "transformation of response count data failed", CountAPIErr)
					http.Error(w, "failed to transform count result", http.StatusInternalServerError)
					return
				}
				var esSearchResponse models.SearchResponse
				if SearchRespErr := json.Unmarshal(responseSearchData, &esSearchResponse); SearchRespErr != nil {
					log.Error(ctx, "failed to un marshal the essearchResponse data due to", SearchRespErr)
					http.Error(w, "failed to un marshal the essearchResponse data due to", http.StatusInternalServerError)
					return
				}
				esSearchResponse.DistinctItemsCount = count
				var responseDataErr error
				responseSearchData, responseDataErr = json.Marshal(esSearchResponse)
				if responseDataErr != nil {
					log.Error(ctx, "failed to marshal the essearchResponse data due to", responseDataErr)
					http.Error(w, "failed to transform search result", http.StatusInternalServerError)
					return
				}
			}
		}

		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		_, err = w.Write(responseSearchData)
		if err != nil {
			log.Error(ctx, "writing response failed", err)
			http.Error(w, "Failed to write http response", http.StatusInternalServerError)
			return
		}
	}
}

// LegacySearchHandlerFunc returns a http handler function handling search api requests.
// TODO: This wil be deleted once the switch over is done to ES 7.10
func LegacySearchHandlerFunc(queryBuilder QueryBuilder, elasticSearchClient ElasticSearcher, transformer ResponseTransformer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		params := req.URL.Query()

		q := params.Get("q")
		sort := paramGet(params, "sort", "relevance")

		highlight := paramGetBool(params, "highlight", true)
		topics := paramGet(params, "topics", "")
		topicSlice := sanitiseURLParams(topics)
		log.Info(ctx, "topic extracted and sanitised from the request url params", log.Data{
			"param": "topics",
			"value": topicSlice,
		})
		limitParam := paramGet(params, "limit", "10")
		limit, err := strconv.Atoi(limitParam)
		if err != nil {
			log.Warn(ctx, "numeric search parameter provided with non numeric characters", log.Data{
				"param": "limit",
				"value": limitParam,
			})
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
		if limit < 0 {
			log.Warn(ctx, "numeric search parameter provided with negative value", log.Data{
				"param": "limit",
				"value": limitParam,
			})
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}

		offsetParam := paramGet(params, "offset", "0")
		offset, err := strconv.Atoi(offsetParam)
		if err != nil {
			log.Warn(ctx, "numeric search parameter provided with non numeric characters", log.Data{
				"param": "from",
				"value": offsetParam,
			})
			http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
			return
		}
		if offset < 0 {
			log.Warn(ctx, "numeric search parameter provided with negative value", log.Data{
				"param": "from",
				"value": offsetParam,
			})
			http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
			return
		}

		typesParam := paramGet(params, "content_type", defaultContentTypes)

		formattedQuery, err := queryBuilder.BuildSearchQuery(ctx, q, typesParam, sort, topicSlice, limit, offset, false)
		if err != nil {
			log.Error(ctx, "creation of search query failed", err, log.Data{"q": q, "sort": sort, "limit": limit, "offset": offset})
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
			responseData, err = transformer.TransformSearchResponse(ctx, responseData, q, highlight)
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

func processSearchQuery(ctx context.Context, elasticSearchClient DpElasticSearcher, queryBuilder QueryBuilder, sanitisedQuery, typesParam, sort string, topicSlice []string, limit, offset int, responseDataChan chan []byte) {
	formattedQuery, err := queryBuilder.BuildSearchQuery(ctx, sanitisedQuery, typesParam, sort, topicSlice, limit, offset, true)
	if err != nil {
		log.Error(ctx, "creation of search query failed", err, log.Data{"q": sanitisedQuery, "sort": sort, "limit": limit, "offset": offset})
		responseDataChan <- nil
	}

	var searches []client.Search

	err = json.Unmarshal(formattedQuery, &searches)
	if err != nil {
		log.Error(ctx, "creation of search query failed", err, log.Data{"q": sanitisedQuery, "sort": sort, "limit": limit, "offset": offset})
		responseDataChan <- nil
	}

	enableTotalHitsCount := true
	responseData, err := elasticSearchClient.MultiSearch(ctx, searches, &client.QueryParams{
		EnableTotalHitsCounter: &enableTotalHitsCount,
	})
	if err != nil {
		log.Error(ctx, "elasticsearch query failed", err)
		responseDataChan <- nil
	}

	if !json.Valid(responseData) {
		log.Error(ctx, "elastic search returned invalid JSON for search query", errors.New("elastic search returned invalid JSON for search query"))
		responseDataChan <- nil
	}
	responseDataChan <- responseData
	return
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
		log.Error(ctx, "call to elastic count api failed", err, log.Data{"q": sanitisedQuery})
		resCountChan <- nil
		return
	}
	resCountChan <- countRes
}
