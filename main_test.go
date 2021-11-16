package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/ONSdigital/dp-search-api/features/steps"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

func InitializeScenario(godogCtx *godog.ScenarioContext) {
	ctx := context.Background()

	apiComponent, err := steps.NewSearchAPIComponent()
	if err != nil {
		fmt.Println(ctx, "failed to create search api component - error: #{err}")
		os.Exit(1)
	}

	apiFeature := apiComponent.InitAPIFeature()

	godogCtx.BeforeScenario(func(*godog.Scenario) {
		apiFeature.Reset()
		apiComponent.Reset()
	})

	apiComponent.RegisterSteps(godogCtx)

	godogCtx.AfterScenario(func(*godog.Scenario, error) {
		if err := apiComponent.Close(); err != nil {
			fmt.Println(ctx, "error occurred while closing the api component - error: #{err}")
			os.Exit(1)
		}
	})
}

func TestMain(t *testing.T) {
	if *componentFlag {
		status := 0

		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Paths:  flag.Args(),
			Format: "pretty",
		}

		status = godog.TestSuite{
			Name:                "component_tests",
			ScenarioInitializer: InitializeScenario,
			Options:             &opts,
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
