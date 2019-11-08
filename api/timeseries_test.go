package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTimeseriesLookupHandlerFunc(t *testing.T) {

	Convey("Should return InternalError for invalid template", t, func() {
		setupTimeseriesTestTemplates("dummy{{.Moo}}")
		esMock := &ElasticSearcherMock{
			SearchFunc: func(index string, docType string, request []byte) ([]byte, error) {
				return nil, nil
			},
		}

		timeSeriesHandler := TimeseriesLookupHandlerFunc(esMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/timeseries/a", nil)
		resp := httptest.NewRecorder()

		timeSeriesHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(resp.Body.String(), ShouldContainSubstring, "Failed to create query")
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 0)
	})

	Convey("Should return InternalError for errors returned from elastic search", t, func() {
		setupTimeseriesTestTemplates("Cdid={{.Cdid}}")
		esMock := &ElasticSearcherMock{
			SearchFunc: func(index string, docType string, request []byte) ([]byte, error) {
				return nil, errors.New("Something")
			},
		}

		timeSeriesHandler := TimeseriesLookupHandlerFunc(esMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/timeseries/a", nil)
		req = mux.SetURLVars(req, map[string]string{"cdid": "a"})
		resp := httptest.NewRecorder()

		timeSeriesHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(resp.Body.String(), ShouldContainSubstring, "Failed to run timeseries query")
		So(esMock.SearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.SearchCalls()[0].Request)
		So(actualRequest, ShouldContainSubstring, "Cdid=a")
	})

	Convey("Should return OK for valid search result", t, func() {
		setupTimeseriesTestTemplates("Cdid={{.Cdid}}")
		esMock := &ElasticSearcherMock{
			SearchFunc: func(index string, docType string, request []byte) ([]byte, error) {
				return []byte(`{"dummy":"response"}`), nil
			},
		}

		timeSeriesHandler := TimeseriesLookupHandlerFunc(esMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/timeseries/a", nil)
		req = mux.SetURLVars(req, map[string]string{"cdid": "a"})
		resp := httptest.NewRecorder()

		timeSeriesHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, `{"dummy":"response"}`)
		So(esMock.SearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.SearchCalls()[0].Request)
		So(actualRequest, ShouldContainSubstring, "Cdid=a")
	})
}

func setupTimeseriesTestTemplates(rawtemplate string) {
	temp, err := template.New("lookup.tmpl").Parse(rawtemplate)
	So(err, ShouldBeNil)
	timeseriesTemplate = temp
}
