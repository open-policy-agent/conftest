package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/open-policy-agent/conftest/downloader"
	"github.com/open-policy-agent/conftest/output"
	"github.com/open-policy-agent/conftest/parser"
	"github.com/open-policy-agent/conftest/policy"
)

// TestRunner is the runner for the Test command, executing
// Rego policy checks against configuration files.
type TestRunner struct {
	Trace              bool
	Strict             bool
	Capabilities       string
	RegoVersion        string `mapstructure:"rego-version"`
	Policy             []string
	Data               []string
	Update             []string
	Ignore             string
	Parser             string
	Namespace          []string
	AllNamespaces      bool `mapstructure:"all-namespaces"`
	FailOnWarn         bool `mapstructure:"fail-on-warn"`
	NoColor            bool `mapstructure:"no-color"`
	NoFail             bool `mapstructure:"no-fail"`
	SuppressExceptions bool `mapstructure:"suppress-exceptions"`
	ShowBuiltinErrors  bool `mapstructure:"show-builtin-errors"`
	Combine            bool
	Quiet              bool
	Output             string
}

// Run executes the TestRunner, verifying all Rego policies against the given
// list of configuration files.
func (t *TestRunner) Run(ctx context.Context, fileList []string) (output.CheckResults, error) {
	files, err := parseFileList(fileList, t.Ignore)
	if err != nil {
		return nil, fmt.Errorf("parse files: %w", err)
	}

	var configurations map[string]any
	if t.Parser != "" {
		configurations, err = parser.ParseConfigurationsAs(files, t.Parser)
	} else {
		configurations, err = parser.ParseConfigurations(files)
	}
	if err != nil {
		return nil, fmt.Errorf("parse configurations: %w", err)
	}

	// When there are policies to download, they are currently placed in the first
	// directory that appears in the list of policies.
	if len(t.Update) > 0 {
		if err := downloader.Download(ctx, t.Policy[0], t.Update); err != nil {
			return nil, fmt.Errorf("update policies: %w", err)
		}
	}

	capabilities, err := policy.LoadCapabilities(t.Capabilities)
	if err != nil {
		return nil, fmt.Errorf("load capabilities: %w", err)
	}
	opts := policy.CompilerOptions{
		Strict:       t.Strict,
		RegoVersion:  t.RegoVersion,
		Capabilities: capabilities,
	}
	engine, err := policy.LoadWithData(t.Policy, t.Data, opts)
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}
	engine.EnableInterQueryCache()

	if t.Trace {
		engine.EnableTracing()
	}

	if t.ShowBuiltinErrors {
		engine.ShowBuiltinErrors()
	}

	namespaces := t.Namespace
	if t.AllNamespaces {
		namespaces = engine.Namespaces()
	}

	var results output.CheckResults
	for _, namespace := range namespaces {
		if t.Combine {
			result, err := engine.CheckCombined(ctx, configurations, namespace)
			if err != nil {
				return nil, fmt.Errorf("check combined: %w", err)
			}

			results = append(results, result)
		} else {
			result, err := engine.Check(ctx, configurations, namespace)
			if err != nil {
				return nil, fmt.Errorf("query rule: %w", err)
			}

			results = append(results, result...)
		}
	}

	return results, nil
}

func parseFileList(fileList []string, ignoreRegex string) ([]string, error) {
	var files []string
	for _, file := range fileList {
		if file == "" {
			continue
		}

		if file == "-" {
			files = append(files, "-")
			continue
		}

		fileInfo, err := os.Stat(file)
		if err != nil {
			return nil, fmt.Errorf("get file info: %w", err)
		}

		if fileInfo.IsDir() {
			directoryFiles, err := getFilesFromDirectory(file, ignoreRegex)
			if err != nil {
				return nil, fmt.Errorf("get files from directory: %w", err)
			}

			files = append(files, directoryFiles...)
		} else {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found")
	}

	return files, nil
}

func getWalkFn(visitedDirs map[string]bool, files *[]string, ignoreRegex string, regexp *regexp.Regexp) filepath.WalkFunc {
	return func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk path: %w", err)
		}

		if info.IsDir() {
			if _, ok := visitedDirs[currentPath]; ok {
				return filepath.SkipDir
			}
			visitedDirs[currentPath] = true
			return nil
		}

		if ignoreRegex != "" && regexp.MatchString(currentPath) {
			return nil
		}

		if info.Mode()&os.ModeSymlink == 0 {
			if parser.FileSupported(currentPath) {
				*files = append(*files, currentPath)
			}
			return nil
		}

		realPath, err := filepath.EvalSymlinks(currentPath)
		if err != nil {
			return err
		}

		ri, err := os.Stat(realPath)
		if err != nil {
			return err
		}

		if ri.IsDir() {
			return filepath.Walk(realPath, getWalkFn(visitedDirs, files, ignoreRegex, regexp))
		}

		if parser.FileSupported(realPath) {
			*files = append(*files, realPath)
		}

		return nil
	}
}

func getFilesFromDirectory(directory string, ignoreRegex string) ([]string, error) {
	regexp, err := regexp.Compile(ignoreRegex)
	if err != nil {
		return nil, fmt.Errorf("given regexp couldn't be parsed :%w", err)
	}

	var files []string
	visitedDirs := make(map[string]bool)
	err = filepath.Walk(directory, getWalkFn(visitedDirs, &files, ignoreRegex, regexp))
	if err != nil {
		return nil, err
	}

	return files, nil
}
