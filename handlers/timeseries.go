package handlers

import (
	"net/http"
	"text/template"
	"bytes"
	"strings"
	elasticsearch "github.com/ONSdigital/dp-search-query/elasticsearch"
)

type TimeseriesLookupRequest struct {
	Cdid string
}

func TimeseriesLookupHandler(w http.ResponseWriter, req *http.Request) {

	params := req.URL.Query()
	reqParams := TimeseriesLookupRequest{Cdid: strings.ToLower(params.Get(":cdid"))}

	tmpl, err := template.ParseFiles("templates/timeseries/lookup.tmpl")

	if err != nil {
		panic(err)
	}
	var doc bytes.Buffer
	err = tmpl.Execute(&doc, reqParams)

	if err != nil {
		panic(err)
	}

	responseData,err := elasticsearch.Search("ons", "timeseries", doc.Bytes())

	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Write(responseData)
}
