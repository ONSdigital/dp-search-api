package query

import (
	"bytes"
	"context"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
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

// SetupSearch loads templates for use by the search handler and should be done only once
func SetupSearch() (*template.Template, error) {
	//Load the templates once, the main entry point for the templates is search.tmpl. The search.tmpl takes
	//the SearchRequest struct and uses the Request to build up the multi-query queries that is used to query elastic.
	templates, err := template.ParseFiles(
		"../templates/search/search.tmpl",
		"../templates/search/contentQuery.tmpl",
		"../templates/search/matchAll.tmpl",
		"../templates/search/contentHeader.tmpl",
		"../templates/search/featuredHeader.tmpl",
		"../templates/search/featuredQuery.tmpl",
		"../templates/search/countHeader.tmpl",
		"../templates/search/countQuery.tmpl",
		"../templates/search/departmentsHeader.tmpl",
		"../templates/search/departmentsQuery.tmpl",
		"../templates/search/coreQuery.tmpl",
		"../templates/search/weightedQuery.tmpl",
		"../templates/search/countFilterLatest.tmpl",
		"../templates/search/contentFilters.tmpl",
		"../templates/search/contentFilterUpcoming.tmpl",
		"../templates/search/contentFilterPublished.tmpl",
		"../templates/search/contentFilterOnLatest.tmpl",
		"../templates/search/contentFilterOnFirstLetter.tmpl",
		"../templates/search/contentFilterOnReleaseDate.tmpl",
		"../templates/search/contentFilterOnUriPrefix.tmpl",
		"../templates/search/contentFilterOnTopic.tmpl",
		"../templates/search/contentFilterOnTopicWildcard.tmpl",
		"../templates/search/sortByTitle.tmpl",
		"../templates/search/sortByRelevance.tmpl",
		"../templates/search/sortByReleaseDate.tmpl",
		"../templates/search/sortByReleaseDateAsc.tmpl",
		"../templates/search/sortByReleaseDateAsc.tmpl",
		"../templates/search/sortByFirstLetter.tmpl",
	)

	return templates, err
}

// BuildSearchQuery creates an elastic search query from the provided search parameters
func (sb *Builder) BuildSearchQuery(ctx context.Context, q, contentTypes, sort string, limit, offset int) ([]byte, error) {

	reqParams := searchRequest{
		Term:             q,
		From:             offset,
		Size:             limit,
		Types:            strings.Split(contentTypes, ","),
		SortBy:           sort,
		Queries:          []string{"search", "counts"},
		AggregationField: "_type",
		Highlight:        true,
		FilterOnLatest:   false,
		Upcoming:         false,
		Published:        false,
		Now:              time.Now().UTC().Format(time.RFC3339),
	}

	var doc bytes.Buffer

	err := sb.searchTemplates.Execute(&doc, reqParams)
	if err != nil {
		return nil, errors.Wrap(err, "creation of search from template failed")
	}

	//Put new lines in for ElasticSearch to determine the headers and the queries are detected
	formattedQuery, err := FormatMultiQuery(doc.Bytes())
	if err != nil {
		return nil, errors.Wrap(err, "formating of query for elasticsearch failed")
	}

	return formattedQuery, nil
}

// HasQuery is a helper method used by certain templates
func (sr searchRequest) HasQuery(query string) bool {
	for _, q := range sr.Queries {
		if q == query {
			return true
		}
	}
	return false
}
