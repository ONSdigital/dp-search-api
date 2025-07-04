package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/nlp/berlin"
	brErr "github.com/ONSdigital/dp-api-clients-go/v2/nlp/berlin/errors"
	brModels "github.com/ONSdigital/dp-api-clients-go/v2/nlp/berlin/models"
	"github.com/ONSdigital/dp-api-clients-go/v2/nlp/category"
	catErr "github.com/ONSdigital/dp-api-clients-go/v2/nlp/category/errors"
	catModels "github.com/ONSdigital/dp-api-clients-go/v2/nlp/category/models"
	"github.com/ONSdigital/dp-elasticsearch/v3/client"
	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/dp-search-api/query"
	scrModels "github.com/ONSdigital/dp-search-scrubber-api/models"
	scr "github.com/ONSdigital/dp-search-scrubber-api/sdk"
	scrErr "github.com/ONSdigital/dp-search-scrubber-api/sdk/errors"
	scrMocks "github.com/ONSdigital/dp-search-scrubber-api/sdk/mocks"
	c "github.com/smartystreets/goconvey/convey"
)

const (
	baseURL                            string = "http://localhost:8080/search?q="
	highlightTrue                      string = "&highlight=true"
	highlightFalse                     string = "&highlight=false"
	limit1                             string = "&limit=1"
	offset2                            string = "&offset=2"
	rawTrue                            string = "&raw=true"
	sortRelevance                      string = "&sort=relevance"
	validQueryParam                    string = "a"
	validQueryDoc                      string = `{"valid":"elastic search query"}`
	validESResponse                    string = `{"raw":"response"}`
	validTransformedResponse           string = `{"count":0,"took":0,"distinct_items_count":0,"topics":null,"content_types":null,"items":null}`
	validTransformedResponseWith2Items string = `{"count":2,"took":0,"distinct_items_count":0,"topics":null,"content_types":null,"items":null}`
	internalServerErrMsg                      = "internal server error"
	defaultNLPSettings                 string = "{\"category_weighting\": 1000000000.0, \"category_limit\": 100, \"default_state\": \"gb\"}"
	nlpParamEnabled                    string = "&nlp_weighting=true"
)

func TestValidateContentTypes(t *testing.T) {
	c.Convey("An array of content types containing a subset of the default content types should be allowed", t, func() {
		disallowed, err := validateContentTypes([]string{
			"dataset",
			"statistical_article",
			"dataset_landing_page",
		})
		c.So(err, c.ShouldBeNil)
		c.So(disallowed, c.ShouldHaveLength, 0)
	})

	c.Convey("An array of content types containing a disallowed content type should return the expected error and list", t, func() {
		disallowed, err := validateContentTypes([]string{
			"dataset",
			"dataset_landing_page",
			"wrong_type",
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldEqual, "content type(s) not allowed")
		c.So(disallowed, c.ShouldResemble, []string{"wrong_type"})
	})
}

func TestValidateURIPrefix(t *testing.T) {
	c.Convey("Should return valid URI prefix when the input is valid", t, func() {
		validURIPrefix := "/economy/"
		result, err := validateURIPrefix(validURIPrefix)
		c.So(err, c.ShouldBeNil)
		c.So(result, c.ShouldEqual, validURIPrefix)
	})

	c.Convey("Should return error when the URI prefix is empty", t, func() {
		emptyURIPrefix := ""
		_, err := validateURIPrefix(emptyURIPrefix)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldEqual, "invalid URI prefix parameter")
	})

	c.Convey("Should return error when the URI prefix does not start with '/'", t, func() {
		invalidURIPrefix := "economy"
		_, err := validateURIPrefix(invalidURIPrefix)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldEqual, "invalid URI prefix parameter")
	})

	c.Convey("Should return the first part of URI prefix when comma separated", t, func() {
		commaSeparatedURIPrefix := "/economy,/invalid"
		result, err := validateURIPrefix(commaSeparatedURIPrefix)
		c.So(err, c.ShouldBeNil)
		c.So(result, c.ShouldEqual, "/economy/")
	})
}

