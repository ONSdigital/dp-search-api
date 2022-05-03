package service

import (
	"context"

	legacyESClient "github.com/ONSdigital/dp-elasticsearch/v3/client/elasticsearch/v2"
	"github.com/ONSdigital/dp-net/v2/awsauth"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-search-api/api"
	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	"github.com/ONSdigital/dp-search-api/query"
	"github.com/ONSdigital/dp-search-api/transformer"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type Service struct {
	api                 *api.SearchAPI
	config              *config.Config
	elasticSearchClient elasticsearch.Client
	healthCheck         HealthChecker
	queryBuilder        api.QueryBuilder
	router              *mux.Router
	server              HTTPServer
	serviceList         *ExternalServiceList
	transformer         *transformer.LegacyTransformer
}

// SetServer sets the http server for a service
func (svc *Service) SetServer(server HTTPServer) {
	svc.server = server
}

// SetHealthCheck sets the healthchecker for a service
func (svc *Service) SetHealthCheck(healthCheck HealthChecker) {
	svc.healthCheck = healthCheck
}

// SetQueryBuilder sets the queryBuilder for a service
func (svc *Service) SetQueryBuilder(queryBuilder api.QueryBuilder) {
	svc.queryBuilder = queryBuilder
}

// SetElasticSearchClient sets the new instance of elasticsearch for a service
func (svc *Service) SetElasticSearchClient(elasticSearchClient elasticsearch.Client) {
	svc.elasticSearchClient = elasticSearchClient
}

// SetTransformer sets the transformer for a service
func (svc *Service) SetTransformer(transformerClient *transformer.LegacyTransformer) {
	svc.transformer = transformerClient
}

// Run the service
func Run(ctx context.Context, cfg *config.Config, serviceList *ExternalServiceList, buildTime, gitCommit, version string, svcErrors chan error) (svc *Service, err error) {
	elasticHTTPClient := dphttp.NewClient()

	// Initialise transformerClient
	transformerClient := transformer.NewLegacy()

	// Initialse AWS signer
	if cfg.SignElasticsearchRequests {
		var awsSignerRT *awsauth.AwsSignerRoundTripper

		awsSignerRT, err = awsauth.NewAWSSignerRoundTripper(cfg.AWS.Filename, cfg.AWS.Profile, cfg.AWS.Region, cfg.AWS.Service, awsauth.Options{TlsInsecureSkipVerify: cfg.AWS.TLSInsecureSkipVerify})
		if err != nil {
			log.Error(ctx, "failed to create aws auth round tripper", err)
			return nil, err
		}

		elasticHTTPClient = dphttp.NewClientWithTransport(awsSignerRT)
	}

	dpESClient := legacyESClient.NewClientWithHTTPClient(cfg.ElasticSearchAPIURL, elasticHTTPClient)

	// Initialise deprecatedESClient
	deprecatedESClient := elasticsearch.New(cfg.ElasticSearchAPIURL, elasticHTTPClient, cfg.AWS.Region, cfg.AWS.Service)

	// Initialise query builder
	queryBuilder, err := query.NewQueryBuilder(cfg.ElasticVersion710)
	if err != nil {
		log.Fatal(ctx, "error initialising query builder", err)
		return nil, err
	}

	// Initialise authorisation handler
	permissions := serviceList.GetAuthorisationHandlers(cfg)

	// Get HealthCheck
	healthCheck, err := serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "could not instantiate healthcheck", err)
		return nil, err
	}

	if regErr := registerCheckers(ctx, healthCheck, dpESClient); regErr != nil {
		return nil, errors.Wrap(regErr, "unable to register checkers")
	}

	router := mux.NewRouter()
	server := serviceList.GetHTTPServer(cfg.BindAddr, router)

	router.StrictSlash(true).Path("/health").HandlerFunc(healthCheck.Handler)
	healthCheck.Start(ctx)

	// Create Search API
	searchAPI, err := api.NewSearchAPI(router, dpESClient, deprecatedESClient, queryBuilder, transformerClient, permissions)
	if err != nil {
		log.Fatal(ctx, "error initialising API", err)
		return nil, err
	}

	// Create the interfaces needed, and add route for the new search releases api
	builder, err := query.NewReleaseBuilder()
	if err != nil {
		log.Fatal(ctx, "error initialising release query builder", err)
		return nil, err
	}

	_ = searchAPI.AddSearchReleaseAPI(query.NewReleaseQueryParamValidator(), builder, deprecatedESClient, transformer.NewReleaseTransformer())

	go func() {
		log.Info(ctx, "search api starting")
		if err := server.ListenAndServe(); err != nil {
			log.Error(ctx, "search api http server returned error", err)
			svcErrors <- err
		}
	}()

	return &Service{
		api:                 searchAPI,
		config:              cfg,
		elasticSearchClient: *deprecatedESClient,
		healthCheck:         healthCheck,
		queryBuilder:        queryBuilder,
		router:              router,
		server:              server,
		serviceList:         serviceList,
		transformer:         transformerClient,
	}, nil
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.config.GracefulShutdownTimeout
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout})
	shutdownContext, cancel := context.WithTimeout(ctx, timeout)
	hasShutdownError := false

	// Gracefully shutdown the application closing any open resources.
	go func() {
		defer cancel()

		// stop healthcheck, as it depends on everything else
		if svc.serviceList.HealthCheck {
			svc.healthCheck.Stop()
		}

		// stop any incoming requests before closing any outbound connections
		if err := svc.server.Shutdown(shutdownContext); err != nil {
			log.Error(shutdownContext, "error closing API", err)
			hasShutdownError = true
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-shutdownContext.Done()

	// timeout expired
	if shutdownContext.Err() == context.DeadlineExceeded {
		log.Error(shutdownContext, "shutdown timed out", shutdownContext.Err())
		return shutdownContext.Err()
	}

	// other error
	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Error(shutdownContext, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(shutdownContext, "graceful shutdown was successful")
	cancel()
	return nil
}

func registerCheckers(ctx context.Context, hc HealthChecker, dpESClient *legacyESClient.Client) (err error) {
	if err = hc.AddCheck("Elasticsearch", dpESClient.Checker); err != nil {
		log.Error(ctx, "error creating elasticsearch health check", err)
		err = errors.New("Error(s) registering checkers for health check")
	}
	return err
}
