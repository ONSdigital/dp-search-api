package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"text/template"

	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
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
			log.Error(ctx, "creation of timeseries query from template failed", err, log.Data{"Params": reqParams})
			http.Error(w, "Failed to create query", http.StatusInternalServerError)
			return
		}

		responseData, err := elasticSearchClient.Search(ctx, "ons", "timeseries", doc.Bytes())
		if err != nil {
			log.Error(ctx, "elasticsearch query failed", err)
			http.Error(w, "Failed to run timeseries query", http.StatusInternalServerError)
			return
		}

		if !json.Valid(responseData) {
			log.Error(ctx, "elastic search returned invalid JSON for timeseries query", errors.New("elastic search returned invalid JSON for timeseries query"))
			http.Error(w, "Failed to process timeseries query", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		if _, err := w.Write(responseData); err != nil {
			log.Error(ctx, "error occured while writing response data", err)
		}
	}
}
