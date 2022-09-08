package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	dpEs "github.com/ONSdigital/dp-elasticsearch/v3"
	dpEsClient "github.com/ONSdigital/dp-elasticsearch/v3/client"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-search-api/query"
	"github.com/ONSdigital/dp-search-api/transformer"
)

func main() {
	var (
		sr = query.ReleaseSearchRequest{
			Size:   10,
			SortBy: query.Relevance,
			Term:   `\"Birth summary\"`,
			//ReleasedAfter:  query.MustParseDate("2015-01-01"),
			//ReleasedBefore: query.MustParseDate("2015-09-22"),
			Type:        query.Published,
			Provisional: true,
			Confirmed:   true,
			Postponed:   true,
			Census:      true,
			Highlight:   true,
		}
		builder      *query.ReleaseBuilder
		searches     []dpEsClient.Search
		responseData []byte
		err          error
		ctx          = context.Background()
		esClient, _  = dpEs.NewClient(dpEsClient.Config{ClientLib: dpEsClient.GoElasticV710, Address: "http://localhost:9200", Transport: dphttp.DefaultTransport})
	)

	flag.Var(&sr, "sr", "a searchRequest object in json format")
	flag.Parse()

	builder, err = query.NewReleaseBuilder()
	if err != nil {
		log.Fatalf("failed to create builder: %s", err)
	}
	searches, err = builder.BuildSearchQuery(ctx, sr)
	if err != nil {
		log.Fatalf("failed to build query: %s", err)
	}

	fmt.Println("\nsearches are:")
	for _, s := range searches {
		fmt.Printf("%s\n%s\n", s.Header.Index, s.Query)
	}

	responseData, err = esClient.MultiSearch(ctx, searches, nil)
	if err != nil {
		log.Fatalf("elasticsearch query failed: %s", err)
	}

	if !json.Valid(responseData) {
		log.Fatal("elastic search returned invalid JSON for search query")
	}
	fmt.Printf("\nresponse is:\n%s", responseData)

	responseData, err = transformer.NewReleaseTransformer().TransformSearchResponse(ctx, responseData, sr, sr.Highlight)
	if err != nil {
		log.Fatalf("transformation of response data failed: %s", err)
	}

	fmt.Printf("\nprocessed response is:\n%s", responseData)
}
