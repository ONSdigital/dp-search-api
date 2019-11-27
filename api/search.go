package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"text/template"
	"time"

	"github.com/ONSdigital/log.go/log"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/js"
)

type searchRequest struct {
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

var searchTemplates *template.Template

// SetupSearch loads templates for use by the search handler and should be done only once
func SetupSearch() error {
	//Load the templates once, the main entry point for the templates is search.tmpl. The search.tmpl takes
	//the SearchRequest struct and uses the Request to build up the multi-query queries that is used to query elastic.
	templates, err := template.ParseFiles(
		"templates/search/search.tmpl",
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
		"templates/search/sortByFirstLetter.tmpl",
	)

	searchTemplates = templates
	return err
}

func (sr searchRequest) HasQuery(query string) bool {
	for _, q := range sr.Queries {
		if q == query {
			return true
		}
	}
	return false
}

func formatMultiQuery(rawQuery []byte) ([]byte, error) {
	//Is minify thread Safe? can I put this as a global?
	m := minify.New()
	m.AddFuncRegexp(regexp.MustCompile("[/+]js$"), js.Minify)

	linearQuery, err := m.Bytes("application/js", rawQuery)

	if err != nil {
		return nil, err
	}

	//Put new lines in for ElasticSearch to determine the headers and the queries are detected
	return bytes.Replace(linearQuery, []byte("$$"), []byte("\n"), -1), nil

}

func paramGet(params url.Values, key, defaultValue string) string {
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

// SearchHandlerFunc returns a http handler function handling search api requests.
func SearchHandlerFunc(elasticSearchClient ElasticSearcher) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		params := req.URL.Query()
		sizeParam := paramGet(params, "size", "10")
		size, err := strconv.Atoi(sizeParam)
		if err != nil {
			log.Event(ctx, "search paramater 'size' provided with non numeric characters",
				log.Data{"Size": sizeParam}, log.INFO)
			http.Error(w, "Invalid size paramater", http.StatusBadRequest)
			return
		}
		fromParam := paramGet(params, "from", "0")
		from, err := strconv.Atoi(fromParam)
		if err != nil {
			log.Event(ctx, "search paramater 'from' provided with non numeric characters",
				log.Data{"From": fromParam}, log.INFO)
			http.Error(w, "Invalid from paramater", http.StatusBadRequest)
			return
		}

		var queries []string

		if nil == params["query"] {
			queries = []string{"search"}
		} else {
			queries = params["query"]
		}

		reqParams := searchRequest{
			Term:                params.Get("term"),
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
		log.Event(ctx, "search handler called", log.Data{"queries": queries, "request": reqParams}, log.INFO)
		var doc bytes.Buffer

		err = searchTemplates.Execute(&doc, reqParams)

		if err != nil {
			log.Event(ctx, "creation of search from template failed", log.Data{"Params": reqParams}, log.Error(err), log.ERROR)
			http.Error(w, "Failed to create query", http.StatusInternalServerError)
			return
		}

		//Put new lines in for ElasticSearch to determine the headers and the queries are detected
		formattedQuery, err := formatMultiQuery(doc.Bytes())
		if err != nil {
			log.Event(ctx, "formating of query for elasticsearch failed", log.Error(err), log.ERROR)
			http.Error(w, "Failed to create query", http.StatusInternalServerError)
			return
		}

		responseData, err := elasticSearchClient.MultiSearch(ctx, "ons", "", formattedQuery)
		if err != nil {
			log.Event(ctx, "elasticsearch query failed", log.Error(err), log.ERROR)
			http.Error(w, "Failed to run search query", http.StatusInternalServerError)
			return
		}

		if !json.Valid([]byte(responseData)) {
			log.Event(ctx, "elastic search returned invalid JSON for search query", log.ERROR)
			http.Error(w, "Failed to process search query", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		w.Write(responseData)
	}
}
