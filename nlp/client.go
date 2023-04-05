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
type Client struct {
	berlinBaseURL   string
	scrubberBaseURL string
	categoryBaseURL string
	client          dphttp.Client
}

// New creates a new elasticsearch client. Any trailing slashes from the URL are removed.
func New(nlp config.NLP) *Client {
	client := dphttp.Client{}
	return &Client{
		berlinBaseURL:   nlp.BerlinAPIURL,
		scrubberBaseURL: nlp.ScrubberAPIURL,
		categoryBaseURL: nlp.CategoryAPIURL,
		client:          client,
	}
}

func (cli *Client) GetBerlin(ctx context.Context, params url.Values) (models.Berlin, error) {
	var berlin models.Berlin

	url, err := buildURL(cli.berlinBaseURL, params)
	if err != nil {
		// TODO: error handling
	}

	resp, err := cli.client.Get(ctx, url.String())
	defer resp.Body.Close()
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

func (cli *Client) GetCategory(ctx context.Context, params url.Values) (models.Category, error) {
	var category models.Category

	url, err := buildURL(cli.categoryBaseURL, params)
	if err != nil {
		// TODO: error handling
	}

	resp, err := cli.client.Get(ctx, url.String())
	defer resp.Body.Close()
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

func (cli *Client) GetScrubber(ctx context.Context, params url.Values) (models.Scrubber, error) {
	var scrubber models.Scrubber

	url, err := buildURL(cli.scrubberBaseURL, params)
	if err != nil {
		// TODO: error handling
	}

	resp, err := cli.client.Get(ctx, url.String())
	defer resp.Body.Close()
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
