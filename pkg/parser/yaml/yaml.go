package yaml

import (
	"fmt"
	"github.com/ghodss/yaml"
)

type Parser struct {
	FileName string
}

func (yp *Parser) Unmarshal(p []byte, v interface{}) error {
	err := yaml.Unmarshal(p, v)
	if err != nil {
		return fmt.Errorf("Unable to parse YAML from %s: %s", yp.FileName, err)
	}

	return nil
}
