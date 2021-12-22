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
		builderObject, err := NewQueryBuilder(testPathToTemplates)

		So(builderObject.searchTemplates, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
