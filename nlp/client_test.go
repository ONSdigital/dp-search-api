package nlp

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBuildURL(t *testing.T) {
	Convey("Given a base URL, query parameters, and a query key", t, func() {
		baseURL := "https://example.com/api"
		params := url.Values{}
		params.Set("q", "example")
		queryKey := "search"

		Convey("When buildURL is called", func() {
			url, err := buildURL(baseURL, params, queryKey)

			Convey("The URL should be built correctly", func() {
				So(url.String(), ShouldEqual, "https://example.com/api?search=example")
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given an invalid base URL", t, func() {
		baseURL := ":"
		params := url.Values{}
		params.Set("q", "example")
		queryKey := "search"

		Convey("When buildURL is called", func() {
			url, err := buildURL(baseURL, params, queryKey)

			Convey("An error should be returned", func() {
				So(url, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given an empty query parameter", t, func() {
		baseURL := "https://example.com/api"
		params := url.Values{}
		queryKey := "search"

		Convey("When buildURL is called", func() {
			url, err := buildURL(baseURL, params, queryKey)

			Convey("The URL should only contain the query key", func() {
				So(url.String(), ShouldEqual, "https://example.com/api?search=")
				So(err, ShouldBeNil)
			})
		})
	})
}
