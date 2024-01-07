package elasticsearch_test

import (
	"encoding/json"
	"testing"

	"github.com/ONSdigital/dp-search-api/elasticsearch"
	c "github.com/smartystreets/goconvey/convey"
)

func TestGetDefaultMappings_ValidJson(t *testing.T) {
	c.Convey("File `search-index-settings.json` is valid jason", t, func() {
		c.Convey("When we get the default search index settings json", func() {
			mappingsJSON := elasticsearch.GetSearchIndexSettings()
			c.Convey("Then the json returned should be valid", func() {
				c.So(json.Valid(mappingsJSON), c.ShouldBeTrue)
			})
		})
	})
}
