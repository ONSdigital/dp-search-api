package steps

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/service"
	mocks "github.com/ONSdigital/dp-search-api/service/mock"
)

var (
	BuildTime = strconv.Itoa(time.Now().Nanosecond())
	GitCommit = "component test commit"
	Version   = "component test version"
)

type Component struct {
	ErrorFeature          *componentTest.ErrorFeature
	AuthFeature           *componentTest.AuthorizationFeature
	APIFeature            *componentTest.APIFeature
	ESDependencyInjection *ESDependency

	svc *service.Service
}

func TestComponent(t *testing.T) *Component {
	cfg, err := config.Get()
	if err != nil {
		t.Fatalf("failed to get configuariion: %s", err)
	}

	c := &Component{
		ErrorFeature:          &componentTest.ErrorFeature{TB: t},
		AuthFeature:           componentTest.NewAuthorizationFeature(),
		ESDependencyInjection: NewESDependency(t, cfg),
	}
	c.APIFeature = componentTest.NewAPIFeature(c.ServiceAPIRouter)

	cfg.ZebedeeURL = c.AuthFeature.FakeAuthService.ResolveURL("")
	cfg.HealthCheckInterval = 30 * time.Second
	cfg.HealthCheckCriticalTimeout = 90 * time.Second
	c.ESDependencyInjection.esServer.NewHandler().Get("/elasticsearch/_cluster/health").Reply(200).Body([]byte(""))

	standardInit := service.Init{}
	initFunctions := &mocks.InitialiserMock{
		DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer {
			return &http.Server{Addr: bindAddr, Handler: router}
		},
		DoGetHealthCheckFunc: func(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
			return &mocks.HealthCheckerMock{
				AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
				StartFunc:    func(ctx context.Context) {},
				StopFunc:     func() {},
			}, nil
		},
		DoGetHealthClientFunc:          standardInit.DoGetHealthClient,
		DoGetAuthorisationHandlersFunc: standardInit.DoGetAuthorisationHandlers,
		DoGetElasticSearchServerFunc:   standardInit.DoGetElasticSearchServer,
	}

	serviceList := service.NewServiceList(initFunctions)
	c.svc, err = service.Run(context.Background(), cfg, serviceList, BuildTime, GitCommit, Version, make(chan error, 1))
	if err != nil {
		t.Fatalf("service failed to run: %s", err)
	}

	return c
}

func (c *Component) Reset() *Component {
	c.APIFeature.Reset()
	c.AuthFeature.Reset()
	c.ESDependencyInjection.Reset()

	return c
}

func (c *Component) Close() error {
	if c.svc != nil {
		_ = c.svc.Close(context.Background())
		c.svc = nil
	}

	return nil
}

func (c *Component) ServiceAPIRouter() (http.Handler, error) {
	return c.svc.API.Router, nil
}
