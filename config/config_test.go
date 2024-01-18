package config_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ONSdigital/dp-search-api/config"
	c "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {
	c.Convey("Given an environment with no environment variables set", t, func() {
		cfg, err := config.Get()
		c.Convey("When the config values are retrieved", func() {
			c.Convey("There should be no error returned", func() {
				c.So(err, c.ShouldBeNil)
			})
			c.Convey("The values should be set to the expected defaults", func() {
				c.So(cfg.AWS.Filename, c.ShouldEqual, "")
				c.So(cfg.AWS.Profile, c.ShouldEqual, "")
				c.So(cfg.AWS.Region, c.ShouldEqual, "eu-west-2")
				c.So(cfg.AWS.Service, c.ShouldEqual, "es")
				c.So(cfg.AWS.TLSInsecureSkipVerify, c.ShouldEqual, false)
				c.So(cfg.BindAddr, c.ShouldEqual, ":23900")
				c.So(cfg.ElasticSearchAPIURL, c.ShouldEqual, "http://localhost:11200")
				c.So(cfg.BerlinAPIURL, c.ShouldEqual, "http://localhost:28900")
				c.So(cfg.CategoryAPIURL, c.ShouldEqual, "http://localhost:28800")
				c.So(cfg.ScrubberAPIURL, c.ShouldEqual, "http://localhost:28700")
				c.So(cfg.GracefulShutdownTimeout, c.ShouldEqual, 5*time.Second)
				c.So(cfg.HealthCheckCriticalTimeout, c.ShouldEqual, 90*time.Second)
				c.So(cfg.HealthCheckInterval, c.ShouldEqual, 30*time.Second)
				c.So(cfg.NlpHubSettings, c.ShouldEqual, "{\"category_weighting\": 100000000.0, \"category_limit\": 100, \"default_state\": \"gb\"}")
				c.So(cfg.NlpToggle, c.ShouldEqual, false)
			})
		})

		c.Convey("When we get the config as a string", func() {
			cfgString := cfg.String()

			c.Convey("The string should be valid JSON", func() {
				c.So(cfgString, c.ShouldNotBeBlank)
				c.So(json.Valid([]byte(cfgString)), c.ShouldBeTrue)
			})

			c.Convey("The string should contain configured data", func() {
				c.So(cfgString, c.ShouldContainSubstring, `"BindAddr"`)
				c.So(cfgString, c.ShouldContainSubstring, `":23900"`)
				c.So(cfgString, c.ShouldContainSubstring, `"{\"category_weighting\": 100000000.0, \"category_limit\": 100, \"default_state\": \"gb\"}"`)
				c.So(cfgString, c.ShouldContainSubstring, `"http://localhost:28700"`)
				c.So(cfgString, c.ShouldContainSubstring, `"http://localhost:28800"`)
				c.So(cfgString, c.ShouldContainSubstring, `"http://localhost:28900"`)
			})
		})
	})
}
