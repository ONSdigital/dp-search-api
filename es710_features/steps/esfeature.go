package steps

import (
	"os"
	"testing"

	"github.com/ONSdigital/dp-search-api/config"

	"github.com/cucumber/godog"
	"github.com/maxcnunes/httpfake"
)

type ESDependency struct {
	esServer *httpfake.HTTPFake
}

func NewESDependency(t *testing.T, cfg *config.Config) *ESDependency {
	es := &ESDependency{esServer: httpfake.New(httpfake.WithTesting(t))}
	cfg.ElasticSearchAPIURL = es.esServer.ResolveURL("")

	return es
}

func (es *ESDependency) Reset() { es.esServer.Reset() }
func (es *ESDependency) Close() { es.esServer.Close() }

func (es *ESDependency) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^elasticsearch7x returns one item in search response$`, es.es7xSuccessfullyReturnSingleSearchResult)
	ctx.Step(`^elasticsearch returns one item in search/release response$`, es.successfullyReturnSingleSearchReleaseResult)
	ctx.Step(`^elasticsearch7x returns multiple items in search response$`, es.es7xSuccessfullyReturnMultipleSearchResults)
	ctx.Step(`^elasticsearch returns multiple items in search/release response$`, es.successfullyReturnMultipleSearchReleaseResults)
	ctx.Step(`^elasticsearch7x returns zero items in search response$`, es.es7xSuccessfullyReturnNoSearchResults)
	ctx.Step(`^elasticsearch returns zero items in search/release response$`, es.successfullyReturnNoSearchReleaseResults)
	ctx.Step(`^elasticsearch7x returns internal server error$`, es.es7xFailureInternalServerError)
}

func (es *ESDependency) es7xSuccessfullyReturnSingleSearchResult() error {
	body, err := os.ReadFile("./es710_features/testdata/es_single_search_result.json")
	if err != nil {
		return err
	}

	es.esServer.NewHandler().Get("/_msearch").Reply(200).Body(body)

	return nil
}

func (es *ESDependency) successfullyReturnSingleSearchReleaseResult() error {
	body, err := os.ReadFile("./es710_features/testdata/es_single_search_release_result.json")
	if err != nil {
		return err
	}

	es.esServer.NewHandler().Get("/_msearch").Reply(200).Body(body)

	return nil
}

func (es *ESDependency) es7xSuccessfullyReturnMultipleSearchResults() error {
	body, err := os.ReadFile("./es710_features/testdata/es_mulitple_search_results.json")
	if err != nil {
		return err
	}

	es.esServer.NewHandler().Get("/_msearch").Reply(200).Body(body)

	return nil
}

func (es *ESDependency) successfullyReturnMultipleSearchReleaseResults() error {
	body, err := os.ReadFile("./es710_features/testdata/es_mulitple_search_release_results.json")
	if err != nil {
		return err
	}

	es.esServer.NewHandler().Get("/_msearch").Reply(200).Body(body)

	return nil
}

func (es *ESDependency) es7xSuccessfullyReturnNoSearchResults() error {
	body, err := os.ReadFile("./es710_features/testdata/es_zero_search_results.json")
	if err != nil {
		return err
	}

	es.esServer.NewHandler().Get("/_msearch").Reply(200).Body(body)

	return nil
}

func (es *ESDependency) successfullyReturnNoSearchReleaseResults() error {
	body, err := os.ReadFile("./es710_features/testdata/es_zero_search_release_results.json")
	if err != nil {
		return err
	}

	es.esServer.NewHandler().Get("/_msearch").Reply(200).Body(body)

	return nil
}

func (es *ESDependency) es7xFailureInternalServerError() error {
	es.esServer.NewHandler().Get("/_msearch").Reply(500)

	return nil
}
