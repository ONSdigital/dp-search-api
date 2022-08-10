package steps

import (
	"context"
	"net/http"
	"time"

	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/service"
)

// LegacyComponent contains all the information to create a component test
type LegacyComponent struct {
	APIFeature           *componentTest.APIFeature
	AuthFeature          *componentTest.AuthorizationFeature
	Cfg                  *config.Config
	ErrorFeature         componentTest.ErrorFeature
	FakeElasticSearchAPI *FakeAPI
	HTTPServer           *http.Server
	ServiceRunning       bool
	svc                  *service.Service
	StartTime            time.Time
}

// InitAPIFeature initialises the ApiFeature that's contained within a specific JobsFeature.
func (c *LegacyComponent) InitAPIFeature() *componentTest.APIFeature {
	c.APIFeature = componentTest.NewAPIFeature(c.InitialiseService)

	return c.APIFeature
}

// Reset resets the search api component (should not reset Fake APIs)
func (c *LegacyComponent) Reset() *LegacyComponent {
	return c
}

// Close closes the search api component
func (c *LegacyComponent) Close() error {
	if c.svc != nil && c.ServiceRunning {
		c.svc.Close(context.Background())
		c.ServiceRunning = false
	}

	c.FakeElasticSearchAPI.Close()

	return nil
}

// InitialiseService returns the http.Handler that's contained within the component.
func (c *LegacyComponent) InitialiseService() (http.Handler, error) {
	return c.HTTPServer.Handler, nil
}
