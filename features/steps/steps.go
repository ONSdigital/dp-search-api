package steps

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ONSdigital/dp-search-api/models"
	"github.com/cucumber/godog"
	"github.com/google/go-cmp/cmp"
)

// RegisterSteps registers the specific steps needed to do component tests for the search api
func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	c.APIFeature.RegisterSteps(ctx)
	ctx.Step(`^elasticsearch returns one item in search response$`, c.es7xSuccessfullyReturnSingleSearchResult)
	ctx.Step(`^elasticsearch returns one item in search/release response$`, c.successfullyReturnSingleSearchReleaseResult)
	ctx.Step(`^the response body is the same as the json in "([^"]*)"$`, c.iShouldReceiveTheFollowingSearchResponsefromes7x)
	ctx.Step(`^elasticsearch returns multiple items in search response$`, c.es7xSuccessfullyReturnMultipleSearchResults)
	ctx.Step(`^elasticsearch returns multiple items in search/release response$`, c.successfullyReturnMultipleSearchReleaseResults)
	ctx.Step(`^elasticsearch returns zero items in search response$`, c.es7xSuccessfullyReturnNoSearchResults)
	ctx.Step(`^elasticsearch returns zero items in search/release response$`, c.successfullyReturnNoSearchReleaseResults)
	ctx.Step(`^elasticsearch returns internal server error$`, c.es7xFailureInternalServerError)
}

func (c *Component) es7xSuccessfullyReturnSingleSearchResult() error {
	body, err := os.ReadFile("./features/testdata/es_single_search_result.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Get("/elasticsearch/_msearch").Reply(200).Body(body)

	return nil
}

func (c *Component) successfullyReturnSingleSearchReleaseResult() error {
	body, err := os.ReadFile("./features/testdata/es_single_search_release_result.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Get("/elasticsearch/_msearch").Reply(200).Body(body)

	return nil
}

func (c *Component) es7xSuccessfullyReturnMultipleSearchResults() error {
	body, err := os.ReadFile("./features/testdata/es_mulitple_search_results.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Get("/elasticsearch/_msearch").Reply(200).Body(body)

	return nil
}

func (c *Component) successfullyReturnMultipleSearchReleaseResults() error {
	body, err := os.ReadFile("./features/testdata/es_mulitple_search_release_results.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Get("/elasticsearch/_msearch").Reply(200).Body(body)

	return nil
}

func (c *Component) es7xSuccessfullyReturnNoSearchResults() error {
	body, err := os.ReadFile("./features/testdata/es_zero_search_results.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Get("/elasticsearch/_msearch").Reply(200).Body(body)

	return nil
}

func (c *Component) successfullyReturnNoSearchReleaseResults() error {
	body, err := os.ReadFile("./features/testdata/es_zero_search_release_results.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Get("/elasticsearch/_msearch").Reply(200).Body(body)

	return nil
}

func (c *Component) iShouldReceiveTheFollowingSearchResponsefromes7x(expectedJSONFile string) error {
	var searchResponse, expectedSearchResponse models.SearchResponse

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

func (c *Component) es7xFailureInternalServerError() error {
	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Get("/elasticsearch/_msearch").Reply(500)

	return nil
}
