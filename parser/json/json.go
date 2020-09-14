package json

import (
	"encoding/json"
	"fmt"
)

// Parser is a JSON parser.
type Parser struct{}

// Unmarshal unmarshals JSON files.
func (p *Parser) Unmarshal(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("unmarshal json: %w", err)
	}

	return nil
}
