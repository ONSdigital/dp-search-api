package config

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
)

var ElasticURL = getEnv("ELASTIC_URL", "http://localhost:9200/")

func getEnv(key string, defaultValue string) string {
	envValue := os.Getenv(key)
	if len(envValue) == 0 {
		envValue = defaultValue
	}
	return envValue
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

func post(index string, docType string, action string, request []byte) ([]byte, error) {
	reader := bytes.NewReader(request)
	req, err := http.NewRequest("POST", ElasticURL+buildContext(index, docType)+action, reader)
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
	ioutil.WriteFile("/tmp/request.json", request, 0644)
	response, err := ioutil.ReadAll(resp.Body)
	ioutil.WriteFile("/tmp/response.json", response, 0644)
	return response, nil
}

func MultiSearch(index string, docType string, request []byte) ([]byte, error) {
	action := "_msearch"
	return post(index, docType, action, request)
}

func Search(index string, docType string, request []byte) ([]byte, error) {
	action := "_search"
	return post(index, docType, action, request)
}
