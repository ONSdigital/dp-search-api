package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-search-api/api"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	"github.com/ONSdigital/dp-search-api/query"
	"github.com/ONSdigital/dp-search-api/transformer"
	"github.com/ONSdigital/log.go/v2/log"
)

func main() {
	var (
		sr = query.ReleaseSearchRequest{
			Size:           10,
			SortBy:         query.RelDateDesc,
			Term:           "Education in Wales",
			ReleasedAfter:  query.MustParseDate("2018-01-01"),
			ReleasedBefore: query.MustParseDate("2018-12-31"),
			//Upcoming: true,
			Published: true,
			Highlight: false,
			Now:       query.Date(time.Now()),
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

	log.Namespace = "dp-search-api/sitewide-test"

	if *file != "" {
		switch *file {
		case "-":
			q, err = io.ReadAll(os.Stdin)
		default:
			q, err = os.ReadFile(*file)
		}
		if err != nil {
			log.Error(ctx, "failed to read query from file", err)
			os.Exit(1)
		}
	} else {
		builder, err = query.NewReleaseBuilder("../../../")
		if err != nil {
			log.Error(ctx, "failed to create builder", err)
			os.Exit(2)
		}

		uq, err = builder.BuildSearchQuery(ctx, sr)
		if err != nil {
			log.Error(ctx, "failed to build query", err)
			os.Exit(3)
		}

		var b bytes.Buffer
		err = json.Compact(&b, uq)
		if err != nil {
			log.Error(ctx, "failed to compact query", err)
			os.Exit(4)
		}
		q = b.Bytes()
	}

	if *multi {
		q, err = query.FormatMultiQuery(append([]byte(`{"index" : "ons", "type": ["release"], "search_type": "dfs_query_then_fetch"}$$`), append(q, []byte(`$$`)...)...))
		if err != nil {
			log.Error(ctx, "failed to format multi query", err)
			os.Exit(5)
		}
		esSearch = esClient.MultiSearch
		esTransformer = transformer.New()
	}

	fmt.Printf("\nformatted query is:\n%s", q)
	responseData, err = esSearch(ctx, "ons", "release", q)
	if err != nil {
		log.Error(ctx, "elasticsearch query failed", err)
		return
	}

	if !json.Valid(responseData) {
		log.Error(ctx, "elastic search returned invalid JSON for search query", errors.New("elastic search returned invalid JSON for search query"))
		return
	}
	fmt.Printf("\nresponse is:\n%s", responseData)

	responseData, err = esTransformer.TransformSearchResponse(ctx, responseData, sr.Term, sr.Highlight)
	if err != nil {
		log.Error(ctx, "transformation of response data failed", err)
		return
	}

	fmt.Printf("\nprocessed response is:\n%s", responseData)
}
