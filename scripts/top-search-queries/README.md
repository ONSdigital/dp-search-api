# Scripts

Although the Scripts directory is housed within the dp-search-api repository, it is totally separate from the Search API itself.

The intention is that any scripts within the directory are ones that call the Search API independently. 

## Script for Getting Top Search Results

The get-top-search-results script aims to compare the results of queries to the Search API when the NLP (Natural 
Language Processing) functionality is either switched on or off.

It calls the live Search API repeatedly using a supplied list of 10 query strings. Each query string is used twice, which 
is once with the NLP weighting set to false and once with it set to true.

The supplied list of query strings comes from the file named search-queries.txt. These are believed to be the top 10 
most commonly used queries (by users of the Search API).

The results of each query are written to a CSV file named top-search-query-results.csv, which is created/overwritten, 
each time the script is run, in the same directory as the script.

From within the Scripts directory, use the following command to run the script and create the csv output:

`go run get-top-search-results.go`
