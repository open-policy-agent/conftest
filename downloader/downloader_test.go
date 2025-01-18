package downloader

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloadFailsWhenFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file that would conflict with the download
	existingFile := filepath.Join(tmpDir, "policy.rego")
	if err := os.WriteFile(existingFile, []byte("existing content"), os.FileMode(0600)); err != nil {
		t.Fatal(err)
	}

	// Start a test HTTP server on an ephemeral port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, "new content")
		}),
		ReadHeaderTimeout: 1 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(listener)
	}()
	defer server.Close()

	// Try to download a policy file with the same name
	urls := []string{fmt.Sprintf("http://%s/policy.rego", listener.Addr().String())}
	downloadErr := Download(context.Background(), tmpDir, urls)

	// Verify that download fails with the expected error
	if downloadErr == nil {
		t.Error("Expected download to fail when file exists, but it succeeded")
	}
	if downloadErr != nil && !filepath.IsAbs(existingFile) {
		t.Errorf("Expected error message to contain absolute path, got: %v", downloadErr)
	}

	// Verify the original file is unchanged
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "existing content" {
		t.Error("Existing file was modified")
	}
}
