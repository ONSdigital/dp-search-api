package transformer

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTransformSearchReleaseResponse(t *testing.T) {
	t.Parallel()
	Convey("With a transformer initialised", t, func() {
		ctx := context.Background()
		transformer := NewReleaseTransformer()
		So(t, ShouldNotBeNil)

		Convey("Throws error on invalid JSON", func() {
			sampleResponse := []byte(`{"invalid":"json"`)
			_, err := transformer.TransformSearchResponse(ctx, sampleResponse, "test-query", true)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "Failed to decode elastic search response: unexpected end of JSON input")
		})

		Convey("Converts an example response with highlighting", func() {
			sampleResponse, err := os.ReadFile("testdata/search_release_es_response.json")
			So(err, ShouldBeNil)
			expected, err := os.ReadFile("testdata/search_release_expected_highlighted.json")
			So(err, ShouldBeNil)

			actual, err := transformer.TransformSearchResponse(ctx, sampleResponse, "test-query", true)
			So(err, ShouldBeNil)
			So(actual, ShouldNotBeEmpty)
			var exp, act SearchReleaseResponse
			So(json.Unmarshal(expected, &exp), ShouldBeNil)
			So(json.Unmarshal(actual, &act), ShouldBeNil)
			So(act, ShouldResemble, exp)
		})

		Convey("Converts an example response without highlighting", func() {
			sampleResponse, err := os.ReadFile("testdata/search_release_es_response.json")
			So(err, ShouldBeNil)
			expected, err := os.ReadFile("testdata/search_release_expected_plain.json")
			So(err, ShouldBeNil)

			actual, err := transformer.TransformSearchResponse(ctx, sampleResponse, "test-query", false)
			So(err, ShouldBeNil)
			So(actual, ShouldNotBeEmpty)
			var exp, act SearchResponse
			So(json.Unmarshal(expected, &exp), ShouldBeNil)
			So(json.Unmarshal(actual, &act), ShouldBeNil)
			So(act, ShouldResemble, exp)
		})
	})
}
