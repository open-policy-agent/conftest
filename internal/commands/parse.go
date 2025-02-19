package commands

import (
	"encoding/json"
	"fmt"

	"github.com/open-policy-agent/conftest/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const parseDesc = `
This command prints the internal representation of your input files.

This can be useful in helping to write Rego policies. It is not always clear how 
your input file will be represented in the Rego policies. The type of the input is inferred
based on the file extension. If inference is not possible (e.g. due to the file coming from stdin)
the '--parser' flag can be used to explicitly set the parser, e.g.:

	$ conftest parse --parser toml <input-file(s)>

See the documentation of the '--parser' flag for the supported parsers.
`

// NewParseCommand creates a parse command.
// This command can be used for printing structured inputs from unstructured configuration inputs.
func NewParseCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "parse [file...]",
		Short: "Print out structured data from your input files",
		Long:  parseDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flagNames := []string{"parser", "combine"}
			for _, name := range flagNames {
				if err := viper.BindPFlag(name, cmd.Flags().Lookup(name)); err != nil {
					return fmt.Errorf("bind flag: %w", err)
				}
			}
			if len(args) == 0 {
				return fmt.Errorf("must supply path to at least one file")
			}

			return nil
		},
		RunE: func(_ *cobra.Command, files []string) error {
			var configurations map[string]any
			var err error
			if viper.GetString("parser") != "" {
				configurations, err = parser.ParseConfigurationsAs(files, viper.GetString("parser"))
			} else {
				configurations, err = parser.ParseConfigurations(files)
			}
			if err != nil {
				return fmt.Errorf("parse configurations: %w", err)
			}

			var output string
			if viper.GetBool("combine") {
				output, err = parser.FormatCombined(configurations)
			} else if len(configurations) == 1 {
				output, err = formatSingleJSON(configurations)
			} else {
				output, err = parser.FormatJSON(configurations)
			}
			if err != nil {
				return fmt.Errorf("format output: %w", err)
			}

			fmt.Println(output)
			return nil
		},
	}

	cmd.Flags().Bool("combine", false, "Combine all config files to be evaluated together")
	cmd.Flags().String("parser", "", fmt.Sprintf("Parser to use to parse the configurations. Valid parsers: %s", parser.Parsers()))

	return &cmd
}

func formatSingleJSON(configurations map[string]any) (string, error) {
	if len(configurations) != 1 {
		return "", fmt.Errorf("formatSingleJSON: only supports one configuration")
	}
	var config any
	for _, cfg := range configurations {
		config = cfg
	}
	marshaled, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}

	return string(marshaled), nil
}
