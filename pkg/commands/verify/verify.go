package verify

import (
	"path/filepath"
	"fmt"
	"context"
	"os"

	"github.com/instrumenta/conftest/pkg/constants"
	"github.com/instrumenta/conftest/pkg/policy"
	"github.com/instrumenta/conftest/pkg/commands/test"

	"github.com/containerd/containerd/log"
	"github.com/open-policy-agent/opa/tester"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewVerifyCommand(getOutputManager func() test.OutputManager) *cobra.Command {
	ctx := context.Background()
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify Rego unit tests",
		Version: fmt.Sprintf("Version: %s\nCommit: %s\nDate: %s\n", constants.Version, constants.Commit, constants.Date),

		Run: func(cmd *cobra.Command, args []string) {
			out := getOutputManager()

			policyPath := viper.GetString("policy")
			compiler, err := policy.BuildCompiler(policyPath)
			if err != nil {
				log.G(ctx).Fatalf("Problem building rego compiler: %s", err)
			}

			runner := tester.NewRunner().
				SetCompiler(compiler)

			ch, err := runner.Run(ctx, compiler.Modules)
			if err != nil {
				log.G(ctx).Fatalf("Problem running rego tests: %s", err)
			}

			results := getResults(ctx, ch)

			for result := range results {
				msg := fmt.Errorf("%s", result.Msg)
				fileName := filepath.Join(policyPath, result.FileName)
				if result.Fail {
					err = out.Put(fileName, test.CheckResult{Failures:[]error{msg}})
				} else {
					err = out.Put(fileName, test.CheckResult{Successes:[]error{msg}})
				}

				if err != nil {
					log.G(ctx).Fatalf("Problem writing to output: %s", err)
				}
			}

			err = out.Flush()
			if err != nil {
				log.G(ctx).Fatal(err)
			}

			os.Exit(0)
		},
	}

	cmd.Flags().StringP("output", "o", "", fmt.Sprintf("output format for conftest results - valid options are: %s", test.ValidOutputs()))
	err := viper.BindPFlag("output", cmd.Flags().Lookup("output"))
	if err != nil {
		log.G(ctx).Fatal("Failed to bind argument:", err)
	}

	return cmd
}

type report struct {
	FileName string
	Msg      string
	Fail 	bool
}

func getResults(ctx context.Context, in <-chan *tester.Result) <-chan report {
	results := make(chan report)
	go func() {
		defer close(results)
		for result := range in {
			if result.Error != nil {
				log.G(ctx).Fatalf("Test failed to execute: %s", result.Error)
			}
			msg := result.Package+"."+result.Name
			results <- report{FileName: result.Location.File, Msg: msg, Fail: result.Fail}
		}
	}()

	return results
}
