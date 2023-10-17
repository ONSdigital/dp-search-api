package nlp

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBuildURL(t *testing.T) {
	baseURL := "https://example.com/api"
	queryKey := "search"
	Convey("Given a base URL, query parameters, and a query key", t, func() {
		Convey("When buildURL is called", func() {
			resultURL, err := buildURL(baseURL, "example", queryKey)

			Convey("The URL should be built correctly", func() {
				So(resultURL.String(), ShouldEqual, "https://example.com/api?search=example")
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given an invalid base URL", t, func() {
		Convey("When buildURL is called", func() {
			resultURL, err := buildURL(":", "example", queryKey)

			Convey("An error should be returned", func() {
				So(resultURL, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given an empty query parameter", t, func() {
		Convey("When buildURL is called", func() {
			resultURL, err := buildURL(baseURL, "", queryKey)

			Convey("The URL should only contain the query key", func() {
				So(resultURL.String(), ShouldEqual, "https://example.com/api?search=")
				So(err, ShouldBeNil)
			})
		})
	})
}
