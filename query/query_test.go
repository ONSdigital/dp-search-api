package query

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
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
}