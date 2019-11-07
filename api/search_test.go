package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHasQuery(t *testing.T) {
	Convey("When called with multiple query values", t, func() {
		sr := searchRequest{
			Queries: []string{"moo", "quack"},
		}

		Convey("Should return true for included queries", func() {
			So(sr.HasQuery("moo"), ShouldBeTrue)
			So(sr.HasQuery("quack"), ShouldBeTrue)
		})
		Convey("Should return false for excluded queries", func() {
			So(sr.HasQuery("oink"), ShouldBeFalse)

		})
	})

	Convey("Should return false when called with zero query values", t, func() {
		sr := searchRequest{
			Queries: []string{},
		}
		So(sr.HasQuery("oink"), ShouldBeFalse)

	})
}

func TestSearchHandlerFunc(t *testing.T) {

	Convey("Should return BadRequest for invalid size parameter", t, func() {
		setupTestTemplates("dummy")
		esMock := &ElasticSearcherMock{
			MultiSearchFunc: func(index string, docType string, request []byte) ([]byte, error) {
				return nil, nil
			},
		}

		searchHandler := SearchHandlerFunc(esMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?size=a", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusBadRequest)
		So(resp.Body.String(), ShouldContainSubstring, "Invalid size paramater")
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 0)
	})

	Convey("Should return BadRequest for invalid from parameter", t, func() {
		setupTestTemplates("dummy")
		esMock := &ElasticSearcherMock{
			MultiSearchFunc: func(index string, docType string, request []byte) ([]byte, error) {
				return nil, nil
			},
		}

		searchHandler := SearchHandlerFunc(esMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?from=b", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusBadRequest)
		So(resp.Body.String(), ShouldContainSubstring, "Invalid from paramater")
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 0)
	})

	Convey("Should return InternalError for invalid template", t, func() {
		setupTestTemplates("dummy{{.Moo}}")
		esMock := &ElasticSearcherMock{
			MultiSearchFunc: func(index string, docType string, request []byte) ([]byte, error) {
				return nil, nil
			},
		}

		searchHandler := SearchHandlerFunc(esMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(resp.Body.String(), ShouldContainSubstring, "Failed to create query")
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 0)
	})

	Convey("Should return InternalError for errors returned from elastic search", t, func() {
		setupTestTemplates("term={{.Term}};")
		esMock := &ElasticSearcherMock{
			MultiSearchFunc: func(index string, docType string, request []byte) ([]byte, error) {
				return nil, errors.New("Something")
			},
		}

		searchHandler := SearchHandlerFunc(esMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?term=a", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(resp.Body.String(), ShouldContainSubstring, "Failed to run search query")
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		So(actualRequest, ShouldContainSubstring, "term=a")
	})

	Convey("Should return OK for valid search result", t, func() {
		setupTestTemplates("term={{.Term}};")
		esMock := &ElasticSearcherMock{
			MultiSearchFunc: func(index string, docType string, request []byte) ([]byte, error) {
				return []byte(`{"dummy":"response"}`), nil
			},
		}

		searchHandler := SearchHandlerFunc(esMock)

		req := httptest.NewRequest("GET", "http://localhost:8080/search?term=a", nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, `{"dummy":"response"}`)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		So(actualRequest, ShouldContainSubstring, "term=a")
	})

	Convey("Should pass all search terms on to elastic search", t, func() {
		setupTestTemplates("Term={{.Term}};" +
			"From={{.From}};" +
			"Size={{.Size}};" +
			"Types={{.Types}};" +
			"Index={{.Index}};" +
			"Queries={{.Queries}};" +
			"SortBy={{.SortBy}};" +
			"AggregationField={{.AggregationField}};" +
			"Highlight={{.Highlight}};" +
			"FilterOnLatest={{.FilterOnLatest}};" +
			"FilterOnFirstLetter={{.FilterOnFirstLetter}};" +
			"ReleasedAfter={{.ReleasedAfter}};" +
			"ReleasedBefore={{.ReleasedBefore}};" +
			"UriPrefix={{.UriPrefix}};" +
			"Topic={{.Topic}};" +
			"TopicWildcard={{.TopicWildcard}};" +
			"Upcoming={{.Upcoming}};" +
			"Published={{.Published}};" +
			"Now={{.Now}}")
		esMock := &ElasticSearcherMock{
			MultiSearchFunc: func(index string, docType string, request []byte) ([]byte, error) {
				return []byte(`{"dummy":"response"}`), nil
			},
		}

		searchHandler := SearchHandlerFunc(esMock)

		req := httptest.NewRequest(
			"GET",
			"http://localhost:8080/search?term=a"+
				"&from=1"+
				"&size=2"+
				"&type=ta"+
				"&type=tb"+
				"&index=i"+
				"&sort=date"+
				"&query=qa"+
				"&query=qb"+
				"&aggField=_af"+
				"&highlight=false"+
				"&latest=true"+
				"&withFirstLetter=f"+
				"&releasedAfter=ra"+
				"&releasedBefore=rb"+
				"&uriPrefix=u"+
				"&topic=t1"+
				"&topic=t2"+
				"&topicWildcard=tw1"+
				"&topicWildcard=tw2"+
				"&upcoming=true"+
				"&published=true",
			nil)
		resp := httptest.NewRecorder()

		searchHandler.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Body.String(), ShouldResemble, `{"dummy":"response"}`)
		So(esMock.MultiSearchCalls(), ShouldHaveLength, 1)
		actualRequest := string(esMock.MultiSearchCalls()[0].Request)
		So(actualRequest, ShouldContainSubstring, "Term=a")
		So(actualRequest, ShouldContainSubstring, "From=1")
		So(actualRequest, ShouldContainSubstring, "Size=2")
		So(actualRequest, ShouldContainSubstring, "Types=[ta tb]")
		So(actualRequest, ShouldContainSubstring, "Index=i")
		So(actualRequest, ShouldContainSubstring, "SortBy=date")
		So(actualRequest, ShouldContainSubstring, "Queries=[qa qb]")
		So(actualRequest, ShouldContainSubstring, "AggregationField=_af")
		So(actualRequest, ShouldContainSubstring, "Highlight=false")
		So(actualRequest, ShouldContainSubstring, "FilterOnLatest=true")
		So(actualRequest, ShouldContainSubstring, "FilterOnFirstLetter=f")
		So(actualRequest, ShouldContainSubstring, "ReleasedAfter=ra")
		So(actualRequest, ShouldContainSubstring, "ReleasedBefore=rb")
		So(actualRequest, ShouldContainSubstring, "UriPrefix=u")
		So(actualRequest, ShouldContainSubstring, "Topic=[t1 t2]")
		So(actualRequest, ShouldContainSubstring, "TopicWildcard=[tw1 tw2]")
		So(actualRequest, ShouldContainSubstring, "Upcoming=true")
		So(actualRequest, ShouldContainSubstring, "Published=true")
		So(actualRequest, ShouldContainSubstring, "Now=20")

	})
}

func setupTestTemplates(rawtemplate string) {
	temp, _ := template.New("search.tmpl").Parse(rawtemplate)
	searchTemplates = temp
}
