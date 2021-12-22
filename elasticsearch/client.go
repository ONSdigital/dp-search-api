package elasticsearch

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	esauth "github.com/ONSdigital/dp-elasticsearch/v2/awsauth"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/pkg/errors"
)

// Client represents an instance of the elasticsearch client - now deprecated
type Client struct {
	awsRegion    string
	awsSDKSigner *esauth.Signer
	awsService   string
	url          string
	client       dphttp.Clienter
	signRequests bool
}

// New creates a new elasticsearch client. Any trailing slashes from the URL are removed.
func New(url string, client dphttp.Clienter, signRequests bool, awsSDKSigner *esauth.Signer, awsService, awsRegion string) *Client {
	return &Client{
		awsSDKSigner: awsSDKSigner,
		awsRegion:    awsRegion,
		awsService:   awsService,
		client:       client,
		signRequests: signRequests,
		url:          url,
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

// GetStatus makes status call for healthcheck purposes
func (cli *Client) GetStatus(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequest("GET", cli.url+"/_cat/health", http.NoBody)
	if err != nil {
		return nil, err
	}

	if cli.signRequests {
		reader := bytes.NewReader([]byte{})
		if awsSignErr := cli.awsSDKSigner.Sign(req, reader, time.Now()); awsSignErr != nil {
			return nil, awsSignErr
		}
	}

	resp, err := cli.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "elaticsearchClient error reading get status response body")
	}

	return response, err
}

func (cli *Client) post(ctx context.Context, index, docType, action string, request []byte) ([]byte, error) {
	reader := bytes.NewReader(request)
	req, err := http.NewRequest("POST", cli.url+"/"+buildContext(index, docType)+action, reader)
	if err != nil {
		return nil, err
	}

	if cli.signRequests {
		if awsSignErr := cli.awsSDKSigner.Sign(req, reader, time.Now()); awsSignErr != nil {
			return nil, awsSignErr
		}
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
	if len(index) > 0 {
		ctx = index + "/"
		if len(docType) > 0 {
			ctx += docType + "/"
		}
	}
	return ctx
}
