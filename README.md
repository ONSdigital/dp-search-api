dp-search-query
================

A Go application microservice to provide query functionality on to the ONSWebsite ElasticSearch :

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

## timeseries/{cdid}


Remember to update the [README](README.md) and [CHANGELOG](CHANGELOG.md) files.

### Configuration

An overview of the configuration options available, either as a table of
environment variables, or with a link to a configuration guide.

| Environment variable | Default | Description
| -------------------- | ------- | -----------
| BIND_ADDR            | :10001  | The host and port to bind to
| ELASTIC_URL	       | "http://localhost:9200/" | Http url of the ElasticSearch server 

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