func TestCheckForSpecialCharacters(t *testing.T) {
	c.Convey("A string containing no special characters should return false", t, func() {
		expected := false
		actual := checkForSpecialCharacters("Test string")
		c.So(actual, c.ShouldEqual, expected)
	})

	c.Convey("A string containing whitelisted special characters should return false", t, func() {
		expected := false
		actual := checkForSpecialCharacters("Test string –‘’")
		c.So(actual, c.ShouldEqual, expected)
	})

	c.Convey("A string containing special characters should return true", t, func() {
		expected := true
		actual := checkForSpecialCharacters("Test 怎么开 string")
		c.So(actual, c.ShouldEqual, expected)
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

	cfg := &config.Config{
		DefaultLimit:        10,
		DefaultOffset:       0,
		DefaultMaximumLimit: 100,
		DefaultSort:         "relevance",
	}

	c.Convey("Should return BadRequest for invalid limit parameter", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newDpElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?limit=a", http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusBadRequest)
		c.So(resp.Body.String(), c.ShouldContainSubstring, "invalid limit parameter")
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 0)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 0)
	})

	c.Convey("Should return BadRequest for negative limit parameter", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newDpElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?limit=-1", http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusBadRequest)
		c.So(resp.Body.String(), c.ShouldContainSubstring, "invalid limit parameter")
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 0)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 0)
	})

	c.Convey("Should return BadRequest for invalid offset parameter", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newDpElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?offset=b", http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusBadRequest)
		c.So(resp.Body.String(), c.ShouldContainSubstring, "invalid offset parameter")
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 0)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 0)
	})

	c.Convey("Should return BadRequest for negative offset parameter", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newDpElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?offset=-1", http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusBadRequest)
		c.So(resp.Body.String(), c.ShouldContainSubstring, "invalid offset parameter")
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 0)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 0)
	})

	c.Convey("ShouldReturn BadRequest for a content_type that is not allowed", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newDpElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?content_type=wrong1,wrong2", http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusBadRequest)
		c.So(resp.Body.String(), c.ShouldContainSubstring, "invalid content_type(s): wrong1,wrong2")
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 0)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 0)
	})

	c.Convey("Should return BadRequest for invalid dataset_ids params", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newDpElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?dataset_ids=q", http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)
		c.So(resp.Code, c.ShouldEqual, http.StatusBadRequest)
		c.So(resp.Body.String(), c.ShouldContainSubstring, "invalid dataset_ids: q")
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 0)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 0)
	})

	c.Convey("Should return OK for valid dataset_ids parameter", t, func() {
		searchBytes, _ := json.Marshal(searches)
		qbMock := newQueryBuilderMock(searchBytes, nil)
		esMock := newDpElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam+"&dataset_ids=QNA", http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusOK)
		c.So(resp.Body.String(), c.ShouldResemble, validTransformedResponse)
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		c.So(actualRequest, c.ShouldResemble, expectedQuery)
		c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
		c.So(trMock.TransformSearchResponseCalls()[0].Highlight, c.ShouldBeTrue)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		c.So(actualResponse, c.ShouldResemble, validESResponse)
	})

	c.Convey("Should return BadRequest for a uri_prefix that is not allowed", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newDpElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?uri_prefix=wrong", http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusBadRequest)
		c.So(resp.Body.String(), c.ShouldContainSubstring, "invalid URI prefix parameter")
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 0)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 0)
	})

	c.Convey("Should return BadRequest for invalid cdid params", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newDpElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		c.Convey("Too short CDID", func() {
			req := httptest.NewRequest("GET", "http://localhost:8080/search?cdids=cd", http.NoBody)
			resp := httptest.NewRecorder()

			searchHandler.ServeHTTP(resp, req)
			c.So(resp.Code, c.ShouldEqual, http.StatusBadRequest)
			c.So(resp.Body.String(), c.ShouldContainSubstring, "invalid cdid(s) found")
			c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 0)
			c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 0)
		})

		c.Convey("Too long CDID", func() {
			req := httptest.NewRequest("GET", "http://localhost:8080/search?cdids=cd1234", http.NoBody)
			resp := httptest.NewRecorder()

			searchHandler.ServeHTTP(resp, req)
			c.So(resp.Code, c.ShouldEqual, http.StatusBadRequest)
			c.So(resp.Body.String(), c.ShouldContainSubstring, "invalid cdid(s) found")
			c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 0)
			c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 0)
		})

		c.Convey("CDID with special character", func() {
			req := httptest.NewRequest("GET", "http://localhost:8080/search?cdids=cd-1", http.NoBody)
			resp := httptest.NewRecorder()

			searchHandler.ServeHTTP(resp, req)
			c.So(resp.Code, c.ShouldEqual, http.StatusBadRequest)
			c.So(resp.Body.String(), c.ShouldContainSubstring, "invalid cdid(s) found")
			c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 0)
			c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 0)
		})

		c.Convey("Invalid CDID in list", func() {
			req := httptest.NewRequest("GET", "http://localhost:8080/search?cdids=cd,cdid", http.NoBody)
			resp := httptest.NewRecorder()

			searchHandler.ServeHTTP(resp, req)
			c.So(resp.Code, c.ShouldEqual, http.StatusBadRequest)
			c.So(resp.Body.String(), c.ShouldContainSubstring, "invalid cdid(s) found")
			c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 0)
			c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 0)
		})
	})

	c.Convey("Should return OK for valid cdid params", t, func() {
		qbMock := newQueryBuilderMock(validQueryDocBytes, nil)
		esMock := newDpElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam+"&cdids=cdid,cdid1", http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusOK)
		c.So(resp.Body.String(), c.ShouldResemble, validTransformedResponse)
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		c.So(actualRequest, c.ShouldResemble, expectedQuery)
		c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
		c.So(trMock.TransformSearchResponseCalls()[0].Highlight, c.ShouldBeTrue)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		c.So(actualResponse, c.ShouldResemble, validESResponse)
	})

	c.Convey("Should return InternalError for errors returned from query builder", t, func() {
		qbMock := newQueryBuilderMock(nil, errors.New("Something"))
		esMock := newDpElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusInternalServerError)
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 0)
	})

	c.Convey("Should return InternalError for errors returned from elastic search", t, func() {
		qbMock := newQueryBuilderMock(validQueryDocBytes, nil)
		esMock := newDpElasticSearcherMock(nil, errors.New("Something"))
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusInternalServerError)
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		c.So(actualRequest, c.ShouldResemble, expectedQuery)
	})

	c.Convey("Should return InternalError for invalid json from elastic search", t, func() {
		qbMock := newQueryBuilderMock(validQueryDocBytes, nil)
		esMock := newDpElasticSearcherMock([]byte(`{"dummy":"response"`), nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusInternalServerError)
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		c.So(actualRequest, c.ShouldResemble, expectedQuery)
	})

	c.Convey("Should return InternalError for transformation failures", t, func() {
		qbMock := newQueryBuilderMock(validQueryDocBytes, nil)
		esMock := newDpElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock(nil, errors.New("Something"))

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusInternalServerError)
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		c.So(actualRequest, c.ShouldResemble, expectedQuery)
		c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		c.So(actualResponse, c.ShouldResemble, validESResponse)
	})

	c.Convey("Should return OK for valid search result with raw=true", t, func() {
		qbMock := newQueryBuilderMock(validQueryDocBytes, nil)
		esMock := newDpElasticSearcherMock([]byte(`{"dummy":"response"}`), nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam+rawTrue, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusOK)
		c.So(resp.Body.String(), c.ShouldResemble, `{"dummy":"response"}`)
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		c.So(actualRequest, c.ShouldResemble, expectedQuery)
		c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 0)
	})

	c.Convey("Should return OK for valid search result", t, func() {
		searchBytes, _ := json.Marshal(searches)
		qbMock := newQueryBuilderMock(searchBytes, nil)
		esMock := newDpElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusOK)
		c.So(resp.Body.String(), c.ShouldResemble, validTransformedResponse)
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		c.So(actualRequest, c.ShouldResemble, expectedQuery)
		c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
		c.So(trMock.TransformSearchResponseCalls()[0].Highlight, c.ShouldBeTrue)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		c.So(actualResponse, c.ShouldResemble, validESResponse)
	})

	c.Convey("Should return OK for valid search result with highlight = true", t, func() {
		qbMock := newQueryBuilderMock(validQueryDocBytes, nil)
		esMock := newDpElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam+highlightTrue, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusOK)
		c.So(resp.Body.String(), c.ShouldResemble, validTransformedResponse)
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		c.So(actualRequest, c.ShouldResemble, expectedQuery)
		c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
		c.So(trMock.TransformSearchResponseCalls()[0].Highlight, c.ShouldBeTrue)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		c.So(actualResponse, c.ShouldResemble, validESResponse)
	})

	c.Convey("Should return OK for valid search result with highlight = false", t, func() {
		qbMock := newQueryBuilderMock(validQueryDocBytes, nil)
		esMock := newDpElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam+highlightFalse, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusOK)
		c.So(resp.Body.String(), c.ShouldResemble, validTransformedResponse)
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		c.So(actualRequest, c.ShouldResemble, expectedQuery)
		c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
		c.So(trMock.TransformSearchResponseCalls()[0].Highlight, c.ShouldBeFalse)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		c.So(actualResponse, c.ShouldResemble, validESResponse)
	})

	c.Convey("Should pass all search terms on to elastic search", t, func() {
		qbMock := newQueryBuilderMock(validQueryDocBytes, nil)
		esMock := newDpElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, &ClientList{DpESClient: esMock}, trMock)

		req := httptest.NewRequest(
			"GET",
			baseURL+validQueryParam+
				"&content_type=dataset,release"+
				sortRelevance+
				limit1+
				offset2+
				"&dimensions=dim1,dim2"+
				"&population_types=pop1,pop2"+
				"&fromDate=2020-10-10"+
				"&toDate=2023-10-10"+
				"&dataset_ids=QNA,QDA"+
				"&uri_prefix=/economy"+
				"&cdids=id01,id02",
			http.NoBody)

		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusOK)
		c.So(resp.Body.String(), c.ShouldResemble, validTransformedResponse)
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Types, c.ShouldResemble, []string{"dataset", "release"})
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.SortBy, c.ShouldResemble, "relevance")
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Size, c.ShouldEqual, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.From, c.ShouldEqual, 2)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.ReleasedAfter, c.ShouldResemble, query.MustParseDate("2020-10-10"))
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.ReleasedBefore, c.ShouldResemble, query.MustParseDate("2023-10-10"))
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Dimensions, c.ShouldResemble, []*query.DimensionRequest{
			{Key: "dim1"}, {Key: "dim2"},
		})
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.PopulationTypes, c.ShouldResemble, []*query.PopulationTypeRequest{
			{Key: "pop1"}, {Key: "pop2"},
		})
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.DatasetIDs, c.ShouldResemble, []string{"QNA", "QDA"})

		c.So(qbMock.BuildSearchQueryCalls()[0].Req.CDIDs, c.ShouldResemble, []string{"id01", "id02"})
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.URIPrefix, c.ShouldEqual, "/economy/")
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
		c.So(actualRequest, c.ShouldResemble, expectedQuery)

		c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		c.So(actualResponse, c.ShouldResemble, validESResponse)
	})

	c.Convey("When NLP features are enabled and a valid request is made", t, func() {
		searchBytes, _ := json.Marshal(searches)
		qbMock := newQueryBuilderMock(searchBytes, nil)

		esMock := newDpElasticSearcherMock([]byte(`{"raw":"response"}`), nil)

		brMock := newBerlinClienterMock(&brModels.Berlin{
			Matches: []brModels.Matches{
				{
					Loc: brModels.Locations{
						Subdivision: []string{
							"subdiv1",
							"subdiv2",
						},
					},
				},
			},
		}, nil)

		catMock := newCategoryClienterMock(&[]catModels.Category{
			{
				Code:  []string{"sth", "sth"},
				Score: 100,
			},
		}, nil)

		scrMock := newScrubberClienterMock(&scrModels.ScrubberResp{
			Results: scrModels.Results{
				Areas: []scrModels.AreaResp{
					{
						Name:   "Area1",
						Region: "region1",
					},
				},
			},
		}, nil)

		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		clList := &ClientList{
			ScrubberClient: scrMock,
			CategoryClient: catMock,
			BerlinClient:   brMock,
			DpESClient:     esMock,
		}

		cfg := &config.Config{
			EnableNLPWeighting: true,
			NLPSettings:        defaultNLPSettings,
			DefaultSort:        "relevance",
		}

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, clList, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam+nlpParamEnabled, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.Convey("Then the request should be processed OK", func() {
			c.So(resp.Code, c.ShouldEqual, http.StatusOK)
			c.So(resp.Body.String(), c.ShouldResemble, validTransformedResponse)
			c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
			c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		})

		c.Convey("And it should make calls to NLP APIs", func() {
			c.So(scrMock.GetScrubberCalls(), c.ShouldHaveLength, 1)
			c.So(catMock.GetCategoryCalls(), c.ShouldHaveLength, 1)
			c.So(brMock.GetBerlinCalls(), c.ShouldHaveLength, 1)
		})

		c.Convey("And the request to ES should be processed correctly", func() {
			c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
			actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
			c.So(actualRequest, c.ShouldResemble, expectedQuery)
		})

		c.Convey("And the response to ES should be processed correctly", func() {
			c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
			c.So(trMock.TransformSearchResponseCalls()[0].Highlight, c.ShouldBeTrue)
			actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
			c.So(actualResponse, c.ShouldResemble, validESResponse)
		})
	})

	c.Convey("When NLP features are enabled, a valid request is made but no nlp settings are set", t, func() {
		searchBytes, _ := json.Marshal(searches)
		qbMock := newQueryBuilderMock(searchBytes, nil)

		esMock := newDpElasticSearcherMock([]byte(`{"raw":"response"}`), nil)

		brMock := newBerlinClienterMock(&brModels.Berlin{
			Matches: []brModels.Matches{
				{
					Loc: brModels.Locations{
						Subdivision: []string{
							"subdiv1",
							"subdiv2",
						},
					},
				},
			},
		}, nil)

		catMock := newCategoryClienterMock(&[]catModels.Category{
			{
				Code:  []string{"sth", "sth"},
				Score: 100,
			},
		}, nil)

		scrMock := newScrubberClienterMock(&scrModels.ScrubberResp{
			Results: scrModels.Results{
				Areas: []scrModels.AreaResp{
					{
						Name:   "Area1",
						Region: "region1",
					},
				},
			},
		}, nil)

		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		clList := &ClientList{
			ScrubberClient: scrMock,
			CategoryClient: catMock,
			BerlinClient:   brMock,
			DpESClient:     esMock,
		}

		cfg := &config.Config{
			EnableNLPWeighting: true,
			NLPSettings:        "",
			DefaultSort:        "relevance",
		}

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, clList, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam+nlpParamEnabled, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.Convey("Then the request should be processed OK", func() {
			c.So(resp.Code, c.ShouldEqual, http.StatusOK)
			c.So(resp.Body.String(), c.ShouldResemble, validTransformedResponse)
			c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
			c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		})

		c.Convey("And it should make calls to NLP APIs", func() {
			c.So(scrMock.GetScrubberCalls(), c.ShouldHaveLength, 1)
			c.So(catMock.GetCategoryCalls(), c.ShouldHaveLength, 1)
			c.So(brMock.GetBerlinCalls(), c.ShouldHaveLength, 1)
		})

		c.Convey("And the request to ES should be processed correctly", func() {
			c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
			actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
			c.So(actualRequest, c.ShouldResemble, expectedQuery)
		})

		c.Convey("And the response to ES should be processed correctly", func() {
			c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
			c.So(trMock.TransformSearchResponseCalls()[0].Highlight, c.ShouldBeTrue)
			actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
			c.So(actualResponse, c.ShouldResemble, validESResponse)
		})
	})

	c.Convey("When NLP features are enabled, a valid request is made but the Scrubber API is unresponsive", t, func() {
		searchBytes, _ := json.Marshal(searches)
		qbMock := newQueryBuilderMock(searchBytes, nil)

		esMock := newDpElasticSearcherMock([]byte(`{"raw":"response"}`), nil)

		brMock := newBerlinClienterMock(&brModels.Berlin{
			Matches: []brModels.Matches{
				{
					Loc: brModels.Locations{
						Subdivision: []string{
							"subdiv1",
							"subdiv2",
						},
					},
				},
			},
		}, nil)

		catMock := newCategoryClienterMock(&[]catModels.Category{
			{
				Code:  []string{"sth", "sth"},
				Score: 100,
			},
		}, nil)

		scrMock := newScrubberClienterMock(nil, scrErr.StatusError{
			Err: errors.New("Scrubber error"),
		})

		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		clList := &ClientList{
			ScrubberClient: scrMock,
			CategoryClient: catMock,
			BerlinClient:   brMock,
			DpESClient:     esMock,
		}

		cfg := &config.Config{
			EnableNLPWeighting: true,
			NLPSettings:        defaultNLPSettings,
			DefaultSort:        "relevance",
		}

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, clList, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam+nlpParamEnabled, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.Convey("Then the request should be processed OK", func() {
			c.So(resp.Code, c.ShouldEqual, http.StatusOK)
			c.So(resp.Body.String(), c.ShouldResemble, validTransformedResponse)
			c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
			c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		})

		c.Convey("Then it should make calls to NLP APIs", func() {
			c.So(scrMock.GetScrubberCalls(), c.ShouldHaveLength, 1)

			c.Convey("When the scrubber call fails, the category and berlin APIs should not be called", func() {
				c.So(catMock.GetCategoryCalls(), c.ShouldHaveLength, 0)
				c.So(brMock.GetBerlinCalls(), c.ShouldHaveLength, 0)
			})
		})

		c.Convey("Then the request to ES should be processed correctly", func() {
			c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
			actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
			c.So(actualRequest, c.ShouldResemble, expectedQuery)
		})

		c.Convey("Then the response to ES should be processed correctly", func() {
			c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
			c.So(trMock.TransformSearchResponseCalls()[0].Highlight, c.ShouldBeTrue)
			actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
			c.So(actualResponse, c.ShouldResemble, validESResponse)
		})
	})

	c.Convey("When NLP features are enabled, a valid request is made but the Berlin API is unresponsive", t, func() {
		searchBytes, _ := json.Marshal(searches)
		qbMock := newQueryBuilderMock(searchBytes, nil)

		esMock := newDpElasticSearcherMock([]byte(`{"raw":"response"}`), nil)

		brMock := newBerlinClienterMock(nil, catErr.StatusError{
			Err: errors.New("Berlin error"),
		})

		catMock := newCategoryClienterMock(&[]catModels.Category{
			{
				Code:  []string{"sth", "sth"},
				Score: 100,
			},
		}, nil)

		scrMock := newScrubberClienterMock(&scrModels.ScrubberResp{
			Results: scrModels.Results{
				Areas: []scrModels.AreaResp{
					{
						Name:   "Area1",
						Region: "region1",
					},
				},
			},
		}, nil)

		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		clList := &ClientList{
			ScrubberClient: scrMock,
			CategoryClient: catMock,
			BerlinClient:   brMock,
			DpESClient:     esMock,
		}

		cfg := &config.Config{
			EnableNLPWeighting: true,
			NLPSettings:        defaultNLPSettings,
			DefaultSort:        "relevance",
		}

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, clList, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam+nlpParamEnabled, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.Convey("Then the request should be processed OK", func() {
			c.So(resp.Code, c.ShouldEqual, http.StatusOK)
			c.So(resp.Body.String(), c.ShouldResemble, validTransformedResponse)
			c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
			c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		})

		c.Convey("Then it should make calls to NLP APIs", func() {
			c.So(scrMock.GetScrubberCalls(), c.ShouldHaveLength, 1)
			c.So(catMock.GetCategoryCalls(), c.ShouldHaveLength, 1)
			c.So(brMock.GetBerlinCalls(), c.ShouldHaveLength, 1)
		})

		c.Convey("Then the request to ES should be processed correctly", func() {
			c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
			actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
			c.So(actualRequest, c.ShouldResemble, expectedQuery)
		})

		c.Convey("Then the response to ES should be processed correctly", func() {
			c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
			c.So(trMock.TransformSearchResponseCalls()[0].Highlight, c.ShouldBeTrue)
			actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
			c.So(actualResponse, c.ShouldResemble, validESResponse)
		})
	})

	c.Convey("When NLP features are enabled, a valid request is made but the Berlin API gives an empty response", t, func() {
		searchBytes, _ := json.Marshal(searches)
		qbMock := newQueryBuilderMock(searchBytes, nil)

		esMock := newDpElasticSearcherMock([]byte(`{"raw":"response"}`), nil)

		brMock := newBerlinClienterMock(&brModels.Berlin{
			Matches: []brModels.Matches{},
		}, nil)

		catMock := newCategoryClienterMock(&[]catModels.Category{
			{
				Code:  []string{"sth", "sth"},
				Score: 100,
			},
		}, nil)

		scrMock := newScrubberClienterMock(&scrModels.ScrubberResp{
			Results: scrModels.Results{
				Areas: []scrModels.AreaResp{
					{
						Name:   "Area1",
						Region: "region1",
					},
				},
			},
		}, nil)

		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		clList := &ClientList{
			ScrubberClient: scrMock,
			CategoryClient: catMock,
			BerlinClient:   brMock,
			DpESClient:     esMock,
		}

		cfg := &config.Config{
			EnableNLPWeighting: true,
			NLPSettings:        defaultNLPSettings,
			DefaultSort:        "relevance",
		}

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, clList, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam+nlpParamEnabled, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.Convey("Then the request should be processed OK", func() {
			c.So(resp.Code, c.ShouldEqual, http.StatusOK)
			c.So(resp.Body.String(), c.ShouldResemble, validTransformedResponse)
			c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
			c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		})

		c.Convey("Then it should make calls to NLP APIs", func() {
			c.So(scrMock.GetScrubberCalls(), c.ShouldHaveLength, 1)
			c.So(catMock.GetCategoryCalls(), c.ShouldHaveLength, 1)
			c.So(brMock.GetBerlinCalls(), c.ShouldHaveLength, 1)
		})

		c.Convey("Then the request to ES should be processed correctly", func() {
			c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
			actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
			c.So(actualRequest, c.ShouldResemble, expectedQuery)
		})

		c.Convey("Then the response to ES should be processed correctly", func() {
			c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
			c.So(trMock.TransformSearchResponseCalls()[0].Highlight, c.ShouldBeTrue)
			actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
			c.So(actualResponse, c.ShouldResemble, validESResponse)
		})
	})
	c.Convey("When NLP features are enabled, a valid request is made but the Category API is unresponsive", t, func() {
		searchBytes, _ := json.Marshal(searches)
		qbMock := newQueryBuilderMock(searchBytes, nil)

		esMock := newDpElasticSearcherMock([]byte(`{"raw":"response"}`), nil)

		brMock := newBerlinClienterMock(&brModels.Berlin{
			Matches: []brModels.Matches{
				{
					Loc: brModels.Locations{
						Subdivision: []string{
							"subdiv1",
							"subdiv2",
						},
					},
				},
			},
		}, nil)

		catMock := newCategoryClienterMock(nil, catErr.StatusError{
			Err: errors.New("Category error"),
		})

		scrMock := newScrubberClienterMock(&scrModels.ScrubberResp{
			Results: scrModels.Results{
				Areas: []scrModels.AreaResp{
					{
						Name:   "Area1",
						Region: "region1",
					},
				},
			},
		}, nil)

		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		clList := &ClientList{
			ScrubberClient: scrMock,
			CategoryClient: catMock,
			BerlinClient:   brMock,
			DpESClient:     esMock,
		}

		cfg := &config.Config{
			EnableNLPWeighting: true,
			NLPSettings:        defaultNLPSettings,
			DefaultSort:        "relevance",
		}

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, clList, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam+nlpParamEnabled, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.Convey("Then the request should be processed OK", func() {
			c.So(resp.Code, c.ShouldEqual, http.StatusOK)
			c.So(resp.Body.String(), c.ShouldResemble, validTransformedResponse)
			c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
			c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		})

		c.Convey("Then it should make calls to NLP APIs", func() {
			c.So(scrMock.GetScrubberCalls(), c.ShouldHaveLength, 1)
			c.So(catMock.GetCategoryCalls(), c.ShouldHaveLength, 1)
			c.So(brMock.GetBerlinCalls(), c.ShouldHaveLength, 1)
		})

		c.Convey("Then the request to ES should be processed correctly", func() {
			c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
			actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
			c.So(actualRequest, c.ShouldResemble, expectedQuery)
		})

		c.Convey("Then the response to ES should be processed correctly", func() {
			c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
			c.So(trMock.TransformSearchResponseCalls()[0].Highlight, c.ShouldBeTrue)
			actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
			c.So(actualResponse, c.ShouldResemble, validESResponse)
		})
	})

	c.Convey("When NLP features and enabled but the API caller does not request NLP weighting", t, func() {
		searchBytes, _ := json.Marshal(searches)
		qbMock := newQueryBuilderMock(searchBytes, nil)

		esMock := newDpElasticSearcherMock([]byte(`{"raw":"response"}`), nil)

		brMock := newBerlinClienterMock(nil, nil)

		catMock := newCategoryClienterMock(nil, nil)

		scrMock := newScrubberClienterMock(nil, nil)

		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		clList := &ClientList{
			ScrubberClient: scrMock,
			CategoryClient: catMock,
			BerlinClient:   brMock,
			DpESClient:     esMock,
		}

		cfg := &config.Config{
			EnableNLPWeighting: true,
			NLPSettings:        defaultNLPSettings,
			DefaultSort:        "relevance",
		}

		searchHandler := SearchHandlerFunc(validator, qbMock, cfg, clList, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.Convey("Then the request should be processed OK", func() {
			c.So(resp.Code, c.ShouldEqual, http.StatusOK)
			c.So(resp.Body.String(), c.ShouldResemble, validTransformedResponse)
			c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
			c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		})

		c.Convey("Then it should not make calls to NLP APIs", func() {
			c.So(scrMock.GetScrubberCalls(), c.ShouldHaveLength, 0)
			c.So(catMock.GetCategoryCalls(), c.ShouldHaveLength, 0)
			c.So(brMock.GetBerlinCalls(), c.ShouldHaveLength, 0)
		})

		c.Convey("Then the request to ES should be processed correctly", func() {
			c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
			actualRequest := string(esMock.MultiSearchCalls()[0].Searches[0].Query)
			c.So(actualRequest, c.ShouldResemble, expectedQuery)
		})

		c.Convey("Then the response to ES should be processed correctly", func() {
			c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
			c.So(trMock.TransformSearchResponseCalls()[0].Highlight, c.ShouldBeTrue)
			actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
			c.So(actualResponse, c.ShouldResemble, validESResponse)
		})
	})
}

