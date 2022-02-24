package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-authorisation/auth"
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

	Convey("Should return BadRequest for negative limit parameter", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?limit=-1", nil)
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

	Convey("Should return BadRequest for negative offset parameter", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?offset=-1", nil)
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
		So(trMock.TransformSearchResponseCalls()[0].Highlight, ShouldBeTrue)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		So(actualResponse, ShouldResemble, validESResponse)
	})

	Convey("Should return OK for valid search result with highlight = true", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := SearchHandlerFunc(qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam+"&highlight=true", nil)
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
		So(trMock.TransformSearchResponseCalls()[0].Highlight, ShouldBeTrue)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		So(actualResponse, ShouldResemble, validESResponse)
	})

	Convey("Should return OK for valid search result with highlight = false", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := SearchHandlerFunc(qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam+"&highlight=false", nil)
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
		So(trMock.TransformSearchResponseCalls()[0].Highlight, ShouldBeFalse)
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

func TestCreateSearchIndexHandlerFunc(t *testing.T) {
	Convey("Given a Search API that is pointing to the Site Wide version of Elastic Search", t, func() {
		//cfg, err := config.Get()
		//So(err, ShouldBeNil)
		//
		//cfg.ElasticSearchAPIURL := "http://localhost:11200"

		dpESClient := newDpElasticSearcherMock(200, nil)
		permissions := newAuthHandlerMock()

		searchAPI := &SearchAPI{dpESClient: dpESClient, permissions: permissions}

		Convey("When a new reindex job is created and stored", func() {
			req := httptest.NewRequest("POST", "http://localhost:23900/search", nil)
			//req, err := http.NewRequest(http.MethodPost, "http://localhost:23900/search", http.NoBody)
			//So(err, ShouldBeNil)

			resp := httptest.NewRecorder()

			searchAPI.CreateSearchIndexHandlerFunc(resp, req)

		})
	})
}

func newElasticSearcherMock(response []byte, err error) *ElasticSearcherMock {
	return &ElasticSearcherMock{
		MultiSearchFunc: func(ctx context.Context, index string, docType string, request []byte) ([]byte, error) {
			return response, err
		},
	}
}

func newDpElasticSearcherMock(status int, err error) *DpElasticSearcherMock {
	return &DpElasticSearcherMock{
		CreateIndexFunc: func(ctx context.Context, indexName string, indexSettings []byte) (int, error) {
			return status, err
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
		TransformSearchResponseFunc: func(ctx context.Context, responseData []byte, query string, highlight bool) ([]byte, error) {
			return response, err
		},
	}
}

func newAuthHandlerMock() *AuthHandlerMock {
	return &AuthHandlerMock{
		RequireFunc: func(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc {
			return handler
		},
	}
}
