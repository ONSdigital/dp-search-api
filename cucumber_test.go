/* file: $GOPATH/src/godogs/godogs_test.go */
package main

import (
	"github.com/DATA-DOG/godog"
	"github.com/ONSdigital/go-ns/log"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"testing"
	"os"
	"github.com/ONSdigital/dp-search-query/elasticsearch"
	"time"
	"strings"
	"fmt"
	"net/url"
)

type Source struct {
	Uri         string
	Description struct {
			    NextRelease string
			    ReleaseDate time.Time
			    DatasetUri  string
		    }
}

type RecordHits struct {
	Id     string  `json:"_id"`
	Score  float64 `json:"_score"`
	Source Source  `json:"_source"`

	Sort   [] float64
}

type Hits struct {
	Total int64
	Hits  []RecordHits
}

type Responses struct {
	Took int
	Hits Hits
}
type HttpResponse struct {
	Responses []Responses
}

var httpResponse HttpResponse

func TestMain(m *testing.M) {
	config.ElasticURL = "http://localhost:9999/"
	go main()

	status := godog.RunWithOptions("search", func(s *godog.Suite) {
		FeatureContext(s)
	}, godog.Options{
		Format: "progress",
		Paths:  []string{"features"},
	})

	if st := m.Run(); st > status {
		status = st
	}

	os.Exit(status)
}

func searchForTerm(term string) error {
	i := "http://localhost:10001/search?term=" + url.PathEscape(term)
	log.Debug("searchForTerm", log.Data{"url":i})
	req, err := http.NewRequest("GET", i, nil)
	if err != nil {
		panic(err)
		return err
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
		panic(err)
		return err
	}
	var response []byte
	response, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
		return err
	}
	ioutil.WriteFile("/tmp/testResponse.json", response, 0644)

	err = json.Unmarshal(response, &httpResponse)
	if err != nil {
		panic(err)
		return err
	}
	log.Debug("indexeddata", log.Data{"structure":httpResponse})

	return nil
}

func onlyReceiveFromDatasetURI(uri string) error {
	for _, r := range httpResponse.Responses {
		for _, hit := range r.Hits.Hits {
			if !strings.HasPrefix(uri, hit.Source.Description.DatasetUri) {
				return fmt.Errorf("URI %s does not match expected %s", hit.Source.Description.DatasetUri, uri)
			}
		}
	}
	return nil
}

func theResultsAreInDateDescendingOrder() error {
	var lastHit RecordHits

	for _, r := range httpResponse.Responses {
		for _, hit := range r.Hits.Hits {
			currentDate := hit.Source.Description.ReleaseDate
			lastTime := lastHit.Source.Description.ReleaseDate
			if lastHit.Id != "" && lastHit.Score == hit.Score  && lastTime.Before(currentDate) {
				return fmt.Errorf("date order is not valid date last Hit %s is before %s current %s which is %s",
					lastHit.Id, lastTime, hit.Id, currentDate, lastTime)
			}
			lastHit = hit
		}
	}
	return nil
}

func FeatureContext(s *godog.Suite) {
	s.Step(`^a user searches for the term\(s\) "([a-zA-Z\s]*)"$`, searchForTerm)
	s.Step(`^the user will receive the first page with documents only from this uri prefix (.*)$`, onlyReceiveFromDatasetURI)
	s.Step(`^the results with the same score are in date descending order$`, theResultsAreInDateDescendingOrder)


}
