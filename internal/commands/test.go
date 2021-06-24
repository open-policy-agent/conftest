package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/open-policy-agent/conftest/internal/runner"
	"github.com/open-policy-agent/conftest/output"
	"github.com/open-policy-agent/conftest/parser"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const testDesc = `
This command tests your configuration files using the Open Policy Agent.

The test command expects one or more input files that will be evaluated 
against Open Policy Agent policies. Directories are also supported as valid
inputs. 

Policies are written in the Rego language. For more
information on how to write Rego policies, see the documentation:
https://www.openpolicyagent.org/docs/latest/policy-language/

The policy location defaults to the policy directory in the local folder.
The location can be overridden with the '--policy' flag, e.g.:

	$ conftest test --policy <my-directory> <input-file(s)/input-folder>

Some policies are dependant on external data. This data is loaded in seperatly 
from policies. The location of any data directory or file can be specified with 
the '--data' flag. If a directory is specified, it will be recursively searched for 
any data files. Right now any '.json' or '.yaml' file will be loaded in 
and made available in the Rego policies. Data will be made available in Rego based on 
the file path where the data was found. For example, if data is stored 
under 'policy/exceptions/my_data.yaml', and we execute the following command:

	$ conftest test --data policy <input-file>

The data is available under 'import data.exceptions'.

The test command supports the '--output' flag to specify the type, e.g.:

	$ conftest test -o table -p examples/kubernetes/policy examples/kubernetes/deployment.yaml

Which will return the following output:
+---------+----------------------------------+--------------------------------+
| RESULT  |               FILE               |            MESSAGE             |
+---------+----------------------------------+--------------------------------+
| success | examples/kubernetes/service.yaml |                                |
| warning | examples/kubernetes/service.yaml | Found service hello-kubernetes |
|         |                                  | but services are not allowed   |
+---------+----------------------------------+--------------------------------+

By default, it will use the regular stdout output. For a full list of available output types, see the of the '--output' flag.

The test command supports the '--update' flag to fetch the latest version of the policy at the given url.
It expects one or more urls to fetch the latest policies from, e.g.:

	$ conftest test --update opa.azurecr.io/test

See the pull command for more details on supported protocols for fetching policies.

When debugging policies it can be useful to use a more verbose policy evaluation output. By using the '--trace' flag
the output will include a detailed trace of how the policy was evaluated, e.g.

	$ conftest test --trace <input-file>
`

// TestRun stores the compiler and store for a test run.
type TestRun struct {
	Compiler *ast.Compiler
	Store    storage.Store
}

// NewTestCommand creates a new test command.
func NewTestCommand(ctx context.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "test <path> [path [...]]",
		Short: "Test your configuration files using Open Policy Agent",
		Long:  testDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flagNames := []string{"all-namespaces", "combine", "data", "fail-on-warn", "ignore", "namespace", "no-color", "no-fail", "suppress-exceptions", "output", "parser", "policy", "trace", "update"}
			for _, name := range flagNames {
				if err := viper.BindPFlag(name, cmd.Flags().Lookup(name)); err != nil {
					return fmt.Errorf("bind flag: %w", err)
				}
			}

			return nil
		},

		RunE: func(cmd *cobra.Command, fileList []string) error {
			if len(fileList) < 1 {
				cmd.Usage() //nolint
				return fmt.Errorf("missing required arguments")
			}

			var runner runner.TestRunner
			if err := viper.Unmarshal(&runner); err != nil {
				return fmt.Errorf("unmarshal parameters: %w", err)
			}

			results, err := runner.Run(ctx, fileList)
			if err != nil {
				return fmt.Errorf("running test: %w", err)
			}

			outputter := output.Get(runner.Output, output.Options{NoColor: runner.NoColor, SuppressExceptions: runner.SuppressExceptions, Tracing: runner.Trace})
			if err := outputter.Output(results); err != nil {
				return fmt.Errorf("output results: %w", err)
			}

			// When the no-fail parameter is set, there is no need to figure out the error code
			// as we always want to return zero.
			if runner.NoFail {
				return nil
			}

			var exitCode int
			if runner.FailOnWarn {
				exitCode = output.ExitCodeFailOnWarn(results)
			} else {
				exitCode = output.ExitCode(results)
			}

			os.Exit(exitCode)
			return nil
		},
	}

	cmd.Flags().Bool("fail-on-warn", false, "Return a non-zero exit code if warnings or errors are found")
	cmd.Flags().Bool("no-fail", false, "Return an exit code of zero even if a policy fails")
	cmd.Flags().Bool("no-color", false, "Disable color when printing")
	cmd.Flags().Bool("suppress-exceptions", false, "Do not include exceptions in output")
	cmd.Flags().Bool("all-namespaces", false, "Test policies found in all namespaces")

	cmd.Flags().BoolP("trace", "", false, "Enable more verbose trace output for Rego queries")
	cmd.Flags().BoolP("combine", "", false, "Combine all config files to be evaluated together")

	cmd.Flags().String("ignore", "", "A regex pattern which can be used for ignoring paths")
	cmd.Flags().String("parser", "", fmt.Sprintf("Parser to use to parse the configurations. Valid parsers: %s", parser.Parsers()))

	cmd.Flags().StringP("output", "o", output.OutputStandard, fmt.Sprintf("Output format for conftest results - valid options are: %s", output.Outputs()))

	cmd.Flags().StringSliceP("policy", "p", []string{"policy"}, "Path to the Rego policy files directory")
	cmd.Flags().StringSliceP("update", "u", []string{}, "A list of URLs can be provided to the update flag, which will download before the tests run")
	cmd.Flags().StringSliceP("namespace", "n", []string{"main"}, "Test policies in a specific namespace")
	cmd.Flags().StringSliceP("data", "d", []string{}, "A list of paths from which data for the rego policies will be recursively loaded")

	return &cmd
}
