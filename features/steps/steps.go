package steps

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ONSdigital/dp-search-api/models"
	"github.com/cucumber/godog"
	"github.com/google/go-cmp/cmp"
)

// RegisterSteps registers the specific steps needed to do component tests for the search api
func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	c.APIFeature.RegisterSteps(ctx)
	ctx.Step(`^elasticsearch is healthy`, c.elasticSearchIsHealthy)
	ctx.Step(`^elasticsearch returns one item in search response$`, c.es7xSuccessfullyReturnSingleSearchResult)
	ctx.Step(`^elasticsearch returns multiple items in search response with topics filter$`, c.es7xSuccessfullyReturnMultipleSearchResultWithTopicFilter)
	ctx.Step(`^elasticsearch returns multiple items in search response with datasetIDs filter$`, c.es7xSuccessfullyReturnMultipleSearchResultWithDatasetIDsFilter)
	ctx.Step(`^elasticsearch returns multiple items with distinct topic count in search response$`, c.es7xSuccessfullyReturnMultipleSearchResultWithDistinctTopicCount)
	ctx.Step(`^elasticsearch returns one item in search response with topics filter$`, c.es7xSuccessfullyReturnSingleSearchResultWithTopicFilter)
	ctx.Step(`^elasticsearch returns one item in search response with datasetIDs filter$`, c.es7xSuccessfullyReturnSingleSearchResultWithDatasetIDsFilter)
	ctx.Step(`^elasticsearch returns one item in search/release response$`, c.successfullyReturnSingleSearchReleaseResult)
	ctx.Step(`^elasticsearch returns one item in search/uris response$`, c.es7xSuccessfullyReturnSingleSearchResult)
	ctx.Step(`^the response body is the same as the json in "([^"]*)"$`, c.iShouldReceiveTheFollowingSearchResponsefromes7x)
	ctx.Step(`^elasticsearch returns multiple items in search response$`, c.es7xSuccessfullyReturnMultipleSearchResults)
	ctx.Step(`^elasticsearch returns multiple items in search/release response$`, c.successfullyReturnMultipleSearchReleaseResults)
	ctx.Step(`^elasticsearch returns zero items in search response$`, c.es7xSuccessfullyReturnNoSearchResults)
	ctx.Step(`^elasticsearch returns zero items in search/release response$`, c.successfullyReturnNoSearchReleaseResults)
	ctx.Step(`^elasticsearch returns internal server error$`, c.es7xFailureInternalServerError)
}

// elasticSearchIsHealthy generates a mocked healthy response for elasticsearch healthecheck
func (c *Component) elasticSearchIsHealthy() error {
	const res = `{"cluster_name": "docker-cluster", "status": "green"}`
	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().
		Get("/_cluster/health").
		Reply(http.StatusOK).
		BodyString(res)
	return nil
}

func (c *Component) es7xSuccessfullyReturnSingleSearchResult() error {
	body, err := os.ReadFile("./features/testdata/es_single_search_result.json")
	if err != nil {
		return err
	}

	countbody, err := os.ReadFile("./features/testdata/es_single_count_result.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Get("/elasticsearch/_msearch").Reply(200).Body(body)
	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Post("/elasticsearch/_count").Reply(200).Body(countbody)

	return nil
}

func (c *Component) es7xSuccessfullyReturnMultipleSearchResultWithTopicFilter() error {
	body, err := os.ReadFile("./features/testdata/es_mulitple_search_topics_results.json")
	if err != nil {
		return err
	}
	countbody, err := os.ReadFile("./features/testdata/es_single_count_result.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Get("/elasticsearch/_msearch").Reply(200).Body(body)
	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Post("/elasticsearch/_count").Reply(200).Body(countbody)

	return nil
}

func (c *Component) es7xSuccessfullyReturnMultipleSearchResultWithDatasetIDsFilter() error {
	body, err := os.ReadFile("./features/testdata/es_multiple_search_dataset_results.json")
	if err != nil {
		return err
	}
	countbody, err := os.ReadFile("./features/testdata/es_single_count_result.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Get("/elasticsearch/_msearch").Reply(200).Body(body)
	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Post("/elasticsearch/_count").Reply(200).Body(countbody)

	return nil
}

func (c *Component) es7xSuccessfullyReturnMultipleSearchResultWithDistinctTopicCount() error {
	body, err := os.ReadFile("./features/testdata/es_mulitple_search_topics_results_with_distinct_topics_count.json")
	if err != nil {
		return err
	}
	countbody, err := os.ReadFile("./features/testdata/es_single_count_result.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Get("/elasticsearch/_msearch").Reply(200).Body(body)
	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Post("/elasticsearch/_count").Reply(200).Body(countbody)

	return nil
}

func (c *Component) es7xSuccessfullyReturnSingleSearchResultWithTopicFilter() error {
	body, err := os.ReadFile("./features/testdata/es_single_search_topic_results.json")
	if err != nil {
		return err
	}
	countbody, err := os.ReadFile("./features/testdata/es_single_count_result.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Get("/elasticsearch/_msearch").Reply(200).Body(body)
	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Post("/elasticsearch/_count").Reply(200).Body(countbody)

	return nil
}

func (c *Component) es7xSuccessfullyReturnSingleSearchResultWithDatasetIDsFilter() error {
	body, err := os.ReadFile("./features/testdata/es_single_search_dataset_results.json")
	if err != nil {
		return err
	}
	countbody, err := os.ReadFile("./features/testdata/es_single_count_result.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Get("/elasticsearch/_msearch").Reply(200).Body(body)
	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Post("/elasticsearch/_count").Reply(200).Body(countbody)

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
	countbody, err := os.ReadFile("./features/testdata/es_single_count_result.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Get("/elasticsearch/_msearch").Reply(200).Body(body)
	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Post("/elasticsearch/_count").Reply(200).Body(countbody)

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
	countbody, err := os.ReadFile("./features/testdata/es_single_count_result.json")
	if err != nil {
		return err
	}

	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Get("/elasticsearch/_msearch").Reply(200).Body(body)
	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Post("/elasticsearch/_count").Reply(200).Body(countbody)

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

	responseBody, err := io.ReadAll(c.APIFeature.HTTPResponse.Body)
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
	c.FakeElasticSearchAPI.fakeHTTP.NewHandler().Post("/elasticsearch/_count").Reply(200)

	return nil
}
