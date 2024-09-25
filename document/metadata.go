package document

import (
	"fmt"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"
	"os"
	"path/filepath"
	"strings"
)

// ParseRegoWithAnnotations parse the rego in the indicated directory
// This can be later used to access the annotation and generate the documentation
func ParseRegoWithAnnotations(directory string) (*ast.Compiler, error) {
	// Recursively find all rego files (ignoring test files), starting at the given directory.
	result, err := loader.NewFileLoader().
		WithProcessAnnotation(true).
		Filtered([]string{directory}, func(_ string, info os.FileInfo, _ int) bool {
			if strings.HasSuffix(info.Name(), "_test.rego") {
				return true
			}

			if !info.IsDir() && filepath.Ext(info.Name()) != ".rego" {
				return true
			}

			return false
		})

	if err != nil {
		return nil, fmt.Errorf("filter rego files: %w", err)
	}

	if _, err := result.Compiler(); err != nil {
		return nil, fmt.Errorf("compile: %w", err)
	}

	compiler := ast.NewCompiler()
	compiler.Compile(result.ParsedModules())

	return compiler, nil
}
