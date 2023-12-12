package transformer

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/ONSdigital/dp-search-api/models"
	c "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-search-api/query"
)

func TestTransformSearchReleaseResponse(t *testing.T) {
	t.Parallel()
	c.Convey("With a transformer initialised", t, func() {
		ctx := context.Background()
		transformer := NewReleaseTransformer()
		c.So(t, c.ShouldNotBeNil)

		c.Convey("Throws error on invalid JSON", func() {
			sampleResponse := []byte(`{"invalid":"json"`)
			_, err := transformer.TransformSearchResponse(ctx, sampleResponse, query.ReleaseSearchRequest{Term: "Education in Wales", Type: query.Upcoming, Size: 2, Provisional: true, Postponed: true}, true)
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldResemble, "Failed to decode elastic search response: unexpected end of JSON input")
		})

		c.Convey("Converts an example response with highlighting", func() {
			sampleResponse, err := os.ReadFile("testdata/search_release_es_response.json")
			c.So(err, c.ShouldBeNil)
			expected, err := os.ReadFile("testdata/search_release_expected_highlighted.json")
			c.So(err, c.ShouldBeNil)

			actual, err := transformer.TransformSearchResponse(ctx, sampleResponse, query.ReleaseSearchRequest{Term: "Education in Wales", Type: query.Upcoming, Size: 2, Provisional: true, Postponed: true}, true)
			c.So(err, c.ShouldBeNil)
			c.So(actual, c.ShouldNotBeEmpty)
			var exp, act SearchReleaseResponse
			c.So(json.Unmarshal(expected, &exp), c.ShouldBeNil)
			c.So(json.Unmarshal(actual, &act), c.ShouldBeNil)
			c.So(act, c.ShouldResemble, exp)
		})

		c.Convey("Converts an example response without highlighting", func() {
			sampleResponse, err := os.ReadFile("testdata/search_release_es_response.json")
			c.So(err, c.ShouldBeNil)
			expected, err := os.ReadFile("testdata/search_release_expected_plain.json")
			c.So(err, c.ShouldBeNil)

			actual, err := transformer.TransformSearchResponse(ctx, sampleResponse, query.ReleaseSearchRequest{Term: "Education in Wales", Type: query.Upcoming, Size: 2, Provisional: true, Postponed: true}, false)
			c.So(err, c.ShouldBeNil)
			c.So(actual, c.ShouldNotBeEmpty)
			var exp, act models.SearchResponseLegacy
			c.So(json.Unmarshal(expected, &exp), c.ShouldBeNil)
			c.So(json.Unmarshal(actual, &act), c.ShouldBeNil)
			c.So(act, c.ShouldResemble, exp)
		})
	})
}
