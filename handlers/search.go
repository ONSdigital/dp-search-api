package handlers

import (
	"net/http"
	"github.com/ONSdigital/go-ns/log"
	"text/template"
	"bytes"
	"strconv"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify"
	"regexp"
	elasticsearch "github.com/ONSdigital/dp-search-query/elasticsearch"
)

type SearchRequest struct {
	Term    string
	From    int
	Size    int
	Types   []string
	Index   string
	Queries []string
	SortBy  string
}

//Load the templates once, the main entry point for the templates is search.tmpl. The search.tmpl takes the SearchRequest struct and
//uses the Request to build up the multi-query queries that is used to query elastic.
var searchTemplates, _ = template.ParseFiles("templates/search/search.tmpl",
	"templates/search/contentQuery.tmpl",
	"templates/search/matchAll.tmpl",
	"templates/search/contentHeader.tmpl",
	"templates/search/featuredHeader.tmpl",
	"templates/search/featuredQuery.tmpl",
	"templates/search/countHeader.tmpl",
	"templates/search/countQuery.tmpl",
	"templates/search/departmentsHeader.tmpl",
	"templates/search/departmentsQuery.tmpl",
	"templates/search/coreQuery.tmpl",
	"templates/search/sortByTitle.tmpl",
	"templates/search/sortByRelevance.tmpl",
	"templates/search/sortByReleaseDate.tmpl",
	"templates/search/sortByReleaseDateAsc.tmpl",
	"templates/search/sortByFirstLetter.tmpl")

func (sr SearchRequest) HasQuery(query string) bool {
	for _, q := range sr.Queries {
		if q == query {
			return true
		}
	}
	return false
}


func formatMultiQuery(rawQuery []byte) []byte {
	//Is minify thread Safe? can I put this as a global?
	m := minify.New()
	m.AddFuncRegexp(regexp.MustCompile("[/+]js$"), js.Minify)


	linearQuery, err := m.Bytes("application/js",rawQuery)

	if err != nil {
		panic(err)
	}

	//Put new lines in for ElasticSearch to determine the headers and the queries are detected
	return bytes.Replace(linearQuery, []byte("$$"), []byte("\n"), -1)

}

func SearchHandler(w http.ResponseWriter, req *http.Request) {

	params := req.URL.Query()
	size, err := strconv.Atoi(params.Get("size"))
	from, err := strconv.Atoi(params.Get("from"))
	var queries []string

	if nil == params["query"] {
		queries = []string{"search"}
	} else {
		queries = params["query"]
	}
	log.Debug("SearchHandler", log.Data{"queries":queries})

	reqParams := SearchRequest{Term: params.Get("term"),
		From: from,
		Size: size,
		Types: params["type"],
		Index: params.Get("index"),
		SortBy: params.Get("sort"),
		Queries: queries}

	var doc bytes.Buffer
	err = searchTemplates.Execute(&doc, reqParams)

	if err != nil {
		panic(err)
	}

	//Put new lines in for ElasticSearch to determine the headers and the queries are detected
	formattedQuery := formatMultiQuery(doc.Bytes())
	responseData, err := elasticsearch.MultiSearch("ons", "", formattedQuery)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Write(responseData)
}
