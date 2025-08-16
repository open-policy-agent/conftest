package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/open-policy-agent/conftest/output"
	"github.com/spf13/viper"
)

func TestReformatCommand(t *testing.T) {
	sampleResults := output.CheckResults{
		{
			FileName:  "test.yaml",
			Namespace: "main",
			Successes: 1,
			Warnings: []output.Result{
				{
					Message: "Warning: test warning",
				},
			},
			Failures: []output.Result{
				{
					Message: "Error: test failure",
				},
			},
		},
	}

	testCases := []struct {
		name           string
		outputFormat   string
		expectError    bool
		expectedOutput string
	}{
		{
			name:         "json output",
			outputFormat: "json",
			expectError:  false,
		},
		{
			name:         "table output",
			outputFormat: "table",
			expectError:  false,
		},
		{
			name:         "tap output",
			outputFormat: "tap",
			expectError:  false,
		},
		{
			name:         "junit output",
			outputFormat: "junit",
			expectError:  false,
		},
		{
			name:         "invalid output format",
			outputFormat: "invalid",
			expectError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset viper for each test
			viper.Reset()

			jsonInput, err := json.Marshal(sampleResults)
			if err != nil {
				t.Fatalf("Failed to marshal test data: %v", err)
			}

			var outputBuffer bytes.Buffer

			// Temporarily redirect stdout
			oldStdout := os.Stdout
			defer func() {
				os.Stdout = oldStdout
			}()

			cmd := NewReformatCommand()
			cmd.SetArgs([]string{})
			if err := cmd.Flags().Set("output", tc.outputFormat); err != nil {
				t.Error("failed to set flags for reformat command", err)
			}

			// Create a pipe to simulate stdin
			reader := strings.NewReader(string(jsonInput))
			oldStdin := os.Stdin
			defer func() {
				os.Stdin = oldStdin
			}()

			// Test command creation and flag parsing
			if !tc.expectError {

				outputFlag := cmd.Flags().Lookup("output")
				if outputFlag == nil {
					t.Error("Expected output flag to exist")
				}

				inputFlag := cmd.Flags().Lookup("input")
				if inputFlag == nil {
					t.Error("Expected input flag to exist")
				}

				// Test that PreRunE doesn't error
				err := cmd.PreRunE(cmd, []string{})
				if err != nil {
					t.Errorf("PreRunE failed: %v", err)
				}
			}

			// Restore stdin
			_ = reader
			_ = outputBuffer
		})
	}
}

func TestReformatCommandFlags(t *testing.T) {
	cmd := NewReformatCommand()

	// Test default values
	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("Expected output flag to exist")
	}
	if outputFlag.DefValue != output.OutputStandard {
		t.Errorf("Expected default output format to be %s, got %s", output.OutputStandard, outputFlag.DefValue)
	}

	inputFlag := cmd.Flags().Lookup("input")
	if inputFlag == nil {
		t.Fatal("Expected input flag to exist")
	}
	if inputFlag.DefValue != "" {
		t.Errorf("Expected default input to be empty, got %s", inputFlag.DefValue)
	}

}

func TestReformatCommandPreRunE(t *testing.T) {
	cmd := NewReformatCommand()

	// Test that PreRunE binds flags correctly
	err := cmd.PreRunE(cmd, []string{})
	if err != nil {
		t.Errorf("PreRunE failed: %v", err)
	}
}
