package api

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDataLookupHandlerFunc(t *testing.T) {

	Convey("Should return InternalError for invalid template", t, func() {
		setupDataTestTemplates("dummy{{.Moo}}")
		esMock := &ElasticSearcherMock{
			SearchFunc: func(ctx context.Context, index string, docType string, request []byte) ([]byte, error) {
				return nil, nil
			},
		}

		dataHandler := DataLookupHandlerFunc(esMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/data", nil)
		resp := httptest.NewRecorder()

		dataHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(resp.Body.String(), ShouldContainSubstring, "Failed to create query")
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 0)
	})

	Convey("Should return InternalError for errors returned from elastic search", t, func() {
		setupDataTestTemplates("Uris={{.Uris}};Types={{.Types}};")
		esMock := &ElasticSearcherMock{
			SearchFunc: func(ctx context.Context, index string, docType string, request []byte) ([]byte, error) {
				return nil, errors.New("Something")
			},
		}

		dataHandler := DataLookupHandlerFunc(esMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/data?uris=u&types=t", nil)
		resp := httptest.NewRecorder()

		dataHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(resp.Body.String(), ShouldContainSubstring, "Failed to run data query")
		So(esMock.SearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.SearchCalls()[0].Request)
		So(actualRequest, ShouldContainSubstring, "Uris=[u]")
		So(actualRequest, ShouldContainSubstring, "Types=[t]")
	})

	Convey("Should return InternalError for invalid json from elastic search", t, func() {
		setupDataTestTemplates("Uris={{.Uris}};Types={{.Types}};")
		esMock := &ElasticSearcherMock{
			SearchFunc: func(ctx context.Context, index string, docType string, request []byte) ([]byte, error) {
				return []byte(`{"dummy":"response"`), nil
			},
		}

		dataHandler := DataLookupHandlerFunc(esMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/data?uris=u&types=t", nil)
		resp := httptest.NewRecorder()

		dataHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(resp.Body.String(), ShouldContainSubstring, "Failed to process data query")
		So(esMock.SearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.SearchCalls()[0].Request)
		So(actualRequest, ShouldContainSubstring, "Uris=[u]")
		So(actualRequest, ShouldContainSubstring, "Types=[t]")
	})

	Convey("Should return OK for valid search result", t, func() {
		setupDataTestTemplates("Uris={{.Uris}};Types={{.Types}};")
		esMock := &ElasticSearcherMock{
			SearchFunc: func(ctx context.Context, index string, docType string, request []byte) ([]byte, error) {
				return []byte(`{"dummy":"response"}`), nil
			},
		}

		dataHandler := DataLookupHandlerFunc(esMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/data?uris=u&types=t", nil)
		resp := httptest.NewRecorder()

		dataHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, `{"responses":[{"dummy":"response"}]}`)
		So(esMock.SearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.SearchCalls()[0].Request)
		So(actualRequest, ShouldContainSubstring, "Uris=[u]")
		So(actualRequest, ShouldContainSubstring, "Types=[t]")
	})

	Convey("Should pass multiple terms to elastic search", t, func() {
		setupDataTestTemplates("Uris={{.Uris}};Types={{.Types}};")
		esMock := &ElasticSearcherMock{
			SearchFunc: func(ctx context.Context, index string, docType string, request []byte) ([]byte, error) {
				return []byte(`{"dummy":"response"}`), nil
			},
		}

		dataHandler := DataLookupHandlerFunc(esMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/data?uris=ua&uris=ub&types=t1&types=t2", nil)
		resp := httptest.NewRecorder()

		dataHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, `{"responses":[{"dummy":"response"}]}`)
		So(esMock.SearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.SearchCalls()[0].Request)
		So(actualRequest, ShouldContainSubstring, "Uris=[ua ub]")
		So(actualRequest, ShouldContainSubstring, "Types=[t1 t2]")
	})
}

func setupDataTestTemplates(rawtemplate string) {
	temp, err := template.New("queryByUri.tmpl").Parse(rawtemplate)
	So(err, ShouldBeNil)
	dataTemplates = temp
}
