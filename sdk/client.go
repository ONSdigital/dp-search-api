package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-search-api/models"
	apiError "github.com/ONSdigital/dp-search-api/sdk/errors"
	"github.com/ONSdigital/dp-search-api/transformer"
)

const (
	service = "dp-search-api"
)

type Client struct {
	hcCli *healthcheck.Client
}

// New creates a new instance of Client with a given search api url
func New(searchAPIURL string) *Client {
	return &Client{
		hcCli: healthcheck.NewClient(service, searchAPIURL),
	}
}

// NewWithHealthClient creates a new instance of search API Client,
// reusing the URL and Clienter from the provided healthcheck client
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		hcCli: healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// URL returns the URL used by this client
func (cli *Client) URL() string {
	return cli.hcCli.URL
}

// Health returns the underlying Healthcheck Client for this search API client
func (cli *Client) Health() *healthcheck.Client {
	return cli.hcCli
}

// Checker calls search api health endpoint and returns a check object to the caller
func (cli *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return cli.hcCli.Checker(ctx, check)
}

// GetReleaseCalendarEntries gets a list of release calendar entries based on the search request
func (cli *Client) GetReleaseCalendarEntries(ctx context.Context, options Options) (*transformer.SearchReleaseResponse, apiError.Error) {
	path := fmt.Sprintf("%s/search/releases", cli.hcCli.URL)
	if options.Query != nil {
		path = path + "?" + options.Query.Encode()
	}

	respInfo, apiErr := cli.callSearchAPI(ctx, path, http.MethodGet, options.Headers, nil)
	if apiErr != nil {
		return nil, apiErr
	}

	var searchResponse transformer.SearchReleaseResponse

	if err := json.Unmarshal(respInfo.Body, &searchResponse); err != nil {
		return nil, apiError.StatusError{
			Err: fmt.Errorf("failed to unmarshal release calendar search response - error is: %v", err),
		}
	}

	return &searchResponse, nil
}

// GetSearch gets a list of search results based on the search request
func (cli *Client) GetSearch(ctx context.Context, options Options) (*models.SearchResponse, apiError.Error) {
	path := fmt.Sprintf("%s/search", cli.hcCli.URL)
	if options.Query != nil {
		path = path + "?" + options.Query.Encode()
	}

	respInfo, apiErr := cli.callSearchAPI(ctx, path, http.MethodGet, options.Headers, nil)
	if apiErr != nil {
		return nil, apiErr
	}

	var searchResponse models.SearchResponse

	if err := json.Unmarshal(respInfo.Body, &searchResponse); err != nil {
		return nil, apiError.StatusError{
			Err: fmt.Errorf("failed to unmarshal search response - error is: %v", err),
		}
	}

	return &searchResponse, nil
}

// PostSearch creates a new search index
func (cli *Client) CreateIndex(ctx context.Context, options Options) (*models.CreateIndexResponse, apiError.Error) {
	path := fmt.Sprintf("%s/search", cli.hcCli.URL)

	respInfo, apiErr := cli.callSearchAPI(ctx, path, http.MethodPost, options.Headers, nil)
	if apiErr != nil {
		return nil, apiErr
	}

	var searchResponse models.CreateIndexResponse

	if err := json.Unmarshal(respInfo.Body, &searchResponse); err != nil {
		return nil, apiError.StatusError{
			Err: fmt.Errorf("failed to unmarshal creating index search response - error is: %v", err),
		}
	}

	return &searchResponse, nil
}

type ResponseInfo struct {
	Body    []byte
	Headers http.Header
	Status  int
}

// callSearchAPI calls the Search API endpoint given by path for the provided REST method, request headers, and body payload.
// It returns the response body and any error that occurred.
func (cli *Client) callSearchAPI(ctx context.Context, path, method string, headers map[header][]string, payload []byte) (*ResponseInfo, apiError.Error) {
	URL, err := url.Parse(path)
	if err != nil {
		return nil, apiError.StatusError{
			Err: fmt.Errorf("failed to parse path: \"%v\" error is: %v", path, err),
		}
	}

	path = URL.String()

	var req *http.Request

	if payload != nil {
		req, err = http.NewRequest(method, path, bytes.NewReader(payload))
	} else {
		req, err = http.NewRequest(method, path, http.NoBody)
	}

	// check req, above, didn't error
	if err != nil {
		return nil, apiError.StatusError{
			Err: fmt.Errorf("failed to create request for call to search api, error is: %v", err),
		}
	}

	// set any headers against request
	setHeaders(req, headers)

	if payload != nil {
		req.Header.Add("Content-type", "application/json")
	}

	resp, err := cli.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, apiError.StatusError{
			Err:  fmt.Errorf("failed to call search api, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}
	defer func() {
		err = closeResponseBody(resp)
	}()

	respInfo := &ResponseInfo{
		Headers: resp.Header.Clone(),
		Status:  resp.StatusCode,
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 400 {
		return respInfo, apiError.StatusError{
			Err:  fmt.Errorf("failed as unexpected code from search api: %v", resp.StatusCode),
			Code: resp.StatusCode,
		}
	}

	if resp.Body == nil {
		return respInfo, nil
	}

	respInfo.Body, err = io.ReadAll(resp.Body)
	if err != nil {
		return respInfo, apiError.StatusError{
			Err:  fmt.Errorf("failed to read response body from call to search api, error is: %v", err),
			Code: resp.StatusCode,
		}
	}

	return respInfo, nil
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(resp *http.Response) apiError.Error {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			return apiError.StatusError{
				Err:  fmt.Errorf("error closing http response body from call to search api, error is: %v", err),
				Code: http.StatusInternalServerError,
			}
		}
	}

	return nil
}
