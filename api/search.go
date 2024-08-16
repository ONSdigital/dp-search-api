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

	brlCli "github.com/ONSdigital/dp-api-clients-go/v2/nlp/berlin"
	brModel "github.com/ONSdigital/dp-api-clients-go/v2/nlp/berlin/models"
	catCli "github.com/ONSdigital/dp-api-clients-go/v2/nlp/category"
	catModel "github.com/ONSdigital/dp-api-clients-go/v2/nlp/category/models"
	"github.com/ONSdigital/dp-elasticsearch/v3/client"
	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	"github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/dp-search-api/query"
	scrModel "github.com/ONSdigital/dp-search-scrubber-api/models"
	scrSdk "github.com/ONSdigital/dp-search-scrubber-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
)

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
	ParamNLPWeighting       = "nlp_weighting"
	ParamDatasetIDs         = "dataset_ids"
	ParamURIPrefix          = "uri_prefix"
	ParamCDIDs              = "cdids"
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

func validateURIPrefix(uriPrefix string) (string, error) {
	parts := strings.Split(uriPrefix, ",")
	uriPrefix = parts[0]

	if len(uriPrefix) < 1 || !strings.HasPrefix(uriPrefix, "/") {
		return "", fmt.Errorf("invalid URI prefix parameter")
	}

	if !strings.HasSuffix(uriPrefix, "/") {
		uriPrefix += "/"
	}

	return uriPrefix, nil
}

func parseURIPrefix(ctx context.Context, params url.Values) (uriPrefix string, err error) {
	uriPrefix = paramGet(params, ParamURIPrefix, "")
	if uriPrefix != "" {
		var uriPrefixErr error
		uriPrefix, uriPrefixErr = validateURIPrefix(uriPrefix)
		if uriPrefixErr != nil {
			log.Warn(ctx, uriPrefixErr.Error(), log.Data{"param": ParamURIPrefix, "value": uriPrefix})
			return "", uriPrefixErr
		}
	}
	return uriPrefix, nil
}

