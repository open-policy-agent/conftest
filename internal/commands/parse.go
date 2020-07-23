package commands

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/conftest/internal/runner"
	"github.com/open-policy-agent/conftest/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const parseDesc = `
This command prints the internal representation of your input files.

This can be useful in helping to write Rego policies. It is not always clear how 
your input file will be represented in the Rego policies. The type of the input is inferred
based on the file extension. If inference is not possible (e.g. due to the file coming from stdin)
the '--input' flag can be used to explicitly set the input type, e.g.:

	$ conftest parse --input toml <input-file(s)>

See the documentation of the '--input' flag for the supported input types.
`

// NewParseCommand creates a parse command.
// This command can be used for printing structured inputs from unstructured configuration inputs.
func NewParseCommand(ctx context.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "parse [file...]",
		Short: "Print out structured data from your input files",
		Long:  parseDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flagNames := []string{"input", "combine"}
			for _, name := range flagNames {
				if err := viper.BindPFlag(name, cmd.Flags().Lookup(name)); err != nil {
					return fmt.Errorf("bind flag: %w", err)
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, fileList []string) error {
			params := &runner.ParseParams{}
			viper.Unmarshal(params)
			configManager := &parser.ConfigManager{}
			runner := runner.ParseRunner{Params: params, ConfigManager: configManager}
			out, err := runner.Run(ctx, fileList)
			if err != nil {
				return fmt.Errorf("failed during parser process: %w", err)
			}

			fmt.Println(out)
			return nil
		},
	}

	cmd.Flags().BoolP("combine", "", false, "combine all given config files to be evaluated together")
	cmd.Flags().StringP("input", "i", "", fmt.Sprintf("input type for given source, especially useful when using conftest with stdin, valid options are: %s", parser.ValidInputs()))
	return &cmd
}
