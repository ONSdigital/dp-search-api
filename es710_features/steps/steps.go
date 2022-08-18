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
	c.AuthFeature.RegisterSteps(ctx)
	c.ESDependencyInjection.RegisterSteps(ctx)

	ctx.Step(`^the response body is the same as the json in "([^"]*)"$`, c.iShouldReceiveTheFollowingSearchResponsefromES7x)
}

func (c *Component) iShouldReceiveTheFollowingSearchResponsefromES7x(expectedJSONFile string) error {
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
