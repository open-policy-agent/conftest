package vcl

import (
	"encoding/json"
	"fmt"

	"github.com/KeisukeYamashita/go-vcl/vcl"
)

// Parser is a VCL parser.
type Parser struct{}

// Unmarshal unmarshals VCL files.
func (p *Parser) Unmarshal(b []byte, v any) error {
	result := make(map[string]any)
	if errs := vcl.Decode(b, &result); len(errs) > 0 {
		return fmt.Errorf("decode vcl: %w", errs[0])
	}

	j, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal vcl to json: %w", err)
	}

	if err := json.Unmarshal(j, v); err != nil {
		return fmt.Errorf("unmarshal vcl json: %w", err)
	}

	return nil
}
