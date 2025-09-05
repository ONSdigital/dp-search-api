package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/dp-search-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

// A searchResponse struct to map the response, from the Search API, to a particular query
type searchResponse struct {
	Items []searchItem `json:"items"`
}

// A SearchItem Struct to map each searchItem of the searchResponse to.
type searchItem struct {
	Type        string `json:"type"`
	Edition     string `json:"edition"`
	ReleaseDate string `json:"release_date"`
	Summary     string `json:"summary"`
	Title       string `json:"title"`
	Uri         string `json:"uri"`
}

type runConfig struct {
	APIURL        string `json:"apiurl"`
	InputFileName string `json:"inputFileName"`
	NumResults    int    `json:"numResults"`
}

// main reads in a list of queries from a text file. These are intended to be the queries that are most commonly used
// in the live Search API service.
//
// It then creates an output CSV file and processes the queries, by calling the live Search API, and outputs
// all the query results to that CSV file.
//
// Each query is called with NLP (Natural Language Processing) initially switched off, then called again with
// NLP switched on. The top 10 results of each query call are added to the output CSV file.
func main() {
	ctx := context.Background()
	log.Info(ctx, "starting script to get results of most common queries from the Search API")

	config := runConfig{}
	flag.StringVar(&config.APIURL, "api_url", "https://api.beta.ons.gov.uk/v1", "the base url for the search api")
	flag.StringVar(&config.InputFileName, "input_file_name", "search-queries.txt", "name of input file including extension")
	flag.IntVar(&config.NumResults, "num_results", 10, "number of results to fetch for each query")
	flag.Parse()
	logData := log.Data{"config: ": config}
	log.Info(ctx, "parsed config", logData)
	run(ctx, config)
	log.Info(ctx, "end of script")
}

func run(ctx context.Context, config runConfig) {
	listOfQueries, err := readQueriesFromFile(ctx, config)
	check(ctx, "failed reading queries file", err)

	csvFile := createOutputCSVFile(ctx)
	defer func() {
		if err = csvFile.Close(); err != nil {
			log.Fatal(ctx, "failed closing csv file", err)
		}
	}()

	log.Info(ctx, "making each request to the Search API with NLP off initially, then NLP on")
	for _, query := range listOfQueries {
		searchResp := getQueryResults(ctx, query, "false", config)
		AddResultsToCSV(ctx, query, csvFile, "false", searchResp)
		searchResp = getQueryResults(ctx, query, "true", config)
		AddResultsToCSV(ctx, query, csvFile, "true", searchResp)
	}
}

// createOutputCSVFile creates the CSV file ready for the query results to be added to. It firstly adds a header row.
func createOutputCSVFile(ctx context.Context) *os.File {
	logData := log.Data{"output_filename": "top-search-query-results.csv"}
	csvFile, err := os.Create("top-search-query-results.csv")
	check(ctx, "failed creating file", err)
	log.Info(ctx, "successfully created csv file", logData)
	writeHeaderRow(csvFile, ctx)
	return csvFile
}

// getQueryResults calls the Search API for a particular query string and NLP weighting (true or false)
// If the NLP weighting is true then NLP is used, if false then NLP is not used. The results of the query call are each
// written to the supplied output CSV file as separate rows.
func getQueryResults(ctx context.Context, querySupplied, nlpWeighting string, cfg runConfig) searchResponse {
	sdkItems := callSearchAPI(ctx, querySupplied, nlpWeighting, cfg.APIURL, cfg.NumResults)

	response := searchResponse{
		Items: make([]searchItem, len(sdkItems)),
	}

	for i, sdkItem := range sdkItems {
		response.Items[i] = searchItem{
			Uri:         sdkItem.URI,
			Edition:     sdkItem.Edition,
			Title:       sdkItem.Title,
			Type:        sdkItem.DataType,
			Summary:     sdkItem.Summary,
			ReleaseDate: sdkItem.ReleaseDate,
		}
	}

	return response
}

// AddResultsToCSV writes the results of a Search API query call to the supplied output CSV file (as separate rows).
func AddResultsToCSV(ctx context.Context, querySupplied string, csvFile *os.File, nlpWeighting string, responseObject searchResponse) {
	now := time.Now()
	dateTimeRequest := now.Format(time.DateTime)
	nlpOnOrOff := "off"
	if nlpWeighting == "true" {
		nlpOnOrOff = "on"
	}
	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	// The number of items in the responseObject will default to 10, but will be whatever is specified by the num_results input parameter, unless fewer are returned.
	for i := 0; i < len(responseObject.Items); i++ {
		item := responseObject.Items[i]
		addItemToCSV(ctx, querySupplied, nlpOnOrOff, dateTimeRequest, item, i, csvWriter)
	}
}

// addItemToCSV adds a row of data, taken from one item of the results (of a particular query), to the output CSV file.
func addItemToCSV(ctx context.Context, querySupplied, nlpOnOrOff, dateTimeRequest string, item searchItem, position int, csvWriter *csv.Writer) {
	rowNum := position + 1
	resultsRow := []string{
		nlpOnOrOff,
		dateTimeRequest,
		querySupplied,
		item.ReleaseDate,
		item.Type,
		strconv.Itoa(rowNum),
		item.Title,
		item.Uri,
		item.Edition,
		item.Summary,
	}

	err := csvWriter.Write(resultsRow)
	check(ctx, fmt.Sprintf("failed writing results row for query '%s' at row %d", querySupplied, rowNum), err)
}

// callSearchAPI calls the live Search API using the supplied values of query and nlpWeighting for the relevant query
// parameters. It specifies a limit of 10 results to be returned in the query response.
func callSearchAPI(ctx context.Context, query, nlpWeighting, apiURL string, numResults int) []models.Item {
	queryVals := url.Values{}
	queryVals.Add("q", query)
	queryVals.Add("limit", strconv.Itoa(numResults))
	queryVals.Add("nlp_weighting", nlpWeighting)
	options := sdk.Options{
		Query: queryVals,
	}
	logData := log.Data{"query_values": queryVals}
	log.Info(ctx, "using SDK to call Search API", logData)
	searchAPIClient := sdk.New(apiURL)
	response, err := searchAPIClient.GetSearch(ctx, options)
	if err != nil {
		log.Fatal(ctx, "failed calling Search API", err, logData)
	}

	return response.Items
}

// readQueriesFromFile opens the text file named "search-queries.txt", in the local directory, and reads it into a
// []string object, which it returns.
func readQueriesFromFile(ctx context.Context, cfg runConfig) (listOfQueries []string, err error) {
	logData := log.Data{"file_name": cfg.InputFileName}
	file, err := os.Open(cfg.InputFileName)
	if err != nil {
		log.Fatal(ctx, "failed opening input file", err, logData)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		listOfQueries = append(listOfQueries, strings.TrimSpace(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(ctx, "failed reading input file", err, logData)
	}

	log.Info(ctx, "queries read from input file", log.Data{"queries": listOfQueries})
	return listOfQueries, err
}

// writeHeaderRow writes a header row to the supplied CSV file, which will be used for the output.
func writeHeaderRow(csvFile *os.File, ctx context.Context) {
	headerRow := []string{
		"NLP on or off?",
		"Date and time of request",
		"Query supplied",
		"Date of release",
		"Type of release",
		"Position in results",
		"Title of release",
		"URI of release",
		"Edition of release",
		"Summary of release",
	}

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()
	err := csvWriter.Write(headerRow)
	check(ctx, "failed writing header row", err)
}

func check(ctx context.Context, msg string, err error) {
	if err != nil {
		log.Fatal(ctx, msg, err)
	}
}
