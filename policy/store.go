package policy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/open-policy-agent/opa/loader"
	"github.com/open-policy-agent/opa/storage"
)

// StoreFromDataFiles returns an Open Policy Agent Store with the
// loaded documents found in the paths. Any JSON or YAML document
// could be a valid document.
func StoreFromDataFiles(paths []string) (storage.Store, error) {
	res, err := loader.Filtered(paths, filterDataFiles)
	if err != nil {
		return nil, fmt.Errorf("load data files: %w", err)
	}

	store, err := res.Store()
	if err != nil {
		return nil, fmt.Errorf("load store from data files: %w", err)
	}

	return store, nil
}

// ReadDataFiles returns the paths to all data documents
// at the given path. Any JSON or YAML document
// could be a valid document.
func ReadDataFiles(path string) ([]string, error) {
	var data []string
	err := filepath.Walk(path, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk path: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(currentPath) == ".json" || filepath.Ext(currentPath) == ".yaml" {
			data = append(data, currentPath)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("find data files: %w", err)
	}

	return data, nil
}

func filterDataFiles(abspath string, info os.FileInfo, depth int) bool {
	pattern := "*.rego"
	match, _ := filepath.Match(pattern, info.Name())
	return match
}
