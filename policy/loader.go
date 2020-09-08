package policy

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/conftest/downloader"

	"github.com/open-policy-agent/opa/loader"
)

// Loader handles the retrieval of all rego policies and related data.
type Loader struct {
	PolicyPaths []string
	DataPaths   []string
	URLs        []string
	Tracing     bool
}

// Load returns an Engine after loading all of the specified policies and data paths.
// If URLs are specified, Load will first download all of the policies at the specified URLs.
func (l *Loader) Load(ctx context.Context) (*Engine, error) {

	// Downloaded policies are put into the first policy directory specified.
	for _, url := range l.URLs {
		sourcedURL, err := downloader.Detect(url, l.PolicyPaths[0])
		if err != nil {
			return nil, fmt.Errorf("detect policies: %w", err)
		}

		if err := downloader.Download(ctx, l.PolicyPaths[0], []string{sourcedURL}); err != nil {
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

	engine := Engine{
		result:   result,
		compiler: compiler,
		store:    store,
		tracing:  l.Tracing,
	}

	return &engine, nil
}
