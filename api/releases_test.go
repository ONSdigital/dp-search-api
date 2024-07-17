package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-elasticsearch/v3/client"
	"github.com/ONSdigital/dp-search-api/query"
	"github.com/smartystreets/goconvey/convey"
)

func TestSearchReleasesHandlerFunc(t *testing.T) {
	validator := query.NewReleaseQueryParamValidator()
	builder := &ReleaseQueryBuilderMock{
		BuildSearchQueryFunc: func(ctx context.Context, request interface{}) ([]client.Search, error) {
			return []client.Search{
				{Query: []byte(`{"query": "test"}`)},
			}, nil
		},
	}
	searcher := &DpElasticSearcherMock{
		MultiSearchFunc: func(ctx context.Context, searches []client.Search, params *client.QueryParams) ([]byte, error) {
			searchResponse := map[string]interface{}{
				"dummy": "response",
			}
			return json.Marshal(searchResponse)
		},
	}
	transformer := &ReleaseResponseTransformerMock{
		TransformSearchResponseFunc: func(ctx context.Context, responseData []byte, req query.ReleaseSearchRequest, highlight bool) ([]byte, error) {
			return responseData, nil
		},
	}

	searchHandler := SearchReleasesHandlerFunc(validator, builder, searcher, transformer)

	convey.Convey("Should return BadRequest for invalid limit parameter", t, func() {
		req := httptest.NewRequest("GET", "http://localhost:8080/search/releases?limit=test", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		convey.So(resp.Code, convey.ShouldEqual, http.StatusBadRequest)
		convey.So(resp.Body.String(), convey.ShouldContainSubstring, "Invalid limit parameter")
	})

	convey.Convey("Should return BadRequest for invalid offset parameter", t, func() {
		req := httptest.NewRequest("GET", "http://localhost:8080/search/releases?offset=test", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		convey.So(resp.Code, convey.ShouldEqual, http.StatusBadRequest)
		convey.So(resp.Body.String(), convey.ShouldContainSubstring, "Invalid offset parameter")
	})

	convey.Convey("Should return BadRequest for invalid sort parameter", t, func() {
		req := httptest.NewRequest("GET", "http://localhost:8080/search/releases?sort=test", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		convey.So(resp.Code, convey.ShouldEqual, http.StatusBadRequest)
		convey.So(resp.Body.String(), convey.ShouldContainSubstring, "Invalid sort parameter")
	})

	convey.Convey("Should return valid response for correct parameters", t, func() {
		req := httptest.NewRequest("GET", "http://localhost:8080/search/releases?query=test", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		convey.So(resp.Code, convey.ShouldEqual, http.StatusOK)
		convey.So(resp.Body.String(), convey.ShouldContainSubstring, `{"dummy":"response"}`)
	})
}
