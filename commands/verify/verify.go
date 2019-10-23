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

func NewVerifyCommand() *cobra.Command {
	ctx := context.Background()
	outputManager := test.GetOutputManager()

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
			policyPath := viper.GetString("policy")

			success, failure, err := RunVerification(ctx, policyPath)
			if err != nil {
				return fmt.Errorf("running verification: %w", err)
			}

			fmt.Println(success)
			fmt.Println(failure)

			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "", fmt.Sprintf("output format for conftest results - valid options are: %s", test.ValidOutputs()))

	return &cmd
}

// RunVerification runs the verify command
func RunVerification(ctx context.Context, path string) ([]test.CheckResult, error) {
	compiler, err := policy.BuildCompiler(path, true)
	if err != nil {
		return nil, fmt.Errorf("build compiler: %w", err)
	}

	runner := tester.NewRunner().SetCompiler(compiler)
	ch, err := runner.RunTests(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("running tests: %w", err)
	}

	var failures []error
	var successes []error
	for result := range ch {
		msg := fmt.Errorf("%s", result.Package+"."+result.Name)
		fileName := filepath.Join(path, result.Location.File)

		if result.Fail {
			failures = append(failures, msg)
		} else {
			successes = append(successes, msg)
		}

		result := test.CheckResult{
			FileName:  fileName,
			Successes: successes,
			Failures:  failures,
		}

		/// SPLIT OUT CONCERNS BETWEEN RESULT AND OUTPUT

	}

	return success, failure, nil
}
