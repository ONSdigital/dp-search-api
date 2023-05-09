package nlp

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBuildURL(t *testing.T) {
	baseURL := "https://example.com/api"
	queryKey := "search"
	Convey("Given a base URL, query parameters, and a query key", t, func() {
		params := url.Values{}
		params.Set("q", "example")

		Convey("When buildURL is called", func() {
			resultURL, err := buildURL(baseURL, params, queryKey)

			Convey("The URL should be built correctly", func() {
				So(resultURL.String(), ShouldEqual, "https://example.com/api?search=example")
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given an invalid base URL", t, func() {
		params := url.Values{}
		params.Set("q", "example")

		Convey("When buildURL is called", func() {
			resultURL, err := buildURL(":", params, queryKey)

			Convey("An error should be returned", func() {
				So(resultURL, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given an empty query parameter", t, func() {
		params := url.Values{}

		Convey("When buildURL is called", func() {
			resultURL, err := buildURL(baseURL, params, queryKey)

			Convey("The URL should only contain the query key", func() {
				So(resultURL.String(), ShouldEqual, "https://example.com/api?search=")
				So(err, ShouldBeNil)
			})
		})
	})
}
