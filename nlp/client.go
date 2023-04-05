package nlp

import (
	"context"
	"encoding/json"
	"io"
	"net/url"

	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/models"
)

// Client represents an instance of the elasticsearch client - now deprecated
type NLPClient struct {
	BerlinBaseURL   string
	ScrubberBaseURL string
	CategoryBaseURL string
	client          dphttp.Client
}

// New creates a new elasticsearch client. Any trailing slashes from the URL are removed.
func New(nlp config.NLP, client dphttp.Client) *NLPClient {
	return &NLPClient{
		BerlinBaseURL:   nlp.BerlinAPIURL,
		ScrubberBaseURL: nlp.ScrubberAPIURL,
		CategoryBaseURL: nlp.CategoryAPIURL,
		client:          client,
	}
}

func (cli *NLPClient) GetBerlin(ctx context.Context, params url.Values) (models.Berlin, error) {
	var berlin models.Berlin

	url, err := buildURL(cli.BerlinBaseURL, params)
	if err != nil {
		// TODO: error handling
	}

	resp, err := cli.client.Get(ctx, url.String())
	if err != nil {
		// TODO: error handling
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		// TODO: error handling
	}

	if err := json.Unmarshal(b, &berlin); err != nil {
		// TODO: error handling
	}

	return berlin, nil
}

func (cli *NLPClient) GetCategory(ctx context.Context, params url.Values) (models.Category, error) {
	var category models.Category

	url, err := buildURL(cli.CategoryBaseURL, params)
	if err != nil {
		// TODO: error handling
	}

	resp, err := cli.client.Get(ctx, url.String())
	if err != nil {
		// TODO: error handling
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		// TODO: error handling
	}

	if err := json.Unmarshal(b, &category); err != nil {
		// TODO: error handling
	}

	return category, nil
}

func (cli *NLPClient) GetScrubber(ctx context.Context, params url.Values) (models.Scrubber, error) {
	var scrubber models.Scrubber

	url, err := buildURL(cli.ScrubberBaseURL, params)
	if err != nil {
		// TODO: error handling
	}

	resp, err := cli.client.Get(ctx, url.String())
	if err != nil {
		// TODO: error handling
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		// TODO: error handling
	}

	if err := json.Unmarshal(b, &scrubber); err != nil {
		// TODO: error handling
	}

	return scrubber, nil
}

func buildURL(baseURL string, params url.Values) (*url.URL, error) {
	var catQuery url.Values
	catQuery.Add("query", params.Get("q"))

	requestURL, err := url.Parse(baseURL)
	if err != nil {
		// TODO: error handling
	}

	requestURL.RawQuery = catQuery.Encode()

	return requestURL, nil
}
