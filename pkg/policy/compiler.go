package policy

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/opa/ast"
)

// BuildCompiler compiles all Rego policies at the given path and returns the Compiler containing
// the compilation state
func BuildCompiler(path string) (*ast.Compiler, error) {
	files, err := recursivelySearchDirForRegoFiles(path)
	if err != nil {
		return nil, err
	}

	modules := map[string]*ast.Module{}

	for _, file := range files {
		out, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		name := filepath.Base(file)
		parsed, err := ast.ParseModule(name, string(out[:]))
		if err != nil {
			return nil, err
		}
		modules[name] = parsed
	}

	compiler := ast.NewCompiler()
	compiler.Compile(modules)

	if compiler.Failed() {
		return nil, compiler.Errors
	}

	return compiler, nil
}

func recursivelySearchDirForRegoFiles(path string) ([]string, error) {
	var filepaths []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".rego") {
			filepaths = append(filepaths, path)
		}

		return nil
	})

	return filepaths, err
}
