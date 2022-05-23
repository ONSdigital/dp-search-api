package steps

import (
	"context"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-authorisation/auth"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-search-api/api"
	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/service"
	mocks "github.com/ONSdigital/dp-search-api/service/mock"
)

const (
	gitCommitHash = "3t7e5s1t4272646ef477f8ed755"
	appVersion    = "v1.2.3"
	buildTime     = "20"
)

// Component contains all the information to create a component test
type Component struct {
	APIFeature           *componentTest.APIFeature
	AuthFeature          *componentTest.AuthorizationFeature
	Cfg                  *config.Config
	ErrorFeature         componentTest.ErrorFeature
	FakeElasticSearchAPI *FakeAPI
	HTTPServer           *http.Server
	ServiceRunning       bool
	svc                  *service.Service
	svcErrors            chan error
	StartTime            time.Time
}

// SearchAPIComponent creates a search api component
func SearchAPIComponent(authFeature *componentTest.AuthorizationFeature) (c *Component, err error) {
	c = &Component{
		HTTPServer: &http.Server{},
		svcErrors:  make(chan error),
	}

	ctx := context.Background()

	svcErrors := make(chan error, 1)

	c.Cfg, err = config.Get()
	if err != nil {
		return nil, err
	}

	c.AuthFeature = authFeature
	c.Cfg.ZebedeeURL = c.AuthFeature.FakeAuthService.ResolveURL("")

	c.FakeElasticSearchAPI = NewFakeAPI(&c.ErrorFeature)
	c.Cfg.ElasticSearchAPIURL = c.FakeElasticSearchAPI.fakeHTTP.ResolveURL("/elasticsearch")
	c.Cfg.ElasticVersion710 = true

	// Setup responses from registered checkers for component
	c.FakeElasticSearchAPI.setJSONResponseForGetHealth("/elasticsearch/_cluster/health", 200)

	c.Cfg.HealthCheckInterval = 30 * time.Second
	c.Cfg.HealthCheckCriticalTimeout = 90 * time.Second

	initFunctions := &mocks.InitialiserMock{
		DoGetHTTPServerFunc:            c.getHTTPServer,
		DoGetHealthCheckFunc:           getHealthCheckOK,
		DoGetHealthClientFunc:          c.getHealthClient,
		DoGetAuthorisationHandlersFunc: c.doGetAuthorisationHandlers,
	}

	serviceList := service.NewServiceList(initFunctions)

	c.svc, err = service.Run(ctx, c.Cfg, serviceList, buildTime, gitCommitHash, appVersion, svcErrors)
	if err != nil {
		return nil, err
	}

	c.StartTime = time.Now()
	c.ServiceRunning = true

	return c, nil
}

// InitAPIFeature initialises the ApiFeature that's contained within a specific JobsFeature.
func (c *Component) InitAPIFeature() *componentTest.APIFeature {
	c.APIFeature = componentTest.NewAPIFeature(c.InitialiseService)

	return c.APIFeature
}

// Reset resets the search api component (should not reset Fake APIs)
func (c *Component) Reset() *Component {
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

func getHealthCheckOK(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
	return &mocks.HealthCheckerMock{
		AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		StartFunc:    func(ctx context.Context) {},
		StopFunc:     func() {},
	}, nil
}

func (c *Component) getHealthClient(name, url string) *health.Client {
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

// DoGetAuthorisationHandlers returns the mock AuthHandler that was created in the NewComponent function.
func (c *Component) doGetAuthorisationHandlers(cfg *config.Config) api.AuthHandler {
	authClient := auth.NewPermissionsClient(dphttp.NewClient())
	authVerifier := auth.DefaultPermissionsVerifier()

	// for checking caller permissions when we only have a user/service token
	permissions := auth.NewHandler(
		auth.NewPermissionsRequestBuilder(cfg.ZebedeeURL),
		authClient,
		authVerifier,
	)

	return permissions
}
