package transformer

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBuildMatches(t *testing.T) {
	Convey("Build matches for translates successfully", t, func() {
		hl := esHighlight{
			DescriptionTitle:     []string{"<strong>value</strong> and double <strong>value</strong>"},
			DescriptionEdition:   []string{"<strong>value</strong>"},
			DescriptionSummary:   []string{"single <strong>value</strong>"},
			DescriptionMeta:      []string{"a <wrong>value</strong> here but not <strong>value</strong> here"},
			DescriptionKeywords:  []string{"one <strong>value</strong>", "<strong>value</strong> another <strong>value</strong> third <strong>value</strong>", "<strong>value</strong> again"},
			DescriptionDatasetID: []string{"       space before <strong>value</strong> and after    "},
		}

		matches := buildMatches(hl)
		So(matches, ShouldNotBeNil)
		So(matches.Description, ShouldNotBeNil)

		So(matches.Description.Title, ShouldNotBeNil)
		titleDetails := matches.Description.Title
		So(titleDetails, ShouldNotBeEmpty)
		So(len(titleDetails), ShouldEqual, 2)
		So(titleDetails[0].Value, ShouldBeBlank)
		So(titleDetails[0].Start, ShouldEqual, 1)
		So(titleDetails[0].End, ShouldEqual, 5)
		So(titleDetails[1].Value, ShouldBeBlank)
		So(titleDetails[1].Start, ShouldEqual, 18)
		So(titleDetails[1].End, ShouldEqual, 22)

		So(matches.Description.Edition, ShouldNotBeNil)
		editionDetails := matches.Description.Edition
		So(editionDetails, ShouldNotBeEmpty)
		So(len(editionDetails), ShouldEqual, 1)
		So(editionDetails[0].Value, ShouldBeBlank)
		So(editionDetails[0].Start, ShouldEqual, 1)
		So(editionDetails[0].End, ShouldEqual, 5)

		So(matches.Description.Summary, ShouldNotBeNil)
		summaryDetails := matches.Description.Summary
		So(summaryDetails, ShouldNotBeEmpty)
		So(len(summaryDetails), ShouldEqual, 1)
		So(summaryDetails[0].Value, ShouldBeBlank)
		So(summaryDetails[0].Start, ShouldEqual, 8)
		So(summaryDetails[0].End, ShouldEqual, 12)

		So(matches.Description.MetaDescription, ShouldNotBeNil)
		metaDetails := matches.Description.MetaDescription
		So(metaDetails, ShouldNotBeEmpty)
		So(len(metaDetails), ShouldEqual, 1)
		So(metaDetails[0].Value, ShouldBeBlank)
		So(metaDetails[0].Start, ShouldEqual, 38)
		So(metaDetails[0].End, ShouldEqual, 42)

		So(matches.Description.Keywords, ShouldNotBeNil)
		keywordsDetails := matches.Description.Keywords
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

		So(matches.Description.DatasetID, ShouldNotBeNil)
		dataSetDetails := matches.Description.DatasetID
		So(dataSetDetails, ShouldNotBeEmpty)
		So(len(dataSetDetails), ShouldEqual, 1)
		So(dataSetDetails[0].Value, ShouldBeBlank)
		So(dataSetDetails[0].Start, ShouldEqual, 21)
		So(dataSetDetails[0].End, ShouldEqual, 25)
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

func TestTransformSearchResponse(t *testing.T) {
	Convey("With a transformer initialised", t, func() {
		ctx := context.Background()
		t := New()
		So(t, ShouldNotBeNil)

		Convey("Throws error on invalid JSON", func() {
			sampleResponse := []byte(`{"invalid":"json"`)
			_, err := t.TransformSearchResponse(ctx, sampleResponse)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "Failed to decode elastic search response: unexpected end of JSON input")
		})

		Convey("Handles missing responses", func() {
			sampleResponse := []byte(`{}`)
			_, err := t.TransformSearchResponse(ctx, sampleResponse)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "Response to be transformed contained 0 items")
		})

		Convey("Converts an example response", func() {
			sampleResponse, err := ioutil.ReadFile("testdata/search_example.json")
			So(err, ShouldBeNil)
			expected, err := ioutil.ReadFile("testdata/search_expected.json")
			So(err, ShouldBeNil)

			actual, err := t.TransformSearchResponse(ctx, sampleResponse)
			So(err, ShouldBeNil)
			So(actual, ShouldNotBeEmpty)
			var exp, act searchResponse
			So(json.Unmarshal(expected, &exp), ShouldBeNil)
			So(json.Unmarshal(actual, &act), ShouldBeNil)
			So(act, ShouldResemble, exp)

		})

	})
}
