package steps

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/service"
	mocks "github.com/ONSdigital/dp-search-api/service/mock"
	"github.com/maxcnunes/httpfake"
)

const (
	gitCommitHash = "3t7e5s1t4272646ef477f8ed755"
	appVersion    = "v1.2.3"
	buildTime     = "20"
)

// Component contains all the information to create a component test
type Component struct {
	APIFeature           *componentTest.APIFeature
	cfg                  *config.Configuration
	ErrorFeature         componentTest.ErrorFeature
	FakeElasticSearchAPI *FakeAPI
	fakeRequest          *httpfake.Request
	HTTPServer           *http.Server
	ServiceRunning       bool
	svc                  *service.Service
	svcErrors            chan error
	StartTime            time.Time
}

// NewSearchAPIComponent creates a search api component
func NewSearchAPIComponent() (c *Component, err error) {
	c = &Component{
		HTTPServer: &http.Server{},
		svcErrors:  make(chan error),
	}

	c.FakeElasticSearchAPI = NewFakeAPI(&c.ErrorFeature)

	ctx := context.Background()

	svcErrors := make(chan error, 1)

	c.cfg, err = config.Get()
	if err != nil {
		return nil, err
	}

	c.cfg.ElasticSearchAPIURL = c.FakeElasticSearchAPI.fakeHTTP.ResolveURL("/ons/_search")

	c.cfg.HealthCheckInterval = 1 * time.Second
	c.cfg.HealthCheckCriticalTimeout = 2 * time.Second

	initFunctions := &mocks.InitialiserMock{
		DoGetHTTPServerFunc:   c.getHTTPServer,
		DoGetHealthCheckFunc:  getHealthCheckOK,
		DoGetHealthClientFunc: c.getHealthClient,
	}

	serviceList := service.NewServiceList(initFunctions)
	c.svc, err = service.Run(ctx, c.cfg, serviceList, buildTime, gitCommitHash, appVersion, svcErrors)
	if err != nil {
		return nil, err
	}

	c.StartTime = time.Now()
	c.ServiceRunning = true
	c.APIFeature = componentTest.NewAPIFeature(c.InitialiseService)

	return c, nil
}

// InitAPIFeature initialises the ApiFeature that's contained within a specific JobsFeature.
func (c *Component) InitAPIFeature() *componentTest.APIFeature {
	c.APIFeature = componentTest.NewAPIFeature(c.InitialiseService)

	return c.APIFeature
}

// Reset resets the search api component
func (c *Component) Reset() *Component {
	c.FakeElasticSearchAPI.Reset()
	c.FakeElasticSearchAPI.setJSONResponseForGet("/ons/_search?q=test", 200, []byte{}) // TODO Maybe url needs to change and add test body for response
	c.FakeElasticSearchAPI.setJSONResponseForGet("/search?q=test2", 200, []byte{})
	return c
}

// Close closes the search api component
func (c *Component) Close() error {
	if c.svc != nil && c.ServiceRunning {
		c.svc.Close(context.Background())
		c.ServiceRunning = false
	}

	c.FakeElasticSearchAPI.Close()

	return nil
}

// InitialiseService returns the http.Handler that's contained within the component.
func (c *Component) InitialiseService() (http.Handler, error) {
	return c.HTTPServer.Handler, nil
}

func getHealthCheckOK(cfg *config.Configuration, buildTime, gitCommit, version string) (service.HealthChecker, error) {
	componentBuildTime := strconv.Itoa(int(time.Now().Unix()))
	versionInfo, err := healthcheck.NewVersionInfo(componentBuildTime, gitCommitHash, appVersion)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

func (c *Component) getHealthClient(name string, url string) *health.Client {
	return &health.Client{
		URL:    url,
		Name:   "elasticsearch",
		Client: c.FakeElasticSearchAPI.getMockAPIHTTPClient(),
	}
}

func (c *Component) getHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	c.HTTPServer.Addr = bindAddr
	c.HTTPServer.Handler = router
	return c.HTTPServer
}

// newMock mocks HTTP Client
func (f *FakeAPI) getMockAPIHTTPClient() *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(paths []string) {},
		GetPathsWithNoRetriesFunc: func() []string { return []string{} },
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return f.fakeHTTP.Server.Client().Do(req)
		},
	}
}
