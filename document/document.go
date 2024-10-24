package document

import (
	"fmt"
	"io"
)

// GenerateDocument generate a documentation file for a given module
// A single page is generated for the module located in the indicated directory this includes all package, subpackages
// and rules of the provided path.
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
		opt = append(opt, ExternalTemplate(tpl))
	}

	err = RenderDocument(out, sec, opt...)
	if err != nil {
		return fmt.Errorf("rendering document: %w", err)
	}

	return nil
}
