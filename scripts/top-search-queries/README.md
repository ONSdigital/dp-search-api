# Scripts

Although the Scripts module is housed within the dp-search-api repository, it is totally separate from the Dissemination Service Search API.

The intention is that any scripts within the module are ones that call the Search API independently. 

## Script for Getting Top Search Results

The get-top-search-results script aims to compare the results of queries to the Search API when the NLP (Natural 
Language Processing) functionality is either switched on or off.

It calls the live Search API repeatedly using a supplied list of query strings. Each query string is used twice, which 
is once with the NLP weighting set to false and once with it set to true.

The supplied list of query strings comes from the first row of the file named top-search-queries.csv. It is so named 
because these are believed to be the queries that are most commonly used, when calling the Search API, at the time of writing.

The results of each query are written to a CSV file named top-search-query-results.csv, which is created/overwritten, each time the script is run,
in the same directory as the script.

From within the Scripts directory, use the following command to run the script and create the csv output:

`go run get-top-search-results.go`
