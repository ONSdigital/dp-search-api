package main

import (
	"context"
	"fmt"

	"github.com/ONSdigital/log.go/v2/log"
)

func main() {
	fmt.Println("Get results of top 10 queries from the Search API")

	ctx := context.Background()

	// TODO Hard coded list of queries for now - find out how to read them in later
	listOfQueries := make([]string, 10)
	listOfQueries[0] = "rpi"
	listOfQueries[1] = "cpi"
	listOfQueries[2] = "population"
	listOfQueries[3] = "life expectancy"
	listOfQueries[4] = "inflation"
	listOfQueries[5] = "gdp"
	listOfQueries[6] = "life expectancy calculator"
	listOfQueries[7] = "crime"
	listOfQueries[8] = "cpih"
	listOfQueries[9] = "ashe"

	logData := log.Data{"List of queries: ": listOfQueries}
	log.Info(ctx, "Take in the list of top 10 queries", logData)
	log.Info(ctx, "Make requests, using the queries, to the Search API, with NLP on and off")

}
