dp-search-api SDK
======================

## Overview

This API contains a client with convenience functions for interacting with the search api from other applications.
It also contains reusable structs representing models used as payloads in API endpoints.

## Example use of the API SDK

Initialise new Search API client

```go
package main

import (
	"context"
	"github.com/ONSdigital/dp-search-api/sdk"
)

func main() {
    ...
	searchAPIClient := sdk.NewClient("http://localhost:23900")
    ...
}
```

### Create Index

Use the CreateIndex method to send a request to generate a new search index. This is a private endpoint and requires authorisation header.

```go
...
    // Set authorisation header
    headers := make(map[header][]string)
	headers[Authorisation] = []string{"Bearer authorised-user"}

    resp, err := searchAPIClient.CreateIndex(ctx, sdk.Options{sdk.Headers: headers})
    if err != nil {
        // handle error
    }

    /* If successdul the resp value will be *CreateIndexResponse struct found in github.com/ONSdigital/dp-search-api/models package

    JSON equivalent:
    {
        "index_name": <value>,
    }
    */
...
```

### Get Search Results

Use the GetSearch method to send a request to find search results based on query parameters. Authorisation header needed if hitting private instance of application.

```go
...
    // Set query parameters - no limit to which keys and values you set - please refer to swagger spec for list of available parameters
    query := url.Values{}
	query.Add("q", "census")

    resp, err := searchAPIClient.GetSearch(ctx, sdk.Options{sdk.Query: query})
    if err != nil {
        // handle error
    }
...
```

### Get Release Calendar Entires

Use the GetReleaseCalendarEntries method to send a request to find release calendar entries based on query parameters. Authorisation header needed if hitting private instance of application.

```go
...
    // Set query parameters - no limit to which keys and values you set - please refer to swagger spec for list of available parameters
    query := url.Values{}
	query.Add("q", "census")

    resp, err := searchAPIClient.GetReleaseCalendarEntries(ctx, sdk.Options{sdk.Query: query})
    if err != nil {
        // handle error
    }
...
```

### Handling errors

The error returned from the method contains status code that can be accessed via `Status()` method and similar to extracting the error message using `Error()` method; see snippet below:

```go
...
    _, err := searchAPIClient.GetSearch(ctx, Options{})
    if err != nil {
        // Retrieve status code from error
        statusCode := err.Status()
        // Retrieve error message from error
        errorMessage := err.Error()

        // log message, below uses "github.com/ONSdigital/log.go/v2/log" package
        log.Error(ctx, "failed to retrieve search results", err, log.Data{"code": statusCode})

        return err
    }
...
```