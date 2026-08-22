package downloader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	getter "github.com/hashicorp/go-getter"
)

var detectors = []getter.Detector{
	new(OCIDetector),
	new(getter.GitHubDetector),
	new(getter.GitDetector),
	new(getter.BitBucketDetector),
	new(getter.S3Detector),
	new(getter.GCSDetector),
	new(getter.FileDetector),
}

var getters = map[string]getter.Getter{
	"file":  new(getter.FileGetter),
	"git":   new(getter.GitGetter),
	"gcs":   new(getter.GCSGetter),
	"hg":    new(getter.HgGetter),
	"s3":    new(getter.S3Getter),
	"oci":   new(OCIGetter),
	"http":  new(getter.HttpGetter),
	"https": new(getter.HttpGetter),
}

type downloadConfig struct {
	overwrite bool
}

// DownloadOption configures a policy download.
type DownloadOption func(*downloadConfig)

// WithOverwrite replaces an existing policy file after its update downloads successfully.
func WithOverwrite() DownloadOption {
	return func(config *downloadConfig) {
		config.overwrite = true
	}
}

// Download downloads the given policies into the given destination.
func Download(ctx context.Context, dst string, urls []string, opts ...DownloadOption) error {
	config := downloadConfig{}
	for _, opt := range opts {
		opt(&config)
	}

	for _, url := range urls {
		detectedURL, err := Detect(url, dst)
		if err != nil {
			return fmt.Errorf("detecting url: %w", err)
		}

		// Check if file already exists
		filename := filepath.Base(detectedURL)
		targetPath := filepath.Join(dst, filename)
		targetInfo, err := os.Stat(targetPath)
		if err == nil {
			if !config.overwrite {
				return fmt.Errorf("policy file already exists at %s, refusing to overwrite", targetPath)
			}
			if !targetInfo.IsDir() {
				if err := overwriteFile(ctx, detectedURL, dst, filename); err != nil {
					return err
				}
				continue
			}
		}

		if config.overwrite && strings.HasPrefix(detectedURL, "git::") {
			if err := removeEmptyDestination(dst); err != nil {
				return err
			}
		}

		if err := get(ctx, detectedURL, dst, dst); err != nil {
			return err
		}
	}

	return nil
}

// removeEmptyDestination removes dst if it exists but is empty. go-getter's
// git getter decides whether to clone or update a destination based solely
// on whether it already exists, so a pre-existing, empty directory (created
// ahead of time, e.g. by Atlantis, but not yet cloned into) is mistaken for
// an existing checkout, and the update fails because it isn't actually a git
// repository. Removing the empty directory lets go-getter clone into it
// fresh, as it would if the directory never existed.
func removeEmptyDestination(dst string) error {
	if entries, err := os.ReadDir(dst); err == nil && len(entries) == 0 {
		if err := os.Remove(dst); err != nil {
			return fmt.Errorf("remove empty policy directory: %w", err)
		}
	}

	return nil
}

func get(ctx context.Context, src string, dst string, pwd string) error {
	opts := []getter.ClientOption{}
	client := &getter.Client{
		Ctx:       ctx,
		Src:       src,
		Dst:       dst,
		Pwd:       pwd,
		Mode:      getter.ClientModeAny,
		Detectors: detectors,
		Getters:   getters,
		Options:   opts,
	}

	if err := client.Get(); err != nil {
		return fmt.Errorf("client get: %w", err)
	}

	return nil
}

func overwriteFile(ctx context.Context, src string, dst string, filename string) error {
	stagingDir, err := os.MkdirTemp("", ".conftest-update-*")
	if err != nil {
		return fmt.Errorf("create update staging directory: %w", err)
	}
	defer os.RemoveAll(stagingDir)

	if err := get(ctx, src, stagingDir, dst); err != nil {
		return err
	}

	stagedPath := filepath.Join(stagingDir, filename)
	targetPath := filepath.Join(dst, filename)
	if err := replaceFile(stagedPath, targetPath); err != nil {
		return fmt.Errorf("replace policy file: %w", err)
	}

	return nil
}

func replaceFile(src string, dst string) error {
	if err := os.Remove(dst); err != nil {
		return fmt.Errorf("remove existing file: %w", err)
	}

	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("install updated file: %w", err)
	}

	return nil
}

// Detect determines whether a url is a known source url from which we can download files.
// If a known source is found, the url is formatted, otherwise an error is returned.
func Detect(url string, dst string) (string, error) {

	// localhost is not considered a valid scheme for the detector which
	// causes pull commands that reference localhost to error.
	//
	// To allow for localhost to be used, replace the localhost reference
	// with the IP address.
	if strings.Contains(url, "localhost") {
		url = strings.ReplaceAll(url, "localhost", "127.0.0.1")
	}

	result, err := getter.Detect(url, dst, detectors)
	if err != nil {
		return "", fmt.Errorf("detect: %w", err)
	}

	return result, nil
}
