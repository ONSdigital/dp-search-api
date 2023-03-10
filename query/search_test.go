package query

import (
	"context"
	"testing"
	"text/template"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBuildSearchQuery(t *testing.T) {
	Convey("Should return InternalError for invalid template", t, func() {
		qb := createQueryBuilderForTemplate("dummy{{.Moo}}")

		reqParams := &SearchRequest{
			Size: 2,
			From: 1,
		}
		query, err := qb.BuildSearchQuery(context.Background(), reqParams, false)

		So(err, ShouldNotBeNil)
		So(query, ShouldBeNil)
		So(err.Error(), ShouldContainSubstring, "creation of search from template failed")
	})

	Convey("Should include all search parameters in elastic search query", t, func() {
		qb := createQueryBuilderForTemplate("Term={{.Term}};" +
			"From={{.From}};" +
			"Size={{.Size}};" +
			"Types={{.Types}};" +
			"SortBy={{.SortBy}};" +
			"AggregationField={{.AggregationField}};" +
			"Highlight={{.Highlight}};" +
			"Now={{.Now}}")

		reqParams := &SearchRequest{
			Term:      "a",
			Types:     []string{"ta", "tb"},
			SortBy:    "relevance",
			Topic:     []string{"test"},
			Size:      2,
			From:      1,
			Highlight: true,
			Now:       time.Date(2023, 03, 10, 12, 15, 04, 05, time.UTC).UTC().Format(time.RFC3339),
		}
		query, err := qb.BuildSearchQuery(context.Background(), reqParams, false)

		So(err, ShouldBeNil)
		So(query, ShouldNotBeNil)
		queryString := string(query)
		So(queryString, ShouldContainSubstring, "Term=a")
		So(queryString, ShouldContainSubstring, "From=1")
		So(queryString, ShouldContainSubstring, "Size=2")
		So(queryString, ShouldContainSubstring, "Types=[ta tb]")
		So(queryString, ShouldContainSubstring, "SortBy=relevance")
		So(queryString, ShouldContainSubstring, "AggregationField=_type")
		So(queryString, ShouldContainSubstring, "Highlight=true")
		So(queryString, ShouldContainSubstring, "Now=2023-03-10T12:15:04")
	})
}

func createQueryBuilderForTemplate(rawTemplate string) *Builder {
	temp, err := template.New("search.tmpl").Parse(rawTemplate)
	So(err, ShouldBeNil)
	return &Builder{
		searchTemplates: temp,
	}
}
