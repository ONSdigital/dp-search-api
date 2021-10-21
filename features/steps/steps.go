package steps

import (
	"encoding/json"
	"fmt"
	"github.com/cucumber/messages-go/v10"
	"github.com/pkg/errors"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
	"github.com/ONSdigital/dp-search-api/transformer"
)


// RegisterSteps registers the specific steps needed to do component tests for the search api
func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	c.APIFeature.RegisterSteps(ctx)
	ctx.Step(`^return one search response$`, c.successfullyReturnOneSearchResult)
	ctx.Step(`^return multiple search responses$`, c.successfullyReturnMultipleSearchResults)
}

func iGET(arg1 string) error {
	return godog.ErrPending
}

func iGETSearch(arg1 string) error {
	return godog.ErrPending
}

func iShouldReceiveTheFollowingJSONResponse(arg1 *messages.PickleStepArgument_PickleDocString) error {
	return godog.ErrPending
}

func theHTTPStatusCodeShouldBe(arg1 string) error {
	return godog.ErrPending
}

func theResponseHeaderShouldBe(arg1, arg2 string) error {
	return godog.ErrPending
}



// delayTimeBySeconds pauses the goroutine for the given seconds
func delayTimeBySeconds(seconds string) error {
	sec, err := strconv.Atoi(seconds)
	if err != nil {
		return err
	}
	time.Sleep(time.Duration(sec) * time.Second)
	return nil
}

func (c *Component) successfullyReturnOneSearchResult() error {
	body, err := ioutil.ReadFile("testdata/single_search_result.json")
	if err != nil {
		return err
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	c.FakeElasticSearchAPI.setJSONResponseForGet("/ons/_search", 200, b)
	return nil
}

func (c *Component) successfullyReturnMultipleSearchResults() error {
	body, err := ioutil.ReadFile("testdata/multiple_search_results.json")
	if err != nil {
		return err
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	c.FakeElasticSearchAPI.setJSONResponseForGet("/ons/_search", 200, b)
	return nil
}

func (c *Component) iShouldReceiveTheFollowingSearchResponse(expectedResponse *godog.DocString) error {
	var searchResponse, expectedSearchResponse transformer.ESResponse

	responseBody, err := ioutil.ReadAll(c.APIFeature.HttpResponse.Body)
	if err != nil {
		return fmt.Errorf("failed to read response of search api component - error: %v", err)
	}

	err = json.Unmarshal(responseBody, &searchResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response of search api component - error: %v", err)
	}
	expectedSearchResults, err := ioutil.ReadFile("testdata/multiple_search_expected_results.json")
	if err != nil {
		return fmt.Errorf("failed to read file of expected results - error: %v", err)
	}

	err = json.Unmarshal(expectedSearchResults, &expectedSearchResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal expected results from file - error: %v", err)
	}

	c.validateSearchResponse(searchResponse, expectedSearchResponse)

	return c.ErrorFeature.StepError()
}

func (c *Component) validateSearchResponse(actualResponse, expectedResponse transformer.ESResponse) error {
	maxExpectedStartTime := c.StartTime.Add((c.cfg.HealthCheckInterval + 1) * time.Second)

	assert.Equal(&c.ErrorFeature, expectedResponse.Responses, actualResponse.Responses)

	//c.validateHealthVersion(healthResponse.Version, expectedResponse.Version, maxExpectedStartTime)

	if len(expectedResponse.Responses) < 1 {
		return fmt.Errorf("Response contained 0 items")
	}
	// TODO: check length first before iterating any arrays; nested for loops...check keywords, items, etc
	for _, checkExpectedResponse := range expectedResponse.Responses {
		for i, checkExpectedHits := range checkExpectedResponse.Hits.Hits {
		}
	}
	return nil
}

func (c *Component) validateHealthVersion(versionResponse healthcheck.VersionInfo, expectedVersion healthcheck.VersionInfo, maxExpectedStartTime time.Time) {
	assert.True(&c.ErrorFeature, versionResponse.BuildTime.Before(maxExpectedStartTime))
	assert.Equal(&c.ErrorFeature, expectedVersion.GitCommit, versionResponse.GitCommit)
	assert.Equal(&c.ErrorFeature, expectedVersion.Language, versionResponse.Language)
	assert.NotEmpty(&c.ErrorFeature, versionResponse.LanguageVersion)
	assert.Equal(&c.ErrorFeature, expectedVersion.Version, versionResponse.Version)
}

func (c *Component) validateHealthCheck(checkResponse *Check, expectedCheck *Check) {
	maxExpectedHealthCheckTime := c.StartTime.Add((c.cfg.HealthCheckInterval + c.cfg.HealthCheckCriticalTimeout + 1) * time.Second)

	assert.Equal(&c.ErrorFeature, expectedCheck.Name, checkResponse.Name)
	assert.Equal(&c.ErrorFeature, expectedCheck.Status, checkResponse.Status)
	assert.Equal(&c.ErrorFeature, expectedCheck.StatusCode, checkResponse.StatusCode)
	assert.Equal(&c.ErrorFeature, expectedCheck.Message, checkResponse.Message)
	assert.True(&c.ErrorFeature, checkResponse.LastChecked.Before(maxExpectedHealthCheckTime))
	assert.True(&c.ErrorFeature, checkResponse.LastChecked.After(c.StartTime))

	if expectedCheck.StatusCode == 200 {
		assert.True(&c.ErrorFeature, checkResponse.LastSuccess.Before(maxExpectedHealthCheckTime))
		assert.True(&c.ErrorFeature, checkResponse.LastSuccess.After(c.StartTime))
		assert.Equal(&c.ErrorFeature, expectedCheck.LastFailure, checkResponse.LastFailure)
	} else {
		assert.Equal(&c.ErrorFeature, expectedCheck.LastSuccess, checkResponse.LastSuccess)
		assert.True(&c.ErrorFeature, checkResponse.LastFailure.Before(maxExpectedHealthCheckTime))
		assert.True(&c.ErrorFeature, checkResponse.LastFailure.After(c.StartTime))
	}
}
