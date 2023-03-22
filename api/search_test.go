package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-elasticsearch/v3/client"
	"github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/dp-search-api/query"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	validQueryParam          string = "a"
	validQueryDoc            string = `{"valid":"elastic search query"}`
	validESResponse          string = `{"raw":"response"}`
	validTransformedResponse string = `{"count":0,"took":0,"distinct_items_count":0,"topics":null,"content_types":null,"items":null}`
	internalServerErrMsg            = "internal server error"
)

func TestValidateContentTypes(t *testing.T) {
	Convey("An array of content types containing a subset of the default content types should be allowed", t, func() {
		disallowed, err := validateContentTypes([]string{
			"dataset",
			"dataset_landing_page",
			"cantabular_flexible_table",
		})
		So(err, ShouldBeNil)
		So(disallowed, ShouldHaveLength, 0)
	})

	Convey("An array of content types containing a disallowed content type should return the expected error and list", t, func() {
		disallowed, err := validateContentTypes([]string{
			"dataset",
			"dataset_landing_page",
			"wrong_type",
		})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "content type(s) not allowed")
		So(disallowed, ShouldResemble, []string{"wrong_type"})
	})
}

func TestSearchHandlerFunc(t *testing.T) {
	expectedQuery := "a valid query"
	searches := []client.Search{
		{
			Header: client.Header{
				Index: "valid index",
			},
			Query: []byte(expectedQuery),
		},
	}
	validQueryDocBytes, _ := json.Marshal(searches)
	validator := query.NewSearchQueryParamValidator()

	Convey("Should return BadRequest for invalid limit parameter", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newDpElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, esMock, trMock)

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
		esMock := newDpElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, esMock, trMock)

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
		esMock := newDpElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, esMock, trMock)

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
		esMock := newDpElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?offset=-1", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusBadRequest)
		So(resp.Body.String(), ShouldContainSubstring, "Invalid offset parameter")
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 0)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 0)
	})

	Convey("ShouldReturn BadRequest for a content_type that is not allowed", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newDpElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?content_type=wrong1,wrong2", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusBadRequest)
		So(resp.Body.String(), ShouldContainSubstring, "Invalid content_type(s): wrong1,wrong2")
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 0)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 0)
	})

	Convey("Should return InternalError for errors returned from query builder", t, func() {
		qbMock := newQueryBuilderMock(nil, errors.New("Something"))
		esMock := newDpElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam, nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 0)
	})

	Convey("Should return InternalError for errors returned from elastic search", t, func() {
		qbMock := newQueryBuilderMock(validQueryDocBytes, nil)
		esMock := newDpElasticSearcherMock(nil, errors.New("Something"))
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam, nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		So(actualRequest, ShouldResemble, expectedQuery)
	})

	Convey("Should return InternalError for invalid json from elastic search", t, func() {
		qbMock := newQueryBuilderMock(validQueryDocBytes, nil)
		esMock := newDpElasticSearcherMock([]byte(`{"dummy":"response"`), nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam, nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		So(actualRequest, ShouldResemble, expectedQuery)
	})

	Convey("Should return InternalError for transformation failures", t, func() {
		qbMock := newQueryBuilderMock(validQueryDocBytes, nil)
		esMock := newDpElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock(nil, errors.New("Something"))

		searchHandler := SearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam, nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		So(actualRequest, ShouldResemble, expectedQuery)
		So(trMock.TransformSearchResponseCalls(), ShouldHaveLength, 1)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		So(actualResponse, ShouldResemble, validESResponse)
	})

	Convey("Should return OK for valid search result with raw=true", t, func() {
		qbMock := newQueryBuilderMock(validQueryDocBytes, nil)
		esMock := newDpElasticSearcherMock([]byte(`{"dummy":"response"}`), nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q=a&raw=true", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, `{"dummy":"response"}`)
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		So(actualRequest, ShouldResemble, expectedQuery)
		So(trMock.TransformSearchResponseCalls(), ShouldHaveLength, 0)
	})

	Convey("Should return OK for valid search result", t, func() {
		searchBytes, _ := json.Marshal(searches)
		qbMock := newQueryBuilderMock(searchBytes, nil)
		esMock := newDpElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam, nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, validTransformedResponse)
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		So(actualRequest, ShouldResemble, expectedQuery)
		So(trMock.TransformSearchResponseCalls(), ShouldHaveLength, 1)
		So(trMock.TransformSearchResponseCalls()[0].Highlight, ShouldBeTrue)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		So(actualResponse, ShouldResemble, validESResponse)
	})

	Convey("Should return OK for valid search result with highlight = true", t, func() {
		qbMock := newQueryBuilderMock(validQueryDocBytes, nil)
		esMock := newDpElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam+"&highlight=true", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, validTransformedResponse)
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		So(actualRequest, ShouldResemble, expectedQuery)
		So(trMock.TransformSearchResponseCalls(), ShouldHaveLength, 1)
		So(trMock.TransformSearchResponseCalls()[0].Highlight, ShouldBeTrue)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		So(actualResponse, ShouldResemble, validESResponse)
	})

	Convey("Should return OK for valid search result with highlight = false", t, func() {
		qbMock := newQueryBuilderMock(validQueryDocBytes, nil)
		esMock := newDpElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam+"&highlight=false", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, validTransformedResponse)
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		So(actualRequest, ShouldResemble, expectedQuery)
		So(trMock.TransformSearchResponseCalls(), ShouldHaveLength, 1)
		So(trMock.TransformSearchResponseCalls()[0].Highlight, ShouldBeFalse)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		So(actualResponse, ShouldResemble, validESResponse)
	})

	Convey("Should pass all search terms on to elastic search", t, func() {
		qbMock := newQueryBuilderMock(validQueryDocBytes, nil)
		esMock := newDpElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest(
			"GET",
			"http://localhost:8080/search?q="+validQueryParam+
				"&content_type=dataset,release"+
				"&sort_order=relevance"+
				"&limit=1"+
				"&offset=2"+
				"&dimensions=dim1,dim2"+
				"&population_type=pop1",
			nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, validTransformedResponse)
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Types, ShouldResemble, []string{"dataset", "release"})
		So(qbMock.BuildSearchQueryCalls()[0].Req.SortBy, ShouldResemble, "relevance")
		So(qbMock.BuildSearchQueryCalls()[0].Req.Size, ShouldEqual, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.From, ShouldEqual, 2)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Dimensions, ShouldResemble, []*query.DimensionRequest{
			{Name: "dim1"}, {Name: "dim2"},
		})
		So(qbMock.BuildSearchQueryCalls()[0].Req.PopulationType, ShouldResemble, &query.PopulationTypeRequest{
			Name: "pop1",
		})

		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		So(actualRequest, ShouldResemble, expectedQuery)

		So(trMock.TransformSearchResponseCalls(), ShouldHaveLength, 1)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		So(actualResponse, ShouldResemble, validESResponse)
	})
}

