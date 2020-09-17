package policy

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/open-policy-agent/conftest/downloader"

	"github.com/open-policy-agent/opa/loader"
)

// Loader handles the retrieval of all rego policies and related data.
type Loader struct {
	PolicyPaths []string
	DataPaths   []string
	URLs        []string
}

// Load returns an Engine after loading all of the specified policies and data paths.
//
// If URLs are specified, it will first download the policies at the specified URLs
// and put them in the first directory that appears in the policy paths.
func (l *Loader) Load(ctx context.Context) (*Engine, error) {
	for _, url := range l.URLs {
		if err := downloader.Download(ctx, l.PolicyPaths[0], []string{url}); err != nil {
			return nil, fmt.Errorf("update policies: %w", err)
		}
	}

	paths := append(l.PolicyPaths, l.DataPaths...)
	result, err := loader.All(paths)
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}

	if len(result.Modules) == 0 {
		return nil, fmt.Errorf("no policies found in %v", l.PolicyPaths)
	}

	compiler, err := result.Compiler()
	if err != nil {
		return nil, fmt.Errorf("get compiler: %w", err)
	}

	store, err := result.Store()
	if err != nil {
		return nil, fmt.Errorf("get store: %w", err)
	}

	docs, err := loadDocuments(l.DataPaths)
	if err != nil {
		return nil, fmt.Errorf("loading docs: %w", err)
	}

	engine := Engine{
		result:   result,
		compiler: compiler,
		store:    store,
		docs:     docs,
	}

	return &engine, nil
}

// The Rego loader is able to take in any number of paths and correctly distinguish between
// data documents (Documents) and policies (Modules).
//
// However, the raw text and the path of the data documents are not preserved.
// Both the path of the data document and its original content is useful to have, especially
// when pushing to OCI registries.
func loadDocuments(paths []string) (map[string]string, error) {
	ignoreFileExtensions := func(abspath string, info os.FileInfo, depth int) bool {
		return !contains([]string{".yaml", ".yml", ".json"}, filepath.Ext(info.Name()))
	}

	documentPaths, err := loader.FilteredPaths(paths, ignoreFileExtensions)
	if err != nil {
		return nil, fmt.Errorf("filter data paths: %w", err)
	}

	documents := make(map[string]string)
	for _, documentPath := range documentPaths {
		contents, err := ioutil.ReadFile(documentPath)
		if err != nil {
			return nil, fmt.Errorf("read file: %w", err)
		}

		documents[documentPath] = string(contents)
	}

	return documents, nil
}
