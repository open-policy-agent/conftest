package xml

import (
	"encoding/json"
	"fmt"
	"io"

	x "github.com/basgys/goxml2json"
)

// Parser is an XML parser.
type Parser struct{}

// Parse parses XML files.
func (xml *Parser) Parse(r io.Reader) ([]any, error) {
	res, err := x.Convert(r)
	if err != nil {
		return nil, fmt.Errorf("unmarshal xml: %w", err)
	}

	var v any
	if err := json.Unmarshal(res.Bytes(), &v); err != nil {
		return nil, fmt.Errorf("convert xml to json: %w", err)
	}

	return []any{v}, nil
}
