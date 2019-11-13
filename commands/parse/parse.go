package parse

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

// NewParseCommand creates a parse command.
// This command can be used for printing structured inputs from unstructured configuration inputs.
func NewParseCommand(ctx context.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "parse [file...]",
		Short: "Print out structured data from your input files",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := viper.BindPFlag("input", cmd.Flags().Lookup("input"))
			if err != nil {
				return fmt.Errorf("failed to bind argument: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, fileList []string) error {
			out, err := parseInput(ctx, viper.GetString("input"), fileList)
			if err != nil {
				return fmt.Errorf("failed during parser process: %w", err)
			}

			fmt.Println(out)
			return nil
		},
	}

	cmd.Flags().StringP("input", "i", "", fmt.Sprintf("input type for given source, especially useful when using conftest with stdin, valid options are: %s", parser.ValidInputs()))
	return &cmd
}

func parseInput(ctx context.Context, input string, fileList []string) (string, error) {
	configurations, err := parser.GetConfigurations(ctx, input, fileList)
	if err != nil {
		return "", fmt.Errorf("calling the parser method: %w", err)
	}

	var output string
	for filename, config := range configurations {
		out, err := json.Marshal(config)
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

		output = output + filename + "\n"
		output = output + prettyJSON.String()
		output = strings.Replace(output, "\\r", "", -1)
	}

	return output, nil
}
