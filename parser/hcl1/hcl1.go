package hcl1

import (
	"fmt"
	"io"

	"github.com/hashicorp/hcl"
)

// Parser is an HCL parser.
type Parser struct{}

// Parse parses HCL files that are using version 1 of
// the HCL language.
func (s *Parser) Parse(r io.Reader) ([]any, error) {
	p, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	var v any
	if err := hcl.Unmarshal(p, &v); err != nil {
		return nil, fmt.Errorf("unmarshal hcl: %w", err)
	}
	return []any{v}, nil
}
