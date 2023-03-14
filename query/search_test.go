package query

import (
	"context"
	"encoding/json"
	"testing"
	"text/template"
	"time"

	"github.com/ONSdigital/dp-elasticsearch/v3/client"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBuildSearchQuery(t *testing.T) {
	Convey("Should return InternalError for invalid template", t, func() {
		qb := createQueryBuilderForSearchTemplate("dummy{{.Moo}}")

		reqParams := &SearchRequest{
			Size: 2,
			From: 1,
		}
		query, err := qb.BuildSearchQuery(context.Background(), reqParams, false)

		So(err, ShouldNotBeNil)
		So(query, ShouldBeNil)
		So(err.Error(), ShouldContainSubstring, "creation of search from template failed")
	})

	Convey("Should include all search parameters in elastic search query", t, func() {
		qb := createQueryBuilderForSearchTemplate("Term={{.Term}};" +
			"From={{.From}};" +
			"Size={{.Size}};" +
			"Types={{.Types}};" +
			"SortBy={{.SortBy}};" +
			"AggregationField={{.AggregationField}};" +
			"Highlight={{.Highlight}};" +
			"Now={{.Now}}")

		reqParams := &SearchRequest{
			Term:      "a",
			Types:     []string{"ta", "tb"},
			SortBy:    "relevance",
			Topic:     []string{"test"},
			Size:      2,
			From:      1,
			Highlight: true,
			Now:       time.Date(2023, 03, 10, 12, 15, 04, 05, time.UTC).UTC().Format(time.RFC3339),
		}
		query, err := qb.BuildSearchQuery(context.Background(), reqParams, false)

		So(err, ShouldBeNil)
		So(query, ShouldNotBeNil)
		queryString := string(query)
		So(queryString, ShouldContainSubstring, "Term=a")
		So(queryString, ShouldContainSubstring, "From=1")
		So(queryString, ShouldContainSubstring, "Size=2")
		So(queryString, ShouldContainSubstring, "Types=[ta tb]")
		So(queryString, ShouldContainSubstring, "SortBy=relevance")
		So(queryString, ShouldContainSubstring, "AggregationField=_type")
		So(queryString, ShouldContainSubstring, "Highlight=true")
		So(queryString, ShouldContainSubstring, "Now=2023-03-10T12:15:04")
	})
}

func TestBuildSearchQueryContent(t *testing.T) {
	Convey("Should generate the expected query to be sent to elasticsearch", t, func() {
		qb, err := NewQueryBuilder()
		So(err, ShouldBeNil)

		reqParams := &SearchRequest{
			Term:      "a",
			Types:     []string{"ta", "tb"},
			SortBy:    "relevance",
			Topic:     []string{"test"},
			Size:      2,
			From:      1,
			Highlight: true,
			Now:       time.Date(2023, 03, 10, 12, 15, 04, 05, time.UTC).UTC().Format(time.RFC3339),
		}
		query, err := qb.BuildSearchQuery(context.Background(), reqParams, true)

		var searches []client.Search
		err = json.Unmarshal(query, &searches)
		So(err, ShouldBeNil)

		So(searches, ShouldHaveLength, 3)
		So(searches[0].Header, ShouldResemble, client.Header{Index: "ons"})
		So(string(searches[0].Query), ShouldEqual, `{"from":1,"size":2,"query":{"bool":{"must":{"function_score":{"query":{"dis_max":{"queries":[{"bool":{"should":[{"match":{"title.title_no_dates":{"query":"a","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"match":{"title.title_no_stem":{"query":"a","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"multi_match":{"query":"a","fields":["title^10","edition","downloads.content^1"],"type":"cross_fields","minimum_should_match":"3<80% 5<60%"}},{"multi_match":{"query":"a","fields":["title^10","summary","metaDescription","edition","downloads.content^1","pageData^1","keywords"],"type":"phrase","boost":10.0,"slop":2}}]}},{"multi_match":{"query":"a","fields":["summary","metaDescription","downloads.content^1","pageData^1","keywords"],"type":"best_fields","minimum_should_match":"75%"}},{"match":{"keywords":{"query":"a","operator":"AND","boost":10.0}}},{"multi_match":{"query":"a","fields":["cdid","dataset_id"]}},{"match":{"searchBoost":{"query":"a","operator":"AND","boost":100.0}}}]}},"functions":[{"filter":{"term":{"type":"bulletin"}},"weight":100},{"filter":{"term":{"type":"dataset_landing_page"}},"weight":70},{"filter":{"terms":{"type":["article","compendium_landing_page","article_download"]}},"weight":50},{"filter":{"term":{"type":"static_adhoc"}},"weight":30},{"filter":{"term":{"type":"timeseries"}},"weight":10}]}},"filter":[{"bool":{"must":[{"bool":{"should":[{"match":{"type":"ta"}},{"match":{"type":"tb"}}]}},{"bool":{"should":[{"match":{"canonical_topic":"test"}},{"match":{"topics":"test"}}]}}]}}]}},"suggest":{"search_suggest":{"text":"a","phrase":{"field":"title.title_no_synonym_no_stem"}}},"_source":{"includes":[],"excludes":["downloads.content","downloads*","pageData"]},"highlight":{"pre_tags":["<em class=\"ons-highlight\">"],"post_tags":["</em>"],"fields":{"terms":{"fragment_size":0,"number_of_fragments":0},"title":{"fragment_size":0,"number_of_fragments":0},"edition":{"fragment_size":0,"number_of_fragments":0},"summary":{"fragment_size":0,"number_of_fragments":0},"meta_description":{"fragment_size":0,"number_of_fragments":0},"keywords":{"fragment_size":0,"number_of_fragments":0},"cdid":{"fragment_size":0,"number_of_fragments":0},"dataset_id":{"fragment_size":0,"number_of_fragments":0},"downloads.content":{"fragment_size":45,"number_of_fragments":5},"pageData":{"fragment_size":45,"number_of_fragments":5}}},"sort":[{"_score":{"order":"desc"}},{"release_date":{"order":"desc"}}]}`)

		So(searches[1].Header, ShouldResemble, client.Header{Index: "ons"})
		So(string(searches[1].Query), ShouldEqual, `{"query":{"bool":{"must":{"dis_max":{"queries":[{"bool":{"should":[{"match":{"title.title_no_dates":{"query":"a","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"match":{"title.title_no_stem":{"query":"a","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"multi_match":{"query":"a","fields":["title^10","edition","downloads.content^1"],"type":"cross_fields","minimum_should_match":"3<80% 5<60%"}},{"multi_match":{"query":"a","fields":["title^10","summary","metaDescription","edition","downloads.content^1","pageData^1","keywords"],"type":"phrase","boost":10.0,"slop":2}}]}},{"multi_match":{"query":"a","fields":["summary","metaDescription","downloads.content^1","pageData^1","keywords"],"type":"best_fields","minimum_should_match":"75%"}},{"match":{"keywords":{"query":"a","operator":"AND","boost":10.0}}},{"multi_match":{"query":"a","fields":["cdid","dataset_id"]}},{"match":{"searchBoost":{"query":"a","operator":"AND","boost":100.0}}}]}}}},"size":0,"aggregations":{"contentTypeCounts":{"terms":{"size":1000,"field":"type"}}}}`)

		So(searches[2].Header, ShouldResemble, client.Header{Index: "ons"})
		So(string(searches[2].Query), ShouldEqual, `{"query":{"bool":{"must":{"dis_max":{"queries":[{"bool":{"should":[{"match":{"title.title_no_dates":{"query":"a","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"match":{"title.title_no_stem":{"query":"a","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"multi_match":{"query":"a","fields":["title^10","edition","downloads.content^1"],"type":"cross_fields","minimum_should_match":"3<80% 5<60%"}},{"multi_match":{"query":"a","fields":["title^10","summary","metaDescription","edition","downloads.content^1","pageData^1","keywords"],"type":"phrase","boost":10.0,"slop":2}}]}},{"multi_match":{"query":"a","fields":["summary","metaDescription","downloads.content^1","pageData^1","keywords"],"type":"best_fields","minimum_should_match":"75%"}},{"match":{"keywords":{"query":"a","operator":"AND","boost":10.0}}},{"multi_match":{"query":"a","fields":["cdid","dataset_id"]}},{"match":{"searchBoost":{"query":"a","operator":"AND","boost":100.0}}}]}}}},"size":0,"aggregations":{"topicCounts":{"terms":{"size":1000,"field":"topics"}}}}`)
	})
}

func TestBuildSearchQueryPopulationType(t *testing.T) {
	Convey("Given a Query builder", t, func() {
		qb, err := NewQueryBuilder()
		So(err, ShouldBeNil)

		Convey("Then the expected search query is generated for a full population type request", func() {
			reqParams := &SearchRequest{
				Size: 2,
				PopulationType: &PopulationTypeRequest{
					Name:  "UR",
					Label: "usual residents",
				},
			}
			query, err := qb.BuildSearchQuery(context.Background(), reqParams, true)
			So(err, ShouldBeNil)

			searches := unmarshal(query)
			expectedQueryString := `{"size":2,"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}},{"bool":{"should":[{"match":{"population_type.name":{"query":"UR"}}},{"match":{"population_type.label":{"query":"usual residents"}}}]}}]}}]}},"suggest":{"search_suggest":{"text":"","phrase":{"field":"title.title_no_synonym_no_stem"}}},"_source":{"includes":[],"excludes":["downloads.content","downloads*","pageData"]},"sort":[{"_score":{"order":"desc"}},{"release_date":{"order":"desc"}}]}`

			So(searches, ShouldHaveLength, 3)
			So(searches[0].Header, ShouldResemble, client.Header{Index: "ons"})
			So(string(searches[0].Query), ShouldEqual, expectedQueryString)
		})

		Convey("Then the expected search query is generated for a population type name-only request", func() {
			reqParams := &SearchRequest{
				Size: 2,
				PopulationType: &PopulationTypeRequest{
					Name: "UR",
				},
			}
			query, err := qb.BuildSearchQuery(context.Background(), reqParams, true)
			So(err, ShouldBeNil)

			searches := unmarshal(query)
			expectedQueryString := `{"size":2,"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}},{"bool":{"should":[{"match":{"population_type.name":{"query":"UR"}}},{"match":{"population_type.label":{"query":""}}}]}}]}}]}},"suggest":{"search_suggest":{"text":"","phrase":{"field":"title.title_no_synonym_no_stem"}}},"_source":{"includes":[],"excludes":["downloads.content","downloads*","pageData"]},"sort":[{"_score":{"order":"desc"}},{"release_date":{"order":"desc"}}]}`

			So(searches, ShouldHaveLength, 3)
			So(searches[0].Header, ShouldResemble, client.Header{Index: "ons"})
			So(string(searches[0].Query), ShouldEqual, expectedQueryString)
		})

		Convey("Then the expected search query is generated for a population type label-only request", func() {
			reqParams := &SearchRequest{
				Size: 2,
				PopulationType: &PopulationTypeRequest{
					Label: "usual residents",
				},
			}
			query, err := qb.BuildSearchQuery(context.Background(), reqParams, true)
			So(err, ShouldBeNil)

			searches := unmarshal(query)
			expectedQueryString := `{"size":2,"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}},{"bool":{"should":[{"match":{"population_type.label":{"query":"usual residents"}}}]}}]}}]}},"suggest":{"search_suggest":{"text":"","phrase":{"field":"title.title_no_synonym_no_stem"}}},"_source":{"includes":[],"excludes":["downloads.content","downloads*","pageData"]},"sort":[{"_score":{"order":"desc"}},{"release_date":{"order":"desc"}}]}`

			So(searches, ShouldHaveLength, 3)
			So(searches[0].Header, ShouldResemble, client.Header{Index: "ons"})
			So(string(searches[0].Query), ShouldEqual, expectedQueryString)
		})
	})
}

func TestBuildSearchQueryDimensions(t *testing.T) {
	Convey("Given a Query builder", t, func() {
		qb, err := NewQueryBuilder()
		So(err, ShouldBeNil)

		Convey("Then the expected search query is generated for a full dimension request", func() {
			reqParams := &SearchRequest{
				Size: 2,
				Dimensions: []*DimensionRequest{
					{
						Name:     "workplace_travel_4a",
						Label:    "Distance travelled to work",
						RawLabel: "Distance travelled to work (4 categories)",
					},
				},
			}
			query, err := qb.BuildSearchQuery(context.Background(), reqParams, true)
			So(err, ShouldBeNil)

			searches := unmarshal(query)
			expectedQueryString := `{"size":2,"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}},{"bool":{"should":[{"match":{"dimensions.name":"workplace_travel_4a"}},{"match":{"dimensions.label":"Distance travelled to work"}},{"match":{"dimensions.raw_label":"Distance travelled to work (4 categories)"}}]}}]}}]}},"suggest":{"search_suggest":{"text":"","phrase":{"field":"title.title_no_synonym_no_stem"}}},"_source":{"includes":[],"excludes":["downloads.content","downloads*","pageData"]},"sort":[{"_score":{"order":"desc"}},{"release_date":{"order":"desc"}}]}`

			So(searches, ShouldHaveLength, 3)
			So(searches[0].Header, ShouldResemble, client.Header{Index: "ons"})
			So(string(searches[0].Query), ShouldEqual, expectedQueryString)
		})

		Convey("Then the expected search query is generated for a name-only dimension request", func() {
			reqParams := &SearchRequest{
				Size: 2,
				Dimensions: []*DimensionRequest{
					{
						Name: "workplace_travel_4a",
					},
				},
			}
			query, err := qb.BuildSearchQuery(context.Background(), reqParams, true)
			So(err, ShouldBeNil)

			searches := unmarshal(query)
			expectedQueryString := `{"size":2,"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}},{"bool":{"should":[{"match":{"dimensions.name":"workplace_travel_4a"}},{"match":{"dimensions.label":""}},{"match":{"dimensions.raw_label":""}}]}}]}}]}},"suggest":{"search_suggest":{"text":"","phrase":{"field":"title.title_no_synonym_no_stem"}}},"_source":{"includes":[],"excludes":["downloads.content","downloads*","pageData"]},"sort":[{"_score":{"order":"desc"}},{"release_date":{"order":"desc"}}]}`

			So(searches, ShouldHaveLength, 3)
			So(searches[0].Header, ShouldResemble, client.Header{Index: "ons"})
			So(string(searches[0].Query), ShouldEqual, expectedQueryString)
		})

		Convey("Then the expected search query is generated for a label-only dimension request", func() {
			reqParams := &SearchRequest{
				Size: 2,
				Dimensions: []*DimensionRequest{
					{
						Label: "Distance travelled to work",
					},
				},
			}
			query, err := qb.BuildSearchQuery(context.Background(), reqParams, true)
			So(err, ShouldBeNil)

			searches := unmarshal(query)
			expectedQueryString := `{"size":2,"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}},{"bool":{"should":[{"match":{"dimensions.label":"Distance travelled to work"}},{"match":{"dimensions.raw_label":""}}]}}]}}]}},"suggest":{"search_suggest":{"text":"","phrase":{"field":"title.title_no_synonym_no_stem"}}},"_source":{"includes":[],"excludes":["downloads.content","downloads*","pageData"]},"sort":[{"_score":{"order":"desc"}},{"release_date":{"order":"desc"}}]}`

			So(searches, ShouldHaveLength, 3)
			So(searches[0].Header, ShouldResemble, client.Header{Index: "ons"})
			So(string(searches[0].Query), ShouldEqual, expectedQueryString)
		})

		Convey("Then the expected search query is generated for a raw-label-only dimension request", func() {
			reqParams := &SearchRequest{
				Size: 2,
				Dimensions: []*DimensionRequest{
					{
						RawLabel: "Distance travelled to work",
					},
				},
			}
			query, err := qb.BuildSearchQuery(context.Background(), reqParams, true)
			So(err, ShouldBeNil)

			searches := unmarshal(query)
			expectedQueryString := `{"size":2,"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}},{"bool":{"should":[{"match":{"dimensions.label":""}},{"match":{"dimensions.raw_label":"Distance travelled to work"}}]}}]}}]}},"suggest":{"search_suggest":{"text":"","phrase":{"field":"title.title_no_synonym_no_stem"}}},"_source":{"includes":[],"excludes":["downloads.content","downloads*","pageData"]},"sort":[{"_score":{"order":"desc"}},{"release_date":{"order":"desc"}}]}`

			So(searches, ShouldHaveLength, 3)
			So(searches[0].Header, ShouldResemble, client.Header{Index: "ons"})
			So(string(searches[0].Query), ShouldEqual, expectedQueryString)
		})
	})
}

func TestBuildCountQuery(t *testing.T) {
	Convey("Should return InternalError for invalid template", t, func() {
		qb := createQueryBuilderForCountTemplate("dummy{{.Moo}}")

		query, err := qb.BuildCountQuery(context.Background(), "someQuery")

		So(err, ShouldNotBeNil)
		So(query, ShouldBeNil)
		So(err.Error(), ShouldContainSubstring, "creation of search from template failed")
	})
}

func TestBuildCountQueryPopulationType(t *testing.T) {
	Convey("Given a Query builder", t, func() {
		qb, err := NewQueryBuilder()
		So(err, ShouldBeNil)

		Convey("Then the expected count query is generated for a full population type request", func() {
			query, err := qb.BuildCountQuery(context.Background(), "abc")
			So(err, ShouldBeNil)
			So(query, ShouldNotBeEmpty)
		})
	})
}

func createQueryBuilderForSearchTemplate(rawTemplate string) *Builder {
	temp, err := template.New("search.tmpl").Parse(rawTemplate)
	So(err, ShouldBeNil)
	return &Builder{
		searchTemplates: temp,
	}
}

func createQueryBuilderForCountTemplate(rawTemplate string) *Builder {
	temp, err := template.New("count.tmpl").Parse(rawTemplate)
	So(err, ShouldBeNil)
	return &Builder{
		countTemplates: temp,
	}
}

func unmarshal(query []byte) []client.Search {
	var searches []client.Search
	err := json.Unmarshal(query, &searches)
	So(err, ShouldBeNil)
	return searches
}
