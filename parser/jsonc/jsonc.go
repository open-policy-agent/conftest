package jsonc

import (
	"fmt"

	"muzzammil.xyz/jsonc"
)

// Parser is a JSON parser.
type Parser struct{}

// Unmarshal unmarshals JSON files.
func (p *Parser) Unmarshal(data []byte, v any) error {
	if err := jsonc.Unmarshal(data, v); err != nil {
		return fmt.Errorf("unmarshal jsonc: %w", err)
	}

	return nil
}