func TestSearchURIsHandlerFunc(t *testing.T) {
	searches := []client.Search{}
	c.Convey("Test SearchURIsHandlerFunc", t, func() {
		clMock := &ClientList{
			DpESClient: newDpElasticSearcherMock([]byte(`{"raw": "response", "hits": {"total": {"value": 2, "relation": "eq"}, "hits": []}}`), nil),
		}
		trMock := newResponseTransformerMock([]byte(validTransformedResponseWith2Items), nil)
		cfg := &config.Config{
			DefaultLimit:        10,
			DefaultOffset:       0,
			DefaultMaximumLimit: 100,
			DefaultSort:         "relevance",
		}
		searchBytes, _ := json.Marshal(searches)
		qbMock := newQueryBuilderMock(searchBytes, nil)
		validator := query.NewSearchQueryParamValidator()

		c.Convey("When valid URIs are provided", func() {
			urisRequest := URIsRequest{
				URIs: []string{
					"/release/1",
					"/economy/dataset/2",
				},
				Limit:  5,
				Offset: 0,
			}
			reqBody, err := json.Marshal(urisRequest)
			c.So(err, c.ShouldBeNil)

			req := httptest.NewRequest(http.MethodPost, "/search/uris", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler := SearchURIsHandlerFunc(validator, qbMock, cfg, clMock, trMock)
			handler.ServeHTTP(rr, req)

			c.So(rr.Code, c.ShouldEqual, http.StatusOK)

			var resp models.SearchResponse
			err = json.Unmarshal(rr.Body.Bytes(), &resp)
			c.So(err, c.ShouldBeNil)

			c.So(resp.Count, c.ShouldEqual, 2)

			c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
			c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
		})

		c.Convey("When request payload is invalid", func() {
			invalidReqBody := []byte(`{invalid}`) // Malformed JSON to trigger decoding error

			req := httptest.NewRequest(http.MethodPost, "/search/uris", bytes.NewReader(invalidReqBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler := SearchURIsHandlerFunc(validator, qbMock, cfg, clMock, trMock)
			handler.ServeHTTP(rr, req)

			c.So(rr.Code, c.ShouldEqual, http.StatusBadRequest)
			c.So(rr.Body.String(), c.ShouldContainSubstring, "Invalid request payload")
		})

		c.Convey("When no URIs are provided", func() {
			emptyURIsRequest := URIsRequest{
				URIs:  []string{},
				Limit: 5,
			}
			reqBody, err := json.Marshal(emptyURIsRequest)
			c.So(err, c.ShouldBeNil)

			req := httptest.NewRequest(http.MethodPost, "/search/uris", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler := SearchURIsHandlerFunc(validator, qbMock, cfg, clMock, trMock)
			handler.ServeHTTP(rr, req)

			c.So(rr.Code, c.ShouldEqual, http.StatusBadRequest)
			c.So(rr.Body.String(), c.ShouldContainSubstring, "No URIs provided")
		})

		c.Convey("When URIs are blank", func() {
			urisRequest := URIsRequest{
				URIs: []string{
					"",
				},
				Limit:  5,
				Offset: 0,
			}
			reqBody, err := json.Marshal(urisRequest)
			c.So(err, c.ShouldBeNil)

			req := httptest.NewRequest(http.MethodPost, "/search/uris", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler := SearchURIsHandlerFunc(validator, qbMock, cfg, clMock, trMock)
			handler.ServeHTTP(rr, req)

			c.So(rr.Code, c.ShouldEqual, http.StatusBadRequest)
			c.So(rr.Body.String(), c.ShouldContainSubstring, "Invalid URI")
		})

		c.Convey("When limit exceeds maximum allowed value", func() {
			urisRequest := URIsRequest{
				URIs: []string{
					"/release/1",
					"/economy/dataset/2",
				},
				Limit:  cfg.DefaultMaximumLimit + 10,
				Offset: 0,
			}
			reqBody, err := json.Marshal(urisRequest)
			c.So(err, c.ShouldBeNil)

			req := httptest.NewRequest(http.MethodPost, "/search/uris", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler := SearchURIsHandlerFunc(validator, qbMock, cfg, clMock, trMock)
			handler.ServeHTTP(rr, req)

			c.So(rr.Code, c.ShouldEqual, http.StatusOK)

			var resp models.SearchResponse
			err = json.Unmarshal(rr.Body.Bytes(), &resp)
			c.So(err, c.ShouldBeNil)

			c.So(resp.Count, c.ShouldEqual, 2)
			c.So(qbMock.BuildSearchQueryCalls()[0].Req.Size, c.ShouldEqual, cfg.DefaultMaximumLimit)
		})

		c.Convey("When a valid sort parameter is provided", func() {
			urisRequest := URIsRequest{
				URIs: []string{
					"/release/1",
					"/economy/dataset/2",
				},
				Limit:  5,
				Offset: 0,
				Sort:   "release_date_asc",
			}
			reqBody, err := json.Marshal(urisRequest)
			c.So(err, c.ShouldBeNil)

			req := httptest.NewRequest(http.MethodPost, "/search/uris", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler := SearchURIsHandlerFunc(validator, qbMock, cfg, clMock, trMock)
			handler.ServeHTTP(rr, req)

			c.So(rr.Code, c.ShouldEqual, http.StatusOK)

			var resp models.SearchResponse
			err = json.Unmarshal(rr.Body.Bytes(), &resp)
			c.So(err, c.ShouldBeNil)

			// Validate that the sort parameter was passed and processed correctly
			c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
			c.So(qbMock.BuildSearchQueryCalls()[0].Req.SortBy, c.ShouldEqual, "release_date_asc")
		})

		c.Convey("When an invalid sort parameter is provided", func() {
			urisRequest := URIsRequest{
				URIs: []string{
					"/release/1",
					"/economy/dataset/2",
				},
				Limit:  5,
				Offset: 0,
				Sort:   "a", // An invalid sort parameter
			}
			reqBody, err := json.Marshal(urisRequest)
			c.So(err, c.ShouldBeNil)

			req := httptest.NewRequest(http.MethodPost, "/search/uris", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler := SearchURIsHandlerFunc(validator, qbMock, cfg, clMock, trMock)
			handler.ServeHTTP(rr, req)

			// Expect a Bad Request status code due to invalid sort
			c.So(rr.Code, c.ShouldEqual, http.StatusOK)
			// TODO: add a check when sort validation has been refactored
		})
	})
}

func TestLegacySearchHandlerFunc(t *testing.T) {
	validator := query.NewSearchQueryParamValidator()
	cfg := &config.Config{
		DefaultLimit:        10,
		DefaultOffset:       0,
		DefaultMaximumLimit: 100,
		DefaultSort:         "relevance",
	}

	c.Convey("Should return BadRequest for invalid limit parameter", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, cfg, &ClientList{DeprecatedESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?limit=a", http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusBadRequest)
		c.So(resp.Body.String(), c.ShouldContainSubstring, "invalid limit parameter")
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 0)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 0)
	})

	c.Convey("Should return BadRequest for negative limit parameter", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, cfg, &ClientList{DeprecatedESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?limit=-1", http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusBadRequest)
		c.So(resp.Body.String(), c.ShouldContainSubstring, "invalid limit parameter")
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 0)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 0)
	})

	c.Convey("Should return BadRequest for invalid offset parameter", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, cfg, &ClientList{DeprecatedESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?offset=b", http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusBadRequest)
		c.So(resp.Body.String(), c.ShouldContainSubstring, "invalid offset parameter")
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 0)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 0)
	})

	c.Convey("Should return BadRequest for negative offset parameter", t, func() {
		qbMock := newQueryBuilderMock(nil, nil)
		esMock := newElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, cfg, &ClientList{DeprecatedESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?offset=-1", http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusBadRequest)
		c.So(resp.Body.String(), c.ShouldContainSubstring, "invalid offset parameter")
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 0)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 0)
	})

	c.Convey("Should return InternalError for errors returned from query builder", t, func() {
		qbMock := newQueryBuilderMock(nil, errors.New("Something"))
		esMock := newElasticSearcherMock(nil, nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, cfg, &ClientList{DeprecatedESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusInternalServerError)
		c.So(resp.Body.String(), c.ShouldContainSubstring, "Failed to create search query")
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 0)
	})

	c.Convey("Should return InternalError for errors returned from elastic search", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock(nil, errors.New("Something"))
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, cfg, &ClientList{DeprecatedESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusInternalServerError)
		c.So(resp.Body.String(), c.ShouldContainSubstring, "Failed to run search query")
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		c.So(actualRequest, c.ShouldResemble, validQueryDoc)
	})

	c.Convey("Should return InternalError for invalid json from elastic search", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock([]byte(`{"dummy":"response"`), nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, cfg, &ClientList{DeprecatedESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusInternalServerError)
		c.So(resp.Body.String(), c.ShouldContainSubstring, "Failed to process search query")
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		c.So(actualRequest, c.ShouldResemble, validQueryDoc)
	})

	c.Convey("Should return InternalError for transformation failures", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock(nil, errors.New("Something"))

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, cfg, &ClientList{DeprecatedESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusInternalServerError)
		c.So(resp.Body.String(), c.ShouldContainSubstring, "Failed to transform search result")
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		c.So(actualRequest, c.ShouldResemble, validQueryDoc)
		c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		c.So(actualResponse, c.ShouldResemble, validESResponse)
	})

	c.Convey("Should return OK for valid search result with raw=true", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock([]byte(`{"dummy":"response"}`), nil)
		trMock := newResponseTransformerMock(nil, nil)

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, cfg, &ClientList{DeprecatedESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam+rawTrue, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusOK)
		c.So(resp.Body.String(), c.ShouldResemble, `{"dummy":"response"}`)
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		c.So(actualRequest, c.ShouldResemble, validQueryDoc)
		c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 0)
	})

	c.Convey("Should return OK for valid search result", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, cfg, &ClientList{DeprecatedESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusOK)
		c.So(resp.Body.String(), c.ShouldResemble, validTransformedResponse)
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		c.So(actualRequest, c.ShouldResemble, validQueryDoc)
		c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
		c.So(trMock.TransformSearchResponseCalls()[0].Highlight, c.ShouldBeTrue)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		c.So(actualResponse, c.ShouldResemble, validESResponse)
	})

	c.Convey("Should return OK for valid search result with highlight = true", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, cfg, &ClientList{DeprecatedESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam+highlightTrue, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusOK)
		c.So(resp.Body.String(), c.ShouldResemble, validTransformedResponse)
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		c.So(actualRequest, c.ShouldResemble, validQueryDoc)
		c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
		c.So(trMock.TransformSearchResponseCalls()[0].Highlight, c.ShouldBeTrue)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		c.So(actualResponse, c.ShouldResemble, validESResponse)
	})

	c.Convey("Should return OK for valid search result with highlight = false", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, cfg, &ClientList{DeprecatedESClient: esMock}, trMock)

		req := httptest.NewRequest("GET", baseURL+validQueryParam+highlightFalse, http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusOK)
		c.So(resp.Body.String(), c.ShouldResemble, validTransformedResponse)
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		c.So(actualRequest, c.ShouldResemble, validQueryDoc)
		c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
		c.So(trMock.TransformSearchResponseCalls()[0].Highlight, c.ShouldBeFalse)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		c.So(actualResponse, c.ShouldResemble, validESResponse)
	})

	c.Convey("Should pass all search terms on to elastic search", t, func() {
		qbMock := newQueryBuilderMock([]byte(validQueryDoc), nil)
		esMock := newElasticSearcherMock([]byte(validESResponse), nil)
		trMock := newResponseTransformerMock([]byte(validTransformedResponse), nil)

		searchHandler := LegacySearchHandlerFunc(validator, qbMock, cfg, &ClientList{DeprecatedESClient: esMock}, trMock)

		req := httptest.NewRequest(
			"GET",
			baseURL+validQueryParam+
				"&content_type=article,release"+
				sortRelevance+
				limit1+
				offset2,
			http.NoBody)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		c.So(resp.Code, c.ShouldEqual, http.StatusOK)
		c.So(resp.Body.String(), c.ShouldResemble, validTransformedResponse)
		c.So(qbMock.BuildSearchQueryCalls(), c.ShouldHaveLength, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Term, c.ShouldResemble, validQueryParam)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Types, c.ShouldResemble, []string{"article", "release"})
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.SortBy, c.ShouldResemble, "relevance")
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.Size, c.ShouldEqual, 1)
		c.So(qbMock.BuildSearchQueryCalls()[0].Req.From, c.ShouldEqual, 2)

		c.So(esMock.MultiSearchCalls(), c.ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		c.So(actualRequest, c.ShouldResemble, validQueryDoc)

		c.So(trMock.TransformSearchResponseCalls(), c.ShouldHaveLength, 1)
		actualResponse := string(trMock.TransformSearchResponseCalls()[0].ResponseData)
		c.So(actualResponse, c.ShouldResemble, validESResponse)
	})
}

