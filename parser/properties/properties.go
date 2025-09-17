package properties

import (
	"encoding/json"
	"fmt"
	"io"

	prop "github.com/magiconair/properties"
)

// Parser is a properties parser.
type Parser struct{}

func (pp *Parser) Parse(r io.Reader) ([]any, error) {
	p, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	rawProps, err := prop.Load(p, prop.UTF8)
	if err != nil {
		return nil, fmt.Errorf("parse properties file: %w", err)
	}

	result := rawProps.Map()

	j, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal properties to json: %w", err)
	}

	var v any
	if err := json.Unmarshal(j, &v); err != nil {
		return nil, fmt.Errorf("unmarshal properties json: %w", err)
	}

	return []any{v}, nil
}
