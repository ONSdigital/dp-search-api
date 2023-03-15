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
	"github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/dp-search-api/transformer"
	. "github.com/smartystreets/goconvey/convey"
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

	Convey("Given clienter.Do returns an error", t, func() {
		clientError := errors.New("unexpected error")
		httpClient := newMockHTTPClient(&http.Response{}, clientError)
		searchAPIClient := newSearchAPIClient(t, httpClient)
		check := initialTestState

		Convey("When search API client Checker is called", func() {
			err := searchAPIClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("Then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, health.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 0)
				So(check.Message(), ShouldEqual, clientError.Error())
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("And client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("Given a 500 response for health check", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusInternalServerError}, nil)
		searchAPIClient := newSearchAPIClient(t, httpClient)
		check := initialTestState

		Convey("When search API client Checker is called", func() {
			err := searchAPIClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("Then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, health.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 500)
				So(check.Message(), ShouldEqual, service+healthcheck.StatusMessage[health.StatusCritical])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("And client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
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

	Convey("Given request is authorised to create a new search index", t, func() {
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

		Convey("When CreateIndex is called", func() {
			resp, err := searchAPIClient.CreateIndex(ctx, Options{Headers: headers})

			Convey("Then the expected response body is returned", func() {
				So(*resp, ShouldResemble, createIndexResponse)

				Convey("And no error is returned", func() {
					So(err, ShouldBeNil)

					Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						So(doCalls, ShouldHaveLength, 1)
						So(doCalls[0].Req.Method, ShouldEqual, "POST")
						So(doCalls[0].Req.URL.Path, ShouldEqual, "/search")
						So(doCalls[0].Req.Header["Authorization"], ShouldResemble, []string{"Bearer authorised-user"})
					})
				})
			})
		})
	})

	Convey("Given a 401 response from search api", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusUnauthorized}, nil)
		searchAPIClient := newSearchAPIClient(t, httpClient)

		Convey("When CreateIndex is called", func() {
			resp, err := searchAPIClient.CreateIndex(ctx, Options{})

			Convey("Then an error should be returned ", func() {
				So(err, ShouldNotBeNil)
				So(err.Status(), ShouldEqual, http.StatusUnauthorized)
				So(err.Error(), ShouldEqual, "failed as unexpected code from search api: 401")

				Convey("And the expected responde body should be nil", func() {
					So(resp, ShouldBeNil)

					Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						So(doCalls, ShouldHaveLength, 1)
						So(doCalls[0].Req.Method, ShouldEqual, "POST")
						So(doCalls[0].Req.URL.Path, ShouldEqual, "/search")
						So(doCalls[0].Req.Header["Authorization"], ShouldBeEmpty)
					})
				})
			})
		})
	})

	Convey("Given a 500 response from search api", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusInternalServerError}, nil)
		searchAPIClient := newSearchAPIClient(t, httpClient)

		Convey("When CreateIndex is called", func() {
			resp, err := searchAPIClient.CreateIndex(ctx, Options{Headers: headers})

			Convey("Then an error should be returned ", func() {
				So(err, ShouldNotBeNil)
				So(err.Status(), ShouldEqual, http.StatusInternalServerError)

				Convey("And the expected responde body should be nil", func() {
					So(resp, ShouldBeNil)

					Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						So(doCalls, ShouldHaveLength, 1)
						So(doCalls[0].Req.Method, ShouldEqual, "POST")
						So(doCalls[0].Req.URL.Path, ShouldEqual, "/search")
						So(doCalls[0].Req.Header["Authorization"], ShouldResemble, []string{"Bearer authorised-user"})
					})
				})
			})
		})
	})
}

func TestGetSearch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("Given request to find search results", t, func() {
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

		Convey("When GetSearch is called", func() {
			query := url.Values{}
			query.Add("q", "census")
			resp, err := searchAPIClient.GetSearch(ctx, Options{Query: query})

			Convey("Then the expected response body is returned", func() {
				So(*resp, ShouldResemble, searchResults)

				Convey("And no error is returned", func() {
					So(err, ShouldBeNil)

					Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						So(doCalls, ShouldHaveLength, 1)
						So(doCalls[0].Req.Method, ShouldEqual, "GET")
						So(doCalls[0].Req.URL.Path, ShouldEqual, "/search")
						So(doCalls[0].Req.URL.Query().Get("q"), ShouldEqual, "census")
						So(doCalls[0].Req.Header["Authorization"], ShouldBeEmpty)
					})
				})
			})
		})
	})
}

func TestGetReleaseCalendar(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("Given request to find release calendar results", t, func() {
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

		Convey("When GetReleaseCalendarEntries is called", func() {
			query := url.Values{}
			query.Add("q", "census")
			resp, err := searchAPIClient.GetReleaseCalendarEntries(ctx, Options{Query: query})

			Convey("Then the expected response body is returned", func() {
				So(*resp, ShouldResemble, releaseCalendarResults)

				Convey("And no error is returned", func() {
					So(err, ShouldBeNil)

					Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						So(doCalls, ShouldHaveLength, 1)
						So(doCalls[0].Req.Method, ShouldEqual, "GET")
						So(doCalls[0].Req.URL.Path, ShouldEqual, "/search/releases")
						So(doCalls[0].Req.URL.Query().Get("q"), ShouldEqual, "census")
						So(doCalls[0].Req.Header["Authorization"], ShouldBeEmpty)
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

func newSearchAPIClient(t *testing.T, httpClient *dphttp.ClienterMock) *Client {
	healthClient := healthcheck.NewClientWithClienter(service, testHost, httpClient)
	return NewWithHealthClient(healthClient)
}
