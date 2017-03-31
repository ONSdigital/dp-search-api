/* file: $GOPATH/src/godogs/godogs_test.go */
package main

import (
	"github.com/DATA-DOG/godog"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"testing"
	"os"
	"time"
	"strings"
	"fmt"
	"net/url"
)

func resolveBindAddr() string {
	bindAddr := os.Getenv("BIND_ADDR")
	if len(bindAddr) == 0 {
		bindAddr = ":10001"
	}
	return bindAddr
}

var bindAddr string = resolveBindAddr()
type Description struct {
	NextRelease string
	ReleaseDate time.Time
	DatasetUri  string
	Published   *bool `json:published,omitempty`
	Cancelled   *bool `json:cancelled,omitempty`
	Title string
}

type Source struct {
	Uri         string
	Description Description
}

type RecordHits struct {
	Id     string  `json:"_id"`
	Score  float64 `json:"_score"`
	Source Source  `json:"_source"`
	Type   string `json:"_type"`
	Index  string `json:"_index"`
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

func buildURL(params map[string]string) string {
	query := "http://localhost" + bindAddr + "/search?"
	for key, value := range params {
		fmt.Println("Key:", key, "Value:", value)
		query = query + "&" + url.PathEscape(key) + "=" + url.PathEscape(value)
	}
	return query
}
func search(params map[string]string) error {

	i := buildURL(params)
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
	httpResponse = HttpResponse{}


	f,_  := ioutil.TempFile("/tmp","responseParsed")
	str,err := json.Marshal(httpResponse)
	if err != nil {
		panic(err)
		return err
	}
	f.Write(str)


	o,err  := ioutil.TempFile("/tmp","responseOrigin")
	o.Write(response)
	if err != nil {
		panic(err)
		return err
	}
	return nil
}

func searchForTerm(term string) error {
	return search(map[string]string{"term": term})

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

func filterReleaseCalendar(pubOrUpComing string) error {
	params := map[string]string{pubOrUpComing:"true", "size":"1000"}
	return search(params)
}
/**
Upcoming means that the documents are not release
not published and not cancelled OR are published and are due
 */
func checkUpComing() error {
	for _, r := range httpResponse.Responses {
		for _, hit := range r.Hits.Hits {
			description := hit.Source.Description
			if !((!*description.Cancelled && !*description.Published) ||
				!(*description.Published && description.ReleaseDate.Before(time.Now()))) {
				str,_ := json.Marshal(hit)
				return fmt.Errorf("Document is not Upcoming", string(str))
			}
		}
	}
	//Upcomm
	return nil
}
/**
Published means that the documents are published and not cancelled OR are cancelled and are due
 */
func checkPublished() error {
	for _, r := range httpResponse.Responses {
		for _, hit := range r.Hits.Hits {
			description := hit.Source.Description
			if !((!*description.Cancelled && *description.Published) ||
				(*description.Cancelled && description.ReleaseDate.Before(time.Now()))) {
				str,_ := json.Marshal(hit)
				return fmt.Errorf("Document is not Published", string(str))
			}
		}
	}

	return nil
}

func FeatureContext(s *godog.Suite) {
	s.Step(`^a user searches for the term\(s\) "([a-zA-Z\s]*)"$`, searchForTerm)
	s.Step(`^the user will receive the first page with documents only from this uri prefix (.*)$`, onlyReceiveFromDatasetURI)
	s.Step(`^the results with the same score are in date descending order$`, theResultsAreInDateDescendingOrder)
	s.Step(`^a user filters the release calendar for "([^"]*)" documents$`, filterReleaseCalendar)
	s.Step(`^user will receive a list of the documents are upcoming$`, checkUpComing)
	s.Step(`^user will receive a list of the documents are published$`, checkPublished)

}