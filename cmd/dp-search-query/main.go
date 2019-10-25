package main

import (
	"net/http"
	"os"
	"time"

	"github.com/ONSdigital/dp-search-query/config"
	"github.com/ONSdigital/dp-search-query/elasticsearch"
	"github.com/ONSdigital/dp-search-query/handlers"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/pat"
)

var server *http.Server

func getEnv(key string, defaultValue string) string {
	envValue := os.Getenv(key)
	if len(envValue) == 0 {
		envValue = defaultValue
	}
	return envValue
}

func main() {
	log.Namespace = "dp-search-query"

	cfg, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	// sensitive fields are omitted from config.String().
	log.Info("config on startup", log.Data{"config": cfg})

	log.Debug("Starting server", log.Data{"Port": cfg.BindAddr, "ElasticSearchUrl": cfg.ElasticSearchAPIURL})

	// Setup libraries and handlers
	elasticsearch.Setup(cfg.ElasticSearchAPIURL)
	errSearch := handlers.SetupSearch()
	if errSearch != nil {
		log.ErrorC("Failed to setup search templates", errSearch, log.Data{})
	}
	errData := handlers.SetupData()
	if errData != nil {
		log.ErrorC("Failed to setup data templates", errData, log.Data{})
	}
	errTimeseries := handlers.SetupTimeseries()
	if errTimeseries != nil {
		log.ErrorC("Failed to setup timeseries templates", errTimeseries, log.Data{})
	}

	// Setup web handlers for the search query services
	router := pat.New()
	router.Get("/search", handlers.SearchHandler)
	router.Get("/timeseries/{cdid}", handlers.TimeseriesLookupHandler)
	router.Get("/data", handlers.DataLookupHandler)
	router.Get("/healthcheck", handlers.HealthCheckHandlerCreator()) // TODO Replace with healthcheck middleware
	server = &http.Server{
		Addr:         cfg.BindAddr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.ErrorC("Failed to bind to port address", err, log.Data{"Port": cfg.BindAddr})
	}
}
