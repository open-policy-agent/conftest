package parse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/instrumenta/conftest/commands/test"
	"github.com/instrumenta/conftest/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//NewParseCommand creates a parse command.
//This command can be used for printing structured inputs from unstructured configuration inputs.
//Can be used with '--input' or '-i' flag.
func NewParseCommand(ctx context.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "parse <filename>",
		Short: "Print out structured data from your input file",
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
			print(out)
			return nil
		},
	}

	cmd.Flags().StringP("input", "i", "", fmt.Sprintf("input type for given source, especially useful when using conftest with stdin, valid options are: %s", parser.ValidInputs()))
	return &cmd
}

func parseInput(ctx context.Context, fileList []string) ([]byte, error) {
	configurations, err := test.GetConfigurations(ctx, fileList)
	var bundle []byte
	if err != nil {
		return nil, fmt.Errorf("Your inputs could not be parsed : %w", err)
	}
	for filename, config := range configurations {
		filename = filename + "\n"
		var prettyJSON bytes.Buffer
		out, err := json.Marshal(config)
		if err != nil {
			return nil, fmt.Errorf("Error converting config to JSON out: %s", err)
		}
		err = json.Indent(&prettyJSON, out, "", "\t")
		if err != nil {
			return nil, fmt.Errorf("Error indenting JSON output: %s", err)
		}
		bundle = append(bundle, filename...)
		bundle = append(bundle, prettyJSON.Bytes()...)
	}
	return bundle, nil
}

//Will be replaced with OutputManager
func print(output []byte) {
	os.Stdout.Write(output)
}
