package query

import (
	"bytes"
	"encoding/json"
	"regexp"
	"text/template"

	"github.com/ONSdigital/dp-elasticsearch/v4/client"
	"github.com/pkg/errors"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/js"
)

// Builder represents an instance of a query builder
type Builder struct {
	searchTemplates *template.Template
}

// NewQueryBuilder loads the elastic search templates and returns a query builder instance
func NewQueryBuilder(esVersion710 bool) (*Builder, error) {
	var searchTemplates *template.Template
	var err error
	if esVersion710 {
		searchTemplates, err = SetupV710Search()
	} else {
		searchTemplates, err = SetupSearch()
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to load search templates")
	}
	return &Builder{
		searchTemplates: searchTemplates,
	}, nil
}

// FormatMultiQuery minifies and reformats an elasticsearch MultiQuery
func FormatMultiQuery(rawQuery []byte) ([]byte, error) {
	// Is minify thread Safe? can I put this as a global?
	m := minify.New()
	m.AddFuncRegexp(regexp.MustCompile("[/+]js$"), js.Minify)

	linearQuery, err := m.Bytes("application/js", rawQuery)
	if err != nil {
		return nil, err
	}
	newBytes := bytes.Split(linearQuery, []byte("$$"))
	var searches []client.Search
	for i := 0; i < len(newBytes)-1; i += 2 {
		var header client.Header
		byteHeader := newBytes[i]
		query := newBytes[i+1]
		if marshalErr := json.Unmarshal(byteHeader, &header); marshalErr != nil {
			return nil, marshalErr
		}
		searches = append(searches, client.Search{
			Header: header,
			Query:  query,
		})
	}
	searchBytes, err := json.Marshal(searches)
	if err != nil {
		return nil, err
	}
	// Put new lines in for ElasticSearch to determine the headers and the queries are detected
	return searchBytes, nil
}

// LegacyFormatMultiQuery minifies and reformats an elasticsearch MultiQuery
func LegacyFormatMultiQuery(rawQuery []byte) ([]byte, error) {
	// Is minify thread Safe? can I put this as a global?
	m := minify.New()
	m.AddFuncRegexp(regexp.MustCompile("[/+]js$"), js.Minify)

	linearQuery, err := m.Bytes("application/js", rawQuery)
	if err != nil {
		return nil, err
	}

	// Put new lines in for ElasticSearch to determine the headers and the queries are detected
	return bytes.Replace(linearQuery, []byte("$$"), []byte("\n"), -1), nil
}
