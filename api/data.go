package api

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/ONSdigital/dp-search-query/elasticsearch"
	"github.com/ONSdigital/go-ns/log"
)

type dataParams struct {
	Uris  []string
	Types []string
}

var dataTemplates *template.Template

func SetupData() error {
	templates, err := template.ParseFiles("templates/data/queryByUri.tmpl")
	dataTemplates = templates
	return err
}

func DataLookupHandler(w http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	templateParams := dataParams{
		Uris:  params["uris"],
		Types: params["types"],
	}
	var doc bytes.Buffer

	err := dataTemplates.Execute(&doc, templateParams)
	if err != nil {
		log.Debug("Failed to create search from template", log.Data{"Error": err.Error(), "Params": templateParams})
		http.Error(w, "Failed to create query", http.StatusInternalServerError)
		return
	}
	formattedQuery, err := formatMultiQuery(doc.Bytes())
	if err != nil {
		log.Debug("Failed to format query for elasticsearch", log.Data{"Error": err.Error()})
		http.Error(w, "Failed to create query", http.StatusInternalServerError)
		return
	}
	responseData, err := elasticsearch.Search("", "", formattedQuery)
	if err != nil {
		log.Debug("Failed to query elasticsearch", log.Data{"Error": err.Error()})
		http.Error(w, "Failed to run data query", http.StatusInternalServerError)
		return
	}
	dataWithResponse := "{\"responses\":[" + string(responseData) + "]}"
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Write([]byte(dataWithResponse))
}
