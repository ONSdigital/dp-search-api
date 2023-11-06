package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
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

var debug = false

const (
	ParamQ                  = "q"
	ParamSort               = "sort"
	ParamHighlight          = "highlight"
	ParamTopics             = "topics"
	ParamLimit              = "limit"
	ParamOffset             = "offset"
	ParamContentType        = "content_type"
	ParamPopulationTypes    = "population_types"
	ParamDimensions         = "dimensions"
	ParamSubtypeProvisional = "subtype-provisional"
	ParamSubtypeConfirmed   = "subtype-confirmed"
	ParamSubtypePostponed   = "subtype-postponed"
	ParamCensus             = "census"
)

// defaultContentTypes is an array of all valid content types, which is the default param value
var defaultContentTypes = []string{
	"article",
	"article_download",
	"bulletin",
	"compendium_landing_page",
	"compendium_chapter",
	"compendium_data",
	"dataset",
	"dataset_landing_page",
	"product_page",
	"reference_tables",
	"release",
	"static_adhoc",
	"static_article",
	"static_foi",
	"static_landing_page",
	"static_methodology",
	"static_methodology_download",
	"static_page",
	"static_qmi",
	"timeseries",
	"timeseries_dataset",
}

// validateContentTypes checks that all the provided content types are allowed
// returns nil and an empty array if all of them are allowed,
// returns error and a list of content types that are not allowed, if at least one is not allowed
func validateContentTypes(contentTypes []string) (disallowed []string, err error) {
	validContentTypes := map[string]struct{}{}
	for _, valid := range defaultContentTypes {
		validContentTypes[valid] = struct{}{}
	}

	for _, t := range contentTypes {
		if _, ok := validContentTypes[t]; !ok {
			disallowed = append(disallowed, t)
			err = errors.New("content type(s) not allowed")
		}
	}

	return disallowed, err
}

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

// CreateRequests reads the parameters from the request and generates the corresponding SearchRequest and CountRequest
// If any validation fails, the http.Error is already handled, and nil is returned: in this case the caller may return straight away
func CreateRequests(w http.ResponseWriter, req *http.Request, validator QueryParamValidator) (string, *query.SearchRequest, *query.CountRequest) {
	ctx := req.Context()
	params := req.URL.Query()

	q := params.Get(ParamQ)
	sanitisedQuery := sanitiseDoubleQuotes(q)
	queryHasSpecialChars := checkForSpecialCharacters(sanitisedQuery)

	if queryHasSpecialChars {
		log.Info(ctx, "rejecting query as contained special characters", log.Data{"query": sanitisedQuery})
		http.Error(w, "Invalid characters in query", http.StatusBadRequest)
		return "", nil, nil
	}

	sortParam := paramGet(params, ParamSort, "relevance")
	sort, err := validator.Validate(ctx, ParamSort, sortParam)
	if err != nil {
		log.Warn(ctx, err.Error(), log.Data{"param": ParamSort, "value": sortParam})
		http.Error(w, "Invalid sort parameter", http.StatusBadRequest)
		return "", nil, nil
	}

	highlight := paramGetBool(params, ParamHighlight, true)

	topicsParam := paramGet(params, ParamTopics, "")
	topics := sanitiseURLParams(topicsParam)
	if topics != nil {
		log.Info(ctx, "topic extracted and sanitised from the request url params", log.Data{
			"param": ParamTopics,
			"value": topics,
		})
	}

	limitParam := paramGet(params, ParamLimit, "10")
	limit, err := validator.Validate(ctx, ParamLimit, limitParam)
	if err != nil {
		log.Warn(ctx, err.Error(), log.Data{"param": ParamLimit, "value": limitParam})
		http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
		return "", nil, nil
	}

	offsetParam := paramGet(params, ParamOffset, "0")
	offset, err := validator.Validate(ctx, ParamOffset, offsetParam)
	if err != nil {
		log.Warn(ctx, err.Error(), log.Data{"param": ParamOffset, "value": offsetParam})
		http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
		return "", nil, nil
	}

	// read content type (expected CSV value), with default, to make sure some content types are
	contentTypesParam := paramGet(params, ParamContentType, "")
	contentTypes := defaultContentTypes
	if contentTypesParam != "" {
		contentTypes = strings.Split(contentTypesParam, ",")
		disallowed, err := validateContentTypes(contentTypes)
		if err != nil {
			log.Warn(ctx, err.Error(), log.Data{"param": ParamContentType, "value": contentTypesParam, "disallowed": disallowed})
			http.Error(w, fmt.Sprint("Invalid content_type(s): ", strings.Join(disallowed, ",")), http.StatusBadRequest)
			return "", nil, nil
		}
	}

	fromDateParam := paramGet(params, "fromDate", "")
	fromDate, err := validator.Validate(ctx, "date", fromDateParam)
	if err != nil {
		log.Warn(ctx, err.Error(), log.Data{"param": "fromDate", "value": fromDateParam})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return "", nil, nil
	}

	toDateParam := paramGet(params, "toDate", "")
	toDate, err := validator.Validate(ctx, "date", toDateParam)
	if err != nil {
		log.Warn(ctx, err.Error(), log.Data{"param": "toDate", "value": toDateParam})
		http.Error(w, "Invalid dateTo parameter", http.StatusBadRequest)
		return "", nil, nil
	}

	// create SearchRequest with all the compulsory values
	reqSearch := &query.SearchRequest{
		Term:           sanitisedQuery,
		From:           offset.(int),
		Size:           limit.(int),
		Types:          contentTypes,
		ReleasedAfter:  fromDate.(query.Date),
		ReleasedBefore: toDate.(query.Date),
		Topic:          topics,
		SortBy:         sort.(string),
		Highlight:      highlight,
		Now:            time.Now().UTC().Format(time.RFC3339),
	}

	// population types only used if provided
	popTypesParam := paramGet(params, ParamPopulationTypes, "")
	if popTypesParam != "" {
		popTypes := strings.Split(popTypesParam, ",")
		p := make([]*query.PopulationTypeRequest, len(popTypes))
		for i, popType := range popTypes {
			p[i] = &query.PopulationTypeRequest{
				Key: popType,
			}
		}
		reqSearch.PopulationTypes = p
	}

	// dimensions only used if provided
	dimensionsParam := paramGet(params, ParamDimensions, "")
	if dimensionsParam != "" {
		dims := strings.Split(dimensionsParam, ",")
		d := make([]*query.DimensionRequest, len(dims))
		for i, dim := range dims {
			d[i] = &query.DimensionRequest{
				Key: dim,
			}
		}
		reqSearch.Dimensions = d
	}

	// create CountRequest with the sanitized query.
	// Note that this is only used to generate the `distinct_items_count`.
	// Other counts are done as aggregations of the search request.
	reqCount := &query.CountRequest{
		Term:        sanitisedQuery,
		CountEnable: true,
	}

	if debug {
		log.Info(ctx, "[DEBUG]", log.Data{"search_request": reqSearch})
	}

	return q, reqSearch, reqCount
}

