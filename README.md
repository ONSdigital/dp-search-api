# dp-search-api

Digital Publishing Search API

A Go application microservice to provide query functionality on the ONS Website

## Getting started

Set up dependencies locally as follows:

* In [dp-compose](https://github.com/ONSdigital/dp-compose) run `docker-compose up -d` to run ElasticSearch 7.10
    * dp-compose will run Elasticsearch 7.10 on port 11200 to not conflict with ES 2.2 running on port 9200
* If using the POST /search endpoint then authorisation for this requires running Vault and Zebedee as follows:
* In any directory run `vault server -dev` as Zebedee has a dependency on Vault
* In the zebedee directory run `./run.sh` to run Zebedee

Then run `make debug`

## Dependencies

* Requires ElasticSearch running on port 11200
* Requires Zebedee running on port 8082
* No further dependencies other than those defined in `go.mod`

### Validating Specification

To validate the swagger specification you can do this via:

```sh
make validate-specification
```

To run this, you will need to run Node > v20 and have [redocly CLI](https://github.com/Redocly/redocly-cli) installed:

```sh
npm install -g redocly-cli
```

## Configuration

An overview of the configuration options available, either as a table of
environment variables, or with a link to a configuration guide.

| Environment variable         | Default                  | Description                                                                                                        |
|------------------------------|--------------------------|--------------------------------------------------------------------------------------------------------------------|
| AWS_FILENAME                 | ""                       | The AWS file location for finding credentials to sign AWS http requests                                            |
| AWS_PROFILE                  | ""                       | The AWS profile to use from credentials file to sign AWS http requests                                             |
| AWS_REGION                   | eu-west-2                | The AWS region to use when signing requests with AWS SDK                                                           |
| AWS_SERVICE                  | "es"                     | The AWS service that the AWS SDK signing mechanism needs to sign a request                                         |
| AWS_SIGNER                   | false                    | The AWS signer flag will determine if requests to Elasticsearch contain round tripper for signing requests         |
| AWS_TLS_INSECURE_SKIP_VERIFY | false                    | This should never be set to true, as it disables SSL certificate verification. Used only for development           |
| BIND_ADDR                    | :23900                   | The host and port to bind to                                                                                       |
| BERLIN_URL                   | "http://localhost:28900" | HTTP URL of the NLP Berlin API                                                                                     |
| CATEGORY_URL                 | "http://localhost:28800" | HTTP URL of the NLP Category API                                                                                   |
| DEFAULT_LIMIT                | 10                       | The default limit of search results in a page                                                                      |
| DEFAULT_MAXIMUM_LIMIT        | 100                      | The default maximum limit of search results in a page                                                              |
| DEFAULT_OFFSET               | 0                        | The default offset of search results                                                                               |
| DEFAULT_SORT                 | "relevance"              | The default sort for search results                                                                                |
| ELASTIC_SEARCH_URL           | "http://localhost:11200" | Http url of the ElasticSearch server                                                                               |
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                       | The graceful shutdown timeout in seconds (`time.Duration` format)                                                  |
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                      | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format) |
| HEALTHCHECK_INTERVAL         | 30s                      | Time between self-healthchecks (`time.Duration` format)                                                            |
| NLP_SETTINGS                 | ***See below***          | [NLP Settings](#nlp-settings)                                                                                      |
| ENABLE_NLP_WEIGHTING         | false                    | Feature flag for enabling NLP Weighting functionality via Scrubber, Category and Berlin                            |
| OTEL_BATCH_TIMEOUT           | 5s                       | Interval between pushes to OT Collector                                                                            |
| OTEL_EXPORTER_OTLP_ENDPOINT  | "http://localhost:4317"  | URL for OpenTelemetry endpoint                                                                                     |
| OTEL_SERVICE_NAME            | "dp-search-api"          | Service name to report to telemetry tools                                                                          |
| OTEL_ENABLED                 | false                    | Feature flag to enable OpenTelemetry                                                                               |
| SCRUBBER_URL                 | "http://localhost:28700" |                                                                                                                    |
| ZEBEDEE_URL                  | "http://localhost:8082"  | The URL to Zebedee (for authorisation)                                                                             |

### NLP Settings

NLP Hub Settings are set as JSON, of which the default is:

```txt
{\"category_weighting\": 100000000.0, \"category_limit\": 100, \"default_state\": \"gb\"}
```

| Key                | Type   | Description                                                              |
|--------------------|--------|--------------------------------------------------------------------------|
| category_weighting | float  | How important is the category weighting when using them in ElasticSearch |
| category_limit     | int    | Limits how many categories are returned                                  |
| default_state      | string |                                                                          |

## API documentation

Documentation of the API interface is described using swagger 2.0. This specification can be
found [here](./swagger.yaml).

## Go SDK - Client Package

Applications trying to interact with the API can use the Go SDK package which contains a list of client methods that are
maintained to align with
the API. For futher reading on how to use the client follow this [link](./sdk/README.md).

## Connecting to AWS Elasticsearch cluster (dev only)

### Prerequisites

* You will need an user account for the aws account you are trying to connect to
* You will need to be given a policy to allow read and write access to the AWS Elasticsearch cluster

### Connecting

To connect to managed Elasticsearch cluster in AWS, you will want to port forward 11200 to the domain endpoint. Using
the dp tool, one can do this like so:

```shell
  dp ssh develop <ip of aws box> -p 9200:<elasticsearch cluster domain endpoint e.g. "<unique identifier>" + "eu-west-1.es.amazonaws.com:443"
```

Once connected, run the following make target:

```shell
  make local
```

## Reindexing locally

Use the [dp-search-reindex-batch](https://github.com/ONSdigital/dp-search-reindex-batch) script to build a local search
index from data in your local zebedee and dataset api data.

(Note, [an earlier version of this service](https://github.com/ONSdigital/dp-search-api/tree/v1.41.0/cmd/reindex)
included a reindexing script that has now been replaced by the above tool)

### Architecture

See [ARCHITECTURE](architecture/README.md) for details.

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2023, Office for National Statistics (<https://www.ons.gov.uk>)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
