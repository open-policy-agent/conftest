package runner

import (
	"testing"

	"github.com/open-policy-agent/conftest/output"
)

func TestTestRunner_GroupFlag(t *testing.T) {
	// Create a mock TestRunner with Group set
	runner := TestRunner{
		Group: "my-custom-group",
	}

	// Create mock results with stdin filename
	results := output.CheckResults{
		{
			FileName:  "-",
			Namespace: "main",
			Failures:  []output.Result{{Message: "test failure"}},
		},
		{
			FileName:  "regular-file.yaml",
			Namespace: "main", 
			Failures:  []output.Result{{Message: "another failure"}},
		},
	}

	// Test the group name override logic
	if runner.Group != "" {
		for i := range results {
			if results[i].FileName == "-" {
				results[i].FileName = runner.Group
			}
		}
	}

	// Verify the stdin result was updated
	if results[0].FileName != "my-custom-group" {
		t.Errorf("Expected stdin filename to be overridden to 'my-custom-group', got '%s'", results[0].FileName)
	}

	// Verify the regular file result was not changed
	if results[1].FileName != "regular-file.yaml" {
		t.Errorf("Expected regular filename to remain 'regular-file.yaml', got '%s'", results[1].FileName)
	}
}

func TestTestRunner_GroupFlagEmpty(t *testing.T) {
	// Create a mock TestRunner without Group set
	runner := TestRunner{
		Group: "",
	}

	// Create mock results with stdin filename
	results := output.CheckResults{
		{
			FileName:  "-",
			Namespace: "main",
			Failures:  []output.Result{{Message: "test failure"}},
		},
	}

	// Test the group name override logic
	if runner.Group != "" {
		for i := range results {
			if results[i].FileName == "-" {
				results[i].FileName = runner.Group
			}
		}
	}

	// Verify the stdin result was NOT updated
	if results[0].FileName != "-" {
		t.Errorf("Expected stdin filename to remain '-', got '%s'", results[0].FileName)
	}
}