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
func NewQueryBuilder(pathToTemplates string) (*Builder, error) {
	searchTemplates, err := SetupSearch(pathToTemplates)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load search templates")
	}
	return &Builder{
		searchTemplates: searchTemplates,
	}, nil
}

// FormatMultiQuery minifies and reformats an elasticsearch MultiQuery
func FormatMultiQuery(rawQuery []byte) ([]byte, error) {
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
