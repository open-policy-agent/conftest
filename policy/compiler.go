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
func BuildCompiler(path string, withTests bool) (*ast.Compiler, error) {
	files, err := recursivelySearchDirForRegoFiles(path, withTests)
	if err != nil {
		return nil, err
	}

	modules := map[string]*ast.Module{}

	for _, file := range files {
		out, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		relPath, err := filepath.Rel(path, file)
		if err != nil {
			return nil, err
		}

		parsed, err := ast.ParseModule(relPath, string(out[:]))
		if err != nil {
			return nil, err
		}

		modules[relPath] = parsed
	}

	compiler := ast.NewCompiler()
	compiler.Compile(modules)

	if compiler.Failed() {
		return nil, compiler.Errors
	}

	return compiler, nil
}

func recursivelySearchDirForRegoFiles(path string, withTests bool) ([]string, error) {
	var filepaths []string

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if withTests {
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".rego") {
				filepaths = append(filepaths, path)
			}
		} else {
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".rego") && !strings.HasSuffix(info.Name(), "_test.rego") {
				filepaths = append(filepaths, path)
			}

		}

		return nil
	})

	return filepaths, err
}
