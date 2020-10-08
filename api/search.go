package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ONSdigital/log.go/log"
)

const defaultContentTypes string = "bulletin," +
	"article," +
	"article_download," +
	"compendium_landing_page," +
	"compendium_chapter," +
	"compendium_data," +
	"timeseries_dataset," +
	"dataset," +
	"dataset_landing_page," +
	"reference_tables," +
	"static_adhoc," +
	"static_article," +
	"static_foi," +
	"static_landing_page," +
	"static_methodology," +
	"static_methodology_download," +
	"static_page," +
	"static_qmi," +
	"timeseries"

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

		limitParam := paramGet(params, "limit", "10")
		limit, err := strconv.Atoi(limitParam)
		if err != nil {
			log.Event(ctx, "numeric search parameter provided with non numeric characters", log.Data{
				"param": "limit",
				"value": limitParam,
			}, log.WARN)
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
		if limit < 0 {
			log.Event(ctx, "numeric search parameter provided with negative value", log.Data{
				"param": "limit",
				"value": limitParam,
			}, log.WARN)
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}

		offsetParam := paramGet(params, "offset", "0")
		offset, err := strconv.Atoi(offsetParam)
		if err != nil {
			log.Event(ctx, "numeric search parameter provided with non numeric characters", log.Data{
				"param": "from",
				"value": offsetParam,
			}, log.WARN)
			http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
			return
		}
		if offset < 0 {
			log.Event(ctx, "numeric search parameter provided with negative value", log.Data{
				"param": "from",
				"value": offsetParam,
			}, log.WARN)
			http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
			return
		}

		typesParam := paramGet(params, "content_type", defaultContentTypes)

		formattedQuery, err := queryBuilder.BuildSearchQuery(ctx, q, typesParam, sort, limit, offset)
		if err != nil {
			log.Event(ctx, "creation of search query failed", log.Data{"q": q, "sort": sort, "limit": limit, "offset": offset}, log.Error(err), log.ERROR)
			http.Error(w, "Failed to create search query", http.StatusInternalServerError)
			return
		}

		responseData, err := elasticSearchClient.MultiSearch(ctx, "ons", "", formattedQuery)
		if err != nil {
			log.Event(ctx, "elasticsearch query failed", log.Error(err), log.ERROR)
			http.Error(w, "Failed to run search query", http.StatusInternalServerError)
			return
		}

		if !json.Valid([]byte(responseData)) {
			log.Event(ctx, "elastic search returned invalid JSON for search query", log.ERROR)
			http.Error(w, "Failed to process search query", http.StatusInternalServerError)
			return
		}

		if !paramGetBool(params, "raw", false) {
			responseData, err = transformer.TransformSearchResponse(ctx, responseData)
			if err != nil {
				log.Event(ctx, "transformation of response data failed", log.Error(err), log.ERROR)
				http.Error(w, "Failed to transform search result", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		_, err = w.Write(responseData)
		if err != nil {
			log.Event(ctx, "writing response failed", log.Error(err), log.ERROR)
			http.Error(w, "Failed to write http response", http.StatusInternalServerError)
			return
		}

	}
}
