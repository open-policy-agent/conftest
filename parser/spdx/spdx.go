package spdx

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/spdx/tools-golang/tagvalue"
)

// Parser is a SPDX parser.
type Parser struct{}

// Unmarshal unmarshals SPDX files.
func (*Parser) Unmarshal(p []byte, v any) error {
	doc, err := tagvalue.Read(bytes.NewBuffer(p))
	if err != nil {
		return fmt.Errorf("error while parsing %v: %v", p, err)
	}

	out, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("error while marshaling %v: %v", p, err)
	}

	if err := json.Unmarshal(out, v); err != nil {
		return fmt.Errorf("unmarshal SPDX json: %w", err)
	}

	return nil
}
