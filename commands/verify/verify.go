package verify

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/instrumenta/conftest/commands/test"
	"github.com/instrumenta/conftest/policy"
	"github.com/open-policy-agent/opa/tester"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewVerifyCommand creates a new verify command which allows users
// to validate their rego unit tests
func NewVerifyCommand(ctx context.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "verify",
		Short: "Verify Rego unit tests",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := viper.BindPFlag("output", cmd.Flags().Lookup("output"))
			if err != nil {
				return fmt.Errorf("bind output: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			outputManager := test.GetOutputManager()
			policyPath := viper.GetString("policy")

			results, err := runVerification(ctx, policyPath)
			if err != nil {
				return fmt.Errorf("running verification: %w", err)
			}

			for _, result := range results {
				if err := outputManager.Put(result.FileName, result); err != nil {
					return fmt.Errorf("put result: %w", err)
				}
			}

			if err := outputManager.Flush(); err != nil {
				return fmt.Errorf("flushing output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "", fmt.Sprintf("output format for conftest results - valid options are: %s", test.ValidOutputs()))

	return &cmd
}

func runVerification(ctx context.Context, path string) ([]test.CheckResult, error) {
	regoFiles, err := policy.ReadFilesWithTests(path)
	if err != nil {
		return nil, fmt.Errorf("read rego test files: %s", err)
	}

	compiler, err := policy.BuildCompiler(regoFiles)
	if err != nil {
		return nil, fmt.Errorf("build compiler: %w", err)
	}

	runner := tester.NewRunner().SetCompiler(compiler)
	ch, err := runner.RunTests(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("running tests: %w", err)
	}

	var results []test.CheckResult
	for result := range ch {
		msg := fmt.Errorf("%s", result.Package+"."+result.Name)
		fileName := filepath.Join(path, result.Location.File)

		var failure []error
		var success []error

		if result.Fail {
			failure = []error{msg}
		} else {
			success = []error{msg}
		}

		result := test.CheckResult{
			FileName:  fileName,
			Successes: success,
			Failures:  failure,
		}

		results = append(results, result)
	}

	return results, nil
}
