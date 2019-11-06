package api

import (
	"context"

	"github.com/ONSdigital/dp-search-query/elasticsearch"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var httpServer *server.Server

//SearchQueryAPI provides an API around elasticseach
type SearchQueryAPI struct {
	Router *mux.Router
}

// CreateAndInitialise initiates a new Search Query API
func CreateAndInitialise(bindAddr, elasticSearchAPIURL string, errorChan chan error) error {
	router := mux.NewRouter()

	elasticsearch.Setup(elasticSearchAPIURL)

	errSearch := SetupSearch()
	if errSearch != nil {
		return errors.Wrap(errSearch, "Failed to setup search templates")
	}

	errData := SetupData()
	if errData != nil {
		return errors.Wrap(errData, "Failed to setup data templates")
	}

	errTimeseries := SetupTimeseries()
	if errTimeseries != nil {
		return errors.Wrap(errTimeseries, "Failed to setup timeseries templates")
	}

	api := NewSearchQueryAPI(router)

	httpServer = server.New(bindAddr, api.Router)

	// Disable this here to allow service to manage graceful shutdown of the entire app.
	httpServer.HandleOSSignals = false

	go func() {
		log.Debug("Starting search-query api...", nil)
		if err := httpServer.ListenAndServe(); err != nil {
			log.ErrorC("search-query api http server returned error", err, nil)
			errorChan <- err
		}
	}()

	return nil
}

// NewSearchQueryAPI returns a new Search Query API struct after registerig the routes
func NewSearchQueryAPI(router *mux.Router) *SearchQueryAPI {

	api := &SearchQueryAPI{
		Router: router,
	}

	api.routes()
	return api
}

func (api *SearchQueryAPI) routes() {
	api.Router.HandleFunc("/search", SearchHandler).Methods("GET")
	api.Router.HandleFunc("/timeseries/{cdid}", TimeseriesLookupHandler).Methods("GET")
	api.Router.HandleFunc("/data", DataLookupHandler).Methods("GET")
	api.Router.HandleFunc("/healthcheck", HealthCheckHandlerCreator()).Methods("GET")
}

// Close represents the graceful shutting down of the http server
func Close(ctx context.Context) error {
	if err := httpServer.Shutdown(ctx); err != nil {
		return err
	}
	log.Info("graceful shutdown of http server complete", nil)
	return nil
}
