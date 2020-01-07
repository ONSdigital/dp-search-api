package query

import (
	"bytes"
	"regexp"
	"text/template"

	"github.com/pkg/errors"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/js"
)

// Builder represents an instance of a query builder
type Builder struct {
	searchTemplates *template.Template
}

// NewQueryBuilder loads the elastic search templates and returns a query builder instance
func NewQueryBuilder() (*Builder, error) {
	searchTemplates, err := SetupSearch()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load search templates")
	}
	return &Builder{
		searchTemplates: searchTemplates,
	}, nil
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
