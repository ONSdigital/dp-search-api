package query

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
		So(err, ShouldBeNil)

		var searches []client.Search
		err = json.Unmarshal(query, &searches)
		So(err, ShouldBeNil)

		So(searches, ShouldHaveLength, 5)
		So(searches[0].Header, ShouldResemble, client.Header{Index: "ons"})
		So(string(searches[0].Query), ShouldEqual, `{"from":1,"size":2,"query":{"bool":{"must":{"function_score":{"query":{"dis_max":{"queries":[{"bool":{"should":[{"match":{"title.title_no_dates":{"query":"a","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"match":{"title.title_no_stem":{"query":"a","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"multi_match":{"query":"a","fields":["title^10","edition","downloads.content^1"],"type":"cross_fields","minimum_should_match":"3<80% 5<60%"}},{"multi_match":{"query":"a","fields":["title^10","summary","metaDescription","edition","downloads.content^1","pageData^1","keywords"],"type":"phrase","boost":10.0,"slop":2}}]}},{"multi_match":{"query":"a","fields":["summary","metaDescription","downloads.content^1","pageData^1","keywords"],"type":"best_fields","minimum_should_match":"75%"}},{"match":{"keywords":{"query":"a","operator":"AND","boost":10.0}}},{"multi_match":{"query":"a","fields":["cdid","dataset_id"]}},{"match":{"searchBoost":{"query":"a","operator":"AND","boost":100.0}}}]}},"functions":[{"filter":{"term":{"type":"bulletin"}},"weight":100},{"filter":{"term":{"type":"dataset_landing_page"}},"weight":70},{"filter":{"terms":{"type":["article","compendium_landing_page","article_download"]}},"weight":50},{"filter":{"term":{"type":"static_adhoc"}},"weight":30},{"filter":{"term":{"type":"timeseries"}},"weight":10}]}},"filter":[{"bool":{"must":[{"bool":{"should":[{"match":{"type":"ta"}},{"match":{"type":"tb"}}]}},{"bool":{"should":[{"range":{"release_date":{"gte":null,"lte":null}}}]}},{"bool":{"should":[{"match":{"canonical_topic":"test"}},{"match":{"topics":"test"}}]}}]}}]}},"suggest":{"search_suggest":{"text":"a","phrase":{"field":"title.title_no_synonym_no_stem"}}},"_source":{"includes":[],"excludes":["downloads.content","downloads*","pageData"]},"highlight":{"pre_tags":["<em class=\"ons-highlight\">"],"post_tags":["</em>"],"fields":{"terms":{"fragment_size":0,"number_of_fragments":0},"title":{"fragment_size":0,"number_of_fragments":0},"edition":{"fragment_size":0,"number_of_fragments":0},"summary":{"fragment_size":0,"number_of_fragments":0},"meta_description":{"fragment_size":0,"number_of_fragments":0},"keywords":{"fragment_size":0,"number_of_fragments":0},"cdid":{"fragment_size":0,"number_of_fragments":0},"dataset_id":{"fragment_size":0,"number_of_fragments":0},"downloads.content":{"fragment_size":45,"number_of_fragments":5},"pageData":{"fragment_size":45,"number_of_fragments":5}}},"sort":[{"_score":{"order":"desc"}},{"release_date":{"order":"desc"}}]}`)

		So(searches[1].Header, ShouldResemble, client.Header{Index: "ons"})
		So(string(searches[1].Query), ShouldEqual, `{"query":{"bool":{"must":{"dis_max":{"queries":[{"bool":{"should":[{"match":{"title.title_no_dates":{"query":"a","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"match":{"title.title_no_stem":{"query":"a","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"multi_match":{"query":"a","fields":["title^10","edition","downloads.content^1"],"type":"cross_fields","minimum_should_match":"3<80% 5<60%"}},{"multi_match":{"query":"a","fields":["title^10","summary","metaDescription","edition","downloads.content^1","pageData^1","keywords"],"type":"phrase","boost":10.0,"slop":2}}]}},{"multi_match":{"query":"a","fields":["summary","metaDescription","downloads.content^1","pageData^1","keywords"],"type":"best_fields","minimum_should_match":"75%"}},{"match":{"keywords":{"query":"a","operator":"AND","boost":10.0}}},{"multi_match":{"query":"a","fields":["cdid","dataset_id"]}},{"match":{"searchBoost":{"query":"a","operator":"AND","boost":100.0}}}]}},"filter":[{"bool":{"must":[{"bool":{"should":[{"match":{"type":"ta"}},{"match":{"type":"tb"}}]}}]}}]}},"size":0,"aggregations":{"topic":{"terms":{"size":1000,"field":"topics"}}}}`)

		So(searches[2].Header, ShouldResemble, client.Header{Index: "ons"})
		So(string(searches[2].Query), ShouldEqual, `{"query":{"bool":{"must":{"dis_max":{"queries":[{"bool":{"should":[{"match":{"title.title_no_dates":{"query":"a","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"match":{"title.title_no_stem":{"query":"a","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"multi_match":{"query":"a","fields":["title^10","edition","downloads.content^1"],"type":"cross_fields","minimum_should_match":"3<80% 5<60%"}},{"multi_match":{"query":"a","fields":["title^10","summary","metaDescription","edition","downloads.content^1","pageData^1","keywords"],"type":"phrase","boost":10.0,"slop":2}}]}},{"multi_match":{"query":"a","fields":["summary","metaDescription","downloads.content^1","pageData^1","keywords"],"type":"best_fields","minimum_should_match":"75%"}},{"match":{"keywords":{"query":"a","operator":"AND","boost":10.0}}},{"multi_match":{"query":"a","fields":["cdid","dataset_id"]}},{"match":{"searchBoost":{"query":"a","operator":"AND","boost":100.0}}}]}},"filter":[{"bool":{"must":[{"bool":{"should":[{"match":{"canonical_topic":"test"}},{"match":{"topics":"test"}}]}}]}}]}},"size":0,"aggregations":{"content_types":{"terms":{"size":1000,"field":"type"}}}}`)

		So(searches[3].Header, ShouldResemble, client.Header{Index: "ons"})
		So(string(searches[3].Query), ShouldEqual, `{"query":{"bool":{"must":{"dis_max":{"queries":[{"bool":{"should":[{"match":{"title.title_no_dates":{"query":"a","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"match":{"title.title_no_stem":{"query":"a","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"multi_match":{"query":"a","fields":["title^10","edition","downloads.content^1"],"type":"cross_fields","minimum_should_match":"3<80% 5<60%"}},{"multi_match":{"query":"a","fields":["title^10","summary","metaDescription","edition","downloads.content^1","pageData^1","keywords"],"type":"phrase","boost":10.0,"slop":2}}]}},{"multi_match":{"query":"a","fields":["summary","metaDescription","downloads.content^1","pageData^1","keywords"],"type":"best_fields","minimum_should_match":"75%"}},{"match":{"keywords":{"query":"a","operator":"AND","boost":10.0}}},{"multi_match":{"query":"a","fields":["cdid","dataset_id"]}},{"match":{"searchBoost":{"query":"a","operator":"AND","boost":100.0}}}]}},"filter":[{"bool":{"must":[{"bool":{"should":[{"match":{"type":"ta"}},{"match":{"type":"tb"}}]}},{"bool":{"should":[{"match":{"canonical_topic":"test"}},{"match":{"topics":"test"}}]}}]}}]}},"size":0,"aggregations":{"population_type":{"terms":{"size":1000,"field":"population_type.agg_key"}}}}`)

		So(searches[4].Header, ShouldResemble, client.Header{Index: "ons"})
		So(string(searches[4].Query), ShouldEqual, `{"query":{"bool":{"must":{"dis_max":{"queries":[{"bool":{"should":[{"match":{"title.title_no_dates":{"query":"a","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"match":{"title.title_no_stem":{"query":"a","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"multi_match":{"query":"a","fields":["title^10","edition","downloads.content^1"],"type":"cross_fields","minimum_should_match":"3<80% 5<60%"}},{"multi_match":{"query":"a","fields":["title^10","summary","metaDescription","edition","downloads.content^1","pageData^1","keywords"],"type":"phrase","boost":10.0,"slop":2}}]}},{"multi_match":{"query":"a","fields":["summary","metaDescription","downloads.content^1","pageData^1","keywords"],"type":"best_fields","minimum_should_match":"75%"}},{"match":{"keywords":{"query":"a","operator":"AND","boost":10.0}}},{"multi_match":{"query":"a","fields":["cdid","dataset_id"]}},{"match":{"searchBoost":{"query":"a","operator":"AND","boost":100.0}}}]}},"filter":[{"bool":{"must":[{"bool":{"should":[{"match":{"type":"ta"}},{"match":{"type":"tb"}}]}},{"bool":{"should":[{"match":{"canonical_topic":"test"}},{"match":{"topics":"test"}}]}}]}}]}},"size":0,"aggregations":{"dimensions":{"terms":{"size":1000,"field":"dimensions.agg_key"}}}}`)
	})
}

