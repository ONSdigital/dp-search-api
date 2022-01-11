package elasticsearch_test

import (
	"encoding/json"
	"testing"

	"github.com/ONSdigital/dp-search-api/elasticsearch"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetDefaultMappings_ValidJson(t *testing.T) {
	Convey("File `search-index-settings.json` is valid jason", t, func() {
		Convey("When we get the default search index settings json", func() {
			mappingsJSON := elasticsearch.GetSearchIndexSettings()
			Convey("Then the json returned should be valid", func() {
				So(json.Valid(mappingsJSON), ShouldBeTrue)
			})
		})
	})
}
