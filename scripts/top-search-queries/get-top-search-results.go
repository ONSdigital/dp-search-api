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

// A SearchResponse struct to map the response, from the Search API, to a particular query
type SearchResponse struct {
	Items []SearchItem `json:"items"`
}

// A SearchItem Struct to map each SearchItem of the SearchResponse to.
type SearchItem struct {
	Type        string `json:"type"`
	Edition     string `json:"edition"`
	ReleaseDate string `json:"release_date"`
	Summary     string `json:"summary"`
	Title       string `json:"title"`
	Uri         string `json:"uri"`
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

	apiURL := flag.String("api_url", "https://api.beta.ons.gov.uk/v1", "the base url for the search api")
	flag.Parse()

	listOfQueries, err := readQueriesFromFile(ctx)
	check(ctx, "failed reading queries file", err)

	csvFile := createOutputCSVFile(ctx)
	defer csvFile.Close()

	log.Info(ctx, "make each request to the Search API with NLP off initially, then NLP on")
	for _, query := range listOfQueries {
		searchResp := getQueryResults(ctx, query, "false", *apiURL)
		AddResultsToCSV(ctx, query, csvFile, "false", searchResp)
		searchResp = getQueryResults(ctx, query, "true", *apiURL)
		AddResultsToCSV(ctx, query, csvFile, "true", searchResp)
	}
	log.Info(ctx, "end of script")
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
func getQueryResults(ctx context.Context, querySupplied, nlpWeighting, apiURL string) SearchResponse {
	sdkItems := callSearchAPI(ctx, querySupplied, nlpWeighting, apiURL)

	var searchItems []SearchItem
	for _, item := range sdkItems {
		var searchItem SearchItem
		searchItem.Uri = item.URI
		searchItem.Edition = item.Edition
		searchItem.Title = item.Title
		searchItem.Type = item.DataType
		searchItem.Summary = item.Summary
		searchItem.ReleaseDate = item.ReleaseDate
		searchItems = append(searchItems, searchItem)
	}

	var responseObject SearchResponse
	responseObject.Items = searchItems
	return responseObject
}

// AddResultsToCSV writes the results of a Search API query call to the supplied output CSV file (as separate rows).
func AddResultsToCSV(ctx context.Context, querySupplied string, csvFile *os.File, nlpWeighting string, responseObject SearchResponse) {
	now := time.Now()
	dateTimeRequest := now.Format(time.DateTime)
	nlpOnOrOff := "off"
	if nlpWeighting == "true" {
		nlpOnOrOff = "on"
	}
	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	// The number of items in the responseObject will be 10 or fewer (as 10 was specified as the limit in the query)
	for i := 0; i < len(responseObject.Items); i++ {
		item := responseObject.Items[i]
		addItemToCSV(ctx, querySupplied, nlpOnOrOff, dateTimeRequest, item, i, csvWriter)
	}
}

// addItemToCSV adds a row of data, taken from one item of the results (of a particular query), to the output CSV file.
func addItemToCSV(ctx context.Context, querySupplied, nlpOnOrOff, dateTimeRequest string, item SearchItem, position int, csvWriter *csv.Writer) {
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
func callSearchAPI(ctx context.Context, query, nlpWeighting, apiURL string) []models.Item {
	urlStr := fmt.Sprintf("%s/search?q=%s&limit=10&nlp_weighting=%s", apiURL, query, nlpWeighting)
	logData := log.Data{"query to construct: ": urlStr}
	log.Info(ctx, "Using SDK to call Search API", logData)

	queryVals := url.Values{}
	queryVals.Add("q", query)
	queryVals.Add("limit", "10")
	queryVals.Add("nlp_weighting", nlpWeighting)
	options := sdk.Options{
		Query: queryVals,
	}
	logData = log.Data{"query values: ": queryVals}
	searchAPIClient := sdk.New(apiURL)
	response, err := searchAPIClient.GetSearch(ctx, options)
	if err != nil {
		log.Fatal(ctx, "failed calling Search API", err, logData)
	}

	return response.Items
}

// readQueriesFromFile opens the text file named "search-queries.txt", in the local directory, and reads it into a
// []string object, which it returns.
func readQueriesFromFile(ctx context.Context) (listOfQueries []string, err error) {
	logData := log.Data{"File name: ": "search-queries.txt"}
	file, err := os.Open("search-queries.txt")
	if err != nil {
		log.Fatal(ctx, "failed opening file", err, logData)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(ctx, "failed closing file", err, logData)
		}
	}()
	bufferedReader := bufio.NewReader(file)

	for i := 0; i < 10; i++ {
		readStr, err := bufferedReader.ReadString('\n')
		check(ctx, "failed reading row of text", err)
		queryString := strings.TrimSpace(readStr)
		listOfQueries = append(listOfQueries, queryString)
	}

	logData = log.Data{"list of queries: ": listOfQueries}
	log.Info(ctx, "take in list of top 10 queries", logData)
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
