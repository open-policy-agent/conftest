package policy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/opa/ast"
)

// BuildCompiler compiles all of the given Rego policies
// and returns the Compiler containing the compilation state
func BuildCompiler(files []string) (*ast.Compiler, error) {
	modules := map[string]*ast.Module{}

	for _, file := range files {
		out, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		parsed, err := ast.ParseModule(file, string(out[:]))
		if err != nil {
			return nil, err
		}

		modules[file] = parsed
	}

	compiler := ast.NewCompiler()
	compiler.Compile(modules)
	if compiler.Failed() {
		return nil, compiler.Errors
	}

	return compiler, nil
}

// ReadRegoFiles returns all of the rego files at the given path
// including its subdirectories.
func ReadRegoFiles(path string) ([]string, error) {
	files, err := recursivelySearchDirForRegoFiles(path, false)
	if err != nil {
		return nil, fmt.Errorf("search rego files: %w", err)
	}

	return files, nil
}

// ReadRegoTestFiles returns all of the rego test files at
// the given path including its subdirectories.
// Rego test files are Rego files that have a suffix of _test.rego
func ReadRegoTestFiles(path string) ([]string, error) {
	files, err := recursivelySearchDirForRegoFiles(path, true)
	if err != nil {
		return nil, fmt.Errorf("search rego test files: %w", err)
	}

	return files, nil
}

func recursivelySearchDirForRegoFiles(path string, includeTests bool) ([]string, error) {
	var filepaths []string
	err := filepath.Walk(path, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk path: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(currentPath) == "rego" && !strings.HasPrefix(info.Name(), "_test.rego") {
			filepaths = append(filepaths, path)
		}

		if includeTests && strings.HasPrefix(info.Name(), "_test.rego") {
			filepaths = append(filepaths, path)
		}

		return nil
	})

	return filepaths, err
}
