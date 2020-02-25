package terraform

import (
	"fmt"

	"github.com/hashicorp/hcl"
)

// Parser is a parser for TF and HCL files
type Parser struct{}

// Unmarshal unmarshals TF and HCL files
func (s *Parser) Unmarshal(p []byte, v interface{}) error {
	if err := hcl.Unmarshal(p, v); err != nil {
		return fmt.Errorf("unmarshal hcl: %w", err)
	}

	return nil
}
