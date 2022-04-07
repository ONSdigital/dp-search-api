package config_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ONSdigital/dp-search-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {
	Convey("Given an environment with no environment variables set", t, func() {
		cfg, err := config.Get()
		Convey("When the config values are retrieved", func() {
			Convey("There should be no error returned", func() {
				So(err, ShouldBeNil)
			})
			Convey("The values should be set to the expected defaults", func() {
				So(cfg.AWS.Filename, ShouldEqual, "")
				So(cfg.AWS.Profile, ShouldEqual, "")
				So(cfg.AWS.Region, ShouldEqual, "eu-west-1")
				So(cfg.AWS.Service, ShouldEqual, "es")
				So(cfg.AWS.TLSInsecureSkipVerify, ShouldEqual, false)
				So(cfg.BindAddr, ShouldEqual, ":23900")
				So(cfg.ElasticSearchAPIURL, ShouldEqual, "http://localhost:9200")
				So(cfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.SignElasticsearchRequests, ShouldEqual, false)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
			})
		})

		Convey("When we get the config as a string", func() {
			cfgString := cfg.String()

			Convey("The string should be valid JSON", func() {
				So(cfgString, ShouldNotBeBlank)
				So(json.Valid([]byte(cfgString)), ShouldBeTrue)
			})

			Convey("The string should contain configured data", func() {
				So(cfgString, ShouldContainSubstring, `"BindAddr"`)
				So(cfgString, ShouldContainSubstring, `":23900"`)
			})
		})
	})
}
