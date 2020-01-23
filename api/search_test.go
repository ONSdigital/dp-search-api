package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const validQueryParam string = "a"
const validQueryDoc string = `{"valid":"elastic search query"}`
const validESResponse string = `{"raw":"response"}`
const validTransformedResponse string = `{"transformed":"response"}`

func TestSearchHandlerFunc(t *testing.T) {

	Convey("Should return BadRequest for invalid limit parameter", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?limit=a", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusBadRequest)
		So(resp.Body.String(), ShouldContainSubstring, "Invalid limit parameter")
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 0)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 0)
	})

	Convey("Should return BadRequest for invalid offset parameter", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?offset=b", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusBadRequest)
		So(resp.Body.String(), ShouldContainSubstring, "Invalid offset parameter")
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 0)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 0)
	})

	Convey("Should return InternalError for errors returned from query builder", t, func() {
		qbMock := newQueryBuilderMock(nil, errors.New("Something"))
		esMock := newElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam, nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(resp.Body.String(), ShouldContainSubstring, "Failed to create search query")
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Q, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 0)
	})

	Convey("Should return InternalError for errors returned from elastic search", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock(nil, errors.New("Something"))
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam, nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(resp.Body.String(), ShouldContainSubstring, "Failed to run search query")
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Q, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		So(actualRequest, ShouldResemble, validQueryDoc)
	})

	Convey("Should return InternalError for invalid json from elastic search", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock([]byte(`{"dummy":"response"`), nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam, nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(resp.Body.String(), ShouldContainSubstring, "Failed to process search query")
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Q, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		So(actualRequest, ShouldResemble, validQueryDoc)
	})

	Convey("Should return InternalError for transformation failures", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock(nil, errors.New("Something"))

		searchHandler := SearchHandlerFunc(qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam, nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(resp.Body.String(), ShouldContainSubstring, "Failed to transform search result")
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Q, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		So(actualRequest, ShouldResemble, validQueryDoc)
		So(trMock.TransformSearchResponseCalls(), ShouldHaveLength, 1)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		So(actualResponse, ShouldResemble, validESResponse)
	})

	Convey("Should return OK for valid search result with raw=true", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock([]byte(`{"dummy":"response"}`), nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q=a&raw=true", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, `{"dummy":"response"}`)
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Q, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		So(actualRequest, ShouldResemble, validQueryDoc)
		So(trMock.TransformSearchResponseCalls(), ShouldHaveLength, 0)
	})

	Convey("Should return OK for valid search result", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := SearchHandlerFunc(qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam, nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, validTransformedResponse)
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Q, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		So(actualRequest, ShouldResemble, validQueryDoc)
		So(trMock.TransformSearchResponseCalls(), ShouldHaveLength, 1)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		So(actualResponse, ShouldResemble, validESResponse)
	})

	Convey("Should pass all search terms on to elastic search", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := SearchHandlerFunc(qbMock, esMock, trMock)

		req := httptest.NewRequest(
			"GET",
			"http://localhost:8080/search?q="+validQueryParam+
				"&content_type=ta,tb"+
				"&sort_order=relevance"+
				"&limit=1"+
				"&offset=2",
			nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, validTransformedResponse)
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Q, ShouldResemble, validQueryParam)
		So(qbMock.BuildSearchQueryCalls()[0].ContentTypes, ShouldResemble, "ta,tb")
		So(qbMock.BuildSearchQueryCalls()[0].Sort, ShouldResemble, "relevance")
		So(qbMock.BuildSearchQueryCalls()[0].Limit, ShouldEqual, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Offset, ShouldEqual, 2)

		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		So(actualRequest, ShouldResemble, validQueryDoc)

		So(trMock.TransformSearchResponseCalls(), ShouldHaveLength, 1)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		So(actualResponse, ShouldResemble, validESResponse)
	})
}

func newElasticSearcherMock(response []byte, err error) *ElasticSearcherMock {
	return &ElasticSearcherMock{
		MultiSearchFunc: func(ctx context.Context, index string, docType string, request []byte) ([]byte, error) {
			return response, err
		},
	}
}

func newQueryBuilderMock(query []byte, err error) *QueryBuilderMock {
	return &QueryBuilderMock{
		BuildSearchQueryFunc: func(ctx context.Context, q, contentTypes, sort string, limit, offset int) ([]byte, error) {
			return query, err
		},
	}
}

func newResponseTransformerMock(response []byte, err error) *ResponseTransformerMock {
	return &ResponseTransformerMock{
		TransformSearchResponseFunc: func(ctx context.Context, responseData []byte) ([]byte, error) {
			return response, err
		},
	}
}
