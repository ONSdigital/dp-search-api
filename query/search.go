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
	URIPrefix           string
	Topic               []string
	TopicWildcard       []string
	Upcoming            bool
	Published           bool
	Now                 string
}

// SetupSearch loads templates for use by the search handler and should be done only once
func SetupSearch(pathToTemplates string) (*template.Template, error) {
	// Load the templates once, the main entry point for the templates is search.tmpl. The search.tmpl takes
	// the SearchRequest struct and uses the Request to build up the multi-query queries that is used to query elastic.

	templates, err := template.ParseFiles(
		pathToTemplates+"templates/search/search.tmpl",
		pathToTemplates+"templates/search/contentQuery.tmpl",
		pathToTemplates+"templates/search/matchAll.tmpl",
		pathToTemplates+"templates/search/contentHeader.tmpl",
		pathToTemplates+"templates/search/featuredHeader.tmpl",
		pathToTemplates+"templates/search/featuredQuery.tmpl",
		pathToTemplates+"templates/search/countHeader.tmpl",
		pathToTemplates+"templates/search/countQuery.tmpl",
		pathToTemplates+"templates/search/departmentsHeader.tmpl",
		pathToTemplates+"templates/search/departmentsQuery.tmpl",
		pathToTemplates+"templates/search/coreQuery.tmpl",
		pathToTemplates+"templates/search/weightedQuery.tmpl",
		pathToTemplates+"templates/search/countFilterLatest.tmpl",
		pathToTemplates+"templates/search/contentFilters.tmpl",
		pathToTemplates+"templates/search/contentFilterUpcoming.tmpl",
		pathToTemplates+"templates/search/contentFilterPublished.tmpl",
		pathToTemplates+"templates/search/contentFilterOnLatest.tmpl",
		pathToTemplates+"templates/search/contentFilterOnFirstLetter.tmpl",
		pathToTemplates+"templates/search/contentFilterOnReleaseDate.tmpl",
		pathToTemplates+"templates/search/contentFilterOnURIPrefix.tmpl",
		pathToTemplates+"templates/search/contentFilterOnTopic.tmpl",
		pathToTemplates+"templates/search/contentFilterOnTopicWildcard.tmpl",
		pathToTemplates+"templates/search/sortByTitle.tmpl",
		pathToTemplates+"templates/search/sortByRelevance.tmpl",
		pathToTemplates+"templates/search/sortByReleaseDate.tmpl",
		pathToTemplates+"templates/search/sortByReleaseDateAsc.tmpl",
		pathToTemplates+"templates/search/sortByReleaseDateAsc.tmpl",
		pathToTemplates+"templates/search/sortByFirstLetter.tmpl",
	)

	return templates, err
}

// SetupV710Search loads v710 templates for use by the search handler and should be done only once
func SetupV710Search(pathToTemplates string) (*template.Template, error) {
	// Load the templates once, the main entry point for the templates is search.tmpl. The search.tmpl takes
	// the SearchRequest struct and uses the Request to build up the multi-query queries that is used to query elastic.

	templates, err := template.ParseFiles(
		pathToTemplates+"templates/search/v710/search.tmpl",
		pathToTemplates+"templates/search/v710/contentQuery.tmpl",
		pathToTemplates+"templates/search/v710/contentHeader.tmpl",
	)

	return templates, err
}

// BuildSearchQuery creates an elastic search query from the provided search parameters
func (sb *Builder) BuildSearchQuery(ctx context.Context, q, contentTypes, sort string, topics []string, limit, offset int) ([]byte, error) {
	reqParams := searchRequest{
		Term:  q,
		From:  offset,
		Size:  limit,
		Types: strings.Split(contentTypes, ","),
		//Topic:            topics, // Todo: This needs to be reintroduced when migrating to ES 7.10
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

	// Put new lines in for ElasticSearch to determine the headers and the queries are detected
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
