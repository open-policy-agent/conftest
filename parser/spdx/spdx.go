package spdx

import (
	"bytes"
	"fmt"

	"github.com/spdx/tools-golang/tagvalue"

	"github.com/open-policy-agent/conftest/parser/internal/convert"
)

// Parser is a SPDX parser.
type Parser struct{}

// Unmarshal unmarshals SPDX files.
func (*Parser) Unmarshal(p []byte, v any) error {
	doc, err := tagvalue.Read(bytes.NewBuffer(p))
	if err != nil {
		return fmt.Errorf("error while parsing %v: %v", p, err)
	}

	return convert.Remarshal("spdx", doc, v)
}
