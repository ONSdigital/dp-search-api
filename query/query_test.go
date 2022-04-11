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
		So(err, ShouldBeNil)
	})
}
