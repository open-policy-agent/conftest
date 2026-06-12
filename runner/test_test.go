package runner

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseFileList(t *testing.T) {
	dir := t.TempDir()
	main := filepath.Join(dir, "main.tf")
	provider := filepath.Join(dir, "provider.tf")
	for _, file := range []string{main, provider} {
		if err := os.WriteFile(file, []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name        string
		fileList    []string
		ignoreRegex string
		expected    []string
	}{
		{
			name:     "without an ignore regex all explicit files are kept",
			fileList: []string{main, provider},
			expected: []string{main, provider},
		},
		{
			name:        "the ignore regex is applied to explicitly provided files",
			fileList:    []string{main, provider},
			ignoreRegex: `.*/provider\.tf`,
			expected:    []string{main},
		},
		{
			name:        "stdin is never ignored",
			fileList:    []string{"-"},
			ignoreRegex: ".*",
			expected:    []string{"-"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseFileList(tt.fileList, tt.ignoreRegex)
			if err != nil {
				t.Fatalf("parse file list: %v", err)
			}
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, actual)
			}
		})
	}
}

func TestParseFileListInvalidIgnoreRegex(t *testing.T) {
	if _, err := parseFileList([]string{"-"}, "("); err == nil {
		t.Error("expected an error for an invalid ignore regexp, but none occurred")
	}
}
