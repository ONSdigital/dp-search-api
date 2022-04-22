package query

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var testPathToTemplates = "../"

func TestSetupSearch(t *testing.T) {
	Convey("Should return templates and error should be nil", t, func() {
		searchTemplates, err := SetupSearch(testPathToTemplates)

		So(err, ShouldBeNil)
		So(searchTemplates, ShouldNotBeNil)
	})
}

func TestNewQueryBuilder(t *testing.T) {
	Convey("Should return a Builder object with templates", t, func() {
		builderObject, err := NewQueryBuilder(testPathToTemplates, false)

		So(builderObject.searchTemplates, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	Convey("Should return a Builder object with elastic v710 templates", t, func() {
		builderObject, err := NewQueryBuilder(testPathToTemplates, true)

		So(builderObject.searchTemplates, ShouldNotBeNil)
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "search.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentQuery.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentHeader.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "matchAll.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentQuery.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentHeader.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "countHeader.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "countQuery.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "coreQuery.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "weightedQuery.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentFilters.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentFilterUpcoming.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentFilterPublished.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentFilterOnFirstLetter.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentFilterOnReleaseDate.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentFilterOnURIPrefix.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentFilterOnTopic.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "contentFilterOnTopicWildcard.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "sortByTitle.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "sortByRelevance.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "sortByReleaseDate.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "sortByReleaseDateAsc.tmpl")
		So(builderObject.searchTemplates.DefinedTemplates(), ShouldContainSubstring, "sortByFirstLetter.tmpl")

		So(err, ShouldBeNil)
	})
}
