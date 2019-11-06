package parse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/instrumenta/conftest/commands/test"
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
			out, err := parseInput(ctx, fileList)
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

func parseInput(ctx context.Context, fileList []string) ([]byte, error) {
	configurations, err := test.GetConfigurations(ctx, fileList)
	if err != nil {
		return nil, fmt.Errorf("calling the parser method: %w", err)
	}

	var bundle []byte
	for filename, config := range configurations {
		out, err := json.Marshal(config)
		if err != nil {
			return nil, fmt.Errorf("marshal output to json: %w", err)
		}

		var prettyJSON bytes.Buffer
		if err = json.Indent(&prettyJSON, out, "", "\t"); err != nil {
			return nil, fmt.Errorf("indentation: %w", err)
		}

		if _, err := prettyJSON.WriteString("\n"); err != nil {
			return nil, fmt.Errorf("adding line break: %w", err)
		}

		filename = filename + "\n"
		bundle = append(bundle, filename...)
		bundle = append(bundle, prettyJSON.Bytes()...)
	}

	return bundle, nil
}
