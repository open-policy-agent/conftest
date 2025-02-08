package hcl1

import (
	"fmt"

	"github.com/hashicorp/hcl"
)

// Parser is an HCL parser.
type Parser struct{}

// Unmarshal unmarshals HCL files that are using version 1 of
// the HCL language.
func (s *Parser) Unmarshal(p []byte, v any) error {
	if err := hcl.Unmarshal(p, v); err != nil {
		return fmt.Errorf("unmarshal hcl: %w", err)
	}

	return nil
}
