package execute

import (
	"context"
	"os"

	"github.com/containerd/containerd/log"
	"github.com/instrumenta/conftest/pkg/policy"
	"github.com/instrumenta/conftest/pkg/report"
	"github.com/open-policy-agent/opa/tester"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewExecuteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute Rego unit tests",

		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			compiler, err := policy.BuildCompiler(viper.GetString("policy"))
			if err != nil {
				log.G(ctx).Fatalf("Problem building rego compiler: %s", err)
			}

			runner := tester.NewRunner().
				SetCompiler(compiler)

			reporter := report.GetReporter(viper.GetString("output"), !viper.GetBool("color"))

			ch, err := runner.Run(ctx, compiler.Modules)
			if err != nil {
				log.G(ctx).Fatalf("Problem running rego tests: %s", err)
			}

			results := getResults(ctx, ch)

			err = reporter.Report(results)
			if err != nil {
				log.G(ctx).Fatalf("Problem writing to output: %s", err)
			}

			os.Exit(0)
		},
	}

	return cmd
}

func getResults(ctx context.Context, in <-chan *tester.Result) <-chan report.Result {
	results := make(chan report.Result)
	go func() {
		defer close(results)
		for result := range in {
			if result.Error != nil {
				log.G(ctx).Fatalf("Test failed to execute: %s", result.Error)
			}
			results <- report.Result{report.Error, result.Location.File, result.Name}
		}
	}()

	return results
}