// validateCDIDs checks that all the provided CDIDs are not blank and 4 characters long
// returns nil and an empty array if all of them are valid,
// returns error and a list of CDIDs that are not valid, if at least one is not valid
func validateCDIDs(cdids []string) (invalidCDIDs []string, err error) {
	if len(cdids) == 0 {
		return nil, nil
	}

	validationRegex := regexp.MustCompile(`^[a-zA-Z0-9]{4,5}$`)

	for _, cdid := range cdids {
		if !validationRegex.MatchString(cdid) {
			if cdid == "" {
				invalidCDIDs = append(invalidCDIDs, "<blank>")
			} else {
				invalidCDIDs = append(invalidCDIDs, cdid)
			}
		}
	}

	if len(invalidCDIDs) > 0 {
		err = fmt.Errorf("CDID(s) not valid: %v", invalidCDIDs)
	}

	return invalidCDIDs, err
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
func CreateRequests(w http.ResponseWriter, req *http.Request, cfg *config.Config, validator QueryParamValidator, nlpCriteria *query.NlpCriteria) (string, *query.SearchRequest, *query.CountRequest) {
	ctx := req.Context()
	params := req.URL.Query()

	// Sanitise and validate the query string
	sanitisedQuery, sanitiseErr := sanitiseAndValidateQuery(ctx, params)
	if sanitiseErr != nil {
		http.Error(w, sanitiseErr.Error(), http.StatusBadRequest)
		return "", nil, nil
	}

	// Parse and validate other query parameters
	highlight := paramGetBool(params, ParamHighlight, true)

	topics, topicErr := parseTopics(ctx, params)
	if topicErr != nil {
		http.Error(w, topicErr.Error(), http.StatusBadRequest)
		return "", nil, nil
	}

	limit, limitErr := parseLimit(ctx, params, validator)
	if limitErr != nil {
		http.Error(w, limitErr.Error(), http.StatusBadRequest)
		return "", nil, nil
	}

	offset, offsetErr := parseOffset(ctx, params, validator)
	if offsetErr != nil {
		http.Error(w, offsetErr.Error(), http.StatusBadRequest)
		return "", nil, nil
	}

	contentTypes, contentTypesErr := parseAndValidateContentTypes(ctx, params)
	if contentTypesErr != nil {
		http.Error(w, contentTypesErr.Error(), http.StatusBadRequest)
		return "", nil, nil
	}

	sort, sortErr := parseAndValidateSort(ctx, params, validator)
	if sortErr != nil {
		http.Error(w, sortErr.Error(), http.StatusBadRequest)
		return "", nil, nil
	}

	fromDateParam := paramGet(params, "fromDate", "")
	fromDate, err := validator.Validate(ctx, "date", fromDateParam)
	if err != nil {
		log.Warn(ctx, err.Error(), log.Data{"param": "fromDate", "value": fromDateParam})
		http.Error(w, "Invalid fromDate parameter", http.StatusBadRequest)
		return "", nil, nil
	}

	toDateParam := paramGet(params, "toDate", "")
	toDate, err := validator.Validate(ctx, "date", toDateParam)
	if err != nil {
		log.Warn(ctx, err.Error(), log.Data{"param": "toDateParam", "value": toDateParam})
		http.Error(w, "Invalid toDate parameter", http.StatusBadRequest)
		return "", nil, nil
	}

	if fromAfterTo(fromDate.(query.Date), toDate.(query.Date)) {
		log.Warn(ctx, "fromDate after toDate", log.Data{"fromDate": fromDateParam, "toDate": toDateParam})
		http.Error(w, "invalid dates - 'from' after 'to'", http.StatusBadRequest)
		return "", nil, nil
	}

	uriPrefix, uriPrefixErr := parseURIPrefix(ctx, params)
	if uriPrefixErr != nil {
		http.Error(w, uriPrefixErr.Error(), http.StatusBadRequest)
		return "", nil, nil
	}

	cdids, cdidErr := parseCDID(ctx, params)
	if cdidErr != nil {
		http.Error(w, cdidErr.Error(), http.StatusBadRequest)
		return "", nil, nil
	}

	datasetIDs, datasetErr := parseDatasetIDs(ctx, params)
	if datasetErr != nil {
		http.Error(w, datasetErr.Error(), http.StatusBadRequest)
		return "", nil, nil
	}

	// Create SearchRequest
	reqSearch := createSearchRequest(sanitisedQuery, offset, limit, contentTypes, fromDate.(query.Date), toDate.(query.Date), topics, sort, highlight, datasetIDs, uriPrefix, cdids, nlpCriteria)

	// Process additional parameters like Population Types, Dimensions, and Dataset IDs
	reqSearch.PopulationTypes = parsePopulationTypes(params)
	reqSearch.Dimensions = parseDimensions(params)

	// Create CountRequest
	reqCount := createCountRequest(sanitisedQuery)

	if cfg.DebugMode {
		log.Info(ctx, "[DEBUG]", log.Data{"search_request": reqSearch})
	}

	return params.Get(ParamQ), reqSearch, reqCount
}

func sanitiseAndValidateQuery(ctx context.Context, params url.Values) (string, error) {
	q := params.Get(ParamQ)
	sanitisedQuery := sanitiseDoubleQuotes(q)
	queryHasSpecialChars := checkForSpecialCharacters(sanitisedQuery)

	if queryHasSpecialChars {
		log.Info(ctx, "rejecting query as it contains special characters", log.Data{"query": sanitisedQuery})
		return "", errors.New("invalid characters in query")
	}
	return sanitisedQuery, nil
}

func processLimit(ctx context.Context, params url.Values, validator QueryParamValidator) (limit int, err string) {
	limitParam := paramGet(params, ParamLimit, "10")
	validatedLimit, validationErr := validator.Validate(ctx, ParamLimit, limitParam)
	if validationErr != nil {
		log.Warn(ctx, validationErr.Error(), log.Data{"param": ParamLimit, "value": limitParam})
		return 0, "Invalid limit parameter"
	}
	return validatedLimit.(int), ""
}

func processContentTypes(ctx context.Context, params url.Values) (contentTypes []string, err string) {
	// read content type (expected CSV value), with default, to make sure some content types are
	contentTypesParam := paramGet(params, ParamContentType, "")
	contentTypes = defaultContentTypes
	if contentTypesParam != "" {
		contentTypes = strings.Split(contentTypesParam, ",")
		disallowed, validationErr := validateContentTypes(contentTypes)
		if validationErr != nil {
			log.Warn(ctx, validationErr.Error(), log.Data{"param": ParamContentType, "value": contentTypesParam, "disallowed": disallowed})
			return nil, fmt.Sprint("Invalid content_type(s): ", strings.Join(disallowed, ","))
		}
	}
	return contentTypes, ""
}

func processSort(ctx context.Context, params url.Values, validator QueryParamValidator) (sort, err string) {
	sortParam := paramGet(params, ParamSort, "relevance")
	validatedSort, validationErr := validator.Validate(ctx, ParamSort, sortParam)
	if validationErr != nil {
		log.Warn(ctx, validationErr.Error(), log.Data{"param": ParamSort, "value": sortParam})
		return "", "Invalid sort parameter"
	}
	return validatedSort.(string), ""
}

func parseCDID(ctx context.Context, params url.Values) (cdid []string, err error) {
	cdidParam := paramGet(params, ParamCDIDs, "")
	if cdidParam != "" {
		cdid = strings.Split(cdidParam, ",")
		disallowed, validationErr := validateCDIDs(cdid)
		if validationErr != nil {
			log.Warn(ctx, validationErr.Error(), log.Data{"param": ParamCDIDs, "value": cdidParam, "disallowed": disallowed})
			return nil, validationErr
		}
	}
	return cdid, nil
}

func parseLimit(ctx context.Context, params url.Values, validator QueryParamValidator) (limit int, err error) {
	limitParam := paramGet(params, ParamLimit, "10")
	validatedLimit, validationErr := validator.Validate(ctx, ParamLimit, limitParam)
	if validationErr != nil {
		log.Warn(ctx, validationErr.Error(), log.Data{"param": ParamLimit, "value": limitParam})
		return 0, errors.New("invalid limit parameter")
	}
	return validatedLimit.(int), nil
}

func parseOffset(ctx context.Context, params url.Values, validator QueryParamValidator) (offset int, err error) {
	offsetParam := paramGet(params, ParamOffset, "0")
	validatedOffset, validationErr := validator.Validate(ctx, ParamOffset, offsetParam)
	if validationErr != nil {
		log.Warn(ctx, validationErr.Error(), log.Data{"param": ParamOffset, "value": offsetParam})
		return 0, errors.New("invalid offset parameter")
	}
	return validatedOffset.(int), nil
}

func parseAndValidateContentTypes(ctx context.Context, params url.Values) (contentTypes []string, err error) {
	// read content type (expected CSV value), with default, to make sure some content types are
	contentTypesParam := paramGet(params, ParamContentType, "")
	contentTypes = defaultContentTypes
	if contentTypesParam != "" {
		contentTypes = strings.Split(contentTypesParam, ",")
		disallowed, validationErr := validateContentTypes(contentTypes)
		if validationErr != nil {
			log.Warn(ctx, validationErr.Error(), log.Data{"param": ParamContentType, "value": contentTypesParam, "disallowed": disallowed})
			return nil, fmt.Errorf("invalid content_type(s): %s", strings.Join(disallowed, ","))
		}
	}
	return contentTypes, nil
}

func parseAndValidateSort(ctx context.Context, params url.Values, validator QueryParamValidator) (sort string, err error) {
	sortParam := paramGet(params, ParamSort, "relevance")
	validatedSort, validationErr := validator.Validate(ctx, ParamSort, sortParam)
	if validationErr != nil {
		log.Warn(ctx, validationErr.Error(), log.Data{"param": ParamSort, "value": sortParam})
		return "", errors.New("invalid sort parameter")
	}
	return validatedSort.(string), nil
}

func createSearchRequest(sanitisedQuery string, offset, limit int, contentTypes []string, fromDate, toDate query.Date, topics []string, sort string, highlight bool, datasetIDs []string, uriPrefix string, cdids []string, nlpCriteria *query.NlpCriteria) *query.SearchRequest {
	reqSearch := &query.SearchRequest{
		Term:           sanitisedQuery,
		From:           offset,
		Size:           limit,
		Types:          contentTypes,
		ReleasedAfter:  fromDate,
		ReleasedBefore: toDate,
		Topic:          topics,
		SortBy:         sort,
		Highlight:      highlight,
		Now:            time.Now().UTC().Format(time.RFC3339),
		DatasetIDs:     datasetIDs,
		URIPrefix:      uriPrefix,
		CDIDs:          cdids,
	}

	if nlpCriteria != nil {
		if nlpCriteria.UseCategory {
			reqSearch.NlpCategories = nlpCriteria.Categories
		}
		if nlpCriteria.UseSubdivision {
			reqSearch.NlpSubdivisionWords = nlpCriteria.SubdivisionWords
		}
	}

	return reqSearch
}

func parsePopulationTypes(params url.Values) []*query.PopulationTypeRequest {
	popTypesParam := paramGet(params, ParamPopulationTypes, "")
	if popTypesParam == "" {
		return nil
	}
	popTypes := strings.Split(popTypesParam, ",")
	p := make([]*query.PopulationTypeRequest, len(popTypes))
	for i, popType := range popTypes {
		p[i] = &query.PopulationTypeRequest{Key: popType}
	}
	return p
}

func parseDimensions(params url.Values) []*query.DimensionRequest {
	dimensionsParam := paramGet(params, ParamDimensions, "")
	if dimensionsParam == "" {
		return nil
	}
	dims := strings.Split(dimensionsParam, ",")
	d := make([]*query.DimensionRequest, len(dims))
	for i, dim := range dims {
		d[i] = &query.DimensionRequest{Key: dim}
	}
	return d
}

func parseDatasetIDs(ctx context.Context, params url.Values) (datasetIDs []string, err error) {
	datasetIDParam := paramGet(params, ParamDatasetIDs, "")
	if datasetIDParam != "" {
		datasetIDs = sanitiseURLParams(datasetIDParam)
		disallowed, validationErr := validateDatasetIDs(datasetIDs)
		if validationErr != nil {
			log.Warn(ctx, validationErr.Error(), log.Data{"param": ParamDatasetIDs, "value": datasetIDParam, "disallowed": disallowed})
			return nil, fmt.Errorf("invalid dataset_ids: %s", strings.Join(disallowed, ","))
		}
	}
	return datasetIDs, nil
}

func validateDatasetIDs(datasetIDs []string) (invalidDatasetIDs []string, err error) {
	if len(datasetIDs) == 0 {
		return nil, nil
	}

	for _, did := range datasetIDs {
		if did == "" {
			invalidDatasetIDs = append(invalidDatasetIDs, "<blank>")
		} else if len(did) <= 1 {
			invalidDatasetIDs = append(invalidDatasetIDs, did)
		}
	}

	if len(invalidDatasetIDs) > 0 {
		err = fmt.Errorf("datasetID(s) not valid: %v", invalidDatasetIDs)
	}

	return invalidDatasetIDs, err
}

func parseTopics(ctx context.Context, params url.Values) (topics []string, err error) {
	topicsParam := paramGet(params, ParamTopics, "")
	if topicsParam != "" {
		topics = sanitiseURLParams(topicsParam)
		disallowed, validationErr := validateTopics(topics)
		if validationErr != nil {
			log.Warn(ctx, validationErr.Error(), log.Data{"param": ParamTopics, "value": topicsParam, "disallowed": disallowed})
			return nil, fmt.Errorf("invalid topics: %s", strings.Join(disallowed, ","))
		}
	}
	return topics, nil
}

func validateTopics(topics []string) (invalidTopics []string, err error) {
	// If no topics are provided, there's nothing to validate, so return nil for both results.
	if len(topics) == 0 {
		return nil, nil
	}

	// Iterate through each topic ID to validate it.
	for _, topicID := range topics {
		// Check if the topic ID is blank.
		if topicID == "" {
			invalidTopics = append(invalidTopics, "<blank>") // Mark blank topic IDs as invalid.
		} else if len(topicID) != 4 { // Check if the topic ID length is 4 characters.
			invalidTopics = append(invalidTopics, topicID) // Mark topic IDs that are too short or too long as invalid.
		}
	}

	// If any invalid topics were found, return an error listing the invalid topics.
	if len(invalidTopics) > 0 {
		err = fmt.Errorf("topic(s) not valid: %v", invalidTopics)
	}

	// Return the slice of invalid topics and any error.
	return invalidTopics, err
}

func createCountRequest(sanitisedQuery string) *query.CountRequest {
	// create CountRequest with the sanitised query.
	// Note that this is only used to generate the `distinct_items_count`.
	// Other counts are done as aggregations of the search request.
	return &query.CountRequest{
		Term:        sanitisedQuery,
		CountEnable: true,
	}
}

// SearchHandlerFunc returns a http handler function handling search api requests.
func SearchHandlerFunc(validator QueryParamValidator, queryBuilder QueryBuilder, cfg *config.Config, clList *ClientList, transformer ResponseTransformer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		params := req.URL.Query()

		nlpCriteria := getNLPCriteria(ctx, params, cfg, queryBuilder, clList)

		q, searchReq, countReq := CreateRequests(w, req, cfg, validator, nlpCriteria)
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
			processCountQuery(ctx, clList.DpESClient, queryBuilder, countReq, resCountChan)
		}()

		go func() {
			processSearchQuery(ctx, cfg, clList.DpESClient, queryBuilder, searchReq, resDataChan)
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
func LegacySearchHandlerFunc(validator QueryParamValidator, queryBuilder QueryBuilder, cfg *config.Config, clList *ClientList, transformer ResponseTransformer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		params := req.URL.Query()

		nlpCriteria := getNLPCriteria(ctx, params, cfg, queryBuilder, clList)

		q, searchReq, countReq := CreateRequests(w, req, cfg, validator, nlpCriteria)
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

		responseData, err := clList.DeprecatedESClient.MultiSearch(ctx, "ons", "", formattedQuery)
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

func getNLPCriteria(ctx context.Context, params url.Values, cfg *config.Config, queryBuilder QueryBuilder, clList *ClientList) *query.NlpCriteria {
	nlpWeightingRequested := paramGetBool(params, ParamNLPWeighting, false)

	if cfg.EnableNLPWeighting && nlpWeightingRequested {
		nlpSettings := query.NlpSettings{}

		log.Info(ctx, "Employing advanced natural language processing techniques to optimize Elasticsearch querying for enhanced result relevance.")

		if err := json.Unmarshal([]byte(cfg.NLPSettings), &nlpSettings); err != nil {
			log.Error(ctx, "problem unmarshaling NLPSettings", err)
		}

		return AddNlpToSearch(ctx, queryBuilder, params, nlpSettings, clList)
	}

	return nil
}

func (a SearchAPI) CreateSearchIndexHandlerFunc(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	indexName := createIndexName("ons")

	err := a.clList.DpESClient.CreateIndex(ctx, indexName, elasticsearch.GetSearchIndexSettings())
	if err != nil {
		log.Error(ctx, "creating index failed with this error", err)
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
func processSearchQuery(ctx context.Context, cfg *config.Config, elasticSearchClient DpElasticSearcher, queryBuilder QueryBuilder, reqParams *query.SearchRequest, responseDataChan chan []byte) {
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

	if cfg.DebugMode {
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

func AddNlpToSearch(ctx context.Context, queryBuilder QueryBuilder, params url.Values, nlpSettings query.NlpSettings, clList *ClientList) *query.NlpCriteria {
	var berlin *brModel.Berlin
	var category *[]catModel.Category
	var scrubber *scrModel.ScrubberResp

	scrOpt := scrSdk.OptInit()

	// If scrubber is down for any reason, we need to stop the NLP feature from interfering with regular dp-search-api resp
	scrubber, err := clList.ScrubberClient.GetScrubber(ctx, scrOpt.Q(params.Get("q")))
	if err != nil {
		log.Error(ctx, "error making request to scrubber", err)
		return nil
	}

	brOpt := brlCli.OptInit()

	berlin, err = clList.BerlinClient.GetBerlin(ctx, *brOpt.Q(scrubber.Query))
	if err != nil || berlin == nil {
		log.Error(ctx, "error making request to berlin", err)
		// If berlin isn't working or gives an empty response
		// We need to make sure the query is accessible to category
		berlin = &brModel.Berlin{
			Query: scrubber.Query,
		}
	}

	catOpt := catCli.OptInit()

	category, err = clList.CategoryClient.GetCategory(ctx, *catOpt.Q(berlin.Query))
	if err != nil {
		log.Error(ctx, "error making request to category", err)
	}

	var nlpCriteria *query.NlpCriteria

	log.Info(ctx, "NLP full response", log.Data{
		"Does category exist": category != nil,
		"Berlin":              berlin,
		"Scrubber":            scrubber,
		"Category":            category,
	})

	// Process NLP Criteria based on the provided category data.
	// If categories exist, iterate through them, limiting the loop based on the configuration
	// NLP category limit. For each category, build NLP criteria to be used in the query to ElasticSearch.
	if category != nil {
		for i, cat := range *category {
			if nlpSettings.CategoryLimit > 0 && nlpSettings.CategoryLimit <= i {
				break
			}
			log.Info(ctx, "category codes", log.Data{
				"category_code_1": cat.Code[0],
				"category_code_2": cat.Code[1],
			})
			nlpCriteria = queryBuilder.AddNlpCategorySearch(
				nlpCriteria,
				cat.Code[0],
				cat.Code[1],
				nlpSettings.CategoryWeighting,
			)
		}
	}

	// If berlin exists, add the subdivisions to NLP criteria.
	// They'll be used later in the query to ElasticSearch
	if len(berlin.Matches) > 0 && len(berlin.Matches[0].Loc.Subdivision) == 2 {
		nlpCriteria = queryBuilder.AddNlpSubdivisionSearch(nlpCriteria, berlin.Matches[0].Loc.Subdivision[1])
	}

	return nlpCriteria
}
