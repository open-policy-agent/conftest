package spdx

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spdx/tools-golang/tagvalue"
)

// Parser is a SPDX parser.
type Parser struct{}

// Parse parses SPDX files.
func (*Parser) Parse(r io.Reader) ([]any, error) {
	doc, err := tagvalue.Read(r)
	if err != nil {
		return nil, fmt.Errorf("error while parsing %v: %v", r, err)
	}

	out, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling %v: %v", r, err)
	}

	var v any
	if err := json.Unmarshal(out, &v); err != nil {
		return nil, fmt.Errorf("unmarshal SPDX json: %w", err)
	}

	return []any{v}, nil
}
