package elasticsearch

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strings"

	net "github.com/ONSdigital/dp-net"
	"github.com/pkg/errors"
)

// Client represents an instance of the elasticsearch client
type Client struct {
	url          string
	netClient net.Clienter
}

// New creates a new elasticsearch client. Any trailing slashes from the URL are removed.
func New(url string, netClient net.Clienter) *Client {
	return &Client{
		url:          strings.TrimRight(url, "/"),
		netClient: netClient,
	}
}

// Search is a method that wraps the Search function of the elasticsearch package
func (cli *Client) Search(ctx context.Context, index string, docType string, request []byte) ([]byte, error) {
	return cli.post(ctx, index, docType, "_search", request)
}

// MultiSearch is a method that wraps the MultiSearch function of the elasticsearch package
func (cli *Client) MultiSearch(ctx context.Context, index string, docType string, request []byte) ([]byte, error) {
	return cli.post(ctx, index, docType, "_msearch", request)
}

// GetStatus makes status call for healthcheck purposes
func (cli *Client) GetStatus(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequest("GET", cli.url+"/_cat/health", nil)
	if err != nil {
		return nil, err
	}
	resp, err := cli.netClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "elaticsearchClient error reading get status response body")
	}

	return response, err
}

func (cli *Client) post(ctx context.Context, index string, docType string, action string, request []byte) ([]byte, error) {
	reader := bytes.NewReader(request)
	req, err := http.NewRequest("POST", cli.url+"/"+buildContext(index, docType)+action, reader)
	if err != nil {
		return nil, err
	}

	resp, err := cli.netClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "elaticsearchClient error reading post response body")
	}

	return response, err
}

func buildContext(index string, docType string) string {
	context := ""
	if len(index) > 0 {
		context = index + "/"
		if len(docType) > 0 {
			context += docType + "/"
		}
	}
	return context
}
