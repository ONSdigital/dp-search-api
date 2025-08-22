package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ONSdigital/log.go/v2/log"
)

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
	// need to call https://api.beta.ons.gov.uk/v1/search?q=rpi for the first query

	listOfQueries := queries[0]
	fmt.Println("The first query is: " + listOfQueries[0])

	resultStr := callSearchAPI(ctx, listOfQueries[0])
	logData = log.Data{"Search API results: ": resultStr}
	log.Info(ctx, "Successfully called Search API", logData)

	// add first results to csv
	// start by deleting everything in the resultStr before the second occurrence of "items"
	firstSegments := strings.Split(resultStr, "items")
	//for index, segment := range firstSegments {
	//	fmt.Println("The number " + strconv.Itoa(index) + " segment is: " + segment)
	//}

	// discard the first 2 segments but rejoin the remaining ones
	resultStr2 := "{\"items" + firstSegments[2]
	for i := 3; i < len(firstSegments); i++ {
		resultStr2 = resultStr2 + "items" + firstSegments[i]
	}
	fmt.Println("New result string: " + resultStr2)

	//The new result string is valid json and it only contains the information that we're interested in
}

func callSearchAPI(ctx context.Context, query string) string {
	urlStr := fmt.Sprintf("https://api.beta.ons.gov.uk/v1/search?q=%s", query)
	response, err := http.Get(urlStr)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(ctx, "failed reading API response", err)
	}
	searchResults := string(responseData)
	return searchResults
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
