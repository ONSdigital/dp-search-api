package elasticsearch

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

var elasticURL string

func Setup(url string) {
	elasticURL = url
}

func MultiSearch(index string, docType string, request []byte) ([]byte, error) {
	action := "_msearch"
	return post(index, docType, action, request)
}

func Search(index string, docType string, request []byte) ([]byte, error) {
	action := "_search"
	return post(index, docType, action, request)
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
	return get("_cat", "health", "", nil)
}

func post(index string, docType string, action string, request []byte) ([]byte, error) {
	reader := bytes.NewReader(request)
	req, err := http.NewRequest("POST", elasticURL+buildContext(index, docType)+action, reader)
	if err != nil {
		return nil, err
	}
	// For control over HTTP client headers,
	// redirect policy, and other settings,
	// create a Client
	// A Client is an HTTP client
	client := &http.Client{}

	// Send the request via a client
	// Do sends an HTTP request and
	// returns an HTTP response
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	response, err := ioutil.ReadAll(resp.Body)

	return response, err
}

func get(index string, docType string, action string, request []byte) ([]byte, error) {
	reader := bytes.NewReader(request)
	req, err := http.NewRequest("GET", elasticURL+buildContext(index, docType)+action, reader)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	response, err := ioutil.ReadAll(resp.Body)

	return response, err
}
