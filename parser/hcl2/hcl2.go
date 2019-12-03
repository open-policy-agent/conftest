package hcl2

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Parser struct{}

func (h *Parser) Unmarshal(p []byte, v interface{}) error {
	file, diags := hclsyntax.ParseConfig(p, "", hcl.Pos{Byte: 0, Line: 1, Column: 1})

	if diags.HasErrors() {
		return fmt.Errorf("parse hcl2 config: %s", diags.Error())
	}

	content, err := convertFile(file)
	if err != nil {
		return fmt.Errorf("convert hcl2 to json: %w", err)
	}

	j, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("marshal hcl2 to json: %w", err)
	}

	if err := json.Unmarshal(j, v); err != nil {
		return fmt.Errorf("unmarshal hcl2 json: %w", err)
	}

	return nil
}
