package steps

import (
	"encoding/json"
	"fmt"
	"github.com/cucumber/messages-go/v10"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

// HealthCheckTest represents a test healthcheck struct that mimics the real healthcheck struct
type HealthCheckTest struct {
	Status    string                  `json:"status"`
	Version   healthcheck.VersionInfo `json:"version"`
	Uptime    time.Duration           `json:"uptime"`
	StartTime time.Time               `json:"start_time"`
	Checks    []*Check                `json:"checks"`
}

// Check represents a health status of a registered app that mimics the real check struct
// As the component test needs to access fields that are not exported in the real struct
type Check struct {
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	StatusCode  int        `json:"status_code"`
	Message     string     `json:"message"`
	LastChecked *time.Time `json:"last_checked"`
	LastSuccess *time.Time `json:"last_success"`
	LastFailure *time.Time `json:"last_failure"`
}

// RegisterSteps registers the specific steps needed to do component tests for the search api
func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	c.APIFeature.RegisterSteps(ctx)
	ctx.Step(`^all of the downstream services are healthy$`, c.allOfTheDownstreamServicesAreHealthy)
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

func (c *Component) allOfTheDownstreamServicesAreHealthy() error {
	c.FakeElasticSearchAPI.setJSONResponseForGet("/search", 200)
	return nil
}

func (c *Component) oneOfTheDownstreamServicesIsWarning() error {
	c.FakeElasticSearchAPI.setJSONResponseForGet("/search", 429)
	return nil
}

func (c *Component) oneOfTheDownstreamServicesIsFailing() error {
	c.FakeElasticSearchAPI.setJSONResponseForGet("/search", 500)
	return nil
}

func (c *Component) iShouldReceiveTheFollowingHealthJSONResponse(expectedResponse *godog.DocString) error {
	var healthResponse, expectedHealth HealthCheckTest

	responseBody, err := ioutil.ReadAll(c.APIFeature.HttpResponse.Body)
	if err != nil {
		return fmt.Errorf("failed to read response of search controller component - error: %v", err)
	}

	err = json.Unmarshal(responseBody, &healthResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response of search controller component - error: %v", err)
	}

	err = json.Unmarshal([]byte(expectedResponse.Content), &expectedHealth)
	if err != nil {
		return fmt.Errorf("failed to unmarshal expected health response - error: %v", err)
	}

	c.validateHealthCheckResponse(healthResponse, expectedHealth)

	return c.ErrorFeature.StepError()
}

func (c *Component) validateHealthCheckResponse(healthResponse HealthCheckTest, expectedResponse HealthCheckTest) {
	maxExpectedStartTime := c.StartTime.Add((c.cfg.HealthCheckInterval + 1) * time.Second)

	assert.Equal(&c.ErrorFeature, expectedResponse.Status, healthResponse.Status)
	assert.True(&c.ErrorFeature, healthResponse.StartTime.After(c.StartTime))
	assert.True(&c.ErrorFeature, healthResponse.StartTime.Before(maxExpectedStartTime))
	assert.Greater(&c.ErrorFeature, healthResponse.Uptime.Seconds(), float64(0))

	c.validateHealthVersion(healthResponse.Version, expectedResponse.Version, maxExpectedStartTime)

	for i, checkResponse := range healthResponse.Checks {
		c.validateHealthCheck(checkResponse, expectedResponse.Checks[i])
	}
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
