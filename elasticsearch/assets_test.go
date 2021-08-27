package elasticsearch_test

import (
	"encoding/json"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGetDefaultMappings_ValidJson(t *testing.T) {
	Convey("File `matchers.json` is valid jason", t, func() {
		Convey("When we get the default mappings json", func() {
			mappingsJson := elasticsearch.GetDefaultMappings()

			Convey("Then the json returned should be valid", func() {

				So(json.Valid(mappingsJson), ShouldBeTrue)
			})

		})
	})
}