func TestCreateSearchIndexHandlerFunc(t *testing.T) {
	c.Convey("Given a Search API that is pointing to the Site Wide version of Elastic Search", t, func() {
		dpESClient := newDpElasticSearcherMock(nil, nil)

		searchAPI := &SearchAPI{
			clList: &ClientList{
				DpESClient: dpESClient,
			},
		}

		c.Convey("When a new elastic search index is created", func() {
			req := httptest.NewRequest("POST", "http://localhost:23900/search", http.NoBody)
			resp := httptest.NewRecorder()

			searchAPI.CreateSearchIndexHandlerFunc(resp, req)

			c.Convey("Then the newly created search index name is returned with status code 201", func() {
				c.So(resp.Code, c.ShouldEqual, http.StatusCreated)
				payload, err := io.ReadAll(resp.Body)
				c.So(err, c.ShouldBeNil)
				indexCreated := models.CreateIndexResponse{}
				err = json.Unmarshal(payload, &indexCreated)
				c.So(err, c.ShouldBeNil)

				c.Convey("And the index name has the expected name format", func() {
					re := regexp.MustCompile(`(ons)(\d*)`)
					indexName := indexCreated.IndexName
					c.So(indexName, c.ShouldNotBeNil)
					wordWithExpectedPattern := re.FindString(indexName)
					c.So(wordWithExpectedPattern, c.ShouldEqual, indexName)
				})
			})
		})
	})

	c.Convey("Given a Search API that is pointing to the old version of Elastic Search", t, func() {
		// The new ES client will return an error if the Search API config is pointing at the old version of ES
		dpESClient := newDpElasticSearcherMock(nil, errors.New("unexpected status code from api"))

		searchAPI := &SearchAPI{
			clList: &ClientList{
				DpESClient: dpESClient,
			},
		}

		c.Convey("When a new elastic search index is created", func() {
			req := httptest.NewRequest("POST", "http://localhost:23900/search", http.NoBody)
			resp := httptest.NewRecorder()

			searchAPI.CreateSearchIndexHandlerFunc(resp, req)

			c.Convey("Then an internal server error is returned with status code 500", func() {
				c.So(resp.Code, c.ShouldEqual, http.StatusInternalServerError)
				c.So(strings.Trim(resp.Body.String(), "\n"), c.ShouldResemble, internalServerErrMsg)
			})
		})
	})
}

