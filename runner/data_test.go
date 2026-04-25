package runner

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDefaultDataPaths(t *testing.T) {
	root := t.TempDir()

	withData := filepath.Join(root, "with-data")
	if err := os.MkdirAll(filepath.Join(withData, "data"), 0o755); err != nil {
		t.Fatal(err)
	}

	withoutData := filepath.Join(root, "without-data")
	if err := os.MkdirAll(withoutData, 0o755); err != nil {
		t.Fatal(err)
	}

	withDataAsFile := filepath.Join(root, "data-is-file")
	if err := os.MkdirAll(withDataAsFile, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(withDataAsFile, "data"), []byte("noop"), 0o600); err != nil {
		t.Fatal(err)
	}

	policyFile := filepath.Join(root, "policy.rego")
	if err := os.WriteFile(policyFile, []byte("package main"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		policyPaths []string
		want        []string
	}{
		{
			name:        "discovers sibling data directory",
			policyPaths: []string{withData},
			want:        []string{filepath.Join(withData, "data")},
		},
		{
			name:        "skips policy directory without data",
			policyPaths: []string{withoutData},
			want:        nil,
		},
		{
			name:        "skips when data is a file, not a directory",
			policyPaths: []string{withDataAsFile},
			want:        nil,
		},
		{
			name:        "skips when policy path is a file",
			policyPaths: []string{policyFile},
			want:        nil,
		},
		{
			name:        "skips missing policy paths",
			policyPaths: []string{filepath.Join(root, "does-not-exist")},
			want:        nil,
		},
		{
			name:        "dedupes when the same policy is passed twice",
			policyPaths: []string{withData, withData},
			want:        []string{filepath.Join(withData, "data")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := defaultDataPaths(tt.policyPaths)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
