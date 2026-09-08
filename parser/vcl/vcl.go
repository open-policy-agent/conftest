package vcl

import (
	"fmt"

	"github.com/KeisukeYamashita/go-vcl/vcl"

	"github.com/open-policy-agent/conftest/parser/internal/convert"
)

// Parser is a VCL parser.
type Parser struct{}

// Unmarshal unmarshals VCL files.
func (p *Parser) Unmarshal(b []byte, v any) error {
	result := make(map[string]any)
	if errs := vcl.Decode(b, &result); len(errs) > 0 {
		return fmt.Errorf("decode vcl: %w", errs[0])
	}

	return convert.Remarshal("vcl", result, v)
}
