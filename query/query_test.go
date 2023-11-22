package query

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSetupSearch(t *testing.T) {
	Convey("Should return templates and error should be nil", t, func() {
		searchTemplates, err := SetupSearch()

		So(err, ShouldBeNil)
		So(searchTemplates, ShouldNotBeNil)
	})
}

func TestNewQueryBuilder(t *testing.T) {
	Convey("Should return a Builder object with templates", t, func() {
		builderObject, err := NewQueryBuilder()

		So(builderObject.searchTemplates, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	Convey("Should return a Builder object with elastic v710 templates", t, func() {
		builderObject, err := NewQueryBuilder()
		So(err, ShouldBeNil)

		So(builderObject.searchTemplates, ShouldNotBeNil)
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "search.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentQuery.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentHeader.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "matchAll.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentQuery.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentHeader.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "countContentTypeHeader.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "countContentTypeQuery.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "countContentTypeFilters.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "countTopicHeader.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "countTopicQuery.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "countTopicFilters.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "countPopulationTypeHeader.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "countPopulationTypeQuery.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "countPopulationTypeFilters.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "countDimensionsHeader.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "countDimensionsQuery.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "countDimensionsFilters.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "coreQuery.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "weightedQuery.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentFilters.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentFilterOnURIPrefix.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentFilterOnTopicWildcard.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "sortByTitle.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "sortByRelevance.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "sortByReleaseDate.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "sortByReleaseDateAsc.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "sortByFirstLetter.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "topicFilters.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "canonicalFilters.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "subTopicsFilters.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentTypeFilter.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "nlpLocation.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "nlpCategory.tmpl")

		So(builderObject.countTemplates, ShouldNotBeNil)
		So(builderObject.countTemplates.DefinedTemplates(), ShouldContainSubstring, "distinctItemCountQuery.tmpl")
		So(builderObject.countTemplates.DefinedTemplates(), ShouldContainSubstring, "coreQuery.tmpl")
		So(builderObject.countTemplates.DefinedTemplates(), ShouldContainSubstring, "matchAll.tmpl")
	})
}
