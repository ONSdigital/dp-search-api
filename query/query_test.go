package query

import (
	"testing"

	c "github.com/smartystreets/goconvey/convey"
)

func TestSetupSearch(t *testing.T) {
	c.Convey("Should return templates and error should be nil", t, func() {
		searchTemplates, err := SetupSearch()

		c.So(err, c.ShouldBeNil)
		c.So(searchTemplates, c.ShouldNotBeNil)
	})
}

func TestNewQueryBuilder(t *testing.T) {
	c.Convey("Should return a Builder object with templates", t, func() {
		builderObject, err := NewQueryBuilder()

		c.So(builderObject.searchTemplates, c.ShouldNotBeNil)
		c.So(err, c.ShouldBeNil)
	})

	c.Convey("Should return a Builder object with elastic v710 templates", t, func() {
		builderObject, err := NewQueryBuilder()
		c.So(err, c.ShouldBeNil)

		c.So(builderObject.searchTemplates, c.ShouldNotBeNil)
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "search.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "contentQuery.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "contentHeader.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "matchAll.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "contentQuery.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "contentHeader.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "countContentTypeHeader.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "countContentTypeQuery.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "countContentTypeFilters.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "countTopicHeader.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "countTopicQuery.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "countTopicFilters.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "countPopulationTypeHeader.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "countPopulationTypeQuery.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "countPopulationTypeFilters.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "countDimensionsHeader.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "countDimensionsQuery.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "countDimensionsFilters.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "coreQuery.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "weightedQuery.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "contentFilters.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "contentFilterOnURIPrefix.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "contentFilterOnTopicWildcard.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "sortByTitle.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "sortByRelevance.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "sortByReleaseDate.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "sortByReleaseDateAsc.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "sortByFirstLetter.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "topicFilters.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "canonicalFilters.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "subTopicsFilters.tmpl")
		c.So(builderObject.searchTemplates.DefinedTemplates(), c.ShouldContainSubstring, "contentTypeFilter.tmpl")

		c.So(builderObject.countTemplates, c.ShouldNotBeNil)
		c.So(builderObject.countTemplates.DefinedTemplates(), c.ShouldContainSubstring, "distinctItemCountQuery.tmpl")
		c.So(builderObject.countTemplates.DefinedTemplates(), c.ShouldContainSubstring, "coreQuery.tmpl")
		c.So(builderObject.countTemplates.DefinedTemplates(), c.ShouldContainSubstring, "matchAll.tmpl")
	})
}
