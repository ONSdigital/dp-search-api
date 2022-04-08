package transformer

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/ONSdigital/dp-search-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLegacyTransformer(t *testing.T) {
	Convey("Transforms unmarshalled search responses successfully", t, func() {
		transformer := NewLegacy()
		Convey("Zero suggestions creates empty array", func() {
			es := models.ESResponseLegacy{
				Responses: []models.ESResponseItemLegacy{models.ESResponseItemLegacy{
					Suggest: models.ESSuggestLegacy{
						SearchSuggest: []models.ESSearchSuggestLegacy{models.ESSearchSuggestLegacy{
							Options: []models.ESSearchSuggestOptionsLegacy{},
						}},
					},
				}},
			}
			sr := transformer.legayTransform(&es, false)
			So(sr.Suggestions, ShouldBeEmpty)
		})

		Convey("One suggestion creates a populated array", func() {
			es := models.ESResponseLegacy{
				Responses: []models.ESResponseItemLegacy{models.ESResponseItemLegacy{
					Suggest: models.ESSuggestLegacy{
						SearchSuggest: []models.ESSearchSuggestLegacy{models.ESSearchSuggestLegacy{
							Options: []models.ESSearchSuggestOptionsLegacy{
								models.ESSearchSuggestOptionsLegacy{Text: "option1"},
							},
						}},
					},
				}},
			}
			sr := transformer.legayTransform(&es, true)
			So(sr.Suggestions, ShouldNotBeEmpty)
			So(len(sr.Suggestions), ShouldEqual, 1)
			So(sr.Suggestions[0], ShouldResemble, "option1")
		})
		Convey("Multiple suggestions creates a populated array incorrect order", func() {
			es := models.ESResponseLegacy{
				Responses: []models.ESResponseItemLegacy{models.ESResponseItemLegacy{
					Suggest: models.ESSuggestLegacy{
						SearchSuggest: []models.ESSearchSuggestLegacy{
							models.ESSearchSuggestLegacy{
								Options: []models.ESSearchSuggestOptionsLegacy{
									models.ESSearchSuggestOptionsLegacy{Text: "option1"},
								},
							},
							models.ESSearchSuggestLegacy{
								Options: []models.ESSearchSuggestOptionsLegacy{
									models.ESSearchSuggestOptionsLegacy{Text: "option2"},
								},
							},
							models.ESSearchSuggestLegacy{
								Options: []models.ESSearchSuggestOptionsLegacy{
									models.ESSearchSuggestOptionsLegacy{Text: "option3"},
								},
							},
						},
					},
				}},
			}
			sr := transformer.legayTransform(&es, true)
			So(sr.Suggestions, ShouldNotBeEmpty)
			So(len(sr.Suggestions), ShouldEqual, 3)
			So(sr.Suggestions[0], ShouldResemble, "option1")
			So(sr.Suggestions[1], ShouldResemble, "option2")
			So(sr.Suggestions[2], ShouldResemble, "option3")
		})
	})
}

func TestLegacyBuildAdditionalSuggestionsList(t *testing.T) {
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

func TestLegacyTransformSearchResponse(t *testing.T) {
	Convey("With a transformer initialised", t, func() {
		ctx := context.Background()
		transformer := NewLegacy()
		So(t, ShouldNotBeNil)

		Convey("Throws error on invalid JSON", func() {
			sampleResponse := []byte(`{"invalid":"json"`)
			_, err := transformer.TransformSearchResponse(ctx, sampleResponse, "test-query", true)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "Failed to decode elastic search response: unexpected end of JSON input")
		})

		Convey("Handles missing responses", func() {
			sampleResponse := []byte(`{}`)
			_, err := transformer.TransformSearchResponse(ctx, sampleResponse, "test-query", true)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "Response to be transformed contained 0 items")
		})

		Convey("Converts an example response with highlighting", func() {
			sampleResponse, err := os.ReadFile("testdata/search_example.json")
			So(err, ShouldBeNil)
			expected, err := os.ReadFile("testdata/search_expected_highlighted.json")
			So(err, ShouldBeNil)

			actual, err := transformer.TransformSearchResponse(ctx, sampleResponse, "test-query", true)
			So(err, ShouldBeNil)
			So(actual, ShouldNotBeEmpty)
			var exp, act models.SearchResponseLegacy
			So(json.Unmarshal(expected, &exp), ShouldBeNil)
			So(json.Unmarshal(actual, &act), ShouldBeNil)
			So(act, ShouldResemble, exp)
		})

		Convey("Converts an example response without highlighting", func() {
			sampleResponse, err := os.ReadFile("testdata/search_example.json")
			So(err, ShouldBeNil)
			expected, err := os.ReadFile("testdata/search_expected_plain.json")
			So(err, ShouldBeNil)

			actual, err := transformer.TransformSearchResponse(ctx, sampleResponse, "test-query", false)
			So(err, ShouldBeNil)
			So(actual, ShouldNotBeEmpty)
			var exp, act models.SearchResponseLegacy
			So(json.Unmarshal(expected, &exp), ShouldBeNil)
			So(json.Unmarshal(actual, &act), ShouldBeNil)
			So(act, ShouldResemble, exp)
		})

		Convey("Calls buildAdditionalSuggestionsList if zero search results", func() {
			sampleResponse, err := os.ReadFile("testdata/zero_search_example.json")
			So(err, ShouldBeNil)
			expected, err := os.ReadFile("testdata/zero_search_expected.json")
			So(err, ShouldBeNil)

			actual, err := transformer.TransformSearchResponse(ctx, sampleResponse, "test query \"with quote marks\"", false)
			So(err, ShouldBeNil)
			So(actual, ShouldNotBeEmpty)
			var exp, act models.SearchResponseLegacy
			So(json.Unmarshal(expected, &exp), ShouldBeNil)
			So(json.Unmarshal(actual, &act), ShouldBeNil)
			So(act, ShouldResemble, exp)
		})
	})
}

