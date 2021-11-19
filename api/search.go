package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	dpelastic "github.com/ONSdigital/dp-elasticsearch/v2/elasticsearch"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
)

const defaultContentTypes string = "bulletin," +
	"article," +
	"article_download," +
	"compendium_landing_page," +
	"reference_tables," +
	"dataset_landing_page," +
	"static_adhoc," +
	"static_article," +
	"static_foi," +
	"static_landing_page," +
	"static_methodology," +
	"static_methodology_download," +
	"static_page," +
	"static_qmi," +
	"timeseries"

var serverErrorMessage = "internal server error"

type CreateIndexResponse struct {
	IndexName string
}

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
func SearchHandlerFunc(queryBuilder QueryBuilder, elasticSearchClient ElasticSearcher, transformer ResponseTransformer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		params := req.URL.Query()

		q := params.Get("q")
		sort := paramGet(params, "sort", "relevance")

		highlight := paramGetBool(params, "highlight", true)

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

		formattedQuery, err := queryBuilder.BuildSearchQuery(ctx, q, typesParam, sort, limit, offset)
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

		if !json.Valid([]byte(responseData)) {
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

func CreateSearchIndexHandlerFunc(dpESClient *dpelastic.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		indexName := createIndexName("ons")
		fmt.Printf("Index created: %s\n", indexName)
		indexCreated := true

		status, err := dpESClient.CreateIndex(ctx, indexName, elasticsearch.GetSearchIndexSettings())
		if err != nil {
			log.Error(ctx, "error creating index", err, log.Data{"response_status": status, "index_name": indexName})
			indexCreated = false
		}

		if status != http.StatusOK {
			log.Error(ctx, "unexpected http status when creating index", err, log.Data{"response_status": status, "index_name": indexName})
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
		createIndexResponse := CreateIndexResponse{IndexName: indexName}
		jsonResponse, _ := json.Marshal(createIndexResponse)

		_, err = w.Write(jsonResponse)
		if err != nil {
			log.Error(ctx, "writing response failed", err)
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
			return
		}
	}
}

func createIndexName(s string) string {
	now := time.Now()
	return fmt.Sprintf("%s%d", s, now.UnixMicro())
}
