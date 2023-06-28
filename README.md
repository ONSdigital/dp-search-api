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

## Configuration

An overview of the configuration options available, either as a table of
environment variables, or with a link to a configuration guide.

| Environment variable | Default | Description
| -------------------- | ------- | -----------
| AWS_FILENAME                 | ""                        | The AWS file location for finding credentials to sign AWS http requests
| AWS_PROFILE                  | ""                        | The AWS profile to use from credentials file to sign AWS http requests
| AWS_REGION                   | eu-west-2                 | The AWS region to use when signing requests with AWS SDK
| AWS_SERVICE                  | "es"                      | The AWS service that the AWS SDK signing mechanism needs to sign a request
| AWS_SIGNER                   | false                     | The AWS signer flag will determine if requests to Elasticsearch contain round tripper for signing requests
| AWS_TLS_INSECURE_SKIP_VERIFY | false                     | This should never be set to true, as it disables SSL certificate verification. Used only for development
| BIND_ADDR                    | :23900                    | The host and port to bind to
| ELASTIC_SEARCH_URL           | "<http://localhost:11200>"| Http url of the ElasticSearch server
| NLP_BERLIN_API_URL	         | "http://localhost:28900"  | Http url of the Berlin server
| NLP_CATEGORY_API_URL	       | "http://localhost:28800"  | Http url of the Category server
| NLP_SCRUBBER_API_URL	       | "http://localhost:28700"  | Http url of the Scrubber server
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                        | The graceful shutdown timeout in seconds (`time.Duration` format)
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                       | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)
| HEALTHCHECK_INTERVAL         | 30s                       | Time between self-healthchecks (`time.Duration` format)
| SERVICE_AUTH_TOKEN           | ""                        | The service auth token only gets used by the bulk indexer [Running Bulk Indexer](#running-bulk-indexer)
| ZEBEDEE_URL                  | <http://localhost:8082>   | The URL to Zebedee (for authorisation)

## API documentation

Documentation of the API interface is described using swagger 2.0. This specification can be found [here](./swagger.yaml).

## Go SDK - Client Package

Applications trying to interact with the API can use the Go SDK package which contains a list of client methods that are maintained to align with
the API. For futher reading on how to use the client follow this [link](./sdk/README.md).

## Connecting to AWS Elasticsearch cluster (dev only)

### Prerequisites

* You will need an user account for the aws account you are trying to connect to
* You will need to be given a policy to allow read and write access to the AWS Elasticsearch cluster

### Connecting

To connect to managed Elasticsearch cluster in AWS, you will want to port forward 11200 to the domain endpoint. Using the dp tool, one can do this like so:

```shell
  dp ssh develop <ip of aws box> -p 9200:<elasticsearch cluster domain endpoint e.g. "<unique identifier>" + "eu-west-1.es.amazonaws.com:443"
```

Once connected, run the following make target:

```shell
  make local
```

## Running Bulk Indexer

The Bulk Indexer creates an Elastic Search 7.10 index with the alias "ons". It loads data from zebedee, and the dataset API, into the ES 7.10 index.

### Locally

#### Local Prerequisites

* Requires ElasticSearch 7.10 running on port 11200
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

Please make sure your elasticsearch server is running locally on localhost:11200 and version of the server is 7.10, which is the current supported version. You may use `dp-compose/v2/stacks/search` stack for this.

If you want to run the reindex script locally but loading data from an environment (e.g. `sandbox`), you may run `dp ssh` with port forwarding for dataset-api and zebedee (please check the services IPs and ports in `https://consul.dp.aws.onsdigital.uk/ui/eu/services`) For example:

```shell
dp ssh sandbox publishing 2 -p 22000:10.30.138.234:26020
dp ssh sandbox publishing 1 -p 8082:10.30.138.93:25108
```

If you do this the auth token will need to be a valid token accepted in the environment you are using.

#### Steps

Navigate to the root of the 'Search API' repository. Build and run the bulk indexer by running the following command

```shell
  make reindex
```

NB. To run the executable ./reindex is not necessary in the local environment

### Remote Environment

#### Remote Prerequisites

* Requires dp tool setup locally. For more info on setting up the dp tool: <https://github.com/ONSdigital/dp-cli#build-and-run>
* Requires the domain endpoint, for the relevant AWS Open Search site cluster, to be put in line 18 (the esURL) of cmd/reindex/aws.go  (see below)
* Requires the Search API service auth token value, for the remote environment, to be put in line 21 of cmd/reindex/aws.go  (see below)

##### Getting info for above two prerequisites for Prod

1. Getting the `Domain endpoint`

    1. Log into the AWS console and select the ons-dp-prod Account
    1. Select the "Amazon OpenSearch Service"
    1. On the left side bar, select "Domains"
    1. Then select the Domain: "prod-site"
    1. Then under the title `Domain endpoint (VPC)` select the url for use in line 18 of cmd/reindex/aws.go

1. Getting the `Search API service auth token`

    1. In a terminal, select the AWS Prod environment
    2. Log into `Publishing Mount 1` with:

    ```shell
    dp ssh prod publishing_mount 1
    ```

    3. Then when logged in, do:

    ```shell
    cd /var/florence/zebedee/services/

    find . -type f -exec grep -H 'dp-search-api' {} \;
    ```

    4. The last command should return something like:

    ```text
    ./3YjV9H9R< text redacted               >1hotSHMvZ1EsronRHNbTReayA.json:{"id":"dp-search-api"}
    ```

    From the above, use the string starting with `3YjV9H` and ending with `TReayA` (that is a string that is 64 characters long) for the Search API service auth token in line 21 of cmd/reindex/aws.go

#### Steps 2

Navigate to the root of the 'Search API' repository and build the bulk indexer by running the following command

```shell
  make build-reindex
```

Then create a build directory and move the executable into it

```shell
  mkdir build
  mv reindex build
```

Then run the executable

```shell
  ./build/reindex
```

### Architecture

See [ARCHITECTURE](architecture/README.md) for details.

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2023, Office for National Statistics (<https://www.ons.gov.uk>)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
