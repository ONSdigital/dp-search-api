package elasticsearch_test

import (
	"encoding/json"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGetDefaultMappings_ValidJson(t *testing.T) {
	Convey("File `search-index-settings.json` is valid jason", t, func() {
		Convey("When we get the default search index settings json", func() {
			mappingsJson := elasticsearch.GetSearchIndexSettings()

			Convey("Then the json returned should be valid", func() {

				So(json.Valid(mappingsJson), ShouldBeTrue)
			})

		})
	})
}
