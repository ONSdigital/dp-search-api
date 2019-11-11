package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"text/template"

	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

type timeseriesLookupRequest struct {
	Cdid string
}

var timeseriesTemplate *template.Template

// SetupTimeseries loads templates for use by the timeseries lookup handler and should be done only once
func SetupTimeseries() error {
	templates, err := template.ParseFiles("templates/timeseries/lookup.tmpl")
	timeseriesTemplate = templates
	return err
}

// TimeseriesLookupHandlerFunc returns a http handler function handling search api requests.
func TimeseriesLookupHandlerFunc(elasticSearchClient ElasticSearcher) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		vars := mux.Vars(req)
		reqParams := timeseriesLookupRequest{Cdid: strings.ToLower(vars["cdid"])}

		var doc bytes.Buffer
		err := timeseriesTemplate.Execute(&doc, reqParams)
		if err != nil {
			log.Debug("Failed to create timeseries query from template", log.Data{"Error": err.Error(), "Params": reqParams})
			http.Error(w, "Failed to create query", http.StatusInternalServerError)
			return
		}

		responseData, err := elasticSearchClient.Search("ons", "timeseries", doc.Bytes())
		if err != nil {
			log.Debug("Failed to query elasticsearch", log.Data{"Error": err.Error()})
			http.Error(w, "Failed to run timeseries query", http.StatusInternalServerError)
			return
		}

		if !json.Valid([]byte(responseData)) {
			log.Debug("Invlid JSON returned by elastic search for timeseries query", nil)
			http.Error(w, "Failed to process timeseries query", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		w.Write(responseData)
	}
}
