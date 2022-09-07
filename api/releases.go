package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/ONSdigital/dp-elasticsearch/v3/client"
	"github.com/ONSdigital/dp-search-api/query"
	"github.com/ONSdigital/log.go/v2/log"
)

// SearchReleasesHandlerFunc returns a http handler function handling release calendar search api requests.
func SearchReleasesHandlerFunc(validator QueryParamValidator, builder ReleaseQueryBuilder, searcher DpElasticSearcher, transformer ReleaseResponseTransformer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		params := req.URL.Query()

		queryString, err := url.QueryUnescape(params.Get("query"))
		if err != nil {
			log.Warn(ctx, err.Error(), log.Data{"param": "query", "value": params.Get("query")})
			http.Error(w, "Bad url encoding of the query parameter", http.StatusBadRequest)
			return
		}
		sanitisedQuery := sanitiseDoubleQuotes(queryString)

		limitParam := paramGet(params, "limit", "10")
		limit, err := validator.Validate(ctx, "limit", limitParam)
		if err != nil {
			log.Warn(ctx, err.Error(), log.Data{"param": "limit", "value": limitParam})
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}

		offsetParam := paramGet(params, "offset", "0")
		offset, err := validator.Validate(ctx, "offset", offsetParam)
		if err != nil {
			log.Warn(ctx, err.Error(), log.Data{"param": "offset", "value": offsetParam})
			http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
			return
		}

		sortParam := paramGet(params, "sort", query.RelDateAsc.String())
		sort, err := validator.Validate(ctx, "sort", sortParam)
		if err != nil {
			log.Warn(ctx, err.Error(), log.Data{"param": "sort", "value": sortParam})
			http.Error(w, "Invalid sort parameter", http.StatusBadRequest)
			return
		}

		fromDateParam := paramGet(params, "fromDate", "")
		fromDate, err := validator.Validate(ctx, "date", fromDateParam)
		if err != nil {
			log.Warn(ctx, err.Error(), log.Data{"param": "fromDate", "value": fromDateParam})
			http.Error(w, "Invalid dateFrom parameter", http.StatusBadRequest)
			return
		}

		toDateParam := paramGet(params, "toDate", "")
		toDate, err := validator.Validate(ctx, "date", toDateParam)
		if err != nil {
			log.Warn(ctx, err.Error(), log.Data{"param": "toDate", "value": toDateParam})
			http.Error(w, "Invalid dateTo parameter", http.StatusBadRequest)
			return
		}

		if fromAfterTo(fromDate.(query.Date), toDate.(query.Date)) {
			log.Warn(ctx, "fromDate after toDate", log.Data{"fromDate": fromDateParam, "toDate": toDateParam})
			http.Error(w, "invalid dates - 'from' after 'to'", http.StatusBadRequest)
			return
		}

		relTypeParam := paramGet(params, "release-type", query.Published.String())
		relType, err := validator.Validate(ctx, "release-type", relTypeParam)
		if err != nil {
			log.Warn(ctx, err.Error(), log.Data{"param": "release-type", "value": relTypeParam})
			http.Error(w, "Invalid release-type parameter", http.StatusBadRequest)
			return
		}
		provisional := paramGetBool(params, "subtype-provisional", false)
		confirmed := paramGetBool(params, "subtype-confirmed", false)
		postponed := paramGetBool(params, "subtype-postponed", false)
		highlight := paramGetBool(params, "highlight", true)
		census := paramGetBool(params, "census", false)

		searchReq := query.ReleaseSearchRequest{
			Term:           sanitisedQuery,
			From:           offset.(int),
			Size:           limit.(int),
			SortBy:         sort.(query.Sort),
			ReleasedAfter:  fromDate.(query.Date),
			ReleasedBefore: toDate.(query.Date),
			Type:           relType.(query.ReleaseType),
			Provisional:    provisional,
			Confirmed:      confirmed,
			Postponed:      postponed,
			Census:         census,
			Highlight:      highlight,
		}

		formattedQuery, err := builder.BuildSearchQuery(ctx, searchReq)
		if err != nil {
			log.Error(ctx, "creation of search release query failed", err, log.Data{"q": sanitisedQuery, "sort": sort, "limit": limit, "offset": offset})
			http.Error(w, "Failed to create search release query", http.StatusInternalServerError)
			return
		}

		var searches []client.Search
		err = json.Unmarshal(formattedQuery, &searches)
		if err != nil {
			log.Error(ctx, "creation of search release query failed", err, log.Data{"q": sanitisedQuery, "sort": sort, "limit": limit, "offset": offset})
			http.Error(w, "Failed to create search release query", http.StatusInternalServerError)
			return
		}

		responseData, err := searcher.MultiSearch(ctx, searches, nil)
		if err != nil {
			log.Error(ctx, "elasticsearch query failed", err)
			http.Error(w, "Failed to run search query", http.StatusInternalServerError)
			return
		}

		if !json.Valid(responseData) {
			log.Error(ctx, "elastic search returned invalid JSON for search release query", errors.New("elastic search returned invalid JSON for search release query"))
			http.Error(w, "Failed to process search release query", http.StatusInternalServerError)
			return
		}

		if !paramGetBool(params, "raw", false) {
			responseData, err = transformer.TransformSearchResponse(ctx, responseData, searchReq, highlight)
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

func fromAfterTo(from, to query.Date) bool {
	if !time.Time(from).IsZero() && !time.Time(to).IsZero() && time.Time(from).After(time.Time(to)) {
		return true
	}

	return false
}