func newBerlinClienterMock(response *brModels.Berlin, err brErr.Error) *berlin.ClienterMock {
	return &berlin.ClienterMock{
		GetBerlinFunc: func(ctx context.Context, options berlin.Options) (*brModels.Berlin, brErr.Error) {
			return response, err
		},
	}
}

func newCategoryClienterMock(response *[]catModels.Category, err catErr.Error) *category.ClienterMock {
	return &category.ClienterMock{
		GetCategoryFunc: func(ctx context.Context, options category.Options) (*[]catModels.Category, catErr.Error) {
			return response, err
		},
	}
}

func newScrubberClienterMock(response *scrModels.ScrubberResp, err catErr.Error) *scrMocks.ClienterMock {
	return &scrMocks.ClienterMock{
		GetScrubberFunc: func(ctx context.Context, options *scr.Options) (*scrModels.ScrubberResp, scrErr.Error) {
			return response, err
		},
	}
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
		AddNlpCategorySearchFunc: func(nlpCriteria *query.NlpCriteria, category string, subCategory string, categoryWeighting float32) *query.NlpCriteria {
			return &query.NlpCriteria{
				UseCategory: true,
				Categories: []query.NlpCriteriaCategory{
					{
						Category:    category,
						SubCategory: subCategory,
						Weighting:   categoryWeighting,
					},
				},
				UseSubdivision: true,
			}
		},
		AddNlpSubdivisionSearchFunc: func(nlpCriteria *query.NlpCriteria, subdivisionWords string) *query.NlpCriteria {
			if nlpCriteria == nil {
				nlpCriteria = new(query.NlpCriteria)
			}

			nlpCriteria.SubdivisionWords = subdivisionWords
			return nlpCriteria
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
	c.Convey("Given a query term with quoted terms", t, func() {
		queryWithQuotes := `"education results for Wales" "education results for England"`

		c.Convey("when sanitised the individual quotes in the query term should be escaped", func() {
			sanitised := sanitiseDoubleQuotes(queryWithQuotes)
			c.So(sanitised, c.ShouldEqual, `\"education results for Wales\" \"education results for England\"`)
		})
	})
}
