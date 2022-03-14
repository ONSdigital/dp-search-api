package service

//go:generate moq -out mock/initialiser.go -pkg mocks . Initialiser
//go:generate moq -out mock/healthcheck.go -pkg mocks . HealthChecker
//go:generate moq -out mock/httpserver.go -pkg mocks . HTTPServer

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-search-api/api"
	"github.com/ONSdigital/dp-search-api/clients"
	"github.com/ONSdigital/dp-search-api/config"
)

// Initialiser defines the methods to initialise external services
type Initialiser interface {
	DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error)
	DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer
	DoGetHealthClient(name, url string) *health.Client
	DoGetAuthorisationHandlers(cfg *config.Config) api.AuthHandler
	DoGetDatasetClient(cfg *config.Config) clients.DatasetAPIClient
}

// HealthChecker defines the required methods from Healthcheck
type HealthChecker interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddCheck(name string, checker healthcheck.Checker) (err error)
}

// HTTPServer defines the required methods from the HTTP server
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}
