package api

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/ONSdigital/go-ns/log"
)

type dataLookupRequest struct {
	Uris  []string
	Types []string
}

type dataLookupResponse struct {
	Responses []interface{} `json:"responses"`
}

var dataTemplates *template.Template

// SetupTimeseries loads templates for use by the timeseries lookup handler and should be done only once
func SetupData() error {
	templates, err := template.ParseFiles("templates/data/queryByUri.tmpl")
	dataTemplates = templates
	return err
}

// DataLookupHandlerFunc returns a http handler function handling search api requests.
func DataLookupHandlerFunc(elasticSearchClient ElasticSearcher) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		params := req.URL.Query()
		reqParams := dataLookupRequest{
			Uris:  params["uris"],
			Types: params["types"],
		}
		var doc bytes.Buffer

		err := dataTemplates.Execute(&doc, reqParams)
		if err != nil {
			log.Debug("Failed to create search from template", log.Data{"Error": err.Error(), "Params": reqParams})
			http.Error(w, "Failed to create query", http.StatusInternalServerError)
			return
		}

		formattedQuery, err := formatMultiQuery(doc.Bytes())
		if err != nil {
			log.Debug("Failed to format query for elasticsearch", log.Data{"Error": err.Error()})
			http.Error(w, "Failed to create query", http.StatusInternalServerError)
			return
		}

		responseString, err := elasticSearchClient.Search("", "", formattedQuery)
		if err != nil {
			log.Debug("Failed to query elasticsearch", log.Data{"Error": err.Error()})
			http.Error(w, "Failed to run data query", http.StatusInternalServerError)
			return
		}

		responseData := dataLookupResponse{Responses: make([]interface{}, 1)}
		if err := json.Unmarshal([]byte(responseString), &responseData.Responses[0]); err != nil {
			log.Debug("Failed to unmarshal response from elasticsearch", log.Data{"Error": err.Error()})
			http.Error(w, "Failed to process data query", http.StatusInternalServerError)
			return
		}

		dataWithResponse, err := json.Marshal(responseData)
		if err != nil {
			log.Debug("Failed to marshal response data", log.Data{"Error": err.Error()})
			http.Error(w, "Failed to encode data query response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		w.Write([]byte(dataWithResponse))
	}
}
