package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/dimchansky/utfbom"
)

// Parser is a JSON parser.
type Parser struct{}

// Parse parses JSON files.
func (p *Parser) Parse(r io.Reader) ([]any, error) {
	r = utfbom.SkipOnly(r) // Strip UTF BOM, see https://www.rfc-editor.org/rfc/rfc8259#section-8.1
	decoder := json.NewDecoder(r)
	var documents []any
	for {
		var document any
		if err := decoder.Decode(&document); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("unmarshal json: %w", err)
		}
		documents = append(documents, document)
	}
	return documents, nil
}
