# Search Service Architecture

Source of truth for the application architecture of the ons search service including backend processing of data. C4 diagrams available via [ONS google drive](https://drive.google.com/drive/folders/15Qq3wgaULer96kXGuMDUo6D0-vAz7um1).

## Contents

- [Search Service Architecture](#search-service-architecture)
  - [Contents](#contents)
  - [Sitesearch and other aggregated pages](#sitesearch-and-other-aggregated-pages)
    - [Sitesearch UI \& API](#sitesearch-ui--api)
    - [Release Calendar UI \& API](#release-calendar-ui--api)
  - [Search Data Pipeline](#search-data-pipeline)
    - [Add/update search documents when a collection is published](#addupdate-search-documents-when-a-collection-is-published)
      - [Steps](#steps)
  - [Sequence diagrams](#sequence-diagrams)

## Sitesearch and other aggregated pages

### Sitesearch UI & API

![Sitesearch](./sequence-diagrams/search-ui/sitesearch-ui.png)

Sitesearch dataflow from user on the website making a search with query term and/or selecting filters to retrieving data from backing services and rendering the results.

### Release Calendar UI & API

![Release Calendar](./sequence-diagrams/search-ui/release-calendar-ui.png)

Dataflow from user on the website hitting the release calendar page on the ONS website, or selecting filters on this page to retrieving data from backing services and rendering the results.

## Search Data Pipeline

### Add/update search documents when a collection is published

![Publish Search Data](./sequence-diagrams/publish-search-pipeline/publish-search-pipeline.png)

The data pipleine from Florence user publishing some new or updated content or data to being stored in Elasticsearch, ready for web or API users to query (via API, e.g. [sitesearch sequence diagram](#sitesearch-ui--api))

#### Steps

Pre-requisite steps to publishing data not in this workflow as we will not be changing that part of the process. The following steps follow
on from a florence (DP internal user) publishes a collection that can contain 1 to many ONS webpages or datasets by making a request to zebedee
which does the publishing of new webpages or updates to existing pages and lastly triggering updates to sitewide search; which is where the flow begins.

**1: Consume Kafka messages**

When a user publishes a collection via Florence, the backend collection process in Zebedee will
trigger an event to the content-updated topic for consumption by the search data extractor.

```
Datastore: kafka

Topic: content-updated
Record: {
    "uri": string, // mandatory
    "search_index": string, // mandatory
    "data_type": string, // mandatory
    "collection_id": string, // optional
}
```

See [full schema here](https://github.com/onsdigital/dp-search-data-extractor/blob/develop/schema/schema.go#L7)

The `search_index` field should be set to the search index alias, `ons` for publishing new content. Only search reindexes will make use of the actual index name, e.g. `ons_<timestamp>`.

**2: Retrieve Zebedee resource**

If message is of type `legacy` then retrieve resource from Zebedee API endpoint based on the uri in kafka message.

**3: Read JSON file from disc**

Follows from step 3 only - retrieve single document from zebedee content (files on disc).

**4: Retrieve dataset resource**

If message is of type `datasest` retrieve resource from Dataset API endpoint based on the uri in kafka message.

*Note: No parellisation with 2 as each of these are triggered by separate kafka messages*

**5: Find latest version of dataset**

Follows from step 4 only - find one document from Mongo db representing the resource on given uri.

**6: Produce event to search-data-import**

```
Datastore: kafka

Topic: search-data-import
Record: {
    "data_type": string,
    "job_id": string, // empty
    "search_index": string, // Should use search alias ons
    "cdid": string,
    "dataset_id": string,
    "description": string,
    "edition:: string,
    "keywords": string,
    "meta_description": string,
    "release_date": string // date format: ISO8601 or strict_date_optional_time||epoch_millis to match existing docs in search?
    "summary": string,
    "title": string,
    ... other fields we decide need to be in search
}
```

See [full schema here](https://github.com/onsdigital/dp-search-data-extractor/blob/develop/schema/schema.go#L25)

**7: Consume event from search-data-import**

Kafka message consumed by Search Data Importer. Documents received via consumption of kafka topic `search-data-import`, store documents in memory until 500 messages consumed or time from first message exceeded 5 seconds before making bulk request. 5 second limit will allow for the last set of messages to still be reindexed.

**Note: For search reindex process, we create separate batches based on the search index name; this way we can insert documents to the correct index for both reindex jobs and publishing new content and data**

**8: Multi-operational request to update/add docs to an index**

```
Datastore: Elasticsearch
Index: ons *or* ons_<timestamp> // should be using the value in consumed kafka event

Method: POST
Header: 'Content-Type: application/json'
Path: /_bulk
Body: { "update": { <search doc - Depends on data type>, "_index": "ons" } }
```

The search document should set an `_id` field that matches the unique identifier for that document, either uuid or compound identifier (multiple fields to represent a documents uniqueness). Either way this should be determined by the data stored in Zebedee content and the dataset API.


See [Bulk API](https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html)

## Sequence diagrams

To update the diagrams, see [sequence diagrams documentation](sequence-diagrams/README.md)
