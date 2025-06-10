package document

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/loader"
)

var (
	ErrNoAnnotations = errors.New("no annotations found")
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
		return nil, fmt.Errorf("load rego files: %w", err)
	}

	compiler, err := result.Compiler()
	if err != nil {
		return nil, fmt.Errorf("compile: %w", err)
	}
	as := compiler.GetAnnotationSet().Flatten()

	if len(as) == 0 {
		return nil, ErrNoAnnotations
	}

	return as, nil
}

// Document represent a page of the documentation
type Document []Section

// Section is a sequential piece of documentation comprised of ast.Annotations and some pre-processed fields
// This struct exist because some fields of ast.Annotations are not easy to manipulate in go-template
type Section struct {
	// RegoPackageName is the string representation of ast.Annotations.Path
	RegoPackageName string
	// Depth represent title depth for this section (h1, h2, h3, etc.). This values is derived from len(ast.Annotations.RegoPackageName)
	// and smoothed such that subsequent section only defer by +/- 1
	Depth int
	// MarkdownHeading represent the markdown title symbol #, ##, ###, etc. (produced by strings.Repeat("#", depth))
	MarkdownHeading string
	// Annotations is the raw metada provided by OPA compiler
	Annotations *ast.Annotations
}

// Equal is only relevant for tests and assert that two sections are partially Equal
func (s Section) Equal(s2 Section) bool {
	if s.MarkdownHeading == s2.MarkdownHeading &&
		s.RegoPackageName == s2.RegoPackageName &&
		s.Annotations.Title == s2.Annotations.Title {
		return true
	}

	return false
}

// ConvertAnnotationsToSections generates a more convenient struct that can be used to generate the doc
// First concern is to build a coherent title structure; the ideal case is that each package and each rule has a doc,
// but this is not guaranteed. I couldn't find a way to call `strings.Repeat` inside go-template; thus, the title symbol is
// directly provided as markdown (#, ##, ###, etc.)
// Second, the attribute RegoPackageName of ast.Annotations are not easy to use on go-template; thus, we extract it as a string
func ConvertAnnotationsToSections(as ast.FlatAnnotationsRefSet) (Document, error) {

	var d Document
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

		d = append(d, Section{
			Depth:           depth,
			MarkdownHeading: h,
			RegoPackageName: path,
			Annotations:     entry.Annotations,
		})
	}

	return d, nil
}
