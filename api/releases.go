package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-search-api/query"
	"github.com/ONSdigital/log.go/v2/log"
)

func CreateReleaseRequest(w http.ResponseWriter, req *http.Request, validator QueryParamValidator) (string, *query.ReleaseSearchRequest) {
	ctx := req.Context()
	params := req.URL.Query()

	queryString := params.Get("query")
	term, template := query.ParseQuery(queryString)

	limitParam := paramGet(params, ParamLimit, "10")
	limit, err := validator.Validate(ctx, ParamLimit, limitParam)
	if err != nil {
		log.Warn(ctx, err.Error(), log.Data{"param": ParamLimit, "value": limitParam})
		http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
		return "", nil
	}

	offsetParam := paramGet(params, ParamOffset, "0")
	offset, err := validator.Validate(ctx, ParamOffset, offsetParam)
	if err != nil {
		log.Warn(ctx, err.Error(), log.Data{"param": ParamOffset, "value": offsetParam})
		http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
		return "", nil
	}

	sortParam := paramGet(params, ParamSort, query.RelDateAsc.String())
	sort, err := validator.Validate(ctx, ParamSort, sortParam)
	if err != nil {
		log.Warn(ctx, err.Error(), log.Data{"param": ParamSort, "value": sortParam})
		http.Error(w, "Invalid sort parameter", http.StatusBadRequest)
		return "", nil
	}

	fromDateParam := paramGet(params, "fromDate", "")
	fromDate, err := validator.Validate(ctx, "date", fromDateParam)
	if err != nil {
		log.Warn(ctx, err.Error(), log.Data{"param": "fromDate", "value": fromDateParam})
		http.Error(w, "Invalid fromDate parameter", http.StatusBadRequest)
		return "", nil
	}

	toDateParam := paramGet(params, "toDate", "")
	toDate, err := validator.Validate(ctx, "date", toDateParam)
	if err != nil {
		log.Warn(ctx, err.Error(), log.Data{"param": "toDate", "value": toDateParam})
		http.Error(w, "Invalid toDate parameter", http.StatusBadRequest)
		return "", nil
	}

	if fromAfterTo(fromDate.(query.Date), toDate.(query.Date)) {
		log.Warn(ctx, "fromDate after toDate", log.Data{"fromDate": fromDateParam, "toDate": toDateParam})
		http.Error(w, "invalid dates - 'from' after 'to'", http.StatusBadRequest)
		return "", nil
	}

	relTypeParam := paramGet(params, "release-type", query.Published.String())
	relType, err := validator.Validate(ctx, "release-type", relTypeParam)
	if err != nil {
		log.Warn(ctx, err.Error(), log.Data{"param": "release-type", "value": relTypeParam})
		http.Error(w, "Invalid release-type parameter", http.StatusBadRequest)
		return "", nil
	}
	provisional := paramGetBool(params, ParamSubtypeProvisional, false)
	confirmed := paramGetBool(params, ParamSubtypeConfirmed, false)
	postponed := paramGetBool(params, ParamSubtypePostponed, false)
	highlight := paramGetBool(params, ParamHighlight, true)
	census := paramGetBool(params, ParamCensus, false)

	return queryString, &query.ReleaseSearchRequest{
		Term:           term,
		Template:       template,
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
}

// SearchReleasesHandlerFunc returns a http handler function handling release calendar search api requests.
func SearchReleasesHandlerFunc(validator QueryParamValidator, builder ReleaseQueryBuilder, searcher DpElasticSearcher, transformer ReleaseResponseTransformer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		params := req.URL.Query()

		queryString, searchReq := CreateReleaseRequest(w, req, validator)
		if searchReq == nil {
			return // error already handled
		}

		searches, err := builder.BuildSearchQuery(ctx, searchReq)
		if err != nil {
			log.Error(ctx, "creation of search release query failed", err, log.Data{
				ParamQ:      queryString,
				ParamSort:   searchReq.SortBy,
				ParamLimit:  searchReq.Size,
				ParamOffset: searchReq.From,
			})
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
			responseData, err = transformer.TransformSearchResponse(ctx, responseData, *searchReq, searchReq.Highlight)
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
