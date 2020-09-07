package policy

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/conftest/downloader"

	"github.com/open-policy-agent/opa/storage"
)

// Loader handles the retrieval of all rego policies and related data.
type Loader struct {
	PolicyPaths []string
	DataPaths   []string
	URLs        []string

	test bool
}

// SetTestLoad configures the loader to load Rego test files as well
func (l *Loader) SetTestLoad(test bool) *Loader {
	l.test = test
	return l
}

// Load retrieves policies from several locations:
// first it checks for any remote sources of policies and downloads
// the policies into the given policy paths.
// After retrieving the policies from the remote sources, all .rego, .json and .yaml
// files are recursively retrieved from disk and loaded into
// a rego Compiler and Store respectively.
func (l *Loader) Load(ctx context.Context) ([]string, storage.Store, error) {
	// Downloaded policies are put into the first policy directory specified
	for _, url := range l.URLs {
		sourcedURL, err := downloader.Detect(url, l.PolicyPaths[0])
		if err != nil {
			return nil, nil, fmt.Errorf("detect policies: %w", err)
		}

		if err := downloader.Download(ctx, l.PolicyPaths[0], []string{sourcedURL}); err != nil {
			return nil, nil, fmt.Errorf("update policies: %w", err)
		}
	}

	var regoFiles []string
	var err error
	if l.test {
		regoFiles, err = ReadFilesWithTests(l.PolicyPaths...)
	} else {
		regoFiles, err = ReadFiles(l.PolicyPaths...)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("read rego files: %w", err)
	}

	store, err := StoreFromDataFiles(l.DataPaths)
	if err != nil {
		return nil, nil, fmt.Errorf("build store: %w", err)
	}

	return regoFiles, store, nil
}
