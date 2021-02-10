package api

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/ONSdigital/dp-search-api/query"
	"github.com/ONSdigital/log.go/log"
)

type dataLookupRequest struct {
	Uris  []string
	Types []string
}

type dataLookupResponse struct {
	Responses []interface{} `json:"responses"`
}

var dataTemplates *template.Template

// SetupData loads templates for use by the timeseries lookup handler and should be done only once
func SetupData() error {
	templates, err := template.ParseFiles("templates/data/queryByUri.tmpl")
	dataTemplates = templates
	return err
}

// DataLookupHandlerFunc returns a http handler function handling search api requests.
func DataLookupHandlerFunc(elasticSearchClient ElasticSearcher) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		params := req.URL.Query()
		reqParams := dataLookupRequest{
			Uris:  params["uris"],
			Types: params["types"],
		}
		var doc bytes.Buffer

		err := dataTemplates.Execute(&doc, reqParams)
		if err != nil {
			log.Event(ctx, "creation of search from template failed", log.Data{"Params": reqParams}, log.Error(err), log.ERROR)
			http.Error(w, "Failed to create query", http.StatusInternalServerError)
			return
		}

		formattedQuery, err := query.FormatMultiQuery(doc.Bytes())
		if err != nil {
			log.Event(ctx, "formating of data query for elasticsearch failed", log.Error(err), log.ERROR)
			http.Error(w, "Failed to create query", http.StatusInternalServerError)
			return
		}

		responseString, err := elasticSearchClient.Search(ctx, "", "", formattedQuery)
		if err != nil {
			log.Event(ctx, "elasticsearch data query failed", log.Error(err), log.ERROR)
			http.Error(w, "Failed to run data query", http.StatusInternalServerError)
			return
		}

		responseData := dataLookupResponse{Responses: make([]interface{}, 1)}
		if err := json.Unmarshal([]byte(responseString), &responseData.Responses[0]); err != nil {
			log.Event(ctx, "failed to unmarshal response from elasticsearch for data query", log.Error(err), log.ERROR)
			http.Error(w, "Failed to process data query", http.StatusInternalServerError)
			return
		}

		dataWithResponse, err := json.Marshal(responseData)
		if err != nil {
			log.Event(ctx, "Failed to marshal response data for data query", log.Error(err), log.ERROR)
			http.Error(w, "Failed to encode data query response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		w.Write(dataWithResponse)
	}
}
