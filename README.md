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
| AWS_REGION                  | eu-west-1               | The AWS region to use when signing requests with AWS SDK
| AWS_SERVICE                 | "es"                    | The aws service that the AWS SDK signing mechanism needs to sign a request
| BIND_ADDR                   | :23900                  | The host and port to bind to
| ELASTIC_SEARCH_URL	      | "http://localhost:9200" | Http url of the ElasticSearch server. For Site Wide ElasticSearch this needs to be set to "http://localhost:11200".
| GRACEFUL_SHUTDOWN_TIMEOUT   | 5s                      | The graceful shutdown timeout in seconds (`time.Duration` format)
| SIGN_ELASTICSEARCH_REQUESTS | false                   | Boolean flag to identify whether elasticsearch requests via elastic API need to be signed if elasticsearch cluster is running in aws
| HEALTHCHECK_CRITICAL_TIMEOUT| 90s                     | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)
| HEALTHCHECK_INTERVAL        | 30s                     | Time between self-healthchecks (`time.Duration` format)
| ZEBEDEE_URL                 | http://localhost:8082   | The URL to Zebedee (for authorisation)

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2021, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
