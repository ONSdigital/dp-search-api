package service

import (
	"context"

	"github.com/ONSdigital/dp-search-api/api"
	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	"github.com/ONSdigital/dp-search-api/query"
	"github.com/ONSdigital/dp-search-api/transformer"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	esauth "github.com/ONSdigital/dp-elasticsearch/v2/awsauth"
	elastic "github.com/ONSdigital/dp-elasticsearch/v2/elasticsearch"
	dphttp "github.com/ONSdigital/dp-net/http"
)

type Service struct {
	api                 *api.SearchAPI
	config              *config.Configuration
	elasticSearchClient elasticsearch.Client
	esSigner            *esauth.Signer
	healthCheck         HealthChecker
	queryBuilder        api.QueryBuilder
	router              *mux.Router
	server              HTTPServer
	serviceList         *ExternalServiceList
	transformer         transformer.Transformer
}

// New creates a new service
func New(cfg *config.Configuration, serviceList *ExternalServiceList) *Service {
	svc := &Service{
		config:      cfg,
		serviceList: serviceList,
	}
	return svc
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

// SetEsSigner sets the AWS signer for a service
func (svc *Service) SetEsSigner(esSigner *esauth.Signer) {
	svc.esSigner = esSigner
}

// SetElasticSearchClient sets the new instance of elasticsearch for a service
func (svc *Service) SetElasticSearchClient(elasticSearchClient elasticsearch.Client) {
	svc.elasticSearchClient = elasticSearchClient
}

// SetTransformer sets the transformer for a service
func (svc *Service) SetTransformer(transformer transformer.Transformer) {
	svc.transformer = transformer
}

// Run the service
func Run(ctx context.Context, cfg *config.Configuration, serviceList *ExternalServiceList, buildTime, gitCommit, version string, svcErrors chan error) (svc *Service, err error) {
	var esSigner *esauth.Signer
	elasticHTTPClient := dphttp.NewClient()

	// Initialise transformer
	transformer := transformer.New()

	// Initialse AWS signer
	if cfg.SignElasticsearchRequests {
		esSigner, err = esauth.NewAwsSigner("", "", cfg.AwsRegion, cfg.AwsService)
		if err != nil {
			log.Error(ctx, "failed to create aws v4 signer", err)
			return nil, err
		}
	}

	// Initialse elasticSearchClient
	elasticSearchClient := elasticsearch.New(cfg.ElasticSearchAPIURL, elasticHTTPClient, cfg.SignElasticsearchRequests, esSigner, cfg.AwsRegion, cfg.AwsService)

	// Initialise query builder
	queryBuilder, err := query.NewQueryBuilder()
	if err != nil {
		log.Fatal(ctx, "error initialising query builder", err)
		return nil, err
	}

	//Get HealthCheck
	healthCheck, err := serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "could not instantiate healthcheck", err)
		return nil, err
	}

	if err := registerCheckers(ctx, *cfg, healthCheck, elasticHTTPClient, esSigner); err != nil {
		return nil, errors.Wrap(err, "unable to register checkers")
	}

	router := mux.NewRouter()
	server := serviceList.GetHTTPServer(cfg.BindAddr, router)

	router.StrictSlash(true).Path("/health").HandlerFunc(healthCheck.Handler)
	healthCheck.Start(ctx)

	// Create Search API
	searchAPI, err := api.NewSearchAPI(router, elasticSearchClient, queryBuilder, transformer)
	if err != nil {
		log.Fatal(ctx, "error initialising API", err)
		return nil, err
	}

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
		elasticSearchClient: *elasticSearchClient,
		esSigner:            esSigner,
		healthCheck:         healthCheck,
		queryBuilder:        queryBuilder,
		router:              router,
		server:              server,
		serviceList:         serviceList,
		transformer:         *transformer,
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

func registerCheckers(ctx context.Context,
	cfg config.Configuration,
	hc HealthChecker,
	elasticHTTPClient dphttp.Clienter,
	esSigner *esauth.Signer) (err error) {

	hasErrors := false

	elasticClient := elastic.NewClientWithHTTPClientAndAwsSigner(cfg.ElasticSearchAPIURL, esSigner, cfg.SignElasticsearchRequests, elasticHTTPClient)
	if err = hc.AddCheck("Elasticsearch", elasticClient.Checker); err != nil {
		log.Error(ctx, "error creating elasticsearch health check", err)
		hasErrors = true
	}

	if hasErrors {
		errors.New("Error(s) registering checkers for healthcheck")
	}

	return nil
}
