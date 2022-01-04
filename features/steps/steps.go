package steps

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ONSdigital/dp-search-api/transformer"
	"github.com/cucumber/godog"
	"github.com/google/go-cmp/cmp"
)

// RegisterSteps registers the specific steps needed to do component tests for the search api
func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	c.APIFeature.RegisterSteps(ctx)
	ctx.Step(`^elasticsearch returns one item in search response$`, c.successfullyReturnSingleSearchResult)
	ctx.Step(`^elasticsearch returns multiple items in search response$`, c.successfullyReturnMultipleSearchResults)
	ctx.Step(`^elasticsearch returns zero items in search response$`, c.successfullyReturnNoSearchResults)
	ctx.Step(`^elasticsearch returns internal server error$`, c.failureInternalServerError)
	ctx.Step(`^the response body is the same as the json in "([^"]*)"$`, c.iShouldReceiveTheFollowingSearchResponse)
}

func (c *Component) successfullyReturnMultipleSearchResults() error {
	body, err := os.ReadFile("./features/testdata/es_mulitple_search_results.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.setJSONResponseForPost("/elasticsearch/ons/_msearch", 200, body)

	return nil
}

func (c *Component) successfullyReturnSingleSearchResult() error {
	body, err := os.ReadFile("./features/testdata/es_single_search_result.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.setJSONResponseForPost("/elasticsearch/ons/_msearch", 200, body)

	return nil
}

func (c *Component) successfullyReturnNoSearchResults() error {
	body, err := os.ReadFile("./features/testdata/es_zero_search_results.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.setJSONResponseForPost("/elasticsearch/ons/_msearch", 200, body)

	return nil
}

func (c *Component) failureInternalServerError() error {
	c.FakeElasticSearchAPI.setJSONResponseForPost("/elasticsearch/ons/_msearch", 500, nil)

	return nil
}

func (c *Component) iShouldReceiveTheFollowingSearchResponse(expectedJSONFile string) error {
	var searchResponse, expectedSearchResponse transformer.SearchResponse

	responseBody, err := io.ReadAll(c.APIFeature.HttpResponse.Body)
	if err != nil {
		return fmt.Errorf("failed to read response of search api component - error: %v", err)
	}

	err = json.Unmarshal(responseBody, &searchResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response of search api component - error: %v", err)
	}
	expectedSearchResults, err := os.ReadFile(expectedJSONFile)
	if err != nil {
		return fmt.Errorf("failed to read file of expected results - error: %v", err)
	}

	err = json.Unmarshal(expectedSearchResults, &expectedSearchResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal expected results from file - error: %v", err)
	}

	if diff := cmp.Diff(expectedSearchResponse, searchResponse); diff != "" {
		return fmt.Errorf("expected response mismatch (-expected +actual):\n%s", diff)
	}

	return c.ErrorFeature.StepError()
}
