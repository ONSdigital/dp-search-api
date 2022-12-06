package sdk

import (
	"net/http"
	"net/url"
)

type header string

const (
	// List of available headers
	Authorisation header = "Authorisation"
)

// Options is a struct containing for customised options for the API client
type Options struct {
	Headers map[header][]string
	Query   url.Values
}

func setHeaders(req *http.Request, headers map[header][]string) {
	for h := range headers {
		for i := range headers[h] {
			req.Header.Add(string(h), headers[h][i])
		}
	}
}
