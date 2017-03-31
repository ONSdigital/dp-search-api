package main

import (
	"os"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/dp-search-query/handlers"
	"github.com/gorilla/pat"
	"net/http"
	"time"
)

var server *http.Server

func main() {
	bindAddr := os.Getenv("BIND_ADDR")
	log.Debug(bindAddr,nil)
	if len(bindAddr) == 0 {
		bindAddr = ":10001"
	}

	log.Namespace = "dp-search-query"

	router := pat.New()

	log.Debug("main", log.Data{"StartingServer":"Start"})
	router.Get("/search", handlers.SearchHandler)
	router.Get("/timeseries/{cdid}", handlers.TimeseriesLookupHandler)
	server = &http.Server{
		Addr:         bindAddr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	var err error
	if err = server.ListenAndServe(); err != nil {
		log.Debug("main bindevent", log.Data{"failed to bind to":bindAddr})
		log.Error(err, nil)
		os.Exit(1)
	}

}

func stop(){
	server.Close()
}

