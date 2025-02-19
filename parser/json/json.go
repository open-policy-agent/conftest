package json

import (
	"encoding/json"
	"fmt"
)

// Parser is a JSON parser.
type Parser struct{}

// Unmarshal unmarshals JSON files.
func (p *Parser) Unmarshal(data []byte, v any) error {
	if len(data) > 2 && data[0] == 0xef && data[1] == 0xbb && data[2] == 0xbf {
		data = data[3:] // Strip UTF-8 BOM, see https://www.rfc-editor.org/rfc/rfc8259#section-8.1
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("unmarshal json: %w", err)
	}

	return nil
}
