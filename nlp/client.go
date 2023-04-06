package nlp

import (
	"context"
	"encoding/json"
	"fmt"
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

	url, err := buildURL(cli.berlinBaseURL, params, "q")
	if err != nil {
		return berlin, err
	}

	resp, err := cli.client.Get(ctx, url.String())
	defer resp.Body.Close()
	if err != nil {
		return berlin, fmt.Errorf("error making a get request to: %s err: %w", url.String(), err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return berlin, fmt.Errorf("error reading response body: %w", err)
	}

	if err := json.Unmarshal(b, &berlin); err != nil {
		return berlin, fmt.Errorf("error unmarshaling resp body to scrubber model: %w", err)
	}

	return berlin, nil
}

func (cli *Client) GetCategory(ctx context.Context, params url.Values) (models.Category, error) {
	var category models.Category

	url, err := buildURL(cli.categoryBaseURL+"/categories", params, "query")
	if err != nil {
		return category, err
	}

	resp, err := cli.client.Get(ctx, url.String())
	defer resp.Body.Close()
	if err != nil {
		return category, fmt.Errorf("error making a get request to: %s err: %w", url.String(), err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return category, fmt.Errorf("error reading response body: %w", err)
	}

	if err := json.Unmarshal(b, &category); err != nil {
		return category, fmt.Errorf("error unmarshaling resp body to scrubber model: %w", err)
	}

	return category, nil
}

func (cli *Client) GetScrubber(ctx context.Context, params url.Values) (models.Scrubber, error) {
	var scrubber models.Scrubber

	url, err := buildURL(cli.scrubberBaseURL+"/scrubber/search", params, "q")
	if err != nil {
		return scrubber, err
	}

	resp, err := cli.client.Get(ctx, url.String())
	defer resp.Body.Close()
	if err != nil {
		return scrubber, fmt.Errorf("error making a get request to: %s err: %w", url.String(), err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return scrubber, fmt.Errorf("error reading response body: %w", err)
	}

	if err := json.Unmarshal(b, &scrubber); err != nil {
		return scrubber, fmt.Errorf("error unmarshaling resp body to scrubber model: %w", err)
	}

	return scrubber, nil
}

func buildURL(baseURL string, params url.Values, queryKey string) (*url.URL, error) {
	query := url.Values{}

	query.Set(queryKey, params.Get("q"))

	requestURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing baseURL: %w", err)
	}

	requestURL.RawQuery = query.Encode()

	return requestURL, err
}
