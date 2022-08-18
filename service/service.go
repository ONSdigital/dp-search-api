package service

import (
	"context"

	dpEsClient "github.com/ONSdigital/dp-elasticsearch/v3/client"
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
	API                 *api.SearchAPI
	config              *config.Config
	healthCheck         HealthChecker
	server              HTTPServer
	serviceList         *ExternalServiceList
	elasticSearchClient elasticsearch.Client
	elasticSearchServer dpEsClient.Client
	queryBuilder        api.QueryBuilder
	transformer         api.ResponseTransformer
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
	// Inject all external services
	healthCheck, err := serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "could not instantiate healthcheck", err)
		return nil, err
	}
	permissions := serviceList.GetAuthorisationHandlers(cfg)
	esClient, err := serviceList.GetElasticSearchServer(cfg)
	if err != nil {
		log.Fatal(ctx, "could not initialise the ElasticSearch server", err)
		return nil, err
	}
	// temporary workaround while we maintain the deprecatedESClient
	deprecatedESClient := elasticsearch.New(cfg.ElasticSearchAPIURL, dphttp.NewClient(), cfg.AWS.Region, cfg.AWS.Service)

	// Create all internal functional components
	queryBuilder, err := query.NewQueryBuilder()
	if err != nil {
		log.Fatal(ctx, "error initialising query builder", err)
		return nil, err
	}

	releasebuilder, err := query.NewReleaseBuilder()
	if err != nil {
		log.Fatal(ctx, "error initialising release query builder", err)
		return nil, err
	}

	transformerClient := transformer.New()

	// Create the router and populate it with handled routes in the api
	router := mux.NewRouter()
	router.StrictSlash(true).Path("/health").HandlerFunc(healthCheck.Handler)

	searchAPI, err := api.NewSearchAPI(router, esClient, deprecatedESClient, queryBuilder, transformerClient, permissions)
	if err != nil {
		log.Fatal(ctx, "error initialising API", err)
		return nil, err
	}
	_ = searchAPI.AddSearchReleaseAPI(query.NewReleaseQueryParamValidator(), releasebuilder, esClient, deprecatedESClient, transformer.NewReleaseTransformer())

	// Register the checkers
	if regErr := registerCheckers(ctx, healthCheck, esClient); regErr != nil {
		return nil, errors.Wrap(regErr, "unable to register checkers")
	}

	// Create the server - *** this is the only place where the idea of injection of 'independent' external services breaks down; because the server
	// explicitly needs the router, which is a major piece of internal plumbing; All the rest of the internal plumbing use the injected external services
	server := serviceList.GetHTTPServer(cfg.BindAddr, router)

	// Start the health check
	healthCheck.Start(ctx)

	// Start the server
	go func() {
		log.Info(ctx, "search api starting")
		if err := server.ListenAndServe(); err != nil {
			log.Error(ctx, "search api http server returned error", err)
			svcErrors <- err
		}
	}()

	return &Service{
		API:                 searchAPI,
		config:              cfg,
		healthCheck:         healthCheck,
		serviceList:         serviceList,
		server:              server,
		elasticSearchClient: *deprecatedESClient,
		elasticSearchServer: esClient,
		queryBuilder:        queryBuilder,
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

func registerCheckers(ctx context.Context, hc HealthChecker, dpESClient dpEsClient.Client) (err error) {
	if err = hc.AddCheck("Elasticsearch", dpESClient.Checker); err != nil {
		log.Error(ctx, "error creating elasticsearch health check", err)
		err = errors.New("Error(s) registering checkers for health check")
	}
	return err
}
