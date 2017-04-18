package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	elasticsearch "github.com/ONSdigital/dp-search-query/elasticsearch"
)

type DataParams struct {
	Uris  []string
	Types []string
}

var dataTemplates, _ = template.ParseFiles("templates/data/queryByUri.tmpl")

func DataLookupHandler(w http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	dataParams := DataParams{
		Uris:  params["uris"],
		Types: params["types"],
	}
	var doc bytes.Buffer

	err := dataTemplates.Execute(&doc, dataParams)
	if err != nil {
		fmt.Errorf("Error %s", err.Error())
	}
	formattedQuery := formatMultiQuery(doc.Bytes())
	responseData, _ := elasticsearch.Search("", "", formattedQuery)
	dataWithResponse := "{\"responses\":[" + string(responseData) + "]}"
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Write([]byte(dataWithResponse))
}
