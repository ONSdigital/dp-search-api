dp-search-api
================
Digital Publishing Search API

A Go application microservice to provide query functionality on the ONS Website

### Getting started

There are now 2 versions of ElasticSearch being used by this service:
* 2.4.2 (the old/existing ElasticSearch)
* 7.10.0 (Site Wide ElasticSearch)

Version 2.4.2 is required by all endpoints except for the POST /search endpoint, which uses 7.10

Set up dependencies locally as follows:

* In dp-compose run `docker-compose up -d` to run both versions of ElasticSearch
* NB. Version 2.4.2 will run on port 9200, version 7.10 will run on port 11200
* If using endpoints that require version 2.4.2, there are no more dependencies to set up.
* NB. What endpoints are available depends on what port the ELASTIC_SEARCH_URL points to. By default, it points to port 9200, so to point to 11200 (if required) run the following:
* `export ELASTIC_SEARCH_URL="http://localhost:11200"`
* If using the POST /search endpoint then authorisation for this requires running Vault and Zebedee as follows:
* In any directory run `vault server -dev` as Zebedee has a dependency on Vault
* In the zebedee directory run `./run.sh` to run Zebedee

Then run `make debug`

### Dependencies

For endpoints that require the old/existing ElasticSearch (version 2.4.2):
* Requires ElasticSearch running on port 9200
* No further dependencies other than those defined in `go.mod`

For endpoints that require the Site Wide ElasticSearch (version 7.10.0):
* Requires ElasticSearch running on port 11200
* Requires Zebedee running on port 8082
* No further dependencies other than those defined in `go.mod`

### Configuration

An overview of the configuration options available, either as a table of
environment variables, or with a link to a configuration guide.

| Environment variable | Default | Description
| -------------------- | ------- | -----------
| AWS_FILENAME                 | ""                       | The AWS file location for finding credentials to sign AWS http requests
| AWS_PROFILE                  | ""                       | The AWS profile to use from credentials file to sign AWS http requests
| AWS_REGION                   | eu-west-1                | The AWS region to use when signing requests with AWS SDK
| AWS_SERVICE                  | "es"                     | The AWS service that the AWS SDK signing mechanism needs to sign a request
| AWS_TLS_INSECURE_SKIP_VERIFY | false                    | This should never be set to true, as it disables SSL certificate verification. Used only for development
| BIND_ADDR                    | :23900                   | The host and port to bind to
| ELASTIC_SEARCH_URL	       | "http://localhost:9200"  | Http url of the ElasticSearch server. For Site Wide ElasticSearch this needs to be set to "http://localhost:11200".
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                       | The graceful shutdown timeout in seconds (`time.Duration` format)
| SIGN_ELASTICSEARCH_REQUESTS  | false                    | Boolean flag to identify whether elasticsearch requests via elastic API need to be signed if elasticsearch cluster is running in aws
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                      | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)
| HEALTHCHECK_INTERVAL         | 30s                      | Time between self-healthchecks (`time.Duration` format)
| ZEBEDEE_URL                  | http://localhost:8082    | The URL to Zebedee (for authorisation)


### Connecting to AWS Elasticsearch cluster (dev only)

#### Prerequisites

- You will need an user account for the aws account you are trying to connect to
- You will need to be given a policy to allow read and write access to the AWS Elasticsearch cluster

#### Connecting

To connect to managed Elasticsearch cluster in AWS, you will want to port forward 9200 to the domain endpoint. Using the dp tool, one can do this like so:

```
  dp ssh develop <ip of aws box> -p 9200:<elasticsearch cluster domain endpoint e.g. "<unique identifier>" + "eu-west-1.es.amazonaws.com:443"
```

Once connected, run the following make target:

```
  make local
```

### Running Bulk Indexer

#### Locally

Build the bulk indexer by running the following command
```
  make reindex
```
Then run the executable  
```
  ./reindex
```
Please make sure your elasticsearch server is running locally on localhost:9200 and version of the server is 7.10, which is the current supported version.

#### Remote Environment

###### Prerequisites

Before attempting the following steps please make sure you have dp tool setup locally.For more info on setting up the dp tool: https://github.com/ONSdigital/dp-cli#build-and-run

###### Steps

Navigate to ```cmd/reindex/aws.go``` and update esurl to correct elastic search 7.10 url in environment and then build the bulk indexer by running the following command
```
  make build-reindex
```
Then copy to your environment build directory  by running the following command
```
dp scp <environment> publishing <box> ./build/reindex <location on publishing box>
```
For example
```
dp scp develop publishing 2 ./build/reindex .
```
Then ssh into the box as follows
```
dp ssh <environment> publishing <box>
```
Then run the executable
```
  ./build/reindex
```

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2022, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
