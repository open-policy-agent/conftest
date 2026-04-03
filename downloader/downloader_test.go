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

func TestDownloadOverwritesExistingFile(t *testing.T) {
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

	// Download should succeed even when the policy file already exists,
	// overwriting the previous content (e.g. when using --update repeatedly).
	urls := []string{fmt.Sprintf("http://%s/policy.rego", listener.Addr().String())}
	if err := Download(context.Background(), tmpDir, urls); err != nil {
		t.Fatalf("Expected download to succeed when file exists, got error: %v", err)
	}

	// Verify the file was overwritten with new content
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new content" {
		t.Errorf("Expected file to contain 'new content', got: %s", string(content))
	}
}
