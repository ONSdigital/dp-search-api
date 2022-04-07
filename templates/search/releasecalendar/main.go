package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-search-api/api"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	"github.com/ONSdigital/dp-search-api/query"
	"github.com/ONSdigital/dp-search-api/transformer"
)

func main() {
	var (
		sr = query.ReleaseSearchRequest{
			Size:           10,
			SortBy:         query.RelDateDesc,
			Term:           "Education in Wales",
			ReleasedAfter:  query.MustParseDate("2018-01-01"),
			ReleasedBefore: query.MustParseDate("2018-12-31"),
			Published:      true,
			Highlight:      false,
			Now:            query.Date(time.Now()),
		}
		builder             *query.ReleaseBuilder
		q, uq, responseData []byte
		err                 error
		ctx                                         = context.Background()
		multi                                       = flag.Bool("multi", false, "use the multi query format when sending the query to ES")
		file                                        = flag.String("file", "", "a file containing the actual query to be sent to ES (in json format)")
		esClient                                    = elasticsearch.New("http://localhost:9200", dphttp.NewClient(), "eu-west-1", "es")
		esSearch                                    = esClient.Search
		esTransformer       api.ResponseTransformer = transformer.NewReleaseTransformer()
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
		builder, err = query.NewReleaseBuilder("../../../")
		if err != nil {
			log.Fatalf("failed to create builder: %s", err)
		}

		uq, err = builder.BuildSearchQuery(ctx, sr)
		if err != nil {
			log.Fatalf("failed to build query: %s", err)
		}

		var b bytes.Buffer
		err = json.Compact(&b, uq)
		if err != nil {
			log.Fatalf("failed to compact query: %s", err)
		}
		q = b.Bytes()
	}

	if *multi {
		var buf bytes.Buffer
		buf.Write([]byte(`{"index" : "ons", "type": ["release"], "search_type": "dfs_query_then_fetch"}$$`))
		buf.Write(q)
		buf.Write([]byte(`$$`))
		q, err = query.FormatMultiQuery(buf.Bytes())
		if err != nil {
			log.Fatalf("failed to format multi query: %s", err)
		}
		esSearch = esClient.MultiSearch
		esTransformer = transformer.NewLegacy()
	}

	fmt.Printf("\nformatted query is:\n%s", q)
	responseData, err = esSearch(ctx, "ons", "release", q)
	if err != nil {
		log.Fatalf("elasticsearch query failed: %s", err)
	}

	if !json.Valid(responseData) {
		log.Fatal("elastic search returned invalid JSON for search query")
	}
	fmt.Printf("\nresponse is:\n%s", responseData)

	responseData, err = esTransformer.TransformSearchResponse(ctx, responseData, sr.Term, sr.Highlight)
	if err != nil {
		log.Fatalf("transformation of response data failed: %s", err)
	}

	fmt.Printf("\nprocessed response is:\n%s", responseData)
}