// SearchHandlerFunc returns a http handler function handling search api requests.
func SearchHandlerFunc(validator QueryParamValidator, queryBuilder QueryBuilder, elasticSearchClient DpElasticSearcher, transformer ResponseTransformer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		params := req.URL.Query()

		q, searchReq, countReq := CreateRequests(w, req, validator)
		if searchReq == nil || countReq == nil {
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
			processCountQuery(ctx, elasticSearchClient, queryBuilder, countReq, resCountChan)
		}()

		go func() {
			processSearchQuery(ctx, elasticSearchClient, queryBuilder, searchReq, resDataChan)
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

			responseSearchData, err = transformer.TransformSearchResponse(ctx, responseSearchData, q, searchReq.Highlight)
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

		q, searchReq, countReq := CreateRequests(w, req, validator)
		if searchReq == nil || countReq == nil {
			return
		}

		formattedQuery, err := queryBuilder.BuildSearchQuery(ctx, searchReq, false)
		if err != nil {
			log.Error(ctx, "creation of search query failed", err, log.Data{
				ParamQ:      q,
				ParamSort:   searchReq.SortBy,
				ParamLimit:  searchReq.Size,
				ParamOffset: searchReq.From,
			})
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
			responseData, err = transformer.TransformSearchResponse(ctx, responseData, q, searchReq.Highlight)
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

func checkForSpecialCharacters(str string) bool {
	re := regexp.MustCompile("[[:^ascii:]]")
	return re.MatchString(str)
}

// func processSearchQuery(ctx context.Context, elasticSearchClient DpElasticSearcher, queryBuilder QueryBuilder, sanitisedQuery, typesParam, sort string, topics []string, limit, offset int, responseDataChan chan []byte) {
func processSearchQuery(ctx context.Context, elasticSearchClient DpElasticSearcher, queryBuilder QueryBuilder, reqParams *query.SearchRequest, responseDataChan chan []byte) {
	formattedQuery, err := queryBuilder.BuildSearchQuery(ctx, reqParams, true)
	if err != nil {
		log.Error(ctx, "creation of search query failed", err, log.Data{
			ParamQ:      reqParams.Term,
			ParamSort:   reqParams.SortBy,
			ParamLimit:  reqParams.Size,
			ParamOffset: reqParams.From,
		})
		responseDataChan <- nil
		return
	}

	var searches []client.Search

	if marshalErr := json.Unmarshal(formattedQuery, &searches); marshalErr != nil {
		log.Error(ctx, "creation of search query failed", marshalErr, log.Data{
			ParamQ:      reqParams.From,
			ParamSort:   reqParams.From,
			ParamLimit:  reqParams.Size,
			ParamOffset: reqParams.From,
		})
		responseDataChan <- nil
		return
	}

	if debug {
		for i, s := range searches {
			log.Info(ctx, "[DEBUG] Search sent to elasticsearch", log.Data{"i": i, "header": s.Header, "query": string(s.Query)})
		}
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

func processCountQuery(ctx context.Context, elasticSearchClient DpElasticSearcher, queryBuilder QueryBuilder, reqParams *query.CountRequest, resCountChan chan []byte) {
	countQBytes, err := queryBuilder.BuildCountQuery(ctx, reqParams)
	if err != nil {
		log.Error(ctx, "creation of count query failed", err, log.Data{
			ParamQ: reqParams.Term})
		resCountChan <- nil
		return
	}

	countRes, err := elasticSearchClient.Count(ctx, client.Count{
		Query: countQBytes,
	})
	if err != nil {
		log.Error(ctx, "call to elasticsearch count api failed", err, log.Data{
			ParamQ: reqParams.Term})
		resCountChan <- nil
		return
	}
	resCountChan <- countRes
}
