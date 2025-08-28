package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ONSdigital/log.go/v2/log"
)

// A Response struct to map the Response, from the Search API, to a particular query
type Response struct {
	Items []Item `json:"items"`
}

// An Item Struct to map each Item of the Response to.
type Item struct {
	Type        string    `json:"type"`
	Edition     string    `json:"edition"`
	ReleaseDate time.Time `json:"release_date"`
	Summary     string    `json:"summary"`
	Title       string    `json:"title"`
	Uri         string    `json:"uri"`
}

// main reads in a list of 10 queries from the first row of a CSV file. These are understood to be the top 10 queries
// that are used in the live Search API service. It processes the queries, by calling the live Search API, and outputs
// all the query results to another CSV file, which it creates.
//
// Each query is called with NLP (Natural Language Processing) initially switched off, then called again with
// NLP switched on. The top 10 results of each query call are added to the output CSV file.
func main() {
	ctx := context.Background()
	log.Info(ctx, "starting script to get results of top 10 queries from the Search API")

	logData := log.Data{"CSV name: ": "top-search-query-results.csv"}
	csvFile, err := os.Create("top-search-query-results.csv")
	defer csvFile.Close()
	check(ctx, "failed creating file", err)
	log.Info(ctx, "successfully created csv file", logData)

	writeHeaderRow(csvFile, err, ctx)
	listOfQueries, err := readQueriesFromFile(ctx)
	check(ctx, "failed reading queries file", err)

	log.Info(ctx, "make each request to the Search API with NLP off initially, then NLP on")
	for _, query := range listOfQueries {
		getQueryResultsAndAddToCSV(ctx, query, csvFile, "false")
		getQueryResultsAndAddToCSV(ctx, query, csvFile, "true")
	}
	log.Info(ctx, "end of script")
}

// getQueryResultsAndAddToCSV calls the Search API for a particular query string and NLP weighting (true or false)
// If the NLP weighting is true then NLP is used, if false then NLP is not used. The results of the query call are each
// written to the supplied output CSV file as separate rows.
func getQueryResultsAndAddToCSV(ctx context.Context, querySupplied string, csvFile *os.File, nlpWeighting string) {
	now := time.Now()
	dateTimeRequest := now.Format(time.DateTime)
	nlpOnOrOff := "off"
	if nlpWeighting == "true" {
		nlpOnOrOff = "on"
	}
	resultsJson := callSearchAPI(ctx, querySupplied, nlpWeighting)

	var responseObject Response
	err := json.Unmarshal(resultsJson, &responseObject)
	check(ctx, "failed to unmarshall response", err)

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	// The number of items in the responseObject will be 10 or fewer (as 10 is specified as the limit in the query)
	for i := 0; i < len(responseObject.Items); i++ {
		item := responseObject.Items[i]
		addItemToCSV(ctx, querySupplied, nlpOnOrOff, dateTimeRequest, item, i, csvWriter)
	}
}

// addItemToCSV adds a row of data, taken from one item of the results (of a particular query), to the output CSV file.
func addItemToCSV(ctx context.Context, querySupplied string, nlpOnOrOff string, dateTimeRequest string, item Item, position int, csvWriter *csv.Writer) {
	rowNum := position + 1
	resultsRow := make([]string, 10)
	resultsRow[0] = nlpOnOrOff
	resultsRow[1] = dateTimeRequest
	resultsRow[2] = querySupplied
	resultsRow[3] = item.ReleaseDate.Format(time.DateTime)
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
func callSearchAPI(ctx context.Context, query string, nlpWeighting string) []byte {
	urlStr := fmt.Sprintf("https://api.beta.ons.gov.uk/v1/search?q=%s&limit=10&nlp_weighting=%s", query, nlpWeighting)
	logData := log.Data{"query: ": urlStr}
	log.Info(ctx, "call Search API", logData)
	response, err := http.Get(urlStr)
	check(ctx, "failed calling Search API", err)

	responseData, err := io.ReadAll(response.Body)
	check(ctx, "failed reading API response", err)
	return responseData
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
func writeHeaderRow(csvFile *os.File, err error, ctx context.Context) {
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
	err = csvWriter.Write(headerRow)
	check(ctx, "failed writing header row", err)
}

func check(ctx context.Context, msg string, err error) {
	if err != nil {
		log.Fatal(ctx, msg, err)
	}
}
