package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/ONSdigital/dp-search-api/es710_features/steps"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

type Component struct {
	*steps.Component
}

func (c *Component) InitializeScenario(godogCtx *godog.ScenarioContext) {
	godogCtx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		c.Reset()
		return ctx, nil
	})
}

func (c *Component) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	c.RegisterSteps(ctx.ScenarioContext())

	ctx.AfterSuite(func() {
		_ = c.Close()
	})
}

func TestComponent(t *testing.T) {
	if !*componentFlag {
		t.Skip("component flag required to run component tests")
	}

	var (
		status int
		c      = Component{Component: steps.TestComponent(t)}
		opts   = godog.Options{
			Output:   colors.Colored(os.Stdout),
			Paths:    []string{"es710_features"},
			Format:   "pretty",
			TestingT: t,
		}
	)

	status = godog.TestSuite{
		Name:                 "component_tests",
		ScenarioInitializer:  c.InitializeScenario,
		TestSuiteInitializer: c.InitializeTestSuite,
		Options:              &opts,
	}.Run()

	fmt.Println("=================================")
	fmt.Printf("Component test coverage: %.2f%%\n", testing.Coverage()*100)
	fmt.Println("=================================")

	if status != 0 {
		t.FailNow()
	}
}
