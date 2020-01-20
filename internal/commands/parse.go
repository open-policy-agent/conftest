package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/instrumenta/conftest/parser"
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
			out, err := parseInput(ctx, viper.GetString("input"), viper.GetBool("combine"), fileList)
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

func parseInput(ctx context.Context, input string, combine bool, fileList []string) (string, error) {
	configurations, err := parser.GetConfigurations(ctx, input, fileList)
	if err != nil {
		return "", fmt.Errorf("calling the parser method: %w", err)
	}

	parsedConfigurations, err := parseConfigurations(configurations, combine)
	if err != nil {
		return "", fmt.Errorf("parsing configs: %w", err)
	}

	return parsedConfigurations, nil

}

func parseConfigurations(configurations map[string]interface{}, combine bool) (string, error) {
	var output string
	if combine {
		content, err := marshal(configurations)
		if err != nil {
			return "", fmt.Errorf("marshal output to json: %w", err)
		}

		output = strings.Replace(output+"\n"+content, "\\r", "", -1)
	} else {
		for filename, config := range configurations {
			content, err := marshal(config)
			if err != nil {
				return "", fmt.Errorf("marshal output to json: %w", err)
			}

			output = strings.Replace(output+filename+"\n"+content, "\\r", "", -1)
		}
	}

	return output, nil
}

func marshal(in interface{}) (string, error) {
	out, err := json.Marshal(in)
	if err != nil {
		return "", fmt.Errorf("marshal output to json: %w", err)
	}

	var prettyJSON bytes.Buffer
	if err = json.Indent(&prettyJSON, out, "", "\t"); err != nil {
		return "", fmt.Errorf("indentation: %w", err)
	}

	if _, err := prettyJSON.WriteString("\n"); err != nil {
		return "", fmt.Errorf("adding line break: %w", err)
	}

	return prettyJSON.String(), nil
}
