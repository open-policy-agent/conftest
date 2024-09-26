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
func ParseRegoWithAnnotations(directory string) (ast.FlatAnnotationsRefSet, error) {
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
	as := compiler.GetAnnotationSet().Flatten()

	return as, nil
}

type Section struct {
	H           string
	Path        string
	Annotations *ast.Annotations
}

func (s Section) Equal(s2 Section) bool {
	if s.H == s2.H &&
		s.Path == s2.Path &&
		s.Annotations.Title == s2.Annotations.Title {
		return true
	}

	return false
}

// GetDocument generate a more convenient struct that can be used to generate the doc
func GetDocument(as ast.FlatAnnotationsRefSet) []Section {

	var s []Section

	for _, entry := range as {

		depth := strings.Repeat("#", len(entry.Path))
		path := strings.TrimPrefix(entry.Path.String(), "data.")

		s = append(s, Section{
			H:           depth,
			Path:        path,
			Annotations: entry.Annotations,
		})
	}

	return s
}
