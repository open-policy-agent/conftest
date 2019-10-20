package hcl2

import (
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
)

type Parser struct {}

func (h *Parser) Unmarshal(p []byte, v interface{}) error {
	file, diags := hclsyntax.ParseConfig(p, "", hcl.Pos{Line: 1, Column: 1})

	if diags.HasErrors() {
		for _, diag := range diags {
			fmt.Println(diag.Error())
		}
		return fmt.Errorf("Error occured while parsing HCL2 config")
	}

	content, err := convertFile(file)

	if err != nil {
		return fmt.Errorf("Unable to convert config %s", err)
	}

	j, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("Unable to marshal config: %s", err)
	}

	err = yaml.Unmarshal(j, v)
	if err != nil {
		return fmt.Errorf("Unable to parse YAML from HCL-json: %s", err)
	}

	return nil
}
