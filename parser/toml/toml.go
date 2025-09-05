package toml

import (
	"fmt"
	"io"

	"github.com/BurntSushi/toml"
)

// Parser is a TOML parser.
type Parser struct{}

// Parse parses TOML files.
func (tp *Parser) Parse(r io.Reader) ([]any, error) {
	var v any
	if _, err := toml.NewDecoder(r).Decode(&v); err != nil {
		return nil, fmt.Errorf("unmarshal toml: %w", err)
	}

	return []any{v}, nil
}
