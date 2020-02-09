package policy

import (
	"context"
	"fmt"
	"github.com/instrumenta/conftest/downloader"
	"os"
	"path/filepath"
	"strings"
)

// Download downloads the given policies into the given destination
func Download(ctx context.Context, dst string, urls []string) error {
	return downloader.Download(ctx, dst, urls)
}

// Detect determines whether a url is a known source url from which we can download files.
// If a known source is found, the url is formatted, otherwise an error is returned.
func Detect(url string, dst string) (string, error) {
	return downloader.Detect(url, dst)
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
			if info.Size() == 0 {
				return fmt.Errorf("empty policy found in %s", currentPath)
			}

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
