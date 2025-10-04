package vcl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/KeisukeYamashita/go-vcl/vcl"
)

// Parser is a VCL parser.
type Parser struct{}

// Parse parses VCL files.
func (p *Parser) Parse(r io.Reader) ([]any, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	result := make(map[string]any)
	if errs := vcl.Decode(data, &result); len(errs) > 0 {
		return nil, fmt.Errorf("decode vcl: %w", errors.Join(errs...))
	}

	j, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal vcl to json: %w", err)
	}

	var v any
	if err := json.Unmarshal(j, &v); err != nil {
		return nil, fmt.Errorf("unmarshal vcl json: %w", err)
	}

	return []any{v}, nil
}