func TestBuildSearchQueryAggregates(t *testing.T) {
	Convey("Given a Query builder", t, func() {
		qb, err := NewQueryBuilder()
		So(err, ShouldBeNil)

		Convey("Then BuildSearchQuery successfully generates 5 queries for an empty request", func() {
			reqParams := &SearchRequest{}
			query, err := qb.BuildSearchQuery(context.Background(), reqParams, true)
			So(err, ShouldBeNil)

			var searches []client.Search
			err = json.Unmarshal(query, &searches)
			So(err, ShouldBeNil)
			So(searches, ShouldHaveLength, 5)

			Convey("And the expected topics aggregation (count) query is generated with no filters", func() {
				expectedQueryString := `{"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}}]}}]}},"size":0,"aggregations":{"topic":{"terms":{"size":1000,"field":"topics"}}}}`
				So(searches[1].Header, ShouldResemble, client.Header{Index: "ons"})
				So(string(searches[1].Query), ShouldEqual, expectedQueryString)
			})

			Convey("And the expected content types aggregation (count) query is generated with no filters", func() {
				expectedQueryString := `{"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}}]}}]}},"size":0,"aggregations":{"content_types":{"terms":{"size":1000,"field":"type"}}}}`
				So(searches[2].Header, ShouldResemble, client.Header{Index: "ons"})
				So(string(searches[2].Query), ShouldEqual, expectedQueryString)
			})

			Convey("And the expected population type aggregation (count) query is generated with no filters", func() {
				expectedQueryString := `{"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}}]}}]}},"size":0,"aggregations":{"population_type":{"terms":{"size":1000,"field":"population_type.agg_key"}}}}`
				So(searches[3].Header, ShouldResemble, client.Header{Index: "ons"})
				So(string(searches[3].Query), ShouldEqual, expectedQueryString)
			})

			Convey("And the expected dimensions aggregation (count) query is generated with no filters", func() {
				expectedQueryString := `{"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}}]}}]}},"size":0,"aggregations":{"dimensions":{"terms":{"size":1000,"field":"dimensions.agg_key"}}}}`
				So(searches[4].Header, ShouldResemble, client.Header{Index: "ons"})
				So(string(searches[4].Query), ShouldEqual, expectedQueryString)
			})
		})

		Convey("Then BuildSearchQuery successfully generates 5 queries for a request with topics, content type, population types and dimensions", func() {
			reqParams := &SearchRequest{
				Topic: []string{"test"},
				Types: []string{"ta", "tb"},
				PopulationTypes: []*PopulationTypeRequest{
					{Name: "pop1"},
					{Label: "lbl1"},
				},
				Dimensions: []*DimensionRequest{
					{Name: "dim1", Label: "lbl1", RawLabel: "rawLbl1"},
				},
			}
			query, err := qb.BuildSearchQuery(context.Background(), reqParams, true)
			So(err, ShouldBeNil)

			var searches []client.Search
			err = json.Unmarshal(query, &searches)
			So(err, ShouldBeNil)
			So(searches, ShouldHaveLength, 5)

			Convey("And the expected topics aggregation (count) query is generated", func() {
				expectedQueryString := `{"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[{"match":{"type":"ta"}},{"match":{"type":"tb"}}]}},{"bool":{"should":[{"match":{"population_type.name":{"query":"pop1"}}},{"match":{"population_type.label":{"query":""}}},{"match":{"population_type.label":{"query":"lbl1"}}}]}},{"bool":{"must":[{"bool":{"should":[{"match":{"dimensions.name":"dim1"}},{"match":{"dimensions.label":"lbl1"}},{"match":{"dimensions.raw_label":"rawLbl1"}}]}}]}}]}}]}},"size":0,"aggregations":{"topic":{"terms":{"size":1000,"field":"topics"}}}}`
				So(searches[1].Header, ShouldResemble, client.Header{Index: "ons"})
				So(string(searches[1].Query), ShouldEqual, expectedQueryString)
			})

			Convey("And the expected content types aggregation (count) query is generated", func() {
				expectedQueryString := `{"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[{"match":{"canonical_topic":"test"}},{"match":{"topics":"test"}}]}},{"bool":{"should":[{"match":{"population_type.name":{"query":"pop1"}}},{"match":{"population_type.label":{"query":""}}},{"match":{"population_type.label":{"query":"lbl1"}}}]}},{"bool":{"must":[{"bool":{"should":[{"match":{"dimensions.name":"dim1"}},{"match":{"dimensions.label":"lbl1"}},{"match":{"dimensions.raw_label":"rawLbl1"}}]}}]}}]}}]}},"size":0,"aggregations":{"content_types":{"terms":{"size":1000,"field":"type"}}}}`
				So(searches[2].Header, ShouldResemble, client.Header{Index: "ons"})
				So(string(searches[2].Query), ShouldEqual, expectedQueryString)
			})

			Convey("And the expected population type aggregation (count) query is generated", func() {
				expectedQueryString := `{"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[{"match":{"type":"ta"}},{"match":{"type":"tb"}}]}},{"bool":{"should":[{"match":{"canonical_topic":"test"}},{"match":{"topics":"test"}}]}},{"bool":{"must":[{"bool":{"should":[{"match":{"dimensions.name":"dim1"}},{"match":{"dimensions.label":"lbl1"}},{"match":{"dimensions.raw_label":"rawLbl1"}}]}}]}}]}}]}},"size":0,"aggregations":{"population_type":{"terms":{"size":1000,"field":"population_type.agg_key"}}}}`
				So(searches[3].Header, ShouldResemble, client.Header{Index: "ons"})
				So(string(searches[3].Query), ShouldEqual, expectedQueryString)
			})

			Convey("And the expected dimensions aggregation (count) query is generated, filtering by the other parameters", func() {
				expectedQueryString := `{"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[{"match":{"type":"ta"}},{"match":{"type":"tb"}}]}},{"bool":{"should":[{"match":{"canonical_topic":"test"}},{"match":{"topics":"test"}}]}},{"bool":{"should":[{"match":{"population_type.name":{"query":"pop1"}}},{"match":{"population_type.label":{"query":""}}},{"match":{"population_type.label":{"query":"lbl1"}}}]}}]}}]}},"size":0,"aggregations":{"dimensions":{"terms":{"size":1000,"field":"dimensions.agg_key"}}}}`
				So(searches[4].Header, ShouldResemble, client.Header{Index: "ons"})
				So(string(searches[4].Query), ShouldEqual, expectedQueryString)
			})
		})
	})
}

