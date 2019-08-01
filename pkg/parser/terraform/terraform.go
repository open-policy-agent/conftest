package terraform

import (
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/hashicorp/terraform/config"
)

type Parser struct {
	FileName string
}

func (s *Parser) Unmarshal(p []byte, v interface{}) error {
	cfg, err := config.LoadJSON(json.RawMessage(p))
	if err != nil {
		return fmt.Errorf("load terraform config failed: %v", err)
	}

	j, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("Unable to marshal config %s: %s", s.FileName, err)
	}

	err = yaml.Unmarshal(j, v)
	if err != nil {
		return fmt.Errorf("Unable to parse YAML from HCL-json %s: %s", s.FileName, err)
	}

	return nil
}
