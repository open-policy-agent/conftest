package hcl2

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/tmccombs/hcl2json/convert"
)

// Parser is an HCL2 parser.
type Parser struct{}

// Parse parses HCL files that are written using
// version 2 of the HCL language.
func (Parser) Parse(r io.Reader) ([]any, error) {
	p, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	hclBytes, err := convert.Bytes(p, "", convert.Options{})
	if err != nil {
		return nil, fmt.Errorf("convert to bytes: %w", err)
	}

	var v any
	if err := json.Unmarshal(hclBytes, &v); err != nil {
		return nil, fmt.Errorf("unmarshal hcl2: %w", err)
	}

	return []any{v}, nil
}