func TestBuildSearchQueryPopulationType(t *testing.T) {
	Convey("Given a Query builder", t, func() {
		qb, err := NewQueryBuilder()
		So(err, ShouldBeNil)

		Convey("Then the expected search query is generated for a full population type request", func() {
			reqParams := &SearchRequest{
				Size: 2,
				PopulationTypes: []*PopulationTypeRequest{
					{Name: "UR"},
					{Label: "usual residents"},
				},
			}
			query, err := qb.BuildSearchQuery(context.Background(), reqParams, true)
			So(err, ShouldBeNil)

			searches := unmarshal(query)
			expectedQueryString := `{"size":2,"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}},{"bool":{"should":[{"range":{"release_date":{"gte":null,"lte":null}}}]}},{"bool":{"should":[{"match":{"population_type.name":{"query":"UR"}}},{"match":{"population_type.label":{"query":""}}},{"match":{"population_type.label":{"query":"usual residents"}}}]}}]}}]}},"suggest":{"search_suggest":{"text":"","phrase":{"field":"title.title_no_synonym_no_stem"}}},"_source":{"includes":[],"excludes":["downloads.content","downloads*","pageData"]},"sort":[{"_score":{"order":"desc"}},{"release_date":{"order":"desc"}}]}`

			So(searches, ShouldHaveLength, 5)
			So(searches[0].Header, ShouldResemble, client.Header{Index: "ons"})
			So(string(searches[0].Query), ShouldEqual, expectedQueryString)
		})

		Convey("Then the expected search query is generated for a population type name-only request", func() {
			reqParams := &SearchRequest{
				Size: 2,
				PopulationTypes: []*PopulationTypeRequest{
					{Name: "UR"},
				},
			}
			query, err := qb.BuildSearchQuery(context.Background(), reqParams, true)
			So(err, ShouldBeNil)

			searches := unmarshal(query)
			expectedQueryString := `{"size":2,"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}},{"bool":{"should":[{"range":{"release_date":{"gte":null,"lte":null}}}]}},{"bool":{"should":[{"match":{"population_type.name":{"query":"UR"}}},{"match":{"population_type.label":{"query":""}}}]}}]}}]}},"suggest":{"search_suggest":{"text":"","phrase":{"field":"title.title_no_synonym_no_stem"}}},"_source":{"includes":[],"excludes":["downloads.content","downloads*","pageData"]},"sort":[{"_score":{"order":"desc"}},{"release_date":{"order":"desc"}}]}`

			So(searches, ShouldHaveLength, 5)
			So(searches[0].Header, ShouldResemble, client.Header{Index: "ons"})
			So(string(searches[0].Query), ShouldEqual, expectedQueryString)
		})

		Convey("Then the expected search query is generated for a population type label-only request", func() {
			reqParams := &SearchRequest{
				Size: 2,
				PopulationTypes: []*PopulationTypeRequest{
					{Label: "usual residents"},
				},
			}
			query, err := qb.BuildSearchQuery(context.Background(), reqParams, true)
			So(err, ShouldBeNil)

			searches := unmarshal(query)
			expectedQueryString := `{"size":2,"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}},{"bool":{"should":[{"range":{"release_date":{"gte":null,"lte":null}}}]}},{"bool":{"should":[{"match":{"population_type.label":{"query":"usual residents"}}}]}}]}}]}},"suggest":{"search_suggest":{"text":"","phrase":{"field":"title.title_no_synonym_no_stem"}}},"_source":{"includes":[],"excludes":["downloads.content","downloads*","pageData"]},"sort":[{"_score":{"order":"desc"}},{"release_date":{"order":"desc"}}]}`

			So(searches, ShouldHaveLength, 5)
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
			expectedQueryString := `{"size":2,"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}},{"bool":{"should":[{"range":{"release_date":{"gte":null,"lte":null}}}]}},{"bool":{"must":[{"bool":{"should":[{"match":{"dimensions.name":"workplace_travel_4a"}},{"match":{"dimensions.label":"Distance travelled to work"}},{"match":{"dimensions.raw_label":"Distance travelled to work (4 categories)"}}]}}]}}]}}]}},"suggest":{"search_suggest":{"text":"","phrase":{"field":"title.title_no_synonym_no_stem"}}},"_source":{"includes":[],"excludes":["downloads.content","downloads*","pageData"]},"sort":[{"_score":{"order":"desc"}},{"release_date":{"order":"desc"}}]}`

			So(searches, ShouldHaveLength, 5)
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
			expectedQueryString := `{"size":2,"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}},{"bool":{"should":[{"range":{"release_date":{"gte":null,"lte":null}}}]}},{"bool":{"must":[{"bool":{"should":[{"match":{"dimensions.name":"workplace_travel_4a"}},{"match":{"dimensions.label":""}},{"match":{"dimensions.raw_label":""}}]}}]}}]}}]}},"suggest":{"search_suggest":{"text":"","phrase":{"field":"title.title_no_synonym_no_stem"}}},"_source":{"includes":[],"excludes":["downloads.content","downloads*","pageData"]},"sort":[{"_score":{"order":"desc"}},{"release_date":{"order":"desc"}}]}`

			So(searches, ShouldHaveLength, 5)
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
			expectedQueryString := `{"size":2,"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}},{"bool":{"should":[{"range":{"release_date":{"gte":null,"lte":null}}}]}},{"bool":{"must":[{"bool":{"should":[{"match":{"dimensions.label":"Distance travelled to work"}},{"match":{"dimensions.raw_label":""}}]}}]}}]}}]}},"suggest":{"search_suggest":{"text":"","phrase":{"field":"title.title_no_synonym_no_stem"}}},"_source":{"includes":[],"excludes":["downloads.content","downloads*","pageData"]},"sort":[{"_score":{"order":"desc"}},{"release_date":{"order":"desc"}}]}`

			So(searches, ShouldHaveLength, 5)
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
			expectedQueryString := `{"size":2,"query":{"bool":{"must":{"match_all":{}},"filter":[{"bool":{"must":[{"bool":{"should":[]}},{"bool":{"should":[{"range":{"release_date":{"gte":null,"lte":null}}}]}},{"bool":{"must":[{"bool":{"should":[{"match":{"dimensions.label":""}},{"match":{"dimensions.raw_label":"Distance travelled to work"}}]}}]}}]}}]}},"suggest":{"search_suggest":{"text":"","phrase":{"field":"title.title_no_synonym_no_stem"}}},"_source":{"includes":[],"excludes":["downloads.content","downloads*","pageData"]},"sort":[{"_score":{"order":"desc"}},{"release_date":{"order":"desc"}}]}`

			So(searches, ShouldHaveLength, 5)
			So(searches[0].Header, ShouldResemble, client.Header{Index: "ons"})
			So(string(searches[0].Query), ShouldEqual, expectedQueryString)
		})
	})
}

