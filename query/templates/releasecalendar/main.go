package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	"github.com/ONSdigital/dp-search-api/query"
	"github.com/ONSdigital/dp-search-api/transformer"
)

func main() {
	var (
		sr = query.LegacyReleaseSearchRequest{
			ReleaseSearchRequest: query.ReleaseSearchRequest{
				Size:   10,
				SortBy: query.TitleAsc,
				Term:   "",
				//ReleasedAfter:  query.MustParseDate("2015-01-01"),
				//ReleasedBefore: query.MustParseDate("2015-09-22"),
				Type:        query.Published,
				Provisional: true,
				Confirmed:   true,
				Postponed:   true,
				//Census:      true,
				Highlight: true,
			}}
		builder         *query.ReleaseBuilder
		q, responseData []byte
		err             error
		ctx             = context.Background()
		file            = flag.String("file", "", "a file containing the actual multi-query to be sent to ES (in json format)")
		esClient        = elasticsearch.New("http://localhost:9200", dphttp.NewClient(), "eu-west-1", "es")
	)

	flag.Var(&sr, "sr", "a searchRequest object in json format")
	flag.Parse()

	if *file != "" {
		switch *file {
		case "-":
			q, err = io.ReadAll(os.Stdin)
		default:
			q, err = os.ReadFile(*file)
		}
		if err != nil {
			log.Fatalf("failed to read query from file: %s", err)
		}
	} else {
		builder, err = query.NewReleaseBuilder()
		if err != nil {
			log.Fatalf("failed to create builder: %s", err)
		}

		q, err = builder.BuildSearchQuery(ctx, sr)
		if err != nil {
			log.Fatalf("failed to build query: %s", err)
		}
	}

	fmt.Printf("\nformatted query is:\n%s", q)
	responseData, err = esClient.MultiSearch(ctx, "ons", "release", q)
	if err != nil {
		log.Fatalf("elasticsearch query failed: %s", err)
	}

	if !json.Valid(responseData) {
		log.Fatal("elastic search returned invalid JSON for search query")
	}
	fmt.Printf("\nresponse is:\n%s", responseData)

	responseData, err = transformer.NewReleaseTransformer().TransformSearchResponse(ctx, responseData, sr.ReleaseSearchRequest, sr.Highlight)
	if err != nil {
		log.Fatalf("transformation of response data failed: %s", err)
	}

	fmt.Printf("\nprocessed response is:\n%s", responseData)
}
