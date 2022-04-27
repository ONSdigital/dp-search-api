package query

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/ONSdigital/log.go/v2/log"
)

//go:embed templates/releasecalendar/*.tmpl
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

var sortNames = map[Sort]string{RelDateAsc: "release_date_asc", RelDateDesc: "release_date_desc", TitleAsc: "title_asc", TitleDesc: "title_desc", Relevance: "relevance", Invalid: "invalid"}
var esSortNames = map[Sort]string{RelDateAsc: `{"description.releaseDate": "asc"}`, RelDateDesc: `{"description.releaseDate": "desc"}`, TitleAsc: `{"description.title.title_raw": "asc"}`, TitleDesc: `{"description.title.title_raw": "desc"}`, Relevance: `{"_score": "desc"}`, Invalid: "invalid"}

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

func MustParseSort(sort string) Sort {
	s, err := ParseSort(sort)
	if err != nil {
		log.Fatal(context.Background(), "MustParseSort", InvalidSortString(sort))
	}

	return s
}

func (s Sort) String() string {
	return sortNames[s]
}

func (s Sort) ESString() string {
	return esSortNames[s]
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
	searchTemplate, err := template.ParseFS(releaseFS,
		"templates/releasecalendar/search.tmpl",
		"templates/releasecalendar/query.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to load search template: %w", err)
	}
	return &ReleaseBuilder{
		searchTemplates: searchTemplate,
	}, nil
}

// BuildSearchQuery builds an elastic search query from the provided search parameters for Release Calendars
func (sb *ReleaseBuilder) BuildSearchQuery(_ context.Context, sr ReleaseSearchRequest) ([]byte, error) {
	var doc bytes.Buffer
	err := sb.searchTemplates.Execute(&doc, sr)
	if err != nil {
		return nil, fmt.Errorf("creation of search from template failed: %w", err)
	}

	formattedQuery, err := FormatMultiQuery(doc.Bytes())
	if err != nil {
		return nil, fmt.Errorf("formating of query for elasticsearch failed: %w", err)
	}

	return formattedQuery, nil
}

type ReleaseSearchRequest struct {
	Term           string
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
			`{"term": {"description.published": true}}`, `{"term": {"description.cancelled": false}}`)
	default:
		return `{"term": {"description.cancelled": true}}`
	}
}

func (sr ReleaseSearchRequest) CensusClause() string {
	if sr.Census {
		return `{"term": {"census":  true}}`
	}

	return `{}`
}

func (sr ReleaseSearchRequest) HighlightClause() string {
	if sr.Census {
		return `
			"highlight":{
				"pre_tags":["<em class=\"ons-highlight\">"],
				"post_tags":["</em>"],
				"fields":{
					"description.title":{"fragment_size":0,"number_of_fragments":0},
					"description.summary":{"fragment_size":0,"number_of_fragments":0},
					"description.keywords":{"fragment_size":0,"number_of_fragments":0}
				}
			}
`
	}

	return `"highlight":{}`
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