func TestLegacySearchHandlerFunc(t *testing.T) {
	validator := query.NewSearchQueryParamValidator()

	Convey("Should return BadRequest for invalid limit parameter", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, esMock, trMock)

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

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, esMock, trMock)

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

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, esMock, trMock)

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

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, esMock, trMock)

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

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam, nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(resp.Body.String(), ShouldContainSubstring, "Failed to create search query")
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 0)
	})

	Convey("Should return InternalError for errors returned from elastic search", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock(nil, errors.New("Something"))
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam, nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(resp.Body.String(), ShouldContainSubstring, "Failed to run search query")
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		So(actualRequest, ShouldResemble, validQueryDoc)
	})

	Convey("Should return InternalError for invalid json from elastic search", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock([]byte(`{"dummy":"response"`), nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam, nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(resp.Body.String(), ShouldContainSubstring, "Failed to process search query")
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		So(actualRequest, ShouldResemble, validQueryDoc)
	})

	Convey("Should return InternalError for transformation failures", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock(nil, errors.New("Something"))

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam, nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(resp.Body.String(), ShouldContainSubstring, "Failed to transform search result")
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
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

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q=a&raw=true", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, `{"dummy":"response"}`)
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		So(actualRequest, ShouldResemble, validQueryDoc)
		So(trMock.TransformSearchResponseCalls(), ShouldHaveLength, 0)
	})

	Convey("Should return OK for valid search result", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam, nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, validTransformedResponse)
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
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

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam+"&highlight=true", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, validTransformedResponse)
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
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

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?q="+validQueryParam+"&highlight=false", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, validTransformedResponse)
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
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

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, esMock, trMock)

		req := httptest.NewRequest(
			"GET",
			"http://localhost:8080/search?q="+validQueryParam+
				"&content_type=article,release"+
				"&sort_order=relevance"+
				"&limit=1"+
				"&offset=2",
			nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, validTransformedResponse)
		So(qbMock.BuildSearchQueryCalls(), ShouldHaveLength, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Term, ShouldResemble, validQueryParam)
		So(qbMock.BuildSearchQueryCalls()[0].Req.Types, ShouldResemble, []string{"article", "release"})
		So(qbMock.BuildSearchQueryCalls()[0].Req.SortBy, ShouldResemble, "relevance")
		So(qbMock.BuildSearchQueryCalls()[0].Req.Size, ShouldEqual, 1)
		So(qbMock.BuildSearchQueryCalls()[0].Req.From, ShouldEqual, 2)

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
		dpESClient := newDpElasticSearcherMock(nil, nil)

		searchAPI := &SearchAPI{dpESClient: dpESClient}

		Convey("When a new elastic search index is created", func() {
			req := httptest.NewRequest("POST", "http://localhost:23900/search", nil)
			resp := httptest.NewRecorder()

			searchAPI.CreateSearchIndexHandlerFunc(resp, req)

			Convey("Then the newly created search index name is returned with status code 201", func() {
				So(resp.Code, ShouldEqual, http.StatusCreated)
				payload, err := io.ReadAll(resp.Body)
				So(err, ShouldBeNil)
				indexCreated := models.CreateIndexResponse{}
				err = json.Unmarshal(payload, &indexCreated)
				So(err, ShouldBeNil)

				Convey("And the index name has the expected name format", func() {
					re := regexp.MustCompile(`(ons)(\d*)`)
					indexName := indexCreated.IndexName
					So(indexName, ShouldNotBeNil)
					wordWithExpectedPattern := re.FindString(indexName)
					So(wordWithExpectedPattern, ShouldEqual, indexName)
				})
			})
		})
	})

	Convey("Given a Search API that is pointing to the old version of Elastic Search", t, func() {
		// The new ES client will return an error if the Search API config is pointing at the old version of ES
		dpESClient := newDpElasticSearcherMock(nil, errors.New("unexpected status code from api"))

		searchAPI := &SearchAPI{dpESClient: dpESClient}

		Convey("When a new elastic search index is created", func() {
			req := httptest.NewRequest("POST", "http://localhost:23900/search", nil)
			resp := httptest.NewRecorder()

			searchAPI.CreateSearchIndexHandlerFunc(resp, req)

			Convey("Then an internal server error is returned with status code 500", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				So(strings.Trim(resp.Body.String(), "\n"), ShouldResemble, internalServerErrMsg)
			})
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

func newDpElasticSearcherMock(response []byte, err error) *DpElasticSearcherMock {
	return &DpElasticSearcherMock{
		CreateIndexFunc: func(ctx context.Context, indexName string, indexSettings []byte) error {
			return err
		},
		MultiSearchFunc: func(ctx context.Context, searches []client.Search, params *client.QueryParams) ([]byte, error) {
			return response, err
		},
		CountFunc: func(ctx context.Context, count client.Count) ([]byte, error) {
			return response, err
		},
	}
}

func newQueryBuilderMock(retQuery []byte, err error) *QueryBuilderMock {
	return &QueryBuilderMock{
		BuildSearchQueryFunc: func(ctx context.Context, req *query.SearchRequest, esVersion710 bool) ([]byte, error) {
			return retQuery, err
		},
		BuildCountQueryFunc: func(ctx context.Context, req *query.CountRequest) ([]byte, error) {
			return retQuery, err
		},
	}
}

func newResponseTransformerMock(response []byte, err error) *ResponseTransformerMock {
	return &ResponseTransformerMock{
		TransformSearchResponseFunc: func(ctx context.Context, responseData []byte, query string, highlight bool) ([]byte, error) {
			return response, err
		},
		TransformCountResponseFunc: func(ctx context.Context, responseData []byte) (int, error) {
			return 0, err
		},
	}
}

func TestSanitise(t *testing.T) {
	Convey("Given a query term with quoted terms", t, func() {
		queryWithQuotes := `"education results for Wales" "education results for England"`

		Convey("when sanitised the individual quotes in the query term should be escaped", func() {
			sanitised := sanitiseDoubleQuotes(queryWithQuotes)
			So(sanitised, ShouldEqual, `\"education results for Wales\" \"education results for England\"`)
		})
	})
}
