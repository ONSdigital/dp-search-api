package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ONSdigital/log.go/v2/log"
)

// A Response struct to map the Response
type Response struct {
	Items []Item `json:"items"`
}

// An Item Struct to map every Item to.
type Item struct {
	Type        string    `json:"type"`
	Edition     string    `json:"edition"`
	ReleaseDate time.Time `json:"release_date"`
	Summary     string    `json:"summary"`
	Title       string    `json:"title"`
	Uri         string    `json:"uri"`
}

func main() {
	fmt.Println("Get results of top 10 queries from the Search API")
	ctx := context.Background()

	logData := log.Data{"CSV name: ": "top-search-query-results.csv"}
	csvFile, err := os.Create("top-search-query-results.csv")
	defer csvFile.Close()
	if err != nil {
		log.Fatal(ctx, "failed creating file", err)
	}
	log.Info(ctx, "Successfully created csv file", logData)

	writeHeaderRow(csvFile, err, ctx)
	queries, err := readListOfQueries(ctx)
	log.Info(ctx, "Make requests, using the queries, to the Search API, with NLP on and off")

	// presumably there's a feature flag for switching NLP on and off in prod - not sure how to use that in this context
	// need to call https://api.beta.ons.gov.uk/v1/search?q=rpi&limit=10 for the first query

	listOfQueries := queries[0]

	for _, query := range listOfQueries {
		getQueryResultsAndAddToCSV(ctx, query, csvFile)
	}
}

func getQueryResultsAndAddToCSV(ctx context.Context, querySupplied string, csvFile *os.File) {
	now := time.Now()
	dateTimeRequest := now.Format(time.DateTime)
	logData := log.Data{"Query supplied: ": querySupplied}
	log.Info(ctx, "calling the Search API", logData)
	resultsJson := callSearchAPI(ctx, querySupplied)

	var responseObject Response
	err := json.Unmarshal(resultsJson, &responseObject)
	if err != nil {
		log.Fatal(ctx, "failed to unmarshall response", err)
	}

	fmt.Println("The number of items is: " + strconv.Itoa(len(responseObject.Items)))

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	for position, item := range responseObject.Items {
		resultsRow := make([]string, 10)
		resultsRow[0] = "NLP either on or off"
		resultsRow[1] = dateTimeRequest
		resultsRow[2] = querySupplied
		resultsRow[3] = item.ReleaseDate.Format(time.DateTime)
		resultsRow[4] = item.Type
		resultsRow[5] = strconv.Itoa(position)
		resultsRow[6] = item.Title
		resultsRow[7] = item.Uri
		resultsRow[8] = item.Edition
		resultsRow[9] = item.Summary

		err = csvWriter.Write(resultsRow)
		if err != nil {
			log.Fatal(ctx, fmt.Sprintf("failed writing results row for query '%s' at position %d", querySupplied, position), err)
		}
	}
}

func callSearchAPI(ctx context.Context, query string) []byte {
	urlStr := fmt.Sprintf("https://api.beta.ons.gov.uk/v1/search?q=%s&limit=10", query)
	response, err := http.Get(urlStr)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(ctx, "failed reading API response", err)
	}
	return responseData
}

func readListOfQueries(ctx context.Context) ([][]string, error) {
	logData := log.Data{"CSV name: ": "top-search-queries.csv"}
	file, err := os.Open("top-search-queries.csv")
	if err != nil {
		log.Fatal(ctx, "failed opening file", err, logData)
	}
	reader := csv.NewReader(file)
	queries, err := reader.ReadAll()
	if err != nil {
		log.Fatal(ctx, "failed reading file", err, logData)
	}

	logData = log.Data{"List of queries: ": queries}
	log.Info(ctx, "Take in the list of top 10 queries", logData)
	return queries, err
}

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
	if err != nil {
		log.Fatal(ctx, "failed writing header row", err)
	}
}
