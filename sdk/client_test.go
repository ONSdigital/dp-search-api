package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-search-api/api"
	"github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/dp-search-api/transformer"
	c "github.com/smartystreets/goconvey/convey"
)

const testHost = "http://localhost:23900"

var (
	initialTestState = healthcheck.CreateCheckState(service)

	createIndexResponse = models.CreateIndexResponse{
		IndexName: "index-name",
	}

	searchResults = models.SearchResponse{
		Count: 1,
		Took:  10,
		Topics: []models.FilterCount{
			{
				Type:  "4443",
				Count: 1,
			},
		},
		ContentTypes: []models.FilterCount{
			{
				Type:  "dataset_landing_page",
				Count: 1,
			},
		},
		Items: []models.Item{
			{
				DataType:  "dataset_landing_page",
				DatasetID: "Census 2021 Age Stats",
				URI:       "https://ons.gov.uk.test/datasets/123",
			},
		},
	}

	releaseCalendarResults = transformer.SearchReleaseResponse{
		Took: 20,
		Breakdown: transformer.Breakdown{
			Published: 1,
		},
		Releases: []transformer.Release{
			{
				URI: "https://ons.gov.uk.test/releases/123",
				Description: transformer.ReleaseDescription{
					Title:   "test title",
					Summary: "test description",
				},
			},
		},
	}
)

func TestHealthCheckerClient(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	timePriorHealthCheck := time.Now().UTC()
	path := "/health"

	c.Convey("Given clienter.Do returns an error", t, func() {
		clientError := errors.New("unexpected error")
		httpClient := newMockHTTPClient(&http.Response{}, clientError)
		searchAPIClient := newSearchAPIClient(t, httpClient)
		check := initialTestState

		c.Convey("When search API client Checker is called", func() {
			err := searchAPIClient.Checker(ctx, &check)
			c.So(err, c.ShouldBeNil)

			c.Convey("Then the expected check is returned", func() {
				c.So(check.Name(), c.ShouldEqual, service)
				c.So(check.Status(), c.ShouldEqual, health.StatusCritical)
				c.So(check.StatusCode(), c.ShouldEqual, 0)
				c.So(check.Message(), c.ShouldEqual, clientError.Error())
				c.So(*check.LastChecked(), c.ShouldHappenAfter, timePriorHealthCheck)
				c.So(check.LastSuccess(), c.ShouldBeNil)
				c.So(*check.LastFailure(), c.ShouldHappenAfter, timePriorHealthCheck)
			})

			c.Convey("And client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				c.So(doCalls, c.ShouldHaveLength, 1)
				c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, path)
			})
		})
	})

	c.Convey("Given a 500 response for health check", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusInternalServerError}, nil)
		searchAPIClient := newSearchAPIClient(t, httpClient)
		check := initialTestState

		c.Convey("When search API client Checker is called", func() {
			err := searchAPIClient.Checker(ctx, &check)
			c.So(err, c.ShouldBeNil)

			c.Convey("Then the expected check is returned", func() {
				c.So(check.Name(), c.ShouldEqual, service)
				c.So(check.Status(), c.ShouldEqual, health.StatusCritical)
				c.So(check.StatusCode(), c.ShouldEqual, 500)
				c.So(check.Message(), c.ShouldEqual, service+healthcheck.StatusMessage[health.StatusCritical])
				c.So(*check.LastChecked(), c.ShouldHappenAfter, timePriorHealthCheck)
				c.So(check.LastSuccess(), c.ShouldBeNil)
				c.So(*check.LastFailure(), c.ShouldHappenAfter, timePriorHealthCheck)
			})

			c.Convey("And client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				c.So(doCalls, c.ShouldHaveLength, 1)
				c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, path)
			})
		})
	})
}

