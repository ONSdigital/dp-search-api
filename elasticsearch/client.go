package elasticsearch

import (
	"bytes"
	"context"
	"io"
	"net/http"

	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/pkg/errors"
)

// Client represents an instance of the elasticsearch client - now deprecated
type Client struct {
	awsRegion  string
	awsService string
	url        string
	client     dphttp.Clienter
}

// New creates a new elasticsearch client. Any trailing slashes from the URL are removed.
func New(url string, client dphttp.Clienter, awsService, awsRegion string) *Client {
	return &Client{
		awsRegion:  awsRegion,
		awsService: awsService,
		client:     client,
		url:        url,
	}
}

// Search is a method that wraps the Search function of the elasticsearch package
func (cli *Client) Search(ctx context.Context, index, docType string, request []byte) ([]byte, error) {
	return cli.post(ctx, index, docType, "_search", request)
}

// MultiSearch is a method that wraps the MultiSearch function of the elasticsearch package
func (cli *Client) MultiSearch(ctx context.Context, index, docType string, request []byte) ([]byte, error) {
	return cli.post(ctx, index, docType, "_msearch", request)
}

func (cli *Client) post(ctx context.Context, index, docType, action string, request []byte) ([]byte, error) {
	reader := bytes.NewReader(request)
	req, err := http.NewRequest("POST", cli.url+"/"+buildContext(index, docType)+action, reader)
	if err != nil {
		return nil, err
	}

	resp, err := cli.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "elaticsearchClient error reading post response body")
	}

	return response, err
}

func buildContext(index, docType string) string {
	ctx := ""
	if index != "" {
		ctx = index + "/"
		if docType != "" {
			ctx += docType + "/"
		}
	}
	return ctx
}
