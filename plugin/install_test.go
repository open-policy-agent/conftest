package plugin

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsDirectory(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{input: "/", expected: true},
		{input: "/abs/path", expected: true},
		{input: "some/path", expected: true},
		{input: "file://some/path", expected: true},
		{input: "C:\\some\\path", expected: true},
		{input: "unknown", expected: true},
		{input: "unknown.com", expected: true},

		{input: "github.com/username/repo", expected: false},
	}

	for _, testCase := range testCases {
		actual, err := isDirectory(testCase.input)
		if err != nil {
			t.Fatal("is directory:", err)
		}

		if actual != testCase.expected {
			t.Errorf("Directory check failed. expected %v, actual %v", testCase.expected, actual)
		}
	}
}

func TestInstall(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(t *testing.T) (source string, wantPluginName string)
		wantErr          bool
		wantErrSubstring string
	}{
		{
			name: "install from directory",
			setup: func(t *testing.T) (string, string) {
				pluginDir := createInstallTestPlugin(t, "test-plugin")
				return pluginDir, "test-plugin"
			},
			wantErr: false,
		},
		{
			name: "install from valid URL",
			setup: func(t *testing.T) (string, string) {
				server := createTestArchiveServer(t, "test-plugin")
				return fmt.Sprintf("http::%s/plugin.tar.gz?archive=tar.gz", server.URL), "test-plugin"
			},
			wantErr: false,
		},
		{
			name: "install from invalid directory",
			setup: func(_ *testing.T) (string, string) {
				return "/nonexistent/directory", ""
			},
			wantErr: true,
		},
		{
			name: "install from URL with invalid plugin contents",
			setup: func(t *testing.T) (string, string) {
				server := createInvalidArchiveServer(t)
				return fmt.Sprintf("http::%s/invalid.tar.gz?archive=tar.gz", server.URL), ""
			},
			wantErr:          true,
			wantErrSubstring: "loading plugin from dir",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			t.Setenv("XDG_DATA_HOME", tmpDir)

			source, wantPluginName := tt.setup(t)
			err := Install(context.Background(), source)

			assertError(t, tt.wantErr, tt.wantErrSubstring, err)
			assertInstallationResult(t, tmpDir, wantPluginName, tt.wantErr)
		})
	}
}

// Helper functions
func createInstallTestPlugin(t *testing.T, name string) string {
	t.Helper()

	pluginDir := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatal(err)
	}

	configContent := []byte(fmt.Sprintf(`name: %s
version: 1.0.0
command: test-command`, name))
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.yaml"), configContent, 0o600); err != nil {
		t.Fatal(err)
	}

	return pluginDir
}

func createTestArchiveServer(t *testing.T, pluginName string) *httptest.Server {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	configContent := []byte(fmt.Sprintf(`name: %s
version: 1.0.0
command: test-command`, pluginName))

	header := &tar.Header{
		Name: "plugin.yaml",
		Mode: 0o600,
		Size: int64(len(configContent)),
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(configContent); err != nil {
		t.Fatal(err)
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/x-gzip")
		_, _ = w.Write(buf.Bytes())
	}))
}

func createInvalidArchiveServer(t *testing.T) *httptest.Server {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	header := &tar.Header{
		Name: "dummy.txt",
		Mode: 0o600,
		Size: 0,
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/x-gzip")
		_, _ = w.Write(buf.Bytes())
	}))
}

func assertError(t *testing.T, wantErr bool, wantMsg string, actual error) {
	t.Helper()

	if wantErr {
		if actual == nil {
			t.Fatal("expected error but got none")
		}
		if wantMsg != "" && !strings.Contains(actual.Error(), wantMsg) {
			t.Errorf("error %q should contain %q", actual.Error(), wantMsg)
		}
	} else if actual != nil {
		t.Fatalf("unexpected error: %v", actual)
	}
}

func assertInstallationResult(t *testing.T, tmpDir, wantPluginName string, wantErr bool) {
	t.Helper()

	if wantErr {
		assertTempFilesCleaned(t, tmpDir)
		return
	}

	pluginDir := filepath.Join(tmpDir, ".conftest", "plugins", wantPluginName)
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		t.Errorf("plugin directory %q not found", pluginDir)
	}

	configPath := filepath.Join(pluginDir, "plugin.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("plugin.yaml not found in plugin directory")
	}
}

func assertTempFilesCleaned(t *testing.T, tmpDir string) {
	t.Helper()

	matches, _ := filepath.Glob(filepath.Join(tmpDir, ".conftest", "plugins", "conftest-plugin-*"))
	if len(matches) > 0 {
		t.Errorf("temporary directories not cleaned up: %v", matches)
	}
}
