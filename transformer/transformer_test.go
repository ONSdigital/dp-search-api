package transformer

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTransform(t *testing.T) {
	Convey("Transforms unmarshalled search responses successfully", t, func() {
		Convey("Zero suggestions creates empty array", func() {
			es := esResponse{
				Responses: []esResponseItem{esResponseItem{
					Suggest: esSuggest{
						SearchSuggest: []esSearchSuggest{esSearchSuggest{
							Options: []esSearchSuggestOptions{},
						}},
					},
				}},
			}
			sr := transform(&es)
			So(sr.Suggestions, ShouldBeEmpty)
		})

		Convey("One suggestion creates a populated array", func() {
			es := esResponse{
				Responses: []esResponseItem{esResponseItem{
					Suggest: esSuggest{
						SearchSuggest: []esSearchSuggest{esSearchSuggest{
							Options: []esSearchSuggestOptions{
								esSearchSuggestOptions{Text: "option1"},
							},
						}},
					},
				}},
			}
			sr := transform(&es)
			So(sr.Suggestions, ShouldNotBeEmpty)
			So(len(sr.Suggestions), ShouldEqual, 1)
			So(sr.Suggestions[0], ShouldResemble, "option1")
		})
		Convey("Multiple suggestions creates a populated array incorrect order", func() {
			es := esResponse{
				Responses: []esResponseItem{esResponseItem{
					Suggest: esSuggest{
						SearchSuggest: []esSearchSuggest{
							esSearchSuggest{
								Options: []esSearchSuggestOptions{
									esSearchSuggestOptions{Text: "option1"},
								},
							},
							esSearchSuggest{
								Options: []esSearchSuggestOptions{
									esSearchSuggestOptions{Text: "option2"},
								},
							},
							esSearchSuggest{
								Options: []esSearchSuggestOptions{
									esSearchSuggestOptions{Text: "option3"},
								},
							},
						},
					},
				}},
			}
			sr := transform(&es)
			So(sr.Suggestions, ShouldNotBeEmpty)
			So(len(sr.Suggestions), ShouldEqual, 3)
			So(sr.Suggestions[0], ShouldResemble, "option1")
			So(sr.Suggestions[1], ShouldResemble, "option2")
			So(sr.Suggestions[2], ShouldResemble, "option3")
		})
	})
}

func TestBuildAdditionalSuggestionsList(t *testing.T) {
	Convey("buildAdditionalSuggestionList successfully", t, func() {

		Convey("returns array of strings", func() {
			query1 := buildAdditionalSuggestionList("test-query")
			So(query1, ShouldHaveLength, 1)
			So(query1[0], ShouldEqual, "test-query")

			query2 := buildAdditionalSuggestionList("test query")
			So(query2, ShouldHaveLength, 2)
			So(query2[0], ShouldEqual, "test")
			So(query2[1], ShouldEqual, "query")

			query3 := buildAdditionalSuggestionList("test query \"with quote marks\"")
			So(query3, ShouldHaveLength, 3)
			So(query3[0], ShouldEqual, "test")
			So(query3[1], ShouldEqual, "query")
			So(query3[2], ShouldEqual, "\"with quote marks\"")
		})
	})
}

func TestTransformSearchResponse(t *testing.T) {
	Convey("With a transformer initialised", t, func() {
		ctx := context.Background()
		t := New()
		So(t, ShouldNotBeNil)

		Convey("Throws error on invalid JSON", func() {
			sampleResponse := []byte(`{"invalid":"json"`)
			_, err := t.TransformSearchResponse(ctx, sampleResponse, "test-query")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "Failed to decode elastic search response: unexpected end of JSON input")
		})

		Convey("Handles missing responses", func() {
			sampleResponse := []byte(`{}`)
			_, err := t.TransformSearchResponse(ctx, sampleResponse, "test-query")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "Response to be transformed contained 0 items")
		})

		Convey("Converts an example response", func() {
			sampleResponse, err := ioutil.ReadFile("testdata/search_example.json")
			So(err, ShouldBeNil)
			expected, err := ioutil.ReadFile("testdata/search_expected.json")
			So(err, ShouldBeNil)

			actual, err := t.TransformSearchResponse(ctx, sampleResponse, "test-query")
			So(err, ShouldBeNil)
			So(actual, ShouldNotBeEmpty)
			var exp, act searchResponse
			So(json.Unmarshal(expected, &exp), ShouldBeNil)
			So(json.Unmarshal(actual, &act), ShouldBeNil)
			So(act, ShouldResemble, exp)

		})

		Convey("Calls buildAdditionalSuggestionsList if zero search results", func() {
			sampleResponse, err := ioutil.ReadFile("testdata/zero_search_example.json")
			So(err, ShouldBeNil)
			expected, err := ioutil.ReadFile("testdata/zero_search_expected.json")
			So(err, ShouldBeNil)

			actual, err := t.TransformSearchResponse(ctx, sampleResponse, "test query \"with quote marks\"")
			So(err, ShouldBeNil)
			So(actual, ShouldNotBeEmpty)
			var exp, act searchResponse
			So(json.Unmarshal(expected, &exp), ShouldBeNil)
			So(json.Unmarshal(actual, &act), ShouldBeNil)
			So(act, ShouldResemble, exp)
		})

	})
}
