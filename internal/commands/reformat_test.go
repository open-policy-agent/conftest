package commands

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/open-policy-agent/conftest/output"
)

func TestReformat(t *testing.T) {
	sampleResults := output.CheckResults{
		{
			FileName:  "test.yaml",
			Namespace: "main",
			Successes: 1,
			Warnings: []output.Result{
				{Message: "Warning: test warning"},
			},
			Failures: []output.Result{
				{Message: "Error: test failure"},
			},
		},
	}

	jsonInput, err := json.Marshal(sampleResults)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	tests := []struct {
		name         string
		input        string
		outputFormat string
		wantErr      bool
	}{
		{
			name:         "valid json input with json output",
			input:        string(jsonInput),
			outputFormat: "json",
			wantErr:      false,
		},
		{
			name:         "valid json input with table output",
			input:        string(jsonInput),
			outputFormat: "table",
			wantErr:      false,
		},
		{
			name:         "valid json input with tap output",
			input:        string(jsonInput),
			outputFormat: "tap",
			wantErr:      false,
		},
		{
			name:         "valid json input with junit output",
			input:        string(jsonInput),
			outputFormat: "junit",
			wantErr:      false,
		},
		{
			name:         "invalid json input",
			input:        "invalid json",
			outputFormat: "json",
			wantErr:      true,
		},
		{
			name:         "unknown output format defaults to standard",
			input:        string(jsonInput),
			outputFormat: "unknown",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			err := reformat(reader, tt.outputFormat)
			if (err != nil) != tt.wantErr {
				t.Errorf("reformat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReformatCommandFlags(t *testing.T) {
	cmd := NewReformatCommand()

	// Test output flag exists and has correct default
	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("Expected output flag to exist")
	}

	if outputFlag.DefValue != output.OutputStandard {
		t.Errorf("Expected default output format to be %q, got %q", output.OutputStandard, outputFlag.DefValue)
	}
}
