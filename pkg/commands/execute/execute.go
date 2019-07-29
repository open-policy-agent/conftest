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
		Use:   "execute <file> [file...]",
		Short: "Execute Rego unit tests",

		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			compiler, err := policy.BuildCompiler(viper.GetString("policy"))
			if err != nil {
				log.G(ctx).Fatalf("Problem building rego compiler: %s", err)
			}

			runner := tester.NewRunner().
				SetCompiler(compiler)

			reporter := report.GetReporter(!viper.GetBool("color"))

			ch, err := runner.Run(ctx, compiler.Modules)
			if err != nil {
				log.G(ctx).Fatalf("Problem running rego tests: %s", err)
			}

			for result := range ch {
				if result.Error != nil {
					log.G(ctx).Fatalf("Test failed to execute: %s", err)
				}
				reportResult(reporter, result)
			}

			os.Exit(0)
		},
	}

	return cmd
}

func reportResult(reporter report.Reporter, result *tester.Result) {
	reporter.Report(report.Error, result.Location.File, result.Name)
}
