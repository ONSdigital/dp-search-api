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
	fmt.Println("The first query is: " + listOfQueries[0])

	log.Info(ctx, "calling the Search API")
	resultsJson := callSearchAPI(ctx, listOfQueries[0])

	var responseObject Response
	err = json.Unmarshal(resultsJson, &responseObject)
	if err != nil {
		log.Fatal(ctx, "failed to unmarshall response", err)
	}

	fmt.Println("The number of items is: " + strconv.Itoa(len(responseObject.Items)))

	for _, item := range responseObject.Items {
		fmt.Println(item.Title)
		fmt.Println(item.ReleaseDate)
		fmt.Println(item.Uri)
		fmt.Println(item.Type)
		fmt.Println(item.Edition)
		fmt.Println(item.Summary)
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
	headerRow[0] = "uri"
	headerRow[1] = "type"
	headerRow[2] = "release_date"
	headerRow[3] = "title"
	headerRow[4] = "edition"
	headerRow[5] = "summary"
	headerRow[6] = "nlp on or off?"
	headerRow[7] = "query supplied"
	headerRow[8] = "position (array index)"
	headerRow[9] = "date and time of the request"

	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()
	err = csvwriter.Write(headerRow)
	if err != nil {
		log.Fatal(ctx, "failed writing header row", err)
	}
}
