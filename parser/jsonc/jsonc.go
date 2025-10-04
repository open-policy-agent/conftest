package jsonc

import (
	"fmt"
	"io"

	"github.com/muhammadmuzzammil1998/jsonc"
)

// Parser is a JSON parser.
type Parser struct{}

// Parse parses JSON files.
func (p *Parser) Parse(r io.Reader) ([]any, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	var v any
	if err := jsonc.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("unmarshal jsonc: %w", err)
	}

	return []any{v}, nil
}
