package policy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/containerd/log"
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

// Download downloads the given policies into the given destination
func Download(ctx context.Context, dst string, urls []string) error {
	opts := []getter.ClientOption{}
	for _, url := range urls {
		log.G(ctx).Debugf("Initializing go-getter client with url %v and dst %v", url, dst)
		client := &getter.Client{
			Ctx:       ctx,
			Src:       url,
			Dst:       dst,
			Pwd:       dst,
			Mode:      getter.ClientModeAny,
			Detectors: detectors,
			Getters:   getters,
			Options:   opts,
		}

		if err := client.Get(); err != nil {
			return err
		}
	}

	return nil
}

// Detect determines whether a url is a known source url from which we can download files.
// If a known source is found, the url is formatted, otherwise an error is returned.
func Detect(url string, dst string) (string, error) {
	result, err := getter.Detect(url, dst, detectors)
	if err != nil {
		return "", err
	}

	return result, err
}

// ReadFiles returns all of the policy files (not including tests)
// at the given path including its subdirectories.
func ReadFiles(path string) ([]string, error) {
	files, err := getPolicyFiles(path)
	if err != nil {
		return nil, fmt.Errorf("search rego files: %w", err)
	}

	return files, nil
}

func getPolicyFiles(path string) ([]string, error) {
	var filepaths []string
	err := filepath.Walk(path, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk path: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(currentPath) == ".rego" && !strings.HasSuffix(info.Name(), "_test.rego") {
			filepaths = append(filepaths, currentPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(filepaths) < 1 {
		return nil, fmt.Errorf("no policies found in %s", path)
	}

	return filepaths, nil
}

// ReadFilesWithTests returns all of the policies and test files
// at the given path including its subdirectories.
// Test files are Rego files that have a suffix of _test.rego
func ReadFilesWithTests(path string) ([]string, error) {
	files, err := getTestFiles(path)
	if err != nil {
		return nil, fmt.Errorf("search rego test files: %w", err)
	}

	return files, nil
}

func getTestFiles(path string) ([]string, error) {
	var filepaths []string
	err := filepath.Walk(path, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk path: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".rego") {
			filepaths = append(filepaths, currentPath)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return filepaths, nil
}
