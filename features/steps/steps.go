package steps

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/cucumber/messages-go/v10"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-search-api/transformer"
	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

// RegisterSteps registers the specific steps needed to do component tests for the search api
func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	c.APIFeature.RegisterSteps(ctx)
	ctx.Step(`^elasticsearch returns one item in search response$`, c.successfullyReturnOneSearchResult)
	ctx.Step(`^elasticsearch returns multiple items in search response$`, c.successfullyReturnMultipleSearchResults)
	ctx.Step(`^the response body is the same as the json in "([^"]*)"$`, c.iShouldReceiveTheFollowingSearchResponse)

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

// TheHTTPStatusCodeShouldBe asserts that the status code of the response matches the expected code
func (c *Component) theHTTPStatusCodeShouldBes(expectedCodeStr string) error {
	expectedCode, err := strconv.Atoi(expectedCodeStr)
	if err != nil {
		return err
	}

	assert.Equal(c.APIFeature, expectedCode, c.APIFeature.HttpResponse.StatusCode)
	return nil
}

func (c *Component) successfullyReturnOneSearchResult() error {
	body, err := ioutil.ReadFile("./features/testdata/single_search_result.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.setJSONResponseForPost("/elasticsearch/ons/_msearch", 200, body)
	return nil
}

func (c *Component) successfullyReturnMultipleSearchResults() error {
	body, err := ioutil.ReadFile("./features/testdata/mulitple_search_results.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.setJSONResponseForPost("/elasticsearch/ons/_msearch", 200, body)

	return nil
}

func (c *Component) iShouldReceiveTheFollowingSearchResponse(expectedJSONFile string) error {
	var searchResponse, expectedSearchResponse transformer.SearchResponse

	responseBody, err := ioutil.ReadAll(c.APIFeature.HttpResponse.Body)
	if err != nil {
		return fmt.Errorf("failed to read response of search api component - error: %v", err)
	}

	err = json.Unmarshal(responseBody, &searchResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response of search api component - error: %v", err)
	}
	expectedSearchResults, err := ioutil.ReadFile(expectedJSONFile)
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

func (c *Component) validateSearchResponse(actualResponse, expectedResponse transformer.SearchResponse) error {

	// assert.Equal(&c.ErrorFeature, expectedResponse.Items, actualResponse.Items)

	// Probably remove this for 0 items being returned (which is a valid scenario which returns 200 OK responses)
	if len(actualResponse.Items) < 1 {
		return fmt.Errorf("Response contained 0 items")
	}

	assert.Equal(&c.ErrorFeature, actualResponse.AdditionSuggestions, expectedResponse.AdditionSuggestions)

	// TODO: check length first before iterating any arrays; nested for loops...check keywords, items, etc
	// for i, _ := range actualResponse.Items {
	// 	checkItems(actualResponse.Items[i], expectedResponse.Items[i])
	// }

	return nil
}

// TODO remove this - kept in so know what assertions to use
func (c *Component) validateHealthVersion(versionResponse healthcheck.VersionInfo, expectedVersion healthcheck.VersionInfo, maxExpectedStartTime time.Time) {
	assert.True(&c.ErrorFeature, versionResponse.BuildTime.Before(maxExpectedStartTime))
	assert.Equal(&c.ErrorFeature, expectedVersion.GitCommit, versionResponse.GitCommit)
	assert.Equal(&c.ErrorFeature, expectedVersion.Language, versionResponse.Language)
	assert.NotEmpty(&c.ErrorFeature, versionResponse.LanguageVersion)
	assert.Equal(&c.ErrorFeature, expectedVersion.Version, versionResponse.Version)
}
