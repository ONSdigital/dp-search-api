package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	dpelasticsearch "github.com/ONSdigital/dp-elasticsearch/v2/elasticsearch"
	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	"github.com/ONSdigital/log.go/v2/log"
)

const zebedeeURI = "http://localhost:8082"

var (
	maxConcurrentExtractions = 20
	maxConcurrentIndexings   = 20
)

type elasticSearchClient interface {
	CreateIndex(ctx context.Context, indexName string, indexSettings []byte) (int, error)
	AddDocument(ctx context.Context, indexName, documentType, documentID string, document []byte) (int, error)
}

type Document struct {
	URI  string
	Body []byte
}

func main() {
	ctx := context.Background()
	cfg, err := config.Get()
	if err != nil {
		log.Fatal(ctx, "error retrieving config", err)
		os.Exit(1)
	}

	esClient := dpelasticsearch.NewClient(cfg.ElasticSearchAPIURL, cfg.SignElasticsearchRequests, 5)

	_ = docIndexer(ctx, esClient)
}

func docIndexer(ctx context.Context, es elasticSearchClient) error {

	indexName := createIndexName("ons")
	fmt.Printf("Index created: %s\n", indexName)
	status, err := es.CreateIndex(ctx, indexName, elasticsearch.GetSearchIndexSettings())
	if err != nil {
		log.Error(ctx, "error creating index", err)
		return err
	}
	if status != http.StatusOK {
		log.Error(ctx, "error creating index http status - ", fmt.Errorf("error, status: %v", err))
		return err
	}

	log.Info(ctx, "successfully created index")

	return nil
}

func createIndexName(s string) string {
	now := time.Now()
	return fmt.Sprintf("%s%d", s, now.UnixMicro())
}
