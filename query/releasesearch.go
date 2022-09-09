package query

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/js"

	esClient "github.com/ONSdigital/dp-elasticsearch/v3/client"
	"github.com/ONSdigital/log.go/v2/log"
)

//go:embed templates/releasecalendar/*.tmpl
//go:embed templates/search/v710/*.tmpl
var releaseFS embed.FS

type ParamValidator map[paramName]validator

func (qpv ParamValidator) Validate(_ context.Context, name, value string) (interface{}, error) {
	if v, ok := qpv[paramName(name)]; ok {
		return v(value)
	}

	return nil, fmt.Errorf("cannot validate: no validator for %s", name)
}

type validator func(param string) (interface{}, error)
type paramName string

func NewReleaseQueryParamValidator() ParamValidator {
	return ParamValidator{
		"limit": func(param string) (interface{}, error) {
			value, err := strconv.Atoi(param)
			if err != nil {
				return 0, errors.New("limit search parameter provided with non numeric characters")
			}
			if value < 0 {
				return 0, errors.New("limit search parameter provided with negative value")
			}
			if value > 1000 {
				return 0, errors.New("limit search parameter provided with a value that is too high")
			}

			return value, nil
		},
		"offset": func(param string) (interface{}, error) {
			value, err := strconv.Atoi(param)
			if err != nil {
				return 0, errors.New("offset search parameter provided with non numeric characters")
			}
			if value < 0 {
				return 0, errors.New("offset search parameter provided with negative value")
			}
			return value, nil
		},
		"date": func(param string) (interface{}, error) {
			value, err := ParseDate(param)
			if err != nil {
				return nil, fmt.Errorf("date search parameter provided is invalid: %w", err)
			}
			return value, nil
		},
		"sort": func(param string) (interface{}, error) {
			value, err := ParseSort(param)
			if err != nil {
				return nil, fmt.Errorf("sort search parameter provided is invalid: %w", err)
			}
			return value, nil
		},
		"release-type": func(param string) (interface{}, error) {
			value, err := ParseReleaseType(param)
			if err != nil {
				return nil, fmt.Errorf("release-type parameter provided is invalid: %w", err)
			}
			return value, nil
		},
	}
}

type Date time.Time

const dateFormat = "2006-01-02"

type InvalidDateString struct {
	value, err string
}

func (ids InvalidDateString) Error() string {
	return fmt.Sprintf("invalid date string (%q): %s", ids.value, ids.err)
}

