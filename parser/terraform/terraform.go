package terraform

import (
	"fmt"

	"github.com/hashicorp/hcl"
)

type Parser struct{}

func (s *Parser) Unmarshal(p []byte, v interface{}) error {
	if err := hcl.Unmarshal(p, v); err != nil {
		return fmt.Errorf("unmarshal hcl: %w", err)
	}

	return nil
}
