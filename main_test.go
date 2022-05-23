package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	componentTest "github.com/ONSdigital/dp-component-test"
	es710_steps "github.com/ONSdigital/dp-search-api/es710_features/steps"
	"github.com/ONSdigital/dp-search-api/features/steps"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

type LegacyComponentTest struct {
	AuthFeature *componentTest.AuthorizationFeature
}

type ComponentTest struct {
	AuthFeature *componentTest.AuthorizationFeature
}

func (c *ComponentTest) InitializeScenario(godogCtx *godog.ScenarioContext) {
	ctx := context.Background()

	apiComponent, err := es710_steps.SearchAPIComponent(c.AuthFeature)
	if err != nil {
		fmt.Println(ctx, "failed to create search api component - error: #{err}")
		os.Exit(1)
	}

	apiFeature := apiComponent.InitAPIFeature()

	godogCtx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		apiFeature.Reset()
		c.AuthFeature.Reset()
		apiComponent.Reset()
		return ctx, nil
	})

	godogCtx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if err := apiComponent.Close(); err != nil {
			fmt.Println(ctx, "error occurred while closing the api component - error: #{err}")
			os.Exit(1)
		}
		return ctx, nil
	})
	apiComponent.RegisterSteps(godogCtx)
	c.AuthFeature.RegisterSteps(godogCtx)
}

func (c *LegacyComponentTest) InitializeScenario(godogCtx *godog.ScenarioContext) {
	ctx := context.Background()

	apiComponent, err := steps.LegacySearchAPIComponent(c.AuthFeature)
	if err != nil {
		fmt.Println(ctx, "failed to create search api component - error: #{err}")
		os.Exit(1)
	}

	apiFeature := apiComponent.InitAPIFeature()

	godogCtx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		apiFeature.Reset()
		c.AuthFeature.Reset()
		apiComponent.Reset()
		return ctx, nil
	})

	godogCtx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if err := apiComponent.Close(); err != nil {
			fmt.Println(ctx, "error occurred while closing the api component - error: #{err}")
			os.Exit(1)
		}
		return ctx, nil
	})
	apiComponent.RegisterSteps(godogCtx)
	c.AuthFeature.RegisterSteps(godogCtx)
}

func (c *LegacyComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		c.AuthFeature = componentTest.NewAuthorizationFeature()
	})
	ctx.AfterSuite(func() {
		c.AuthFeature.Close()
	})
}

func (c *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		c.AuthFeature = componentTest.NewAuthorizationFeature()
	})
	ctx.AfterSuite(func() {
		c.AuthFeature.Close()
	})
}

func TestComponent(t *testing.T) {
	if *componentFlag {
		status := 0

		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Paths:  []string{"es710_features"},
			Format: "pretty",
		}

		c := &ComponentTest{}

		status = godog.TestSuite{
			Name:                 "legacy_component_tests",
			ScenarioInitializer:  c.InitializeScenario,
			TestSuiteInitializer: c.InitializeTestSuite,
			Options:              &opts,
		}.Run()

		fmt.Println("=================================")
		fmt.Printf("LegacyComponent test coverage: %.2f%%\n", testing.Coverage()*100)
		fmt.Println("=================================")

		if status != 0 {
			t.FailNow()
		}
	} else {
		t.Skip("component flag required to run component tests")
	}
}

func TestLegacyComponent(t *testing.T) {
	if *componentFlag {
		status := 0

		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Paths:  flag.Args(),
			Format: "pretty",
		}

		c := &LegacyComponentTest{}

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
	} else {
		t.Skip("component flag required to run component tests")
	}
}
