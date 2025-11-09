//go:build unix

package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetupClientCredentialsError(t *testing.T) {
	configFilePath := filepath.Join(t.TempDir(), "config.json")

	// Write file with no permissions
	err := os.WriteFile(configFilePath, []byte(`{"auths":{}}`), 0o000)
	if err != nil {
		t.Fatalf("failed to write to config file: %v", err)
	}

	repository := mustParseReference("local-test-registry/image:tag")
	t.Setenv("DOCKER_CONFIG", configFilePath)

	if err := SetupClient(repository); err == nil {
		t.Error("expected error when credentials store initialization fails, got nil")
	}
}
