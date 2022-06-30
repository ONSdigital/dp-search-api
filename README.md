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
| AWS_SIGNER                   | false                    | The AWS signer flag will determine if requests to Elasticsearch contain round tripper for signing requests
| AWS_TLS_INSECURE_SKIP_VERIFY | false                    | This should never be set to true, as it disables SSL certificate verification. Used only for development
| BIND_ADDR                    | :23900                   | The host and port to bind to
| ELASTIC_SEARCH_URL	       | "http://localhost:9200"  | Http url of the ElasticSearch server. For Site Wide ElasticSearch this needs to be set to "http://localhost:11200".
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                       | The graceful shutdown timeout in seconds (`time.Duration` format)
| ELASTIC_VERSION_710          | false                    | Boolean flag to switch on the elastic version 7.10. By default this will point to the legacy ES version which is 2.2
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

The Bulk Indexer creates an Elastic Search 7.10 index with the alias "ons". It loads data from zebedee, and the dataset API, into the ES 7.10 index.

#### Locally

###### Prerequisites

* Requires ElasticSearch 7.10 running on port 9200 (NOT 11200)
* Requires Zebedee running on port 8082 (and this has a dependency on vault)
* Requires the Dataset API running on port 22000
* Requires a local service auth token value to be put in line 35 of cmd/reindex/local.go

NB. The Dataset API requires a mongo database named 'datasets', which must contain the following collections:
* contacts
* datasets
* dimension.options
* editions
* instances
* instances_locks

The Dataset API also requires this environment variable to be set to true: DISABLE_GRAPH_DB_DEPENDENCY

Please make sure your elasticsearch server is running locally on localhost:9200 and version of the server is 7.10, which is the current supported version.

###### Steps

Navigate to the root of the 'Search API' repository. Build and run the bulk indexer by running the following command
```
  make reindex
```
NB. To run the executable ./reindex is not necessary in the local environment

#### Remote Environment

###### Prerequisites

* Requires dp tool setup locally. For more info on setting up the dp tool: https://github.com/ONSdigital/dp-cli#build-and-run
* Requires the domain endpoint, for the relevant AWS Open Search site cluster, to be put in line 18 (the esURL) of cmd/reindex/aws.go 
* Requires the Search API service auth token value, for the remote environment, to be put in line 21 of cmd/reindex/aws.go

###### Steps

Navigate to the root of the 'Search API' repository and build the bulk indexer by running the following command
```
  make build-reindex
```
Then copy the reindex executable from your local environment build directory, to the remote server, by running the following command
```
dp scp <environment> publishing <box> ./build/reindex <location on publishing box>
```
For example
```
dp scp sandbox publishing 2 ./build/reindex .
```
Then ssh into the box as follows
```
dp ssh <environment> publishing <box>
```
Then create a build directory and move the executable into it
```
  mkdir build
  mv reindex build
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
