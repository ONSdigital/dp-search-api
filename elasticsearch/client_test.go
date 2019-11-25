package elasticsearch

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	rchttp "github.com/ONSdigital/dp-rchttp"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSearch(t *testing.T) {

	Convey("When Search is called", t, func() {

		Convey("Then a request with the search action should be posted", func() {
			rchttpMock := &rchttp.ClienterMock{
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return newResponse("moo"), nil
				},
			}

			client := New("http://localhost:999", rchttpMock)

			res, err := client.Search(context.Background(), "index", "doctype", []byte("search request"))
			So(err, ShouldBeNil)
			So(res, ShouldNotBeEmpty)
			So(rchttpMock.DoCalls(), ShouldHaveLength, 1)
			actualRequest := rchttpMock.DoCalls()[0].Req
			So(actualRequest.URL.String(), ShouldResemble, "http://localhost:999/index/doctype/_search")
			So(actualRequest.Method, ShouldResemble, "POST")
			body, err := ioutil.ReadAll(actualRequest.Body)
			So(err, ShouldBeNil)
			So(string(body), ShouldResemble, "search request")

		})

		Convey("Then a returned error should be passed back", func() {
			rchttpMock := &rchttp.ClienterMock{
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return nil, errors.New("http error")
				},
			}

			client := New("http://localhost:999", rchttpMock)

			_, err := client.Search(context.Background(), "index", "doctype", []byte("search request"))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "http error")
			So(rchttpMock.DoCalls(), ShouldHaveLength, 1)
			actualRequest := rchttpMock.DoCalls()[0].Req
			So(actualRequest.URL.String(), ShouldResemble, "http://localhost:999/index/doctype/_search")
			So(actualRequest.Method, ShouldResemble, "POST")
			body, err := ioutil.ReadAll(actualRequest.Body)
			So(err, ShouldBeNil)
			So(string(body), ShouldResemble, "search request")

		})

	})
}

func TestMultiSearch(t *testing.T) {

	Convey("When MultiSearch is called", t, func() {

		Convey("Then a request with the multi search action should be posted", func() {

			rchttpMock := &rchttp.ClienterMock{
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return newResponse("moo"), nil
				},
			}

			client := New("http://localhost:999", rchttpMock)

			res, err := client.MultiSearch(context.Background(), "index", "doctype", []byte("multiSearch request"))
			So(err, ShouldBeNil)
			So(res, ShouldNotBeEmpty)
			So(rchttpMock.DoCalls(), ShouldHaveLength, 1)
			actualRequest := rchttpMock.DoCalls()[0].Req
			So(actualRequest.URL.String(), ShouldResemble, "http://localhost:999/index/doctype/_msearch")
			So(actualRequest.Method, ShouldResemble, "POST")
			body, err := ioutil.ReadAll(actualRequest.Body)
			So(err, ShouldBeNil)
			So(string(body), ShouldResemble, "multiSearch request")
		})

		Convey("Then a returned error should be passed back", func() {
			rchttpMock := &rchttp.ClienterMock{
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return nil, errors.New("http error")
				},
			}

			client := New("http://localhost:999", rchttpMock)

			_, err := client.MultiSearch(context.Background(), "index", "doctype", []byte("search request"))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "http error")
			So(rchttpMock.DoCalls(), ShouldHaveLength, 1)
			actualRequest := rchttpMock.DoCalls()[0].Req
			So(actualRequest.URL.String(), ShouldResemble, "http://localhost:999/index/doctype/_msearch")
			So(actualRequest.Method, ShouldResemble, "POST")
			body, err := ioutil.ReadAll(actualRequest.Body)
			So(err, ShouldBeNil)
			So(string(body), ShouldResemble, "search request")

		})
	})
}

func TestGetStatus(t *testing.T) {

	Convey("When GetStatus is called", t, func() {

		Convey("Then a GET request with the status action should be called", func() {

			rchttpMock := &rchttp.ClienterMock{
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return newResponse("moo"), nil
				},
			}

			client := New("http://localhost:999", rchttpMock)

			res, err := client.GetStatus(context.Background())
			So(err, ShouldBeNil)
			So(res, ShouldNotBeEmpty)
			So(rchttpMock.DoCalls(), ShouldHaveLength, 1)
			actualRequest := rchttpMock.DoCalls()[0].Req
			So(actualRequest.URL.String(), ShouldResemble, "http://localhost:999/_cat/health")
			So(actualRequest.Method, ShouldResemble, "GET")
		})

		Convey("Then a returned error should be passed back", func() {
			rchttpMock := &rchttp.ClienterMock{
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return nil, errors.New("http error")
				},
			}

			client := New("http://localhost:999", rchttpMock)

			_, err := client.GetStatus(context.Background())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "http error")
			So(rchttpMock.DoCalls(), ShouldHaveLength, 1)
			actualRequest := rchttpMock.DoCalls()[0].Req
			So(actualRequest.URL.String(), ShouldResemble, "http://localhost:999/_cat/health")
			So(actualRequest.Method, ShouldResemble, "GET")

		})
	})
}

func newResponse(body string) *http.Response {
	recorder := httptest.NewRecorder()
	recorder.WriteString(body)
	return recorder.Result()
}
