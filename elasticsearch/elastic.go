package elasticsearch

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strings"

	rchttp "github.com/ONSdigital/dp-rchttp"
)

var elasticURL string
var cli rchttp.Clienter

//Setup initialises the elasticsearch ,module with a url, stripping any trailing slashes
func Setup(url string, client rchttp.Clienter) {
	elasticURL = strings.TrimRight(url, "/")
	cli = client
}

func MultiSearch(ctx context.Context, index string, docType string, request []byte) ([]byte, error) {
	action := "_msearch"
	return post(ctx, index, docType, action, request)
}

func Search(ctx context.Context, index string, docType string, request []byte) ([]byte, error) {
	action := "_search"
	return post(ctx, index, docType, action, request)
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

func GetStatus() ([]byte, error) {
	req, err := http.NewRequest("GET", elasticURL+"/_cat/health", nil)
	if err != nil {
		return nil, err
	}
	resp, err := cli.Do(context.Background(), req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)

	return response, err
}

func post(ctx context.Context, index string, docType string, action string, request []byte) ([]byte, error) {
	reader := bytes.NewReader(request)
	req, err := http.NewRequest("POST", elasticURL+"/"+buildContext(index, docType)+action, reader)
	if err != nil {
		return nil, err
	}

	resp, err := cli.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)

	return response, err
}
