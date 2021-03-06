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

func TestBuildMatches(t *testing.T) {
	Convey("Build matches for Title translates successfully", t, func() {
		hl := esHighlight{
			DescriptionTitle: &[]string{"<strong>value</strong> and double <strong>value</strong>"},
		}

		matches := buildMatches(hl)
		So(matches, ShouldNotBeNil)
		So(matches.Description, ShouldNotBeNil)

		So(matches.Description.Title, ShouldNotBeNil)
		titleDetails := *matches.Description.Title
		So(titleDetails, ShouldNotBeEmpty)
		So(len(titleDetails), ShouldEqual, 2)
		So(titleDetails[0].Value, ShouldBeBlank)
		So(titleDetails[0].Start, ShouldEqual, 1)
		So(titleDetails[0].End, ShouldEqual, 5)
		So(titleDetails[1].Value, ShouldBeBlank)
		So(titleDetails[1].Start, ShouldEqual, 18)
		So(titleDetails[1].End, ShouldEqual, 22)
	})

	Convey("Build matches for Edition translates successfully", t, func() {
		hl := esHighlight{
			DescriptionEdition: &[]string{"<strong>value</strong>"},
		}

		matches := buildMatches(hl)
		So(matches, ShouldNotBeNil)
		So(matches.Description, ShouldNotBeNil)

		So(matches.Description.Edition, ShouldNotBeNil)
		editionDetails := *matches.Description.Edition
		So(editionDetails, ShouldNotBeEmpty)
		So(len(editionDetails), ShouldEqual, 1)
		So(editionDetails[0].Value, ShouldBeBlank)
		So(editionDetails[0].Start, ShouldEqual, 1)
		So(editionDetails[0].End, ShouldEqual, 5)

	})

	Convey("Build matches for Summary translates successfully", t, func() {
		hl := esHighlight{
			DescriptionSummary: &[]string{"single <strong>value</strong>"},
		}

		matches := buildMatches(hl)
		So(matches, ShouldNotBeNil)
		So(matches.Description, ShouldNotBeNil)

		So(matches.Description.Summary, ShouldNotBeNil)
		summaryDetails := *matches.Description.Summary
		So(summaryDetails, ShouldNotBeEmpty)
		So(len(summaryDetails), ShouldEqual, 1)
		So(summaryDetails[0].Value, ShouldBeBlank)
		So(summaryDetails[0].Start, ShouldEqual, 8)
		So(summaryDetails[0].End, ShouldEqual, 12)
	})

	Convey("Build matches for MetaDescription translates successfully", t, func() {
		hl := esHighlight{
			DescriptionMeta: &[]string{"a <wrong>value</strong> here but not <strong>value</strong> here"},
		}

		matches := buildMatches(hl)
		So(matches, ShouldNotBeNil)
		So(matches.Description, ShouldNotBeNil)

		So(matches.Description.MetaDescription, ShouldNotBeNil)
		metaDetails := *matches.Description.MetaDescription
		So(metaDetails, ShouldNotBeEmpty)
		So(len(metaDetails), ShouldEqual, 1)
		So(metaDetails[0].Value, ShouldBeBlank)
		So(metaDetails[0].Start, ShouldEqual, 38)
		So(metaDetails[0].End, ShouldEqual, 42)
	})

	Convey("Build matches for Keywords translates successfully", t, func() {
		hl := esHighlight{
			DescriptionKeywords: &[]string{"one <strong>value</strong>", "<strong>value</strong> another <strong>value</strong> third <strong>value</strong>", "<strong>value</strong> again"},
		}

		matches := buildMatches(hl)
		So(matches, ShouldNotBeNil)
		So(matches.Description, ShouldNotBeNil)

		So(matches.Description.Keywords, ShouldNotBeNil)
		keywordsDetails := *matches.Description.Keywords
		So(keywordsDetails, ShouldNotBeEmpty)
		So(len(keywordsDetails), ShouldEqual, 5)
		So(keywordsDetails[0].Value, ShouldResemble, "one value")
		So(keywordsDetails[0].Start, ShouldEqual, 5)
		So(keywordsDetails[0].End, ShouldEqual, 9)
		So(keywordsDetails[1].Value, ShouldResemble, "value another value third value")
		So(keywordsDetails[1].Start, ShouldEqual, 1)
		So(keywordsDetails[1].End, ShouldEqual, 5)
		So(keywordsDetails[2].Value, ShouldResemble, "value another value third value")
		So(keywordsDetails[2].Start, ShouldEqual, 15)
		So(keywordsDetails[2].End, ShouldEqual, 19)
		So(keywordsDetails[3].Value, ShouldResemble, "value another value third value")
		So(keywordsDetails[3].Start, ShouldEqual, 27)
		So(keywordsDetails[3].End, ShouldEqual, 31)
		So(keywordsDetails[4].Value, ShouldResemble, "value again")
		So(keywordsDetails[4].Start, ShouldEqual, 1)
		So(keywordsDetails[4].End, ShouldEqual, 5)
	})

	Convey("Build matches for DatasetID translates successfully", t, func() {
		hl := esHighlight{
			DescriptionDatasetID: &[]string{"       space before <strong>value</strong> and after    "},
		}

		matches := buildMatches(hl)
		So(matches, ShouldNotBeNil)
		So(matches.Description, ShouldNotBeNil)

		So(matches.Description.DatasetID, ShouldNotBeNil)
		dataSetDetails := *matches.Description.DatasetID
		So(dataSetDetails, ShouldNotBeEmpty)
		So(len(dataSetDetails), ShouldEqual, 1)
		So(dataSetDetails[0].Value, ShouldBeBlank)
		So(dataSetDetails[0].Start, ShouldEqual, 21)
		So(dataSetDetails[0].End, ShouldEqual, 25)
	})

	Convey("Build matches for all items translates successfully", t, func() {
		hl := esHighlight{
			DescriptionTitle:     &[]string{"<strong>value</strong> and double <strong>value</strong>"},
			DescriptionEdition:   &[]string{"<strong>value</strong>"},
			DescriptionSummary:   &[]string{"single <strong>value</strong>"},
			DescriptionMeta:      &[]string{"a <wrong>value</strong> here but not <strong>value</strong> here"},
			DescriptionKeywords:  &[]string{"one <strong>value</strong>", "<strong>value</strong> another <strong>value</strong> third <strong>value</strong>", "<strong>value</strong> again"},
			DescriptionDatasetID: &[]string{"       space before <strong>value</strong> and after    "},
		}

		matches := buildMatches(hl)
		So(matches, ShouldNotBeNil)
		So(matches.Description, ShouldNotBeNil)

		So(matches.Description.Title, ShouldNotBeNil)
		So(matches.Description.Title, ShouldNotBeEmpty)

		So(matches.Description.Edition, ShouldNotBeNil)
		So(matches.Description.Edition, ShouldNotBeEmpty)

		So(matches.Description.Summary, ShouldNotBeNil)
		So(matches.Description.Summary, ShouldNotBeEmpty)

		So(matches.Description.MetaDescription, ShouldNotBeNil)
		So(matches.Description.MetaDescription, ShouldNotBeEmpty)

		So(matches.Description.Keywords, ShouldNotBeNil)
		So(matches.Description.Keywords, ShouldNotBeEmpty)

		So(matches.Description.DatasetID, ShouldNotBeNil)
		So(matches.Description.DatasetID, ShouldNotBeEmpty)
	})
}

