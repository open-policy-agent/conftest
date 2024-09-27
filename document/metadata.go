package document

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"
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

// ConvertAnnotationsToSections generate a more convenient struct that can be used to generate the doc
// First concern is to build a coherent title structure, the ideal case is that each package and each rule as a doc,
// but this is not guarantied. I couldn't find a way to call strings.Repeat inside go-template, this the title key is
// directly provided as markdown (#, ##, ###, etc.)
// Second the attribute Path of ast.Annotations are not easy to used on go-template, thus we extract it as a string
func ConvertAnnotationsToSections(as ast.FlatAnnotationsRefSet) ([]Section, error) {

	var s []Section
	var currentDepth = 0
	var offset = 1

	for i, entry := range as {
		// offset at least by one because all path starts with `data.`
		depth := len(entry.Path) - offset

		// If the user is targeting a submodule we need to adjust the depth an offset base on the first annotation found
		if i == 0 && depth > 1 {
			offset = depth
		}

		// We need to compensate for unexpected jump in depth
		// otherwise we would start at h3 if no package documentation is present
		// or jump form h2 to h4 unexpectedly in subpackages
		if (depth - currentDepth) > 1 {
			depth = currentDepth + 1
		}

		currentDepth = depth

		h := strings.Repeat("#", depth)
		path := strings.TrimPrefix(entry.Path.String(), "data.")

		s = append(s, Section{
			H:           h,
			Path:        path,
			Annotations: entry.Annotations,
		})
	}

	return s, nil
}
