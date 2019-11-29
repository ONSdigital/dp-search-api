package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"text/template"

	"github.com/ONSdigital/log.go/log"
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
		ctx := req.Context()

		vars := mux.Vars(req)
		reqParams := timeseriesLookupRequest{Cdid: strings.ToLower(vars["cdid"])}

		var doc bytes.Buffer
		err := timeseriesTemplate.Execute(&doc, reqParams)
		if err != nil {
			log.Event(ctx, "creation of timeseries query from template failed", log.Data{"Params": reqParams}, log.Error(err))
			http.Error(w, "Failed to create query", http.StatusInternalServerError)
			return
		}

		responseData, err := elasticSearchClient.Search(ctx, "ons", "timeseries", doc.Bytes())
		if err != nil {
			log.Event(ctx, "elasticsearch query failed", log.Error(err))
			http.Error(w, "Failed to run timeseries query", http.StatusInternalServerError)
			return
		}

		if !json.Valid([]byte(responseData)) {
			log.Event(ctx, "elastic search returned invalid JSON for timeseries query", log.ERROR)
			http.Error(w, "Failed to process timeseries query", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		w.Write(responseData)
	}
}
