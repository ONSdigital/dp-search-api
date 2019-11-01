package api

import (
	"context"

	"github.com/ONSdigital/dp-search-query/elasticsearch"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

var httpServer *server.Server

//SearchQueryAPI provides an API around elasticseach
type SearchQueryAPI struct {
	Router *mux.Router
}

// CreateAndInitialise initiates a new Search Query API
func CreateAndInitialise(bindAddr, elasticSearchAPIURL string, errorChan chan error) {
	router := mux.NewRouter()
	// Setup libraries and handlers
	elasticsearch.Setup(elasticSearchAPIURL)
	errSearch := SetupSearch()
	if errSearch != nil {
		log.ErrorC("Failed to setup search templates", errSearch, log.Data{})
	}
	errData := SetupData()
	if errData != nil {
		log.ErrorC("Failed to setup data templates", errData, log.Data{})
	}
	errTimeseries := SetupTimeseries()
	if errTimeseries != nil {
		log.ErrorC("Failed to setup timeseries templates", errTimeseries, log.Data{})
	}

	api := NewSearchQueryAPI(router)

	httpServer = server.New(bindAddr, api.Router)

	// Disable this here to allow service to manage graceful shutdown of the entire app.
	httpServer.HandleOSSignals = false

	go func() {
		log.Debug("Starting api...", nil)
		if err := httpServer.ListenAndServe(); err != nil {
			log.ErrorC("api http server returned error", err, nil)
			errorChan <- err
		}
	}()
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
	router := api.Router
	router.HandleFunc("/search", SearchHandler).Methods("GET")
	router.HandleFunc("/timeseries/{cdid}", TimeseriesLookupHandler).Methods("GET")
	router.HandleFunc("/data", DataLookupHandler).Methods("GET")
	router.HandleFunc("/healthcheck", HealthCheckHandlerCreator()).Methods("GET") // TODO Replace with healthcheck middleware
}

// Close represents the graceful shutting down of the http server
func Close(ctx context.Context) error {
	if err := httpServer.Shutdown(ctx); err != nil {
		return err
	}
	log.Info("graceful shutdown of http server complete", nil)
	return nil
}
