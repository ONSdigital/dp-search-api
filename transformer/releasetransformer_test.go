package transformer

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/ONSdigital/dp-search-api/models"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-search-api/query"
)

func TestTransformSearchReleaseResponse(t *testing.T) {
	t.Parallel()
	Convey("With a transformer initialised", t, func() {
		ctx := context.Background()
		transformer := NewReleaseTransformer(true)
		So(t, ShouldNotBeNil)

		Convey("Throws error on invalid JSON", func() {
			sampleResponse := []byte(`{"invalid":"json"`)
			_, err := transformer.TransformSearchResponse(ctx, sampleResponse, query.ReleaseSearchRequest{Term: "Education in Wales", Type: query.Upcoming, Size: 2, Provisional: true, Postponed: true}, true)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "Failed to decode elastic search response: unexpected end of JSON input")
		})

		Convey("Converts an example response with highlighting", func() {
			sampleResponse, err := os.ReadFile("testdata/search_release_es_response.json")
			So(err, ShouldBeNil)
			expected, err := os.ReadFile("testdata/search_release_expected_highlighted.json")
			So(err, ShouldBeNil)

			actual, err := transformer.TransformSearchResponse(ctx, sampleResponse, query.ReleaseSearchRequest{Term: "Education in Wales", Type: query.Upcoming, Size: 2, Provisional: true, Postponed: true}, true)
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

			actual, err := transformer.TransformSearchResponse(ctx, sampleResponse, query.ReleaseSearchRequest{Term: "Education in Wales", Type: query.Upcoming, Size: 2, Provisional: true, Postponed: true}, false)
			So(err, ShouldBeNil)
			So(actual, ShouldNotBeEmpty)
			var exp, act models.SearchResponseLegacy
			So(json.Unmarshal(expected, &exp), ShouldBeNil)
			So(json.Unmarshal(actual, &act), ShouldBeNil)
			So(act, ShouldResemble, exp)
		})
	})
}
