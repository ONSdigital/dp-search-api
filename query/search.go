package query

import (
	"bytes"
	"context"
	"embed"
	"text/template"

	"github.com/pkg/errors"
)

//go:embed templates/search/*.tmpl
//go:embed templates/search/v710/*.tmpl
var searchFS embed.FS

const (
	legacyAggregationField = "_type"
)

var es710AggregationField = &AggregationFields{
	Topics:          "topics",
	ContentTypes:    "type",
	PopulationTypes: "population_type.name",
	Dimensions:      "dimensions.name",
}

// SearchRequest holds the values provided by a request against Search API
// The values are used to build the elasticsearch query using the corresponding template/s
type SearchRequest struct {
	Term              string
	From              int
	Size              int
	Types             []string
	Index             string
	SortBy            string
	AggregationField  string // Deprecated (used only in legacy templates for aggregations)
	AggregationFields *AggregationFields
	Highlight         bool
	URIPrefix         string
	Topic             []string
	TopicWildcard     []string
	PopulationTypes   []*PopulationTypeRequest
	Dimensions        []*DimensionRequest
	Now               string
}

type PopulationTypeRequest struct {
	Name  string
	Label string
}

type DimensionRequest struct {
	Name     string
	Label    string
	RawLabel string
}

// AggregationFields are the elasticsearch keys for which the aggregations will be done
type AggregationFields struct {
	Topics          string
	ContentTypes    string
	PopulationTypes string
	Dimensions      string
}

type CountRequest struct {
	Term        string
	CountEnable bool
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
		"templates/search/v710/countContentTypeHeader.tmpl",
		"templates/search/v710/countContentTypeQuery.tmpl",
		"templates/search/v710/countContentTypeFilters.tmpl",
		"templates/search/v710/countTopicHeader.tmpl",
		"templates/search/v710/countTopicQuery.tmpl",
		"templates/search/v710/countTopicFilters.tmpl",
		"templates/search/v710/countPopulationTypeHeader.tmpl",
		"templates/search/v710/countPopulationTypeQuery.tmpl",
		"templates/search/v710/countPopulationTypeFilters.tmpl",
		"templates/search/v710/countDimensionsHeader.tmpl",
		"templates/search/v710/countDimensionsQuery.tmpl",
		"templates/search/v710/countDimensionsFilters.tmpl",
		"templates/search/v710/coreQuery.tmpl",
		"templates/search/v710/weightedQuery.tmpl",
		"templates/search/v710/contentTypeFilter.tmpl",
		"templates/search/v710/contentFilters.tmpl",
		"templates/search/v710/contentFilterOnURIPrefix.tmpl",
		"templates/search/v710/contentFilterOnTopicWildcard.tmpl",
		"templates/search/v710/sortByTitle.tmpl",
		"templates/search/v710/topicFilters.tmpl",
		"templates/search/v710/canonicalFilters.tmpl",
		"templates/search/v710/subTopicsFilters.tmpl",
		"templates/search/v710/sortByRelevance.tmpl",
		"templates/search/v710/sortByReleaseDate.tmpl",
		"templates/search/v710/sortByReleaseDateAsc.tmpl",
		"templates/search/v710/sortByFirstLetter.tmpl",
		"templates/search/v710/populationTypeFilters.tmpl",
		"templates/search/v710/dimensionsFilters.tmpl",
	)

	return templates, err
}

// SetupV710Search loads v710 templates for use by the search handler and should be done only once
func SetupV710Count() (*template.Template, error) {
	// Load the templates once, the main entry point for the templates is search.tmpl. The search.tmpl takes
	// the SearchRequest struct and uses the Request to build up the multi-query queries that is used to query elastic.

	templates, err := template.ParseFS(searchFS,
		"templates/search/v710/distinctItemCountQuery.tmpl",
		"templates/search/v710/coreQuery.tmpl",
		"templates/search/v710/matchAll.tmpl",
	)

	return templates, err
}

// BuildSearchQuery creates an elastic search query from the provided search parameters
func (sb *Builder) BuildSearchQuery(ctx context.Context, reqParams *SearchRequest, esVersion710 bool) ([]byte, error) {
	if esVersion710 {
		reqParams.AggregationFields = es710AggregationField
	} else {
		reqParams.AggregationField = legacyAggregationField
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

// BuildSearchQuery creates an elastic search query from the provided search parameters
func (sb *Builder) BuildCountQuery(ctx context.Context, reqParams *CountRequest) ([]byte, error) {
	var doc bytes.Buffer
	err := sb.countTemplates.Execute(&doc, reqParams)
	if err != nil {
		return nil, errors.Wrap(err, "creation of search from template failed")
	}
	return doc.Bytes(), nil
}
