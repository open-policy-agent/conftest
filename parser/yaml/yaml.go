package yaml

import (
	"errors"
	"fmt"
	"io"

	"go.yaml.in/yaml/v4"
)

// Parser is a YAML parser.
type Parser struct{}

// Parse parses YAML files supporting multi-document files.
func (yp *Parser) Parse(r io.Reader) ([]any, error) {
	decoder := yaml.NewDecoder(r)
	var documents []any
	for {
		var document any
		if err := decoder.Decode(&document); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("unmarshal yaml: %w", err)
		}
		documents = append(documents, document)
	}
	return documents, nil
}