func TestCreateIndex(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	headers := http.Header{
		Authorization: {"Bearer authorised-user"},
	}

	c.Convey("Given request is authorised to create a new search index", t, func() {
		body, err := json.Marshal(createIndexResponse)
		if err != nil {
			t.Errorf("failed to setup test data, error: %v", err)
		}

		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(bytes.NewReader(body)),
			},
			nil)

		searchAPIClient := newSearchAPIClient(t, httpClient)

		c.Convey("When CreateIndex is called", func() {
			resp, err := searchAPIClient.CreateIndex(ctx, Options{Headers: headers})

			c.Convey("Then the expected response body is returned", func() {
				c.So(*resp, c.ShouldResemble, createIndexResponse)

				c.Convey("And no error is returned", func() {
					c.So(err, c.ShouldBeNil)

					c.Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						c.So(doCalls, c.ShouldHaveLength, 1)
						c.So(doCalls[0].Req.Method, c.ShouldEqual, "POST")
						c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, "/search")
						c.So(doCalls[0].Req.Header["Authorization"], c.ShouldResemble, []string{"Bearer authorised-user"})
					})
				})
			})
		})
	})

	c.Convey("Given a 401 response from search api", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusUnauthorized}, nil)
		searchAPIClient := newSearchAPIClient(t, httpClient)

		c.Convey("When CreateIndex is called", func() {
			resp, err := searchAPIClient.CreateIndex(ctx, Options{})

			c.Convey("Then an error should be returned ", func() {
				c.So(err, c.ShouldNotBeNil)
				c.So(err.Status(), c.ShouldEqual, http.StatusUnauthorized)
				c.So(err.Error(), c.ShouldEqual, "failed as unexpected code from search api: 401")

				c.Convey("And the expected responde body should be nil", func() {
					c.So(resp, c.ShouldBeNil)

					c.Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						c.So(doCalls, c.ShouldHaveLength, 1)
						c.So(doCalls[0].Req.Method, c.ShouldEqual, "POST")
						c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, "/search")
						c.So(doCalls[0].Req.Header["Authorization"], c.ShouldBeEmpty)
					})
				})
			})
		})
	})

	c.Convey("Given a 500 response from search api", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusInternalServerError}, nil)
		searchAPIClient := newSearchAPIClient(t, httpClient)

		c.Convey("When CreateIndex is called", func() {
			resp, err := searchAPIClient.CreateIndex(ctx, Options{Headers: headers})

			c.Convey("Then an error should be returned ", func() {
				c.So(err, c.ShouldNotBeNil)
				c.So(err.Status(), c.ShouldEqual, http.StatusInternalServerError)

				c.Convey("And the expected responde body should be nil", func() {
					c.So(resp, c.ShouldBeNil)

					c.Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						c.So(doCalls, c.ShouldHaveLength, 1)
						c.So(doCalls[0].Req.Method, c.ShouldEqual, "POST")
						c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, "/search")
						c.So(doCalls[0].Req.Header["Authorization"], c.ShouldResemble, []string{"Bearer authorised-user"})
					})
				})
			})
		})
	})
}

func TestGetSearch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	c.Convey("Given request to find search results", t, func() {
		body, err := json.Marshal(searchResults)
		if err != nil {
			t.Errorf("failed to setup test data, error: %v", err)
		}

		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(bytes.NewReader(body)),
			},
			nil)

		searchAPIClient := newSearchAPIClient(t, httpClient)

		c.Convey("When GetSearch is called", func() {
			query := url.Values{}
			query.Add("q", "census")
			resp, err := searchAPIClient.GetSearch(ctx, Options{Query: query})

			c.Convey("Then the expected response body is returned", func() {
				c.So(*resp, c.ShouldResemble, searchResults)

				c.Convey("And no error is returned", func() {
					c.So(err, c.ShouldBeNil)

					c.Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						c.So(doCalls, c.ShouldHaveLength, 1)
						c.So(doCalls[0].Req.Method, c.ShouldEqual, "GET")
						c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, "/search")
						c.So(doCalls[0].Req.URL.Query().Get("q"), c.ShouldEqual, "census")
						c.So(doCalls[0].Req.Header["Authorization"], c.ShouldBeEmpty)
					})
				})
			})
		})
	})
	c.Convey("Given a request to find search results with NLP weighting enabled", t, func() {
		body, err := json.Marshal(searchResults)
		if err != nil {
			t.Errorf("failed to setup test data, error: %v", err)
		}

		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(bytes.NewReader(body)),
			},
			nil)

		searchAPIClient := newSearchAPIClient(t, httpClient)

		c.Convey("When GetSearch is called", func() {
			query := url.Values{}
			query.Add("q", "census")
			query.Add("nlp_weighting", "true")
			resp, err := searchAPIClient.GetSearch(ctx, Options{Query: query})

			c.Convey("Then the expected response body is returned", func() {
				c.So(*resp, c.ShouldResemble, searchResults)

				c.Convey("And no error is returned", func() {
					c.So(err, c.ShouldBeNil)

					c.Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						c.So(doCalls, c.ShouldHaveLength, 1)
						c.So(doCalls[0].Req.Method, c.ShouldEqual, "GET")
						c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, "/search")
						c.So(doCalls[0].Req.URL.Query().Get("q"), c.ShouldEqual, "census")
						c.So(doCalls[0].Req.URL.Query().Get("nlp_weighting"), c.ShouldEqual, "true")
						c.So(doCalls[0].Req.Header["Authorization"], c.ShouldBeEmpty)
					})
				})
			})
		})
	})
}

