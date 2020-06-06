package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/open-policy-agent/conftest/internal/runners"
	"github.com/open-policy-agent/conftest/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const verifyDesc = `
This command executes Rego unit tests.

Any file with a '_test' postfix and '.rego' extension will be compiled and 
any Rego tests inside of them will be executed. For more information on how 
to write tests check out the Rego testing documentation: 
https://www.openpolicyagent.org/docs/latest/policy-testing/.

The policy location defaults to the policy directory in the local folder.
The location can be overridden with the '--policy' flag, e.g.:

	$ conftest verify --policy <my-directory>

Some policies are dependant on external data. This data is loaded in seperatly 
from policies. The location of any data directory or file can be specified with 
the '--data' flag. If a directory is specified, it will be recursively searched for 
any data files. Right now any '.json' or '.yaml' file will be loaded in 
and made available in the Rego policies. Data will be made available in Rego based on 
the file path where the data was found. For example, if data is stored 
under 'policy/exceptions/my_data.yaml', and we execute the following command:

	$ conftest verify --data policy

The data is available under 'import data.exceptions'.

As with the test command, verify supports the '--output' flag to specify the type, e.g.:

	$ conftest verify --output json

For a full list of available output types, see the of the '--output' flag.

When debugging policies it can be useful to use a more verbose policy evaluation output. By using the '--trace' flag
the output will include a detailed trace of how the policy was evaluated, e.g.

	$ conftest verify --trace <input-file>
`

// NewVerifyCommand creates a new verify command which allows users
// to validate their rego unit tests
func NewVerifyCommand(ctx context.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "verify",
		Short: "Verify Rego unit tests",
		Long:  verifyDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flagNames := []string{"output", "trace", "data"}
			for _, name := range flagNames {
				if err := viper.BindPFlag(name, cmd.Flags().Lookup(name)); err != nil {
					return fmt.Errorf("bind flag: %w", err)
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			outFmt := viper.GetString("output")
			color := !viper.GetBool("no-color")
			outputManager := output.GetOutputManager(outFmt, color)
			runner := &runners.VerifyRunner{}
			err := viper.Unmarshal(runner)
			if err != nil {
				return fmt.Errorf("unmarshal parameters: %w", err)
			}

			results, err := runner.Run(ctx)
			if err != nil {
				return fmt.Errorf("running verification: %w", err)
			}

			var failures int
			for _, result := range results {
				if err := outputManager.Put(result); err != nil {
					return fmt.Errorf("put result: %w", err)
				}

				if output.IsResultFailure(result, viper.GetBool("fail-on-warn")) {
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

	cmd.Flags().StringP("output", "o", "", fmt.Sprintf("output format for conftest results - valid options are: %s", output.ValidOutputs()))
	cmd.Flags().BoolP("trace", "", false, "enable more verbose trace output for rego queries")
	cmd.Flags().StringSliceP("data", "d", []string{}, "A list of paths from which data for the rego policies will be recursively loaded")

	return &cmd
}
