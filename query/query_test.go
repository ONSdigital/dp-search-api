package query

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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