func ParseDate(date string) (Date, error) {
	if date == "" {
		return Date{}, nil
	}
	d, err := time.Parse(dateFormat, date)
	if err != nil {
		return Date{}, InvalidDateString{date, err.Error()}
	}

	if d.Before(time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return Date{}, InvalidDateString{value: date, err: "date too far in past"}
	}

	if d.After(time.Date(2200, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return Date{}, InvalidDateString{value: date, err: "date too far in future"}
	}

	return Date(d), nil
}

func MustParseDate(date string) Date {
	d, err := ParseDate(date)
	if err != nil {
		log.Fatal(context.Background(), "MustParseDate", InvalidDateString{value: date})
	}

	return d
}

func (d Date) Set() bool {
	return !time.Time(d).IsZero()
}

func (d Date) String() string {
	return time.Time(d).UTC().Format(dateFormat)
}

func (d Date) ESString() string {
	if time.Time(d).IsZero() {
		return "null"
	}
	return fmt.Sprintf("%q", d.String())
}

type Sort int

const (
	Invalid Sort = iota
	RelDateAsc
	RelDateDesc
	TitleAsc
	TitleDesc
	Relevance
)

var (
	sortNames         = map[Sort]string{RelDateAsc: "release_date_asc", RelDateDesc: "release_date_desc", TitleAsc: "title_asc", TitleDesc: "title_desc", Relevance: "relevance", Invalid: "invalid"}
	esSortNames       = map[Sort]string{RelDateAsc: `{"release_date": "asc"}`, RelDateDesc: `{"release_date": "desc"}`, TitleAsc: `{"title.title_raw": "asc"}`, TitleDesc: `{"title.title_raw": "desc"}`, Relevance: `{"_score": "desc"}`, Invalid: "invalid"}
	legacyESSortNames = map[Sort]string{RelDateAsc: `{"description.releaseDate": "asc"}`, RelDateDesc: `{"description.releaseDate": "desc"}`, TitleAsc: `{"description.title.title_raw": "asc"}`, TitleDesc: `{"description.title.title_raw": "desc"}`, Relevance: `{"_score": "desc"}`, Invalid: "invalid"}
)

type InvalidSortString string

func (iss InvalidSortString) Error() string {
	return fmt.Sprintf("invalid sort string: %q", string(iss))
}

func ParseSort(sort string) (Sort, error) {
	for s, sn := range sortNames {
		if strings.EqualFold(sort, sn) {
			return s, nil
		}
	}

	return Invalid, InvalidSortString(sort)
}

func (s Sort) String() string {
	return sortNames[s]
}

func (s Sort) ESString() string {
	return esSortNames[s]
}

func (s Sort) LegacyESString() string {
	return legacyESSortNames[s]
}

type ReleaseType int

const (
	InvalidReleaseType ReleaseType = iota
	Upcoming
	Published
	Cancelled
)

var relTypeNames = map[ReleaseType]string{Upcoming: "type-upcoming", Published: "type-published", Cancelled: "type-cancelled", InvalidReleaseType: "Invalid"}

type InvalidReleaseTypeString string

func (irts InvalidReleaseTypeString) Error() string {
	return fmt.Sprintf("invalid ReleaseType string: %q", string(irts))
}

func ParseReleaseType(s string) (ReleaseType, error) {
	for rt, rtn := range relTypeNames {
		if strings.EqualFold(s, rtn) {
			return rt, nil
		}
	}

	return InvalidReleaseType, InvalidReleaseTypeString(s)
}

func MustParseReleaseType(s string) ReleaseType {
	rt, err := ParseReleaseType(s)
	if err != nil {
		log.Fatal(context.Background(), "MustParseReleaseType", InvalidReleaseTypeString(s))
	}

	return rt
}

func (rt ReleaseType) String() string {
	return relTypeNames[rt]
}

type ReleaseBuilder struct {
	searchTemplates *template.Template
}

func NewReleaseBuilder() (*ReleaseBuilder, error) {
	var (
		searchTemplate *template.Template
		err            error
	)

	searchTemplate, err = template.ParseFS(releaseFS,
		"templates/releasecalendar/search.tmpl",
		"templates/releasecalendar/query.tmpl",
		"templates/releasecalendar/simplequery.tmpl",
		"templates/search/v710/coreQuery.tmpl")

	if err != nil {
		return nil, fmt.Errorf("failed to load search template: %w", err)
	}

	return &ReleaseBuilder{
		searchTemplates: searchTemplate,
	}, nil
}

// BuildSearchQuery builds an elastic search query from the provided search parameters for Release Calendars
func (rb *ReleaseBuilder) BuildSearchQuery(_ context.Context, searchRequest interface{}) ([]esClient.Search, error) {
	var doc bytes.Buffer
	err := rb.searchTemplates.Execute(&doc, searchRequest)
	if err != nil {
		return nil, fmt.Errorf("creation of search from template failed: %w", err)
	}

	m := minify.New()
	m.AddFuncRegexp(regexp.MustCompile("[/+]js$"), js.Minify)
	linearQuery, err := m.Bytes("application/js", doc.Bytes())
	if err != nil {
		return nil, err
	}

	bytes.Replace(linearQuery, []byte("$$"), []byte("\n"), -1)
	lines := bytes.Split(linearQuery, []byte("\n"))
	var searches []esClient.Search
	for i := 0; i < len(lines)-1; i += 2 {
		searches = append(searches, esClient.Search{
			Header: esClient.Header{Index: string(lines[i])},
			Query:  lines[i+1],
		})
	}

	return searches, nil
}

type ReleaseSearchRequest struct {
	Term           string
	Template       string
	From           int
	Size           int
	SortBy         Sort
	ReleasedAfter  Date
	ReleasedBefore Date
	Type           ReleaseType
	Provisional    bool
	Confirmed      bool
	Postponed      bool
	Census         bool
	Highlight      bool
}

const (
	simple         = "!!s:"
	simpleExtended = "!!se:"
	sitewide       = "!!sw:"
	standard       = "!!st"
)

var templateNames = map[string]string{simple: "s", simpleExtended: "ss", sitewide: "sw", standard: "st"}

// ParseQuery :
//
// escapes double quotes (") in the given query q, so that ElasticSearch will accept them
//
// looks for a prefix in the given query q (which can be used to determine the type of query passed to
// ElasticSearch)
//
// returns the escaped query with the prefix removed (if any was prefixed), together
// with the name of the template to use to generate the ElasticSearch query
func ParseQuery(q string) (s1, s2 string) {
	// The following looks horrific but is probably the easiest and most efficient way to escape quotes(") in
	// the query string (regex in golang doesn't allow negative look-behind)
	qb := []byte(strconv.Quote(q))
	q = string(qb[1 : len(qb)-1])

	for ts, tn := range templateNames {
		if strings.HasPrefix(q, ts) {
			return strings.TrimPrefix(q, ts), tn
		}
	}

	return q, templateNames[standard]
}

func (sr *ReleaseSearchRequest) String() string {
	s, err := json.MarshalIndent(sr, "", "  ")
	if err != nil {
		panic("couldn't marshal the searchRequest: " + err.Error())
	}

	return string(s)
}

func (sr ReleaseSearchRequest) Now() string {
	return fmt.Sprintf("%q", time.Now().Format(dateFormat))
}

func (sr ReleaseSearchRequest) SortClause() string {
	if sr.SortBy == Relevance {
		switch sr.Type {
		case Upcoming:
			return fmt.Sprintf("%s, %s", esSortNames[Relevance], esSortNames[RelDateAsc])
		case Published:
			return fmt.Sprintf("%s, %s", esSortNames[Relevance], esSortNames[RelDateDesc])
		case Cancelled:
			return esSortNames[Relevance]
		}
	}

	return sr.SortBy.ESString()
}

// ReleaseTypeClause returns the query clause to select the type of release
// Note that it is possible for a Release to have both its Published and Cancelled flags true (yes indeed!)
// In this case it is deemed cancelled
func (sr ReleaseSearchRequest) ReleaseTypeClause() string {
	switch sr.Type {
	case Upcoming:
		var buf bytes.Buffer
		buf.WriteString(mainUpcomingClause(time.Now()))
		if secondary := supplementaryUpcomingClause(sr); secondary != "" {
			buf.WriteString(Separator + secondary)
		}
		return buf.String()
	case Published:
		return fmt.Sprintf("%s, %s",
			`{"term": {"published": true}}`, `{"term": {"cancelled": false}}`)
	default:
		return `{"term": {"cancelled": true}}`
	}
}

func (sr ReleaseSearchRequest) CensusClause() string {
	if sr.Census {
		return `{"term": {"survey":  "census"}}`
	}

	return EmptyClause
}

func (sr ReleaseSearchRequest) HighlightClause() string {
	if sr.Highlight {
		return `
			"highlight":{
				"pre_tags":["<em class=\"ons-highlight\">"],
				"post_tags":["</em>"],
				"fields":{
					"title":{"fragment_size":0,"number_of_fragments":0},
					"summary":{"fragment_size":0,"number_of_fragments":0},
					"keywords":{"fragment_size":0,"number_of_fragments":0}
				}
			}
`
	}

	return fmt.Sprintf("%q:%s", "highlight", EmptyClause)
}

func (sr *ReleaseSearchRequest) Set(value string) error {
	var sr2 ReleaseSearchRequest
	err := json.Unmarshal([]byte(value), &sr2)
	if err != nil {
		return err
	}

	*sr = sr2
	return nil
}
