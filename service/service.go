package service

import (
	"context"
	esauth "github.com/ONSdigital/dp-elasticsearch/v2/awsauth"
	elastic "github.com/ONSdigital/dp-elasticsearch/v2/elasticsearch"
	dphttp "github.com/ONSdigital/dp-net/http"
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
	config              *config.Configuration
	healthCheck         HealthChecker
	api                 *api.SearchAPI
	router              *mux.Router
	server              HTTPServer
	serviceList         *ExternalServiceList
	esSigner            *esauth.Signer
	queryBuilder        api.QueryBuilder
	elasticSearchClient elasticsearch.Client
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
func (svc *Service) Run(ctx context.Context, buildTime, gitCommit, version string, svcErrors chan error) (err error) {
	elasticHTTPClient := dphttp.NewClient()
	//Get HealthCheck
	svc.healthCheck, err = svc.serviceList.GetHealthCheck(svc.config, buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "could not instantiate healthcheck", err)
		return err
	}

	if err := svc.registerCheckers(ctx, elasticHTTPClient); err != nil {
		return errors.Wrap(err, "unable to register checkers")
	}

	svc.router = mux.NewRouter()
	svc.router.StrictSlash(true).Path("/health").HandlerFunc(svc.healthCheck.Handler)
	svc.healthCheck.Start(ctx)

	// Initialise transformer
	transformer := transformer.New()

	// Initialse elasticSearchClient
	elasticSearchClient := elasticsearch.New(svc.config.ElasticSearchAPIURL, dphttp.NewClient(), svc.config.SignElasticsearchRequests, svc.esSigner, svc.config.AwsRegion, svc.config.AwsService)

	// Initialse AWS signer
	if svc.config.SignElasticsearchRequests {
		svc.esSigner, err = esauth.NewAwsSigner("", "", svc.config.AwsRegion, svc.config.AwsService)
		if err != nil {
			log.Error(ctx, "failed to create aws v4 signer", err)
			return err
		}
	}

	// Initialise query builder
	queryBuilder, err := query.NewQueryBuilder()
	if err != nil {
		log.Fatal(ctx, "error initialising query builder", err)
		return err
	}

	// Create Search API
	if err := api.CreateAndInitialise(svc.config, svc.router, queryBuilder, elasticSearchClient, transformer, svcErrors); err != nil {
		log.Fatal(ctx, "error initialising API", err)
		return err
	}

	return nil
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

func (svc *Service) registerCheckers(ctx context.Context, elasticHTTPClient dphttp.Clienter) (err error) {
	hasErrors := false
	elasticClient := elastic.NewClientWithHTTPClientAndAwsSigner(svc.config.ElasticSearchAPIURL, svc.esSigner, svc.config.SignElasticsearchRequests, elasticHTTPClient)
	if err = svc.healthCheck.AddCheck("Elasticsearch", elasticClient.Checker); err != nil {
		log.Error(ctx, "error creating elasticsearch health check", err)
		hasErrors = true
	}

	if hasErrors {
		errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
