package query

import (
	"bytes"
	"context"
	"embed"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

//go:embed templates/search/*.tmpl
//go:embed templates/search/v710/*.tmpl
var searchFS embed.FS

type searchRequest struct {
	Term             string
	From             int
	Size             int
	Types            []string
	Index            string
	SortBy           string
	AggregationField string
	Highlight        bool
	URIPrefix        string
	Topic            []string
	TopicWildcard    []string
	Now              string
}

// SetupSearch loads templates for use by the search handler and should be done only once
func SetupSearch() (*template.Template, error) {
	// Load the templates once, the main entry point for the templates is search.tmpl. The search.tmpl takes
	// the SearchRequest struct and uses the Request to build up the multi-query queries that is used to query elastic.

	templates, err := template.ParseFS(searchFS,
		"templates/search/search.tmpl",
		"templates/search/contentQuery.tmpl",
		"templates/search/matchAll.tmpl",
		"templates/search/contentHeader.tmpl",
		"templates/search/countHeader.tmpl",
		"templates/search/countQuery.tmpl",
		"templates/search/coreQuery.tmpl",
		"templates/search/weightedQuery.tmpl",
		"templates/search/contentFilters.tmpl",
		"templates/search/contentFilterOnURIPrefix.tmpl",
		"templates/search/contentFilterOnTopic.tmpl",
		"templates/search/contentFilterOnTopicWildcard.tmpl",
		"templates/search/sortByTitle.tmpl",
		"templates/search/sortByRelevance.tmpl",
		"templates/search/sortByReleaseDate.tmpl",
		"templates/search/sortByReleaseDateAsc.tmpl",
		"templates/search/sortByReleaseDateAsc.tmpl",
		"templates/search/sortByFirstLetter.tmpl",
	)

	return templates, err
}

// SetupV710Search loads v710 templates for use by the search handler and should be done only once
func SetupV710Search() (*template.Template, error) {
	// Load the templates once, the main entry point for the templates is search.tmpl. The search.tmpl takes
	// the SearchRequest struct and uses the Request to build up the multi-query queries that is used to query elastic.

	templates, err := template.ParseFS(searchFS,
		"templates/search/v710/search.tmpl",
		"templates/search/v710/contentQuery.tmpl",
		"templates/search/v710/matchAll.tmpl",
		"templates/search/v710/contentHeader.tmpl",
		"templates/search/v710/countHeader.tmpl",
		"templates/search/v710/countQuery.tmpl",
		"templates/search/v710/coreQuery.tmpl",
		"templates/search/v710/weightedQuery.tmpl",
		"templates/search/v710/contentFilters.tmpl",
		"templates/search/v710/contentFilterOnURIPrefix.tmpl",
		"templates/search/v710/contentFilterOnTopic.tmpl",
		"templates/search/v710/contentFilterOnTopicWildcard.tmpl",
		"templates/search/v710/sortByTitle.tmpl",
		"templates/search/v710/sortByRelevance.tmpl",
		"templates/search/v710/sortByReleaseDate.tmpl",
		"templates/search/v710/sortByReleaseDateAsc.tmpl",
		"templates/search/v710/sortByFirstLetter.tmpl",
	)

	return templates, err
}

// BuildSearchQuery creates an elastic search query from the provided search parameters
func (sb *Builder) BuildSearchQuery(ctx context.Context, q, contentTypes, sort string, topics []string, limit, offset int, esVersion710 bool) ([]byte, error) {
	reqParams := searchRequest{
		Term:  q,
		From:  offset,
		Size:  limit,
		Types: strings.Split(contentTypes, ","),
		//Topic:            topics, // Todo: This needs to be reintroduced when migrating to ES 7.10
		SortBy:           sort,
		AggregationField: "type",
		Highlight:        true,
		Now:              time.Now().UTC().Format(time.RFC3339),
	}

	var doc bytes.Buffer

	err := sb.searchTemplates.Execute(&doc, reqParams)
	if err != nil {
		return nil, errors.Wrap(err, "creation of search from template failed")
	}

	var formattedQuery []byte
	// Put new lines in for ElasticSearch to determine the headers and the queries are detected
	if esVersion710 {
		formattedQuery, err = FormatMultiQuery(doc.Bytes())
	} else {
		formattedQuery, err = LegacyFormatMultiQuery(doc.Bytes())
	}
	if err != nil {
		return nil, errors.Wrap(err, "formating of query for elasticsearch failed")
	}

	return formattedQuery, nil
}
