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

	policies, err := loader.AllRegos(l.PolicyPaths)
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	} else if len(policies.Modules) == 0 {
		return nil, fmt.Errorf("no policies found in %v", l.PolicyPaths)
	}

	compiler, err := policies.Compiler()
	if err != nil {
		return nil, fmt.Errorf("get compiler: %w", err)
	}

	// FilteredPaths will recursively find all file paths that contain a valid document
	// extension from the given list of data paths.
	allDocumentPaths, err := loader.FilteredPaths(l.DataPaths, func(abspath string, info os.FileInfo, depth int) bool {
		if info.IsDir() {
			return false
		}
		return !contains([]string{".yaml", ".yml", ".json"}, filepath.Ext(info.Name()))
	})
	if err != nil {
		return nil, fmt.Errorf("filter data paths: %w", err)
	}

	documents, err := loader.NewFileLoader().All(allDocumentPaths)
	if err != nil {
		return nil, fmt.Errorf("load documents: %w", err)
	}
	store, err := documents.Store()
	if err != nil {
		return nil, fmt.Errorf("get documents store: %w", err)
	}

	// The raw text and the path of the data documents are not preserved in the loader.
	// Both the path of the data document and its original contents are useful to have
	// especially when pushing to OCI registries.
	documentContents := make(map[string]string)
	for _, documentPath := range allDocumentPaths {
		contents, err := ioutil.ReadFile(documentPath)
		if err != nil {
			return nil, fmt.Errorf("read file: %w", err)
		}

		documentContents[documentPath] = string(contents)
	}

	engine := Engine{
		modules:  policies.ParsedModules(),
		compiler: compiler,
		store:    store,
		docs:     documentContents,
	}

	return &engine, nil
}
