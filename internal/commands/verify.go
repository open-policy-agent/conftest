package commands

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/instrumenta/conftest/policy"
	"github.com/open-policy-agent/opa/tester"
	"github.com/open-policy-agent/opa/topdown"
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
			flagNames := []string{"output", "trace"}
			for _, name := range flagNames {
				if err := viper.BindPFlag(name, cmd.Flags().Lookup(name)); err != nil {
					return fmt.Errorf("bind flag: %w", err)
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			outputManager := GetOutputManager()
			policyPath := viper.GetString("policy")
			trace := viper.GetBool("trace")

			results, err := runVerification(ctx, policyPath, trace)
			if err != nil {
				return fmt.Errorf("running verification: %w", err)
			}

			var failures int
			for _, result := range results {
				if err := outputManager.Put(result.FileName, result); err != nil {
					return fmt.Errorf("put result: %w", err)
				}

				if isResultFailure(result) {
					failures++
				}
			}

			if err := outputManager.Flush(); err != nil {
				return fmt.Errorf("flushing output: %w", err)
			}

			if failures > 0 {
				os.Exit(1)
			}

			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "", fmt.Sprintf("output format for conftest results - valid options are: %s", ValidOutputs()))
	cmd.Flags().BoolP("trace", "", false, "enable more verbose trace output for rego queries")

	return &cmd
}

func runVerification(ctx context.Context, path string, trace bool) ([]CheckResult, error) {
	regoFiles, err := policy.ReadFilesWithTests(path)
	if err != nil {
		return nil, fmt.Errorf("read rego test files: %s", err)
	}

	compiler, err := policy.BuildCompiler(regoFiles)
	if err != nil {
		return nil, fmt.Errorf("build compiler: %w", err)
	}

	runner := tester.NewRunner().SetCompiler(compiler).EnableTracing(trace)
	ch, err := runner.RunTests(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("running tests: %w", err)
	}

	var results []CheckResult
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

		buf := new(bytes.Buffer)
		topdown.PrettyTrace(buf, result.Trace)
		var traces []error
		for _, line := range strings.Split(buf.String(), "\n") {
			if len(line) > 0 {
				traces = append(traces, errors.New(line))
			}
		}

		result := CheckResult{
			FileName:  fileName,
			Successes: success,
			Failures:  failure,
			Traces:    traces,
		}

		results = append(results, result)
	}

	return results, nil
}
