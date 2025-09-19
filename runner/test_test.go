package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestTestRunner_FileNameOverrideFlag(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create a test policy file
	policyDir := filepath.Join(tmpDir, "policy")
	if err := os.Mkdir(policyDir, 0755); err != nil {
		t.Fatalf("Failed to create policy directory: %v", err)
	}

	policyContent := `package main

deny contains msg if {
	input.kind == "Deployment"
	input.metadata.name == "test"
	msg := "test deployment found"
}
`
	policyFile := filepath.Join(policyDir, "test.rego")
	if err := os.WriteFile(policyFile, []byte(policyContent), 0644); err != nil {
		t.Fatalf("Failed to write policy file: %v", err)
	}

	// Test cases
	tests := []struct {
		name             string
		fileNameOverride string
		expectedFileName string
	}{
		{
			name:             "with file-name-override",
			fileNameOverride: "my-custom-file.yaml",
			expectedFileName: "my-custom-file.yaml",
		},
		{
			name:             "without file-name-override",
			fileNameOverride: "",
			expectedFileName: "-",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create TestRunner with FileNameOverride set
			runner := TestRunner{
				Policy:           []string{policyDir},
				Namespace:        []string{"main"},
				FileNameOverride: tc.fileNameOverride,
				RegoVersion:      "v1",
			}

			// Create stdin input by using "-" as the file path
			fileList := []string{"-"}

			// Mock stdin with test data
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r
			testInput := `{
	"apiVersion": "apps/v1",
	"kind": "Deployment",
	"metadata": {
		"name": "test"
	}
}`
			go func() {
				defer w.Close()
				w.Write([]byte(testInput))
			}()
			defer func() { os.Stdin = oldStdin }()

			// Run the test
			ctx := context.Background()
			results, err := runner.Run(ctx, fileList)
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}

			// Verify results
			if len(results) == 0 {
				t.Fatal("Expected at least one result")
			}

			// Check that the filename was properly overridden
			if results[0].FileName != tc.expectedFileName {
				t.Errorf("Expected filename to be '%s', got '%s'", tc.expectedFileName, results[0].FileName)
			}
		})
	}
}

func TestTestRunner_FileNameOverrideOnlyAffectsStdin(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create a test policy file
	policyDir := filepath.Join(tmpDir, "policy")
	if err := os.Mkdir(policyDir, 0755); err != nil {
		t.Fatalf("Failed to create policy directory: %v", err)
	}

	policyContent := `package main

deny contains msg if {
	input.kind == "Deployment"
	msg := "deployment found"
}
`
	policyFile := filepath.Join(policyDir, "test.rego")
	if err := os.WriteFile(policyFile, []byte(policyContent), 0644); err != nil {
		t.Fatalf("Failed to write policy file: %v", err)
	}

	// Create a test config file
	configContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-file
`
	configFile := filepath.Join(tmpDir, "deployment.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create TestRunner with FileNameOverride set
	runner := TestRunner{
		Policy:           []string{policyDir},
		Namespace:        []string{"main"},
		FileNameOverride: "overridden-name.yaml",
		RegoVersion:      "v1",
	}

	// Run with a regular file (not stdin)
	ctx := context.Background()
	results, err := runner.Run(ctx, []string{configFile})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify results
	if len(results) == 0 {
		t.Fatal("Expected at least one result")
	}

	// Check that the regular file name was NOT overridden
	if results[0].FileName != configFile {
		t.Errorf("Expected filename to remain '%s', got '%s'", configFile, results[0].FileName)
	}
}
