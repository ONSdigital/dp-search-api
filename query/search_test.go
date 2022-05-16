package query

import (
	"context"
	"testing"
	"text/template"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBuildSearchQuery(t *testing.T) {
	Convey("Should return InternalError for invalid template", t, func() {
		qb := createQueryBuilderForTemplate("dummy{{.Moo}}")
		query, err := qb.BuildSearchQuery(context.Background(), "", "", "", nil, 2, 1, false)

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

		query, err := qb.BuildSearchQuery(context.Background(), "a", "ta,tb", "relevance", []string{"test"}, 2, 1, false)

		So(err, ShouldBeNil)
		So(query, ShouldNotBeNil)
		queryString := string(query)
		So(queryString, ShouldContainSubstring, "Term=a")
		So(queryString, ShouldContainSubstring, "From=1")
		So(queryString, ShouldContainSubstring, "Size=2")
		So(queryString, ShouldContainSubstring, "Types=[ta tb]")
		So(queryString, ShouldContainSubstring, "SortBy=relevance")
		So(queryString, ShouldContainSubstring, "AggregationField=type")
		So(queryString, ShouldContainSubstring, "Highlight=true")
		So(queryString, ShouldContainSubstring, "Now=20")
	})
}

func createQueryBuilderForTemplate(rawTemplate string) *Builder {
	temp, err := template.New("search.tmpl").Parse(rawTemplate)
	So(err, ShouldBeNil)
	return &Builder{
		searchTemplates: temp,
	}
}
