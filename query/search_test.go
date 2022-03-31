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
		query, err := qb.BuildSearchQuery(context.Background(), "", "", "", nil, 2, 1)

		So(err, ShouldNotBeNil)
		So(query, ShouldBeNil)
		So(err.Error(), ShouldContainSubstring, "creation of search from template failed")
	})

	Convey("Should include all search parameters in elastic search query", t, func() {
		qb := createQueryBuilderForTemplate("Term={{.Term}};" +
			"From={{.From}};" +
			"Size={{.Size}};" +
			"Types={{.Types}};" +
			"Queries={{.Queries}};" +
			"SortBy={{.SortBy}};" +
			"AggregationField={{.AggregationField}};" +
			"Highlight={{.Highlight}};" +
			"FilterOnLatest={{.FilterOnLatest}};" +
			"Upcoming={{.Upcoming}};" +
			"Published={{.Published}};" +
			"Now={{.Now}}")

		query, err := qb.BuildSearchQuery(context.Background(), "a", "ta,tb", "relevance", []string{"test"}, 2, 1)

		So(err, ShouldBeNil)
		So(query, ShouldNotBeNil)
		queryString := string(query)
		So(queryString, ShouldContainSubstring, "Term=a")
		So(queryString, ShouldContainSubstring, "From=1")
		So(queryString, ShouldContainSubstring, "Size=2")
		So(queryString, ShouldContainSubstring, "Types=[ta tb]")
		So(queryString, ShouldContainSubstring, "SortBy=relevance")
		So(queryString, ShouldContainSubstring, "Queries=[search counts]")
		So(queryString, ShouldContainSubstring, "AggregationField=_type")
		So(queryString, ShouldContainSubstring, "Highlight=true")
		So(queryString, ShouldContainSubstring, "FilterOnLatest=false")
		So(queryString, ShouldContainSubstring, "Upcoming=false")
		So(queryString, ShouldContainSubstring, "Published=false")
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

func TestHasQuery(t *testing.T) {
	Convey("When called with multiple query values", t, func() {
		sr := searchRequest{
			Queries: []string{"moo", "quack"},
		}

		Convey("Should return true for included queries", func() {
			So(sr.HasQuery("moo"), ShouldBeTrue)
			So(sr.HasQuery("quack"), ShouldBeTrue)
		})

		Convey("Should return false for excluded queries", func() {
			So(sr.HasQuery("oink"), ShouldBeFalse)
		})
	})

	Convey("Should return false when called with zero query values", t, func() {
		sr := searchRequest{
			Queries: []string{},
		}
		So(sr.HasQuery("oink"), ShouldBeFalse)
	})
}
