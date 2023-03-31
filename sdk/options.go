package sdk

import (
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-search-api/api"
)

const (
	// List of available headers
	Authorization string = "Authorization"
	CollectionID  string = "Collection-Id"
)

// Options is a struct containing for customised options for the API client
type Options struct {
	Headers http.Header
	Query   url.Values
}

// Q sets the 'q' Query parameter to the request
func (o *Options) Q(val string) *Options {
	o.Query.Set(api.ParamQ, val)
	return o
}

// Sort sets the 'sort' Query parameter to the request
func (o *Options) Sort(val string) *Options {
	o.Query.Set(api.ParamSort, val)
	return o
}

// Highlight sets the 'highlight' Query parameter to the request
func (o *Options) Highlight(val string) *Options {
	o.Query.Set(api.ParamHighlight, val)
	return o
}

// Topics sets the 'topics' Query parameter to the request
func (o *Options) Topics(val string) *Options {
	o.Query.Set(api.ParamTopics, val)
	return o
}

// Limit sets the 'limit' Query parameter to the request
func (o *Options) Limit(val string) *Options {
	o.Query.Add(api.ParamLimit, val)
	return o
}

// Offset sets the 'offset' Query parameter to the request
func (o *Options) Offset(val string) *Options {
	o.Query.Add(api.ParamOffset, val)
	return o
}

// ContentType sets the 'content_type' Query parameter to the request
func (o *Options) ContentType(val string) *Options {
	o.Query.Set(api.ParamContentType, val)
	return o
}

// PopulationTypes sets the 'population_types' Query parameter to the request
func (o *Options) PopulationTypes(val string) *Options {
	o.Query.Add(api.ParamPopulationTypes, val)
	return o
}

// Dimensions sets the 'dimensions' Query parameter to the request
func (o *Options) Dimensions(val string) *Options {
	o.Query.Set(api.ParamDimensions, val)
	return o
}

// SubtypeProvisional sets the 'subtype-provisional' Query parameter to the request
func (o *Options) SubtypeProvisional(val string) *Options {
	o.Query.Set(api.ParamSubtypeProvisional, val)
	return o
}

// SubtypeConfirmed sets the 'subtype-confirmed' Query parameter to the request
func (o *Options) SubtypeConfirmed(val string) *Options {
	o.Query.Set(api.ParamSubtypeConfirmed, val)
	return o
}

// SubtypePostponed sets the 'subtype-postponed' Query parameter to the request
func (o *Options) SubtypePostponed(val string) *Options {
	o.Query.Set(api.ParamSubtypePostponed, val)
	return o
}

// Census sets the 'census' Query parameter to the request
func (o *Options) Census(val string) *Options {
	o.Query.Set(api.ParamCensus, val)
	return o
}

func setHeaders(req *http.Request, headers http.Header) {
	for name, values := range headers {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}
}
