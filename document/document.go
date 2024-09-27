package document

import (
	"fmt"
	"io"
)

// GenerateDocument generated a documentation file for a given module by parting
// A single page is generated for the module located in the indicated directory this includes the package subpackages
// and rules of the provided path, if you want to split the documentation.
func GenerateDocument(dir string, tpl string, out io.Writer) error {

	as, err := ParseRegoWithAnnotations(dir)
	if err != nil {
		return fmt.Errorf("parse rego annotations: %w", err)
	}

	sec, err := ConvertAnnotationsToSections(as)
	if err != nil {
		return fmt.Errorf("validating annotations: %w", err)
	}

	var opt []RenderDocumentOption
	if tpl != "" {
		opt = append(opt, WithTemplate(tpl))
	}

	err = RenderDocument(out, sec, opt...)
	if err != nil {
		return fmt.Errorf("rendering document: %w", err)
	}

	return nil
}
