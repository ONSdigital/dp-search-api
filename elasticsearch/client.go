package elasticsearch

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	esauth "github.com/ONSdigital/dp-elasticsearch/v2/awsauth"
	elastic "github.com/ONSdigital/dp-elasticsearch/v2/elasticsearch"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
)

// Client represents an instance of the elasticsearch client
type Client struct {
	awsRegion    string
	awsSDKSigner *esauth.Signer
	awsService   string
	url          string
	client       dphttp.Clienter
	signRequests bool
	cfg          *config.Config
	esClient     *elastic.Client
}

// New creates a new elasticsearch client. Any trailing slashes from the URL are removed.
func New(url string, client dphttp.Clienter, signRequests bool, awsSDKSigner *esauth.Signer, awsService string, awsRegion string, cfg *config.Config, esClient *elastic.Client) *Client {
	return &Client{
		awsSDKSigner: awsSDKSigner,
		awsRegion:    awsRegion,
		awsService:   awsService,
		url:          strings.TrimRight(url, "/"),
		client:       client,
		signRequests: signRequests,
		cfg:          cfg,
		esClient:     esClient,
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

	if cli.signRequests {
		reader := bytes.NewReader([]byte{})
		if err = cli.awsSDKSigner.Sign(req, reader, time.Now()); err != nil {
			return nil, err
		}
	}

	resp, err := cli.client.Do(ctx, req)
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

	if cli.signRequests {
		if err = cli.awsSDKSigner.Sign(req, reader, time.Now()); err != nil {
			return nil, err
		}
	}

	resp, err := cli.client.Do(ctx, req)
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

//CreateNewEmptyIndex is a method that creates an empty Elasticsearch index with the given indexName
//It returns true if the index was created successfully, otherwise false
func (cli *Client) CreateNewEmptyIndex(ctx context.Context, indexName string) (bool, error) {
	indexCreated := false
	status, err := cli.esClient.CreateIndex(ctx, indexName, GetSearchIndexSettings())
	if err != nil {
		log.Error(ctx, "error creating index. Make sure that ELASTIC_SEARCH_URL is set to http://localhost:11200", err)
		return indexCreated, err
	}
	if status != http.StatusOK {
		log.Error(ctx, "error creating index http status - "+strconv.Itoa(status), err)
		return indexCreated, err
	}
	indexCreated = true
	return indexCreated, err
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
