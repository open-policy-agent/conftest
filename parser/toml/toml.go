package toml

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Parser is a TOML parser.
type Parser struct{}

// Unmarshal unmarshals TOML files.
func (tp *Parser) Unmarshal(p []byte, v any) error {
	if err := toml.Unmarshal(p, v); err != nil {
		return fmt.Errorf("unmarshal toml: %w", err)
	}

	return nil
}