func TestBuildCountQuery(t *testing.T) {
	Convey("Should return InternalError for invalid template", t, func() {
		qb := createQueryBuilderForCountTemplate("dummy{{.Moo}}")

		reqParams := &CountRequest{
			Term:        "someQuery",
			CountEnable: true,
		}
		query, err := qb.BuildCountQuery(context.Background(), reqParams)
		So(err, ShouldNotBeNil)
		So(query, ShouldBeNil)
		So(err.Error(), ShouldContainSubstring, "creation of search from template failed")
	})
}

func TestBuildCountQueryPopulationType(t *testing.T) {
	Convey("Given a Query builder", t, func() {
		qb, err := NewQueryBuilder()
		So(err, ShouldBeNil)

		Convey("Then the expected count query is generated for an empty request", func() {
			reqParams := &CountRequest{
				// Term:        "someQuery",
				CountEnable: true,
			}
			b, err := qb.BuildCountQuery(context.Background(), reqParams)
			So(err, ShouldBeNil)

			query, err := minifyJSON(b)
			So(err, ShouldBeNil)

			expectedQueryString := `{"query":{"bool":{"must":[{"match_all":{}},{"bool":{"must":{"exists":{"field":"topics"}}}}]}}}`
			So(string(query), ShouldResemble, expectedQueryString)
		})

		Convey("Then the expected count query is generated for a request with a term", func() {
			reqParams := &CountRequest{
				Term:        "someQuery",
				CountEnable: true,
			}
			b, err := qb.BuildCountQuery(context.Background(), reqParams)
			So(err, ShouldBeNil)

			query, err := minifyJSON(b)
			So(err, ShouldBeNil)

			expectedQueryString := `{"query":{"bool":{"must":[{"dis_max":{"queries":[{"bool":{"should":[{"match":{"title.title_no_dates":{"query":"someQuery","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"match":{"title.title_no_stem":{"query":"someQuery","boost":10.0,"minimum_should_match":"1<-2 3<80% 5<60%"}}},{"multi_match":{"query":"someQuery","fields":["title^10","edition","downloads.content^1"],"type":"cross_fields","minimum_should_match":"3<80% 5<60%"}},{"multi_match":{"query":"someQuery","fields":["title^10","summary","metaDescription","edition","downloads.content^1","pageData^1","keywords"],"type":"phrase","boost":10.0,"slop":2}}]}},{"multi_match":{"query":"someQuery","fields":["summary","metaDescription","downloads.content^1","pageData^1","keywords"],"type":"best_fields","minimum_should_match":"75%"}},{"match":{"keywords":{"query":"someQuery","operator":"AND","boost":10.0}}},{"multi_match":{"query":"someQuery","fields":["cdid","dataset_id"]}},{"match":{"searchBoost":{"query":"someQuery","operator":"AND","boost":100.0}}}]}},{"bool":{"must":{"exists":{"field":"topics"}}}}]}}}`
			So(string(query), ShouldResemble, expectedQueryString)
		})
	})
}

// return a minified JSON input string
// return an error encountered during minifiying or reading minified bytes
func minifyJSON(jsonB []byte) ([]byte, error) {
	buff := new(bytes.Buffer)
	errCompact := json.Compact(buff, jsonB)
	if errCompact != nil {
		newErr := fmt.Errorf("failure encountered compacting json := %v", errCompact)
		return []byte{}, newErr
	}

	b, err := io.ReadAll(buff)
	if err != nil {
		readErr := fmt.Errorf("read buffer error encountered := %v", err)
		return []byte{}, readErr
	}

	return b, nil
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