func TestGetReleaseCalendar(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	c.Convey("Given request to find release calendar results", t, func() {
		body, err := json.Marshal(releaseCalendarResults)
		if err != nil {
			t.Errorf("failed to setup test data, error: %v", err)
		}

		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(bytes.NewReader(body)),
			},
			nil)

		searchAPIClient := newSearchAPIClient(t, httpClient)

		c.Convey("When GetReleaseCalendarEntries is called", func() {
			query := url.Values{}
			query.Add("q", "census")
			resp, err := searchAPIClient.GetReleaseCalendarEntries(ctx, Options{Query: query})

			c.Convey("Then the expected response body is returned", func() {
				c.So(*resp, c.ShouldResemble, releaseCalendarResults)

				c.Convey("And no error is returned", func() {
					c.So(err, c.ShouldBeNil)

					c.Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						c.So(doCalls, c.ShouldHaveLength, 1)
						c.So(doCalls[0].Req.Method, c.ShouldEqual, "GET")
						c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, "/search/releases")
						c.So(doCalls[0].Req.URL.Query().Get("q"), c.ShouldEqual, "census")
						c.So(doCalls[0].Req.Header["Authorization"], c.ShouldBeEmpty)
					})
				})
			})
		})
	})
}

func TestGetSearchURIs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	headers := http.Header{
		Authorization: {"Bearer authorised-user"},
	}

	urisRequest := api.URIsRequest{
		URIs: []string{
			"/economy",
		},
	}

	c.Convey("Given a request to get search URIs", t, func() {
		body, err := json.Marshal(searchResults) // Assuming the response structure is similar
		if err != nil {
			t.Errorf("failed to setup test data, error: %v", err)
		}

		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(body)),
			},
			nil)

		searchAPIClient := newSearchAPIClient(t, httpClient)

		c.Convey("When GetSearchURIs is called", func() {
			resp, err := searchAPIClient.PostSearchURIs(ctx, Options{Headers: headers}, urisRequest)

			c.Convey("Then the expected response body is returned", func() {
				c.So(*resp, c.ShouldResemble, searchResults)

				c.Convey("And no error is returned", func() {
					c.So(err, c.ShouldBeNil)

					c.Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						c.So(doCalls, c.ShouldHaveLength, 1)
						c.So(doCalls[0].Req.Method, c.ShouldEqual, "POST")
						c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, "/search/uris")
						c.So(doCalls[0].Req.Header["Authorization"], c.ShouldResemble, []string{"Bearer authorised-user"})
					})
				})
			})
		})
	})

	c.Convey("Given a 400 response from search API", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusBadRequest}, nil)
		searchAPIClient := newSearchAPIClient(t, httpClient)

		c.Convey("When GetSearchURIs is called", func() {
			resp, err := searchAPIClient.PostSearchURIs(ctx, Options{Headers: headers}, urisRequest)

			c.Convey("Then an error should be returned", func() {
				c.So(err, c.ShouldNotBeNil)
				c.So(err.Status(), c.ShouldEqual, http.StatusBadRequest)
				c.So(err.Error(), c.ShouldEqual, "failed as unexpected code from search api: 400")

				c.Convey("And the expected response body should be nil", func() {
					c.So(resp, c.ShouldBeNil)

					c.Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						c.So(doCalls, c.ShouldHaveLength, 1)
						c.So(doCalls[0].Req.Method, c.ShouldEqual, "POST")
						c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, "/search/uris")
						c.So(doCalls[0].Req.Header["Authorization"], c.ShouldResemble, []string{"Bearer authorised-user"})
					})
				})
			})
		})
	})

	c.Convey("Given a 500 response from search API", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusInternalServerError}, nil)
		searchAPIClient := newSearchAPIClient(t, httpClient)

		c.Convey("When GetSearchURIs is called", func() {
			resp, err := searchAPIClient.PostSearchURIs(ctx, Options{Headers: headers}, urisRequest)

			c.Convey("Then an error should be returned", func() {
				c.So(err, c.ShouldNotBeNil)
				c.So(err.Status(), c.ShouldEqual, http.StatusInternalServerError)

				c.Convey("And the expected response body should be nil", func() {
					c.So(resp, c.ShouldBeNil)

					c.Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						c.So(doCalls, c.ShouldHaveLength, 1)
						c.So(doCalls[0].Req.Method, c.ShouldEqual, "POST")
						c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, "/search/uris")
						c.So(doCalls[0].Req.Header["Authorization"], c.ShouldResemble, []string{"Bearer authorised-user"})
					})
				})
			})
		})
	})
}

func newMockHTTPClient(r *http.Response, err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(paths []string) {
		},
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return r, err
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
	}
}

func newSearchAPIClient(_ *testing.T, httpClient *dphttp.ClienterMock) *Client {
	healthClient := healthcheck.NewClientWithClienter(service, testHost, httpClient)
	return NewWithHealthClient(healthClient)
}
