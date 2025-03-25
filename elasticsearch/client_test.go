package elasticsearch

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	dphttp "github.com/ONSdigital/dp-net/v3/http"
	c "github.com/smartystreets/goconvey/convey"
)

func TestSearch(t *testing.T) {
	c.Convey("When Search is called", t, func() {
		// Define a mock struct to be used in your unit tests of myFunc.
		c.Convey("Then a request with the search action should be posted", func() {
			dphttpMock := &dphttp.ClienterMock{
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return newResponse("moo"), nil
				},
			}

			client := New("http://localhost:999", dphttpMock, "es", "eu-west-1")

			res, err := client.Search(context.Background(), "index", "doctype", []byte("search request"))
			c.So(err, c.ShouldBeNil)
			c.So(res, c.ShouldNotBeEmpty)
			c.So(dphttpMock.DoCalls(), c.ShouldHaveLength, 1)
			actualRequest := dphttpMock.DoCalls()[0].Req
			c.So(actualRequest.URL.String(), c.ShouldResemble, "http://localhost:999/index/doctype/_search")
			c.So(actualRequest.Method, c.ShouldResemble, "POST")
			body, err := io.ReadAll(actualRequest.Body)
			c.So(err, c.ShouldBeNil)
			c.So(string(body), c.ShouldResemble, "search request")
		})

		c.Convey("Then a returned error should be passed back", func() {
			dphttpMock := &dphttp.ClienterMock{
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return nil, errors.New("http error")
				},
			}

			client := New("http://localhost:999", dphttpMock, "es", "eu-west-1")

			_, err := client.Search(context.Background(), "index", "doctype", []byte("search request"))
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldResemble, "http error")
			c.So(dphttpMock.DoCalls(), c.ShouldHaveLength, 1)
			actualRequest := dphttpMock.DoCalls()[0].Req
			c.So(actualRequest.URL.String(), c.ShouldResemble, "http://localhost:999/index/doctype/_search")
			c.So(actualRequest.Method, c.ShouldResemble, "POST")
			body, err := io.ReadAll(actualRequest.Body)
			c.So(err, c.ShouldBeNil)
			c.So(string(body), c.ShouldResemble, "search request")
		})
	})
}

func TestMultiSearch(t *testing.T) {
	c.Convey("When MultiSearch is called", t, func() {
		c.Convey("Then a request with the multi search action should be posted", func() {
			dphttpMock := &dphttp.ClienterMock{
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return newResponse("moo"), nil
				},
			}

			client := New("http://localhost:999", dphttpMock, "es", "eu-west-1")

			res, err := client.MultiSearch(context.Background(), "index", "doctype", []byte("multiSearch request"))
			c.So(err, c.ShouldBeNil)
			c.So(res, c.ShouldNotBeEmpty)
			c.So(dphttpMock.DoCalls(), c.ShouldHaveLength, 1)
			actualRequest := dphttpMock.DoCalls()[0].Req
			c.So(actualRequest.URL.String(), c.ShouldResemble, "http://localhost:999/index/doctype/_msearch")
			c.So(actualRequest.Method, c.ShouldResemble, "POST")
			body, err := io.ReadAll(actualRequest.Body)
			c.So(err, c.ShouldBeNil)
			c.So(string(body), c.ShouldResemble, "multiSearch request")
		})

		c.Convey("Then a returned error should be passed back", func() {
			dphttpMock := &dphttp.ClienterMock{
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return nil, errors.New("http error")
				},
			}

			client := New("http://localhost:999", dphttpMock, "es", "eu-west-1")

			_, err := client.MultiSearch(context.Background(), "index", "doctype", []byte("search request"))
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldResemble, "http error")
			c.So(dphttpMock.DoCalls(), c.ShouldHaveLength, 1)
			actualRequest := dphttpMock.DoCalls()[0].Req
			c.So(actualRequest.URL.String(), c.ShouldResemble, "http://localhost:999/index/doctype/_msearch")
			c.So(actualRequest.Method, c.ShouldResemble, "POST")
			body, err := io.ReadAll(actualRequest.Body)
			c.So(err, c.ShouldBeNil)
			c.So(string(body), c.ShouldResemble, "search request")
		})
	})
}

func newResponse(body string) *http.Response {
	recorder := httptest.NewRecorder()
	recorder.WriteString(body)
	return recorder.Result()
}
