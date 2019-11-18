package elasticsearch

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strings"

	rchttp "github.com/ONSdigital/dp-rchttp"
)

// Client represents an instance of the elasticsearch client
type Client struct {
	url          string
	rchttpClient rchttp.Clienter
}

// New creates a new elasticsearch client
func New(url string, rchttpClient rchttp.Clienter) *Client {
	return &Client{
		url:          strings.TrimRight(url, "/"),
		rchttpClient: rchttpClient,
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
func (cli *Client) GetStatus() ([]byte, error) {
	req, err := http.NewRequest("GET", cli.url+"/_cat/health", nil)
	if err != nil {
		return nil, err
	}
	resp, err := cli.rchttpClient.Do(context.Background(), req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)

	return response, err
}

func (cli *Client) post(ctx context.Context, index string, docType string, action string, request []byte) ([]byte, error) {
	reader := bytes.NewReader(request)
	req, err := http.NewRequest("POST", cli.url+"/"+buildContext(index, docType)+action, reader)
	if err != nil {
		return nil, err
	}

	resp, err := cli.rchttpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)

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
