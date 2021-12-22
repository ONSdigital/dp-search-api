package api

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/ONSdigital/dp-search-api/query"
	"github.com/ONSdigital/log.go/v2/log"
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
			log.Error(ctx, "creation of search from template failed", err, log.Data{"Params": reqParams})
			http.Error(w, "Failed to create query", http.StatusInternalServerError)
			return
		}

		formattedQuery, err := query.FormatMultiQuery(doc.Bytes())
		if err != nil {
			log.Error(ctx, "formating of data query for elasticsearch failed", err)
			http.Error(w, "Failed to create query", http.StatusInternalServerError)
			return
		}

		responseString, err := elasticSearchClient.Search(ctx, "", "", formattedQuery)
		if err != nil {
			log.Error(ctx, "elasticsearch data query failed", err)
			http.Error(w, "Failed to run data query", http.StatusInternalServerError)
			return
		}

		responseData := dataLookupResponse{Responses: make([]interface{}, 1)}
		if unMarshalErr := json.Unmarshal(responseString, &responseData.Responses[0]); unMarshalErr != nil {
			log.Error(ctx, "failed to unmarshal response from elasticsearch for data query", unMarshalErr)
			http.Error(w, "Failed to process data query", http.StatusInternalServerError)
			return
		}

		dataWithResponse, err := json.Marshal(responseData)
		if err != nil {
			log.Error(ctx, "Failed to marshal response data for data query", err)
			http.Error(w, "Failed to encode data query response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		if _, err := w.Write(dataWithResponse); err != nil {
			log.Error(ctx, "error occured while writing response data", err)
		}
	}
}