func TestFindMatches(t *testing.T) {
	Convey("Find matches successfully", t, func() {

		Convey("No value returns empty array", func() {
			s := "nothing here"
			md, fs := findMatches(s)
			So(md, ShouldBeEmpty)
			So(len(md), ShouldEqual, 0)
			So(fs, ShouldResemble, "nothing here")
		})

		Convey("Single value finds successfully", func() {
			s := "single <strong>value</strong>"
			md, fs := findMatches(s)
			So(md, ShouldNotBeEmpty)
			So(len(md), ShouldEqual, 1)
			So(fs, ShouldResemble, "single value")
			So(md[0].Start, ShouldEqual, 8)
			So(md[0].End, ShouldEqual, 12)
		})

		Convey("Double value finds successfully", func() {
			s := "<strong>value</strong> and double <strong>value</strong>"
			md, fs := findMatches(s)
			So(fs, ShouldResemble, "value and double value")
			So(md, ShouldNotBeEmpty)
			So(len(md), ShouldEqual, 2)
			So(md[0].Start, ShouldEqual, 1)
			So(md[0].End, ShouldEqual, 5)
			So(md[1].Start, ShouldEqual, 18)
			So(md[1].End, ShouldEqual, 22)
		})

		Convey("Tripple values find successfully", func() {
			s := "<strong>value</strong> and double <strong>value</strong> and tripple–<strong>value</strong>"
			md, fs := findMatches(s)
			So(fs, ShouldResemble, "value and double value and tripple–value")
			So(md, ShouldNotBeEmpty)
			So(len(md), ShouldEqual, 3)
			So(md[0].Start, ShouldEqual, 1)
			So(md[0].End, ShouldEqual, 5)
			So(md[1].Start, ShouldEqual, 18)
			So(md[1].End, ShouldEqual, 22)
			So(md[2].Start, ShouldEqual, 38) // 38 bytes. After UTF8 character, actual characters = 36
			So(md[2].End, ShouldEqual, 42)   // 42 bytes. After UTF8 character, actual characters = 40
		})

		Convey("UTF8 characters counted as bytes", func() {
			s := "€–<strong>value</strong>"
			md, fs := findMatches(s)
			So(fs, ShouldResemble, "€–value")
			So(md, ShouldNotBeEmpty)
			So(len(md), ShouldEqual, 1)
			So(md[0].Start, ShouldEqual, 7) // Should be 5 if chars
			So(md[0].End, ShouldEqual, 11)  // Should be 11 if chars
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
