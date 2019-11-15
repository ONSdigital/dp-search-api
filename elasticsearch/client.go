package elasticsearch

import (
	"context"

	rchttp "github.com/ONSdigital/dp-rchttp"
)

// Client provides methods to wrap around the elastcsearch package in order to facilitate unit testing
type Client struct {
	url string
	cli rchttp.Clienter
}

// New creates a new elasticsearch client
func New(url string, cli rchttp.Clienter) *Client {

	Setup(url, cli)

	return &Client{
		url: url,
		cli: cli,
	}
}

// Search is a method that wraps the Search function of the elasticsearch package
func (cli *Client) Search(ctx context.Context, index string, docType string, request []byte) ([]byte, error) {
	return Search(ctx, index, docType, request)
}

// MultiSearch is a method that wraps the MultiSearch function of the elasticsearch package
func (cli *Client) MultiSearch(ctx context.Context, index string, docType string, request []byte) ([]byte, error) {
	return MultiSearch(ctx, index, docType, request)
}
