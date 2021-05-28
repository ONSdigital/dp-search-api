dp-search-api
================
Digital Publishing Search API

A Go application microservice to provide query functionality on to the ONS Website

### Getting started

* Run `make debug`

### Dependencies

* No further dependencies other than those defined in `go.mod`

## /data?

URL Parameters

1. `uris` _[0..1]_
2. `types` _[0..1]_

## /search?

URL Parameters

1. `term` _[0..1]_
2. `size` _[0..1]_ _default_:10
3. `from` _[0..1]_ _default_:0
4. `index` _[0..1]_
5. `type` _[0..*]_
6. `sort` _[0..1]_ _default_: `relevance`, _options_: `relevance`,`release_date`,`release_date_asc`,`first_letter`,`title`
7. `queries` _[0..4]_ _default_: `search` , _options_: `search`,`counts`,`departments`,`featured`
8. `aggField`: _[0..1]_ _default_ : `_type` ; the field that will be used to group on by the `count` query
9. `latest` _[0..1]_, _options_: [true|false], _defaults_: false filters both search and counts to include only the latest version where `description.latestRelease: true`
10. `withFirstLetter` _[0..1]_, _options_:  a..Z and is restricted to `search` queries will only return results where the title starts with the letter `?`
11. `uriPrefix` _[0..1]_, prefix filter on the URI.
12. `topic` _[0..1]_ topic filter , matches exact Topic
13. `topicWildcard` _[0..1]_ topic filter , matches Topics using Wildcard pattern matching
14. `highlight`  _[0..1]_, _options_: [true|false], _defaults_: true, determines whether to include highlighting on the `SEARCH` results
15. `upcoming`  _[0..1]_, _options_: [true|false], _defaults_: true, determines whether items in the `SEARCH` that are not published and not cancelled __OR__ are published and are not due
16. `published`  _[0..1]_, _options_: [true|false], _defaults_: true, determines whether items in the `SEARCH` that are published and not cancelled __OR__ are cancelled and are due

## timeseries/{cdid}


Remember to update the [README](README.md) and [CHANGELOG](CHANGELOG.md) files.

### Configuration

An overview of the configuration options available, either as a table of
environment variables, or with a link to a configuration guide.

| Environment variable      | Default                 | Description
| ------------------------- | ----------------------- | ------------------
| BIND_ADDR                 | :23900                  | The host and port to bind to
| ELASTIC_URL	            | http://localhost:9200 | Http url of the ElasticSearch server
| GRACEFUL_SHUTDOWN_TIMEOUT | 5s                      | The graceful shutdown timeout in seconds (`time.Duration` format)

## Releasing
To package up the API uses `make package`

## Deploying
Export the following variables;
* export `DATA_CENTER` to the nomad datacenter to use.
* export `S3_TAR_FILE` to a S3 location on where a release file can be found.
* export `ELASTIC_SEARCH_URL` to elastic search url.

Then run `make nomad` this shall create a nomad plan within the root directory
called `dp-search-api.nomad`

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2021, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
