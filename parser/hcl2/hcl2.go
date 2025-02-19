package hcl2

import (
	"encoding/json"
	"fmt"

	"github.com/tmccombs/hcl2json/convert"
)

// Parser is an HCL2 parser.
type Parser struct{}

// Unmarshal unmarshals HCL files that are written using
// version 2 of the HCL language.
func (Parser) Unmarshal(p []byte, v any) error {
	hclBytes, err := convert.Bytes(p, "", convert.Options{})
	if err != nil {
		return fmt.Errorf("convert to bytes: %w", err)
	}

	if err := json.Unmarshal(hclBytes, v); err != nil {
		return fmt.Errorf("unmarshal hcl2: %w", err)
	}

	return nil
}
