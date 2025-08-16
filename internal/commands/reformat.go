package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/open-policy-agent/conftest/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const reformatDesc = `
This command reformats conftest JSON output to different formats.

The reformat command takes JSON output from conftest (typically from stdin or a file)
and converts it to other supported output formats. This allows for decoupling test
execution from formatting, enabling multiple output formats from a single test run.

Usage examples:

	# Convert JSON output to table format
	$ conftest test --output json config.yaml | conftest reformat --output table

	# Convert JSON file to JUnit format
	$ conftest reformat --output junit results.json

	# Convert JSON to multiple formats
	$ conftest test --output json config.yaml > results.json
	$ conftest reformat --output table results.json
	$ conftest reformat --output junit results.json

Supported output formats: %s`

// NewReformatCommand creates a reformat command.
// This command can be used for reformatting conftest JSON output to different formats.
func NewReformatCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "reformat [file...]",
		Short: "Reformat conftest JSON output to different formats",
		Long:  fmt.Sprintf(reformatDesc, output.Outputs()),
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			flagNames := []string{
				"output",
			}
			for _, name := range flagNames {
				if err := viper.BindPFlag(name, cmd.Flags().Lookup(name)); err != nil {
					return fmt.Errorf("bind flag: %w", err)
				}
			}

			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			outputFormat := viper.GetString("output")

			// Determine input source: positional args or stdin
			var reader io.Reader
			if len(args) > 0 {
				// Use first positional argument as input file
				file, err := os.Open(args[0])
				if err != nil {
					return fmt.Errorf("failed to open input file: %w", err)
				}
				defer file.Close()
				reader = file
			} else {
				// No positional args, read from stdin
				reader = os.Stdin
			}

			// Parse JSON input
			var results output.CheckResults
			decoder := json.NewDecoder(reader)
			if err := decoder.Decode(&results); err != nil {
				return fmt.Errorf("failed to parse JSON input: %w", err)
			}

			// Create outputter
			outputter := output.Get(outputFormat, output.Options{})

			// Format and output results
			if err := outputter.Output(results); err != nil {
				return fmt.Errorf("failed to output results: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringP("output", "o", output.OutputStandard, fmt.Sprintf("Output format for conftest results - valid options are: %s", output.Outputs()))

	return &cmd
}
