package handlers

import (
	"bytes"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"text/template"
	"time"

	elasticsearch "github.com/ONSdigital/dp-search-query/elasticsearch"
	"github.com/ONSdigital/go-ns/log"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/js"
)

type SearchRequest struct {
	Term                string
	From                int
	Size                int
	Types               []string
	Index               string
	Queries             []string
	SortBy              string
	AggregationField    string
	Highlight           bool
	FilterOnLatest      bool
	FilterOnFirstLetter string
	ReleasedAfter       string
	ReleasedBefore      string
	UriPrefix           string
	Topic               []string
	TopicWildcard       []string
	Upcoming            bool
	Published           bool
	Now                 string
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
	"templates/search/weightedQuery.tmpl",
	"templates/search/countFilterLatest.tmpl",
	"templates/search/contentFilters.tmpl",
	"templates/search/contentFilterUpcoming.tmpl",
	"templates/search/contentFilterPublished.tmpl",
	"templates/search/contentFilterOnLatest.tmpl",
	"templates/search/contentFilterOnFirstLetter.tmpl",
	"templates/search/contentFilterOnReleaseDate.tmpl",
	"templates/search/contentFilterOnUriPrefix.tmpl",
	"templates/search/contentFilterOnTopic.tmpl",
	"templates/search/contentFilterOnTopicWildcard.tmpl",
	"templates/search/sortByTitle.tmpl",
	"templates/search/sortByRelevance.tmpl",
	"templates/search/sortByReleaseDate.tmpl",
	"templates/search/sortByReleaseDateAsc.tmpl",
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

	linearQuery, err := m.Bytes("application/js", rawQuery)

	if err != nil {
		panic(err)
	}

	//Put new lines in for ElasticSearch to determine the headers and the queries are detected
	return bytes.Replace(linearQuery, []byte("$$"), []byte("\n"), -1)

}

func paramGet(params url.Values, key string, defaultValue string) string {
	value := params.Get(key)
	if len(value) < 1 {
		value = defaultValue
	}
	return value
}

func paramGetBool(params url.Values, key string, defaultValue bool) bool {
	value := params.Get(key)
	if len(value) < 1 {
		return defaultValue
	}
	return value == "true"
}

func SearchHandler(w http.ResponseWriter, req *http.Request) {

	params := req.URL.Query()
	size, err := strconv.Atoi(paramGet(params, "size", "10"))
	from, err := strconv.Atoi(paramGet(params, "from", "0"))
	var queries []string

	if nil == params["query"] {
		queries = []string{"search"}
	} else {
		queries = params["query"]
	}

	reqParams := SearchRequest{Term: params.Get("term"),
		From:                from,
		Size:                size,
		Types:               params["type"],
		Index:               params.Get("index"),
		SortBy:              paramGet(params, "sort", "relevance"),
		Queries:             queries,
		AggregationField:    paramGet(params, "aggField", "_type"),
		Highlight:           paramGetBool(params, "highlight", true),
		FilterOnLatest:      paramGetBool(params, "latest", false),
		FilterOnFirstLetter: params.Get("withFirstLetter"),
		ReleasedAfter:       params.Get("releasedAfter"),
		ReleasedBefore:      params.Get("releasedBefore"),
		UriPrefix:           params.Get("uriPrefix"),
		Topic:               params["topic"],
		TopicWildcard:       params["topicWildcard"],
		Upcoming:            paramGetBool(params, "upcoming", false),
		Published:           paramGetBool(params, "published", false),
		Now:                 time.Now().UTC().Format(time.RFC3339),
	}
	log.Debug("SearchHandler", log.Data{"queries": queries, "request": reqParams})
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
	//fmt.Printf("%s", string(responseData))
	w.Write(responseData)
}
