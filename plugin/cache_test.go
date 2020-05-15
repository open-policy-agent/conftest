package plugin

import (
	"io/ioutil"
	"testing"
)

func Test_createPluginCacheDir(t *testing.T) {
	dir, err := ioutil.TempDir("", "plugin_cache")
	if err != nil {
		t.Fatalf("Unexpted error creating temp directory plugin_cache")
	}

	if _, err := createPluginCacheDir(dir); err != nil {
		t.Errorf("Unexpected error creating plugin cache: %w", err)
	}

	// Second time should skip creation
	if _, err := createPluginCacheDir(dir); err != nil {
		t.Errorf("Unexpected error creating plugin cache: %w", err)
	}
}

func Test_convertURLToValidPath(t *testing.T) {
	tests := []struct {
		name string
		url  string
		path string
	}{
		{
			"Should convert https url to valid path",
			"https://raw.githubusercontent.com/open-policy-agent/conftest/plugin",
			"https-raw-githubusercontent-com-open-policy-agent-conftest-plugin",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if path := convertURLToValidPath(tc.url); path != tc.path {
				t.Errorf("Converted url into incorrect path, got: %v, expected: %v", path, tc.path)
			}
		})
	}
}
