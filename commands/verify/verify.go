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
func RunVerification(ctx context.Context, path string) ([]string, []string, error) {
	compiler, err := policy.BuildCompiler(path, true)
	if err != nil {
		return nil, nil, fmt.Errorf("build compiler: %w", err)
	}

	runner := tester.NewRunner().SetCompiler(compiler)
	ch, err := runner.RunTests(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("running tests: %w", err)
	}

	var success []string
	var failure []string
	for result := range ch {
		msg := fmt.Errorf("%s", result.Package+"."+result.Name)
		fileName := filepath.Join(path, result.Location.File)

		if result.Fail {
			failure = append(failure, fmt.Sprintf("[fail] file %s with message %s\n", fileName, msg))
		} else {
			success = append(success, fmt.Sprintf("[success] file %s with message %s\n", fileName, msg))
		}
	}

	return success, failure, nil
}
