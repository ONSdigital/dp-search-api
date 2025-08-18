# dp-search-api

Dissemination Service Search API

A Go application microservice to provide query functionality on the ONS Website

## Getting started

Set up dependencies locally as follows:

* See our [Docker Compose Search Stack](https://github.com/ONSdigital/dp-compose/tree/main/v2/stacks/search) to run all services end to end.

Alternatively, you can just run the `dp-search-api` on it's own with `elasticsearch` v7.10 running on port 11700.

In order to surface results you will also need to populate the index in some way. Please follow the instructions in [dp-search-reindex-batch](https://github.com/ONSdigital/dp-search-reindex-batch?tab=readme-ov-file#getting-started).

To run the NLP enhancement services for search you will also need to run:

* [dp-search-scrubber-api](https://github.com/ONSdigital/dp-search-scrubber-api)
* [dp-nlp-category-api](https://github.com/ONSdigital/dp-nlp-category-api)
* [dp-nlp-berlin-api](https://github.com/ONSdigital/dp-nlp-berlin-api)

For the authorisation of the POST /index endpoint you will also need to run [zebedee](https://github.com/ONSdigital/zebedee)

Then run `make debug`.

## Dependencies

* Requires ElasticSearch running on port 11200
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

[Documentation of the API interface](./swagger.yaml) is described using swagger 2.0.

## Go SDK - Client Package

Applications trying to interact with the API can use [the Go SDK package](./sdk/README.md) which contains a list of client methods that are
maintained to align with
the API.

## Reindexing locally

Use the [dp-search-reindex-batch](https://github.com/ONSdigital/dp-search-reindex-batch) script to build a local search
index from data in your local zebedee and dataset api data.

### Architecture

See [ARCHITECTURE](architecture/README.md) for details.

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2023, Office for National Statistics (<https://www.ons.gov.uk>)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
