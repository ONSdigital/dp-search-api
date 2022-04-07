package query

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"
)

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

func ParseDate(date string) (Date, error) {
	if date == "" {
		return Date{}, nil
	}
	d, err := time.Parse(dateFormat, date)
	if err != nil {
		return Date{}, err
	}

	if d.Before(time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return Date{}, errors.New("date too far in past")
	}

	if d.After(time.Date(2200, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return Date{}, errors.New("date too far in future")
	}

	return Date(d), nil
}

func MustParseDate(date string) Date {
	d, err := ParseDate(date)
	if err != nil {
		panic("invalid date string: " + date)
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
)

var sortNames = map[Sort]string{RelDateAsc: "release_date_asc", RelDateDesc: "release_date_desc", TitleAsc: "title_asc", TitleDesc: "title_desc", Invalid: "invalid"}
var esSortNames = map[Sort]string{RelDateAsc: `{"description.releaseDate": "asc"}`, RelDateDesc: `{"description.releaseDate": "desc"}`, TitleAsc: `{"description.title": "asc"}`, TitleDesc: `{"description.title": "desc"}`, Invalid: "invalid"}

func ParseSort(sort string) (Sort, error) {
	for s, sn := range sortNames {
		if strings.EqualFold(sort, sn) {
			return s, nil
		}
	}

	return Invalid, errors.New("invalid sort option string")
}

func MustParseSort(sort string) Sort {
	s, err := ParseSort(sort)
	if err != nil {
		panic("invalid sort string: " + sort)
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

func ParseReleaseType(s string) (ReleaseType, error) {
	for rt, rtn := range relTypeNames {
		if strings.EqualFold(s, rtn) {
			return rt, nil
		}
	}

	return InvalidReleaseType, errors.New("invalid release type string")
}

func MustParseReleaseType(s string) ReleaseType {
	rt, err := ParseReleaseType(s)
	if err != nil {
		panic("invalid release type string: " + s)
	}

	return rt
}

func (rt ReleaseType) String() string {
	return relTypeNames[rt]
}

type ReleaseBuilder struct {
	searchTemplates *template.Template
}

func NewReleaseBuilder(pathToTemplates string) (*ReleaseBuilder, error) {
	searchTemplate, err := template.ParseFiles(
		pathToTemplates+"templates/search/releasecalendar/search.tmpl",
		pathToTemplates+"templates/search/releasecalendar/query.tmpl",
		pathToTemplates+"templates/search/releasecalendar/upcoming.tmpl",
		pathToTemplates+"templates/search/releasecalendar/published.tmpl")
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

	return doc.Bytes(), nil
}

type ReleaseSearchRequest struct {
	Term           string
	From           int
	Size           int
	SortBy         Sort
	ReleasedAfter  Date
	ReleasedBefore Date
	Upcoming       bool
	Published      bool
	Highlight      bool
	Now            Date
}

func (sr *ReleaseSearchRequest) String() string {
	s, err := json.MarshalIndent(sr, "", "  ")
	if err != nil {
		panic("couldn't marshal the searchRequest: " + err.Error())
	}

	return string(s)
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
