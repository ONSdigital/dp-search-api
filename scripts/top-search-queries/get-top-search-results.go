package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"

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
	NumResults    string `json:"numResults"`
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
	flag.StringVar(&config.NumResults, "num_results", "10", "number of results to fetch for each query")
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

	log.Info(ctx, "make each request to the Search API with NLP off initially, then NLP on")
	for _, query := range listOfQueries {
		searchResp := getQueryResults(ctx, query, "false", config)
		AddResultsToCSV(ctx, query, csvFile, "false", searchResp)
		searchResp = getQueryResults(ctx, query, "true", config)
		AddResultsToCSV(ctx, query, csvFile, "true", searchResp)
	}
}

// createOutputCSVFile creates the CSV file ready for the query results to be added to. It firstly adds a header row.
func createOutputCSVFile(ctx context.Context) *os.File {
	logData := log.Data{"CSV name: ": "top-search-query-results.csv"}
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

	searchItems := make([]searchItem, 0)
	for i := range sdkItems {
		var searchItem searchItem
		searchItem.Uri = sdkItems[i].URI
		searchItem.Edition = sdkItems[i].Edition
		searchItem.Title = sdkItems[i].Title
		searchItem.Type = sdkItems[i].DataType
		searchItem.Summary = sdkItems[i].Summary
		searchItem.ReleaseDate = sdkItems[i].ReleaseDate
		searchItems = append(searchItems, searchItem)
	}

	var responseObject searchResponse
	responseObject.Items = searchItems
	return responseObject
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
	resultsRow := make([]string, 10)
	resultsRow[0] = nlpOnOrOff
	resultsRow[1] = dateTimeRequest
	resultsRow[2] = querySupplied
	resultsRow[3] = item.ReleaseDate
	resultsRow[4] = item.Type
	resultsRow[5] = strconv.Itoa(rowNum)
	resultsRow[6] = item.Title
	resultsRow[7] = item.Uri
	resultsRow[8] = item.Edition
	resultsRow[9] = item.Summary

	err := csvWriter.Write(resultsRow)
	check(ctx, fmt.Sprintf("failed writing results row for query '%s' at row %d", querySupplied, rowNum), err)
}

// callSearchAPI calls the live Search API using the supplied values of query and nlpWeighting for the relevant query
// parameters. It specifies a limit of 10 results to be returned in the query response.
func callSearchAPI(ctx context.Context, query, nlpWeighting, apiURL, numResults string) []models.Item {
	queryVals := url.Values{}
	queryVals.Add("q", query)
	queryVals.Add("limit", numResults)
	queryVals.Add("nlp_weighting", nlpWeighting)
	options := sdk.Options{
		Query: queryVals,
	}
	logData := log.Data{"query values: ": queryVals}
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
	logData := log.Data{"File name: ": cfg.InputFileName}
	file, err := os.Open(cfg.InputFileName)
	if err != nil {
		log.Fatal(ctx, "failed opening file", err, logData)
	}
	bufferedReader := bufio.NewReader(file)

	var queryString string
	for {
		line, err := bufferedReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// add query from last line, of input file, unless it's empty
				queryString = strings.TrimSpace(line)
				if queryString == "" {
					log.Info(ctx, "ignoring last line of input file as it's empty")
					break
				}
				queryString = url.QueryEscape(queryString)
				listOfQueries = append(listOfQueries, queryString)
				break
			}
			log.Fatal(ctx, "error while reading file - failed reading row of text", err, logData)
			break
		}
		// add query to list
		queryString = strings.TrimSpace(line)
		listOfQueries = append(listOfQueries, queryString)
	}
	if err = file.Close(); err != nil {
		log.Fatal(ctx, "failed closing file", err, logData)
	}
	logData = log.Data{"list of queries: ": listOfQueries}
	log.Info(ctx, "take in list of top queries", logData)
	return listOfQueries, err
}

// writeHeaderRow writes a header row to the supplied CSV file, which will be used for the output.
func writeHeaderRow(csvFile *os.File, ctx context.Context) {
	headerRow := make([]string, 10)
	headerRow[0] = "NLP on or off?"
	headerRow[1] = "Date and time of request"
	headerRow[2] = "Query supplied"
	headerRow[3] = "Date of release"
	headerRow[4] = "Type of release"
	headerRow[5] = "Position in results"
	headerRow[6] = "Title of release"
	headerRow[7] = "URI of release"
	headerRow[8] = "Edition of release"
	headerRow[9] = "Summary of release"

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