func TestTransform(t *testing.T) {
	expectedESDocument1 := models.ESSourceDocument{
		DataType:        "anyDataType1",
		JobID:           "",
		CDID:            "",
		DatasetID:       "",
		Keywords:        []string{"anykeyword1"},
		MetaDescription: "",
		Summary:         "",
		ReleaseDate:     "",
		Title:           "anyTitle2",
		Topics:          []string{"anyTopic1"},
	}
	expectedESDocument2 := models.ESSourceDocument{
		DataType:        "anyDataType2",
		JobID:           "",
		CDID:            "",
		DatasetID:       "",
		Keywords:        []string{"anykeyword2"},
		MetaDescription: "",
		Summary:         "",
		ReleaseDate:     "",
		Title:           "anyTitle2",
		Topics:          []string{"anyTopic2"},
	}

	Convey("Given a new instance of Transformer for ES7x with search responses successfully", t, func() {
		transformer := New()
		esResponse := prepareESMockResponse()

		Convey("When calling a transformer", func() {
			transformedResponse := transformer.transform(&esResponse, true)

			Convey("Then transforms unmarshalled search responses successfully", func() {
				So(transformedResponse, ShouldNotBeNil)
				So(transformedResponse.Took, ShouldEqual, 10)
				So(len(transformedResponse.Items), ShouldEqual, 2)
				So(transformedResponse.Items[0], ShouldResemble, expectedESDocument1)
				So(transformedResponse.Items[1], ShouldResemble, expectedESDocument2)
				So(transformedResponse.Suggestions[0], ShouldResemble, "testSuggestion")
			})
		})
	})
}

// Prepare mock ES response
func prepareESMockResponse() models.EsResponses {
	esDocument1 := models.ESSourceDocument{
		DataType:        "anyDataType1",
		JobID:           "",
		CDID:            "",
		DatasetID:       "",
		Keywords:        []string{"anykeyword1"},
		MetaDescription: "",
		Summary:         "",
		ReleaseDate:     "",
		Title:           "anyTitle2",
		Topics:          []string{"anyTopic1"},
	}

	esDocument2 := models.ESSourceDocument{
		DataType:        "anyDataType2",
		JobID:           "",
		CDID:            "",
		DatasetID:       "",
		Keywords:        []string{"anykeyword2"},
		MetaDescription: "",
		Summary:         "",
		ReleaseDate:     "",
		Title:           "anyTitle2",
		Topics:          []string{"anyTopic2"},
	}

	esDocuments := []models.ESSourceDocument{esDocument1, esDocument2}

	hit := models.ESResponseHit{
		Source:    esDocuments,
		Highlight: models.ESHighlight{},
	}

	bucket1 := models.ESBucket{
		Key:   "article",
		Count: 1,
	}
	bucket2 := models.ESBucket{
		Key:   "product_page",
		Count: 1,
	}
	buckets := []models.ESBucket{bucket1, bucket2}

	esDoccount := models.ESDocCounts{
		Buckets: buckets,
	}

	esResponse1 := models.EsResponse{
		Took: 10,
		Hits: models.ESResponseHits{
			Total: 1,
			Hits: []models.ESResponseHit{
				hit,
			},
		},
		Aggregations: models.ESResponseAggregations{
			Doccounts: esDoccount,
		},
		Suggest: []string{"testSuggestion"},
	}

	// Preparing ES response array
	esResponse := models.EsResponses{
		Responses: []models.EsResponse{
			esResponse1,
		},
	}

	return esResponse
}
