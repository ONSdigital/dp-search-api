dp-search-api
================
Digital Publishing Search API

A Go application microservice to provide query functionality on the ONS Website

### Getting started

* Run `make debug`

### Dependencies

Clone and set up the following project following the README instructions:
- [dp-compose](https://github.com/ONSdigital/dp-compose)

No further dependencies other than those defined in `go.mod`

### Configuration

An overview of the configuration options available, either as a table of
environment variables, or with a link to a configuration guide.

| Environment variable | Default | Description
| -------------------------   | ----------------------- | -----------
| BIND_ADDR                   | :23900                  | The host and port to bind to
| ELASTIC_URL	              | "http://localhost:9200" | Http url of the ElasticSearch server
| GRACEFUL_SHUTDOWN_TIMEOUT   | 5s                      | The graceful shutdown timeout in seconds (`time.Duration` format)
| SIGN_ELASTICSEARCH_REQUESTS | false                   | Boolean flag to identify whether elasticsearch requests via elastic API need to be signed if elasticsearch cluster is running in aws

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2021, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
