package api

import (
	"bytes"
	"net/http"
	"strings"
	"text/template"

	"github.com/ONSdigital/dp-search-query/elasticsearch"
	"github.com/ONSdigital/go-ns/log"
)

type TimeseriesLookupRequest struct {
	Cdid string
}

var timeseriesTemplate *template.Template

func SetupTimeseries() error {
	templates, err := template.ParseFiles("templates/timeseries/lookup.tmpl")
	timeseriesTemplate = templates
	return err
}

func TimeseriesLookupHandler(w http.ResponseWriter, req *http.Request) {

	params := req.URL.Query()
	reqParams := TimeseriesLookupRequest{Cdid: strings.ToLower(params.Get(":cdid"))}

	var doc bytes.Buffer
	err := timeseriesTemplate.Execute(&doc, reqParams)
	if err != nil {
		log.Debug("Failed to create timeseries query from template", log.Data{"Error": err.Error(), "Params": reqParams})
		http.Error(w, "Failed to create query", http.StatusInternalServerError)
		return
	}

	responseData, err := elasticsearch.Search("ons", "timeseries", doc.Bytes())
	if err != nil {
		log.Debug("Failed to query elasticsearch", log.Data{"Error": err.Error()})
		http.Error(w, "Failed to run timeseries query", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Write(responseData)
}
